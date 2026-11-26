package systemadmin

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"casino/internal/content"
	"casino/internal/eventbus"
	"casino/internal/storage"

	"gorm.io/gorm"
)

type SupportExportInput struct {
	SiteID         uint64            `json:"siteID"`
	ExportType     string            `json:"exportType"`
	Filters        map[string]string `json:"filters"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

type SupportExportTaskRow struct {
	ID     uint64 `json:"id"`
	TaskNo string `json:"taskNo" gorm:"column:task_no"`

	SiteID        uint64     `json:"siteID" gorm:"column:site_id"`
	ExportType    string     `json:"exportType" gorm:"column:export_type"`
	Filters       string     `json:"filters"`
	Status        string     `json:"status"`
	ArtifactKey   string     `json:"-" gorm:"column:artifact_key"`
	RowCount      uint64     `json:"rowCount" gorm:"column:row_count"`
	CreatedBy     uint64     `json:"createdBy" gorm:"column:created_by"`
	FailureReason string     `json:"failureReason" gorm:"column:failure_reason"`
	CompletedAt   *time.Time `json:"completedAt" gorm:"column:completed_at"`
	ExpiresAt     *time.Time `json:"expiresAt" gorm:"column:expires_at"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func (s *Service) ListSupportExports(q Query, actor Actor) (Result, error) {
	rows := []SupportExportTaskRow{}
	db := s.db.Table("game_support_export_tasks").Where("deleted_at IS NULL")
	var err error
	db, err = operationalScope(db, actor, value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = exact(db, "status", strings.ToUpper(value(q, "status")))
	db = exact(db, "export_type", strings.ToUpper(value(q, "exportType")))
	return paginate(db.Order("created_at DESC,id DESC"), q, &rows)
}

func (s *Service) CreateSupportExport(input SupportExportInput, actor Actor) (SupportExportTaskRow, error) {
	siteID, err := operationalTargetScope(s.db, actor, input.SiteID)
	if err != nil {
		return SupportExportTaskRow{}, err
	}
	if siteID == 0 {
		return SupportExportTaskRow{}, NewError("SITE_REQUIRED", "H5 configuration is required for support export")
	}
	exportType := strings.ToUpper(strings.TrimSpace(input.ExportType))
	if exportType != "PLAYER_TIMELINE" && exportType != "ORDER_SEARCH" {
		return SupportExportTaskRow{}, NewError("INVALID_EXPORT_TYPE", "unsupported support export type")
	}
	filters := sanitizeExportFilters(input.Filters)
	if exportType == "PLAYER_TIMELINE" && parseID(filters["playerID"]) == 0 {
		return SupportExportTaskRow{}, NewError("PLAYER_REQUIRED", "player is required for timeline export")
	}
	key := strings.TrimSpace(input.IdempotencyKey)
	if key == "" {
		return SupportExportTaskRow{}, NewError("IDEMPOTENCY_KEY_REQUIRED", "idempotency key is required")
	}
	var existing SupportExportTaskRow
	err = s.db.Table("game_support_export_tasks").Where("created_by=? AND idempotency_key=? AND deleted_at IS NULL", actor.ID, key).Take(&existing).Error
	if err == nil {
		return existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return SupportExportTaskRow{}, err
	}
	taskNo, err := newSupportExportNo()
	if err != nil {
		return SupportExportTaskRow{}, err
	}
	row := map[string]any{"task_no": taskNo, "site_id": siteID, "export_type": exportType, "filters": marshal(filters), "status": "PENDING", "idempotency_key": key, "created_by": actor.ID, "expires_at": now().Add(24 * time.Hour), "created_at": now(), "updated_at": now()}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_support_export_tasks").Create(row).Error; err != nil {
			return duplicateError(err)
		}
		var taskID uint64
		if err := tx.Table("game_support_export_tasks").Select("id").Where("task_no=?", taskNo).Scan(&taskID).Error; err != nil {
			return err
		}
		if err := audit(tx, targetAuditActor(actor, siteID), "support:export", "support.export.create", "support_export", taskID, nil, map[string]any{"taskNo": taskNo, "exportType": exportType, "filters": filters}, "support data export"); err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: content.Scope{SiteID: siteID, Platform: actor.Platform}, SiteID: siteID,
			EventType: eventbus.EventSupportExportRequested, AggregateType: "support_export", AggregateID: fmt.Sprint(taskID),
			Source: "systemadmin", Producer: "systemadmin", IdempotencyKey: "support-export:" + taskNo,
			RequestID: actor.CorrelationID, TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID,
			TargetConsumers: []string{eventbus.ConsumerSupportExport}, Payload: map[string]any{"taskID": taskID},
		})
		return err
	}); err != nil {
		return SupportExportTaskRow{}, err
	}
	var created SupportExportTaskRow
	if err := s.db.Table("game_support_export_tasks").Where("task_no=?", taskNo).Take(&created).Error; err != nil {
		return SupportExportTaskRow{}, err
	}
	eventbus.SignalDispatch()
	return created, nil
}

func (s *Service) executeSupportExport(id uint64) error {
	claimed := s.db.Table("game_support_export_tasks").Where("id=? AND (status IN ('PENDING','FAILED') OR (status='RUNNING' AND updated_at<?)) AND deleted_at IS NULL", id, now().Add(-30*time.Minute)).Updates(map[string]any{"status": "RUNNING", "failure_reason": "", "updated_at": now()})
	if claimed.Error != nil || claimed.RowsAffected != 1 {
		if claimed.Error != nil {
			return claimed.Error
		}
		var status string
		if err := s.db.Table("game_support_export_tasks").Select("status").Where("id=? AND deleted_at IS NULL", id).Scan(&status).Error; err != nil {
			return err
		}
		if status == "SUCCESS" {
			return nil
		}
		return NewError("SUPPORT_EXPORT_NOT_AVAILABLE", "support export task is not available")
	}
	var task SupportExportTaskRow
	if err := s.db.Table("game_support_export_tasks").Where("id=?", id).Take(&task).Error; err != nil {
		return err
	}
	contents, rows, err := s.buildSupportExport(task)
	if err == nil {
		key := path.Join("support-exports", now().Format("2006/01/02"), task.TaskNo+".csv")
		var asset storage.Asset
		asset, err = storage.PutBytes(context.Background(), key, contents, "text/csv; charset=utf-8")
		if err == nil {
			exists, checkErr := storage.Exists(context.Background(), asset.Key)
			if checkErr != nil || !exists {
				err = NewError("EXPORT_ARTIFACT_MISSING", "export artifact could not be verified")
			} else {
				completed := now()
				err = s.db.Table("game_support_export_tasks").Where("id=? AND status='RUNNING'", id).Updates(map[string]any{"status": "SUCCESS", "artifact_key": asset.Key, "row_count": rows, "completed_at": completed, "updated_at": completed}).Error
			}
		}
	}
	if err != nil {
		completed := now()
		_ = s.db.Table("game_support_export_tasks").Where("id=? AND status='RUNNING'", id).Updates(map[string]any{"status": "FAILED", "failure_reason": sanitizeFailure(err.Error()), "completed_at": completed, "updated_at": completed}).Error
		return err
	}
	return nil
}

func (s *Service) HandleSupportExportEvent(ctx eventbus.Context) error {
	id := uint64FromJSON(ctx.Event.Payload["taskID"])
	if id == 0 {
		return NewError("SUPPORT_EXPORT_TASK_REQUIRED", "support export task is required")
	}
	var count int64
	if err := s.db.Table("game_support_export_tasks").Where("id = ? AND site_id = ? AND deleted_at IS NULL", id, ctx.Event.SiteID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return NewError("SUPPORT_EXPORT_SCOPE_MISMATCH", "support export is outside event scope")
	}
	return s.executeSupportExport(id)
}

func (s *Service) buildSupportExport(task SupportExportTaskRow) ([]byte, uint64, error) {
	filters := map[string]string{}
	if err := json.Unmarshal([]byte(task.Filters), &filters); err != nil {
		return nil, 0, NewError("INVALID_EXPORT_FILTERS", "export filters are invalid")
	}
	if task.SiteID > 0 {
		filters["siteID"] = fmt.Sprint(task.SiteID)
	}
	actor := Actor{SiteID: task.SiteID, Platform: false, Sensitive: false}
	query := Query{Page: 1, PageSize: 100, Values: filters}
	rows := []map[string]any{}
	switch task.ExportType {
	case "PLAYER_TIMELINE":
		playerID := parseID(filters["playerID"])
		for pageNumber := 1; pageNumber <= 100; pageNumber++ {
			query.Page = pageNumber
			result, err := s.PlayerTimeline(playerID, query, actor)
			if err != nil {
				return nil, 0, err
			}
			pageRows := mapsFromAny(result.List)
			rows = append(rows, pageRows...)
			if int64(len(rows)) >= result.Total || len(pageRows) == 0 {
				break
			}
		}
	case "ORDER_SEARCH":
		query.PageSize = 100
		for pageNumber := 1; pageNumber <= 100; pageNumber++ {
			query.Page = pageNumber
			result, err := s.SearchOrders(query, actor)
			if err != nil {
				return nil, 0, err
			}
			pageRows := mapsFromAny(result.List)
			rows = append(rows, pageRows...)
			if int64(len(rows)) >= result.Total || len(pageRows) == 0 {
				break
			}
		}
	}
	contents, err := supportCSV(rows)
	return contents, uint64(len(rows)), err
}

func (s *Service) ReadSupportExport(id uint64, actor Actor) ([]byte, string, error) {
	var row SupportExportTaskRow
	db := s.db.Table("game_support_export_tasks").Where("id=? AND status='SUCCESS' AND deleted_at IS NULL", id)
	db, _ = operationalScope(db, actor, "")
	if err := db.Take(&row).Error; err != nil {
		return nil, "", rowNotFound("support_export")
	}
	if row.ExpiresAt == nil || !row.ExpiresAt.After(now()) {
		return nil, "", NewError("EXPORT_EXPIRED", "export has expired")
	}
	data, err := storage.ReadBytes(context.Background(), row.ArtifactKey, 128<<20)
	if err != nil {
		return nil, "", err
	}
	return data, row.TaskNo + ".csv", nil
}

func supportCSV(rows []map[string]any) ([]byte, error) {
	headers := []string{}
	seen := map[string]bool{}
	for _, row := range rows {
		for key := range row {
			if !seen[key] && !sensitiveKeyPattern.MatchString(key) {
				seen[key] = true
				headers = append(headers, key)
			}
		}
	}
	sort.Strings(headers)
	var output bytes.Buffer
	output.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&output)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for index, header := range headers {
			record[index] = escapeCSVFormula(fmt.Sprint(row[header]))
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return output.Bytes(), writer.Error()
}

func escapeCSVFormula(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

func mapsFromAny(value any) []map[string]any {
	data, _ := json.Marshal(value)
	rows := []map[string]any{}
	_ = json.Unmarshal(data, &rows)
	return rows
}

func sanitizeExportFilters(filters map[string]string) map[string]string {
	allowed := map[string]bool{"playerID": true, "query": true, "status": true, "from": true, "to": true, "eventTypes": true, "providerCode": true, "minAmount": true, "maxAmount": true}
	result := map[string]string{}
	for key, item := range filters {
		if allowed[key] && len(item) <= 256 {
			result[key] = strings.TrimSpace(item)
		}
	}
	return result
}

func newSupportExportNo() (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "EXP" + now().UTC().Format("20060102150405") + strings.ToUpper(hex.EncodeToString(random)), nil
}
