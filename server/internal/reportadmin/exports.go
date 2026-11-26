package reportadmin

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"mime"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/eventbus"
	"casino/internal/storage"
	"casino/internal/tenancy"
)

func reportAudit(tx *gorm.DB, actor Actor, permission, action, reportType string, targetID uint64, before, after any, reason string) error {
	return tx.Table("game_admin_operation_logs").Create(map[string]any{
		"tenant_id": actor.Scope.TenantID, "site_id": actor.Scope.SiteID,
		"operator_id": actor.ID, "operator_username": actor.Username, "permission": permission, "action": action,
		"target_type": "report_export", "target_id": targetID, "before_state": marshal(before), "after_state": marshal(after),
		"reason": reason, "ip_address": actor.IP, "user_agent": actor.UserAgent, "correlation_id": actor.CorrelationID,
		"currency": "", "created_at": time.Now(), "updated_at": time.Now(),
	}).Error
}

func (s *ReportExportService) Create(command ExportCommand, actor Actor) (uint64, error) {
	typeCode, err := reportType(command.ReportType)
	if err != nil {
		return 0, err
	}
	format, err := fileFormat(command.FileFormat)
	if err != nil {
		return 0, err
	}
	if err = validateRange(Query{Values: command.Filters, Scope: actor.Scope}); err != nil {
		return 0, err
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return 0, NewError("IDEMPOTENCY_KEY_REQUIRED", "idempotency key is required")
	}
	now := time.Now()
	taskNo := fmt.Sprintf("REXP%d", now.UnixNano())
	filtersJSON := marshal(command.Filters)
	summary := filtersSummary(command.Filters)
	row := map[string]any{"tenant_id": actor.Scope.TenantID, "site_id": actor.Scope.SiteID, "task_no": taskNo, "report_type": typeCode, "filters": filtersJSON, "filters_summary": summary, "file_format": format, "status": "PENDING", "operator_id": actor.ID, "operator_username": actor.Username, "idempotency_key": strings.TrimSpace(command.IdempotencyKey), "correlation_id": actor.CorrelationID, "created_at": now, "updated_at": now}
	created := false
	err = s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("game_report_export_tasks").Clauses(clause.Insert{Modifier: "IGNORE"}).Create(row)
		if e := result.Error; e != nil {
			return e
		}
		created = result.RowsAffected == 1
		var existing struct {
			ID         uint64 `gorm:"column:id"`
			ReportType string `gorm:"column:report_type"`
			FileFormat string `gorm:"column:file_format"`
			Filters    string `gorm:"column:filters"`
		}
		if e := exportTaskDB(tx, actor.Scope).Select("id,report_type,file_format,filters").Where("operator_id=? AND idempotency_key=? AND deleted_at IS NULL", actor.ID, command.IdempotencyKey).Take(&existing).Error; e != nil {
			return e
		}
		if existing.ReportType != typeCode || existing.FileFormat != format || !reflect.DeepEqual(unmarshalMap(existing.Filters), command.Filters) {
			return NewError("IDEMPOTENCY_CONFLICT", "idempotency key was used with a different request")
		}
		row["id"] = existing.ID
		if !created {
			return nil
		}
		if err := reportAudit(tx, actor, "report:export:create", "report.export.create", typeCode, existing.ID, nil, map[string]any{"filters": command.Filters, "format": format}, ""); err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: actor.Scope, TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID,
			EventType: eventbus.EventReportExportRequested, AggregateType: "report_export", AggregateID: fmt.Sprint(existing.ID),
			Source: "reportadmin", Producer: "reportadmin", IdempotencyKey: "report-export:" + taskNo,
			RequestID: actor.CorrelationID, TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID,
			TargetConsumers: []string{eventbus.ConsumerReportExport}, Payload: map[string]any{"taskID": existing.ID},
		})
		return err
	})
	if err != nil {
		return 0, err
	}
	id := row["id"].(uint64)
	if created {
		eventbus.SignalDispatch()
	}
	return id, nil
}

func (s *ReportExportService) HandleEvent(ctx eventbus.Context) error {
	id := payloadUint64(ctx.Event.Payload["taskID"])
	if id == 0 {
		return NewError("EXPORT_TASK_REQUIRED", "report export task is required")
	}
	var task struct {
		TenantID uint64 `gorm:"column:tenant_id"`
		SiteID   uint64 `gorm:"column:site_id"`
		Status   string
		Failure  string `gorm:"column:failure_reason"`
	}
	if err := s.db.Table("game_report_export_tasks").Select("tenant_id,site_id,status,failure_reason").Where("id=? AND deleted_at IS NULL", id).Take(&task).Error; err != nil {
		return err
	}
	if task.TenantID != ctx.Event.TenantID || task.SiteID != ctx.Event.SiteID {
		return NewError("EXPORT_SCOPE_MISMATCH", "report export is outside event scope")
	}
	if task.Status == "SUCCESS" {
		return nil
	}
	s.process(id)
	if err := s.db.Table("game_report_export_tasks").Select("status,failure_reason").Where("id=?", id).Take(&task).Error; err != nil {
		return err
	}
	if task.Status != "SUCCESS" {
		if strings.TrimSpace(task.Failure) == "" {
			task.Failure = "report export did not complete"
		}
		return NewError("EXPORT_FAILED", task.Failure)
	}
	return nil
}

func payloadUint64(value any) uint64 {
	switch typed := value.(type) {
	case uint64:
		return typed
	case int:
		if typed > 0 {
			return uint64(typed)
		}
	case float64:
		if typed > 0 {
			return uint64(typed)
		}
	case string:
		parsed, _ := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
		return parsed
	}
	return 0
}

func filtersSummary(filters map[string]string) string {
	keys := sortedKeys(filters)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := strings.TrimSpace(filters[key]); value != "" {
			parts = append(parts, key+"="+value)
		}
	}
	return strings.Join(parts, "; ")
}

func (s *ReportExportService) process(id uint64) {
	now := time.Now()
	result := s.db.Table("game_report_export_tasks").Where("id=? AND (status IN ('PENDING','FAILED') OR (status='PROCESSING' AND updated_at<?)) AND deleted_at IS NULL", id, now.Add(-30*time.Minute)).Updates(map[string]any{"status": "PROCESSING", "started_at": now, "failure_reason": "", "updated_at": now})
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}
	var task struct {
		ReportType, Filters, FileFormat, TaskNo string
		TenantID, SiteID                        uint64
	}
	if e := s.db.Table("game_report_export_tasks").Select("report_type,filters,file_format,task_no,tenant_id,site_id").Where("id=?", id).Take(&task).Error; e != nil {
		s.fail(id, e)
		return
	}
	filters := unmarshalMap(task.Filters)
	scope := exportTaskScope(task.TenantID, task.SiteID)
	rows, err := s.exportRows(task.ReportType, filters, scope)
	if err != nil {
		s.fail(id, err)
		return
	}
	extension := strings.ToLower(task.FileFormat)
	temporary, err := os.CreateTemp("", "report-export-*."+extension)
	if err != nil {
		s.fail(id, err)
		return
	}
	path := temporary.Name()
	_ = temporary.Close()
	defer os.Remove(path)
	if task.FileFormat == "CSV" {
		err = writeCSV(path, rows)
	} else {
		err = writeXLSX(path, rows)
	}
	if err != nil {
		s.fail(id, err)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		s.fail(id, err)
		return
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		s.fail(id, err)
		return
	}
	contentType := mime.TypeByExtension("." + extension)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	asset, err := storage.Put(context.Background(), fmt.Sprintf("report-exports/tenant-%d/site-%d/%s.%s", task.TenantID, task.SiteID, task.TaskNo, extension), file, stat.Size(), contentType)
	_ = file.Close()
	if err != nil {
		s.fail(id, err)
		return
	}
	completed := time.Now()
	expires := completed.Add(24 * time.Hour)
	s.db.Table("game_report_export_tasks").Where("id=? AND status='PROCESSING'", id).Updates(map[string]any{"status": "SUCCESS", "file_path": asset.Key, "row_count": len(rows), "completed_at": completed, "expires_at": expires, "updated_at": completed})
}
func (s *ReportExportService) fail(id uint64, err error) {
	now := time.Now()
	s.db.Table("game_report_export_tasks").Where("id=? AND status='PROCESSING'", id).Updates(map[string]any{"status": "FAILED", "failure_reason": truncate(err.Error(), 500), "completed_at": now, "updated_at": now})
}
func truncate(v string, n int) string {
	if len(v) > n {
		return v[:n]
	}
	return v
}

func (s *ReportExportService) exportRows(kind string, filters map[string]string, scope tenancy.Scope) ([]map[string]any, error) {
	q := exportPageQuery(filters, scope)
	all := []map[string]any{}
	for {
		r, e := s.run(kind, q)
		if e != nil {
			return nil, e
		}
		pageRows := toMaps(r.List)
		all = append(all, pageRows...)
		if len(all) >= r.Total || len(pageRows) == 0 {
			break
		}
		q.Page++
	}
	return all, nil
}

func exportTaskScope(tenantID, siteID uint64) tenancy.Scope {
	return tenancy.Scope{TenantID: tenantID, SiteID: siteID, Platform: tenantID == 0}
}

func exportTaskDB(db *gorm.DB, scope tenancy.Scope) *gorm.DB {
	return tenancy.ApplySite(db.Table("game_report_export_tasks"), "", scope)
}

func exportPageQuery(filters map[string]string, scope tenancy.Scope) Query {
	return Query{Page: 1, PageSize: 100, Values: filters, Scope: scope}
}

func (s *ReportExportService) run(kind string, q Query) (Result, error) {
	switch kind {
	case "FINANCIAL":
		return s.reports.Financial.Report(q)
	case "GAME":
		return s.reports.Game.Report(q)
	case "PROVIDER":
		return s.reports.Provider.Report(q)
	case "USER":
		return s.reports.User.Report(q)
	case "CHANNEL":
		return s.reports.Channel.Report(q)
	case "AGENT":
		return s.reports.Agent.Report(q)
	default:
		return s.reports.Promotion.Report(q)
	}
}
func toMaps(value any) []map[string]any {
	data, _ := json.Marshal(value)
	out := []map[string]any{}
	_ = json.Unmarshal(data, &out)
	return out
}

func writeCSV(path string, rows []map[string]any) error {
	file, e := os.Create(path)
	if e != nil {
		return e
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	headers := exportHeaders(rows)
	if e = writer.Write(headers); e != nil {
		return e
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, key := range headers {
			record[i] = fmt.Sprint(row[key])
		}
		if e = writer.Write(record); e != nil {
			return e
		}
	}
	return writer.Error()
}
func exportHeaders(rows []map[string]any) []string {
	seen := map[string]bool{}
	keys := []string{}
	for _, row := range rows {
		for key := range row {
			if !seen[key] && !sensitiveExportKey(key) {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}
	sort.Strings(keys)
	return keys
}
func sensitiveExportKey(key string) bool {
	lower := strings.ToLower(key)
	for _, fragment := range []string{"password", "token", "secret", "signature", "rawmetadata", "launchurl"} {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

func writeXLSX(path string, rows []map[string]any) error {
	file, e := os.Create(path)
	if e != nil {
		return e
	}
	defer file.Close()
	z := zip.NewWriter(file)
	headers := exportHeaders(rows)
	files := map[string]string{"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`, "_rels/.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`, "xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Report" sheetId="1" r:id="rId1"/></sheets></workbook>`, "xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`, "xl/worksheets/sheet1.xml": sheetXML(headers, rows)}
	for name, content := range files {
		w, err := z.Create(name)
		if err != nil {
			return err
		}
		if _, err = w.Write([]byte(content)); err != nil {
			return err
		}
	}
	return z.Close()
}
func sheetXML(headers []string, rows []map[string]any) string {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	writeSheetRow(&b, 1, headers)
	for i, row := range rows {
		values := make([]string, len(headers))
		for j, key := range headers {
			values[j] = fmt.Sprint(row[key])
		}
		writeSheetRow(&b, i+2, values)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String()
}
func writeSheetRow(b *bytes.Buffer, row int, values []string) {
	b.WriteString(`<row r="` + strconv.Itoa(row) + `">`)
	for col, value := range values {
		b.WriteString(`<c r="` + cellRef(col, row) + `" t="inlineStr"><is><t>` + html.EscapeString(value) + `</t></is></c>`)
	}
	b.WriteString(`</row>`)
}
func cellRef(col, row int) string {
	result := ""
	for col++; col > 0; col = (col - 1) / 26 {
		result = string(rune('A'+(col-1)%26)) + result
	}
	return result + strconv.Itoa(row)
}

func (s *ReportExportService) List(q Query) (Result, error) {
	p, ps := normalizedPage(q)
	db := exportTaskDB(s.db, q.Scope).Where("deleted_at IS NULL")
	db = applyEqual(db, "task_no", val(q, "taskID"))
	db = applyEqual(db, "report_type", val(q, "reportType"))
	db = applyEqual(db, "operator_username", val(q, "operator"))
	db = applyEqual(db, "status", val(q, "status"))
	if v := val(q, "createdFrom"); v != "" {
		db = db.Where("created_at>=?", v)
	}
	if v := val(q, "createdTo"); v != "" {
		db = db.Where("created_at<=?", v)
	}
	if v := val(q, "completedFrom"); v != "" {
		db = db.Where("completed_at>=?", v)
	}
	if v := val(q, "completedTo"); v != "" {
		db = db.Where("completed_at<=?", v)
	}
	var total int64
	if e := db.Count(&total).Error; e != nil {
		return Result{}, e
	}
	rows := []ExportRow{}
	if e := db.Select("id,task_no,report_type,filters_summary,file_format,row_count,status,operator_id,operator_username,failure_reason,created_at,completed_at,expires_at").Order("created_at DESC,id DESC").Offset((p - 1) * ps).Limit(ps).Scan(&rows).Error; e != nil {
		return Result{}, e
	}
	return Result{List: rows, Total: int(total), Page: p, PageSize: ps, SummaryByCurrency: []map[string]any{}, MetricBasis: metricBasis()}, nil
}
func (s *ReportExportService) Detail(id uint64, actor Actor, allowAll bool) (map[string]any, error) {
	var row ExportRow
	db := exportTaskDB(s.db, actor.Scope).Select("id,task_no,report_type,filters_summary,file_format,row_count,status,operator_id,operator_username,failure_reason,created_at,completed_at,expires_at").Where("id=? AND deleted_at IS NULL", id)
	if !allowAll {
		db = db.Where("operator_id=?", actor.ID)
	}
	if e := db.Take(&row).Error; e != nil {
		return nil, e
	}
	var filters string
	exportTaskDB(s.db, actor.Scope).Select("filters").Where("id=?", id).Scan(&filters)
	return map[string]any{"task": row, "filters": unmarshalMap(filters)}, nil
}
func (s *ReportExportService) Download(id uint64, actor Actor, allowAll bool) (string, string, error) {
	var row struct {
		FilePath, FileFormat, Status string
		ExpiresAt                    *time.Time
	}
	db := exportTaskDB(s.db, actor.Scope).Select("file_path,file_format,status,expires_at").Where("id=? AND deleted_at IS NULL", id)
	if !allowAll {
		db = db.Where("operator_id=?", actor.ID)
	}
	if e := db.Take(&row).Error; e != nil {
		return "", "", e
	}
	if row.Status != "SUCCESS" {
		return "", "", NewError("EXPORT_NOT_READY", "export is not ready")
	}
	if row.ExpiresAt == nil || row.ExpiresAt.Before(time.Now()) {
		return "", "", NewError("EXPORT_EXPIRED", "export has expired")
	}
	url, err := storage.PublicURL(context.Background(), row.FilePath)
	if err != nil || strings.TrimSpace(url) == "" {
		return "", "", NewError("EXPORT_FILE_NOT_FOUND", "export file not found")
	}
	_ = reportAudit(s.db, actor, "report:export:download", "report.export.download", "", id, nil, nil, "")
	return url, row.FileFormat, nil
}
func (s *ReportExportService) Retry(id uint64, reason string, actor Actor) error {
	if strings.TrimSpace(reason) == "" {
		return NewError("REASON_REQUIRED", "reason is required")
	}
	requestedAt := time.Now()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		result := exportTaskDB(tx, actor.Scope).
			Where("id=? AND status IN ('FAILED','CANCELLED','EXPIRED') AND deleted_at IS NULL", id).
			Updates(map[string]any{"status": "PENDING", "failure_reason": "", "completed_at": nil, "updated_at": requestedAt})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return NewError("INVALID_EXPORT_STATUS", "export cannot be retried")
		}
		if err := reportAudit(tx, actor, "report:export:retry", "report.export.retry", "", id, nil, map[string]any{"status": "PENDING"}, reason); err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: actor.Scope, TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID,
			EventType: eventbus.EventReportExportRequested, AggregateType: "report_export", AggregateID: fmt.Sprint(id),
			Source: "reportadmin", Producer: "reportadmin", IdempotencyKey: fmt.Sprintf("report-export-retry:%d:%d", id, requestedAt.UnixNano()),
			RequestID: actor.CorrelationID, TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID,
			TargetConsumers: []string{eventbus.ConsumerReportExport}, Payload: map[string]any{"taskID": id},
		})
		return err
	}); err != nil {
		return err
	}
	eventbus.SignalDispatch()
	return nil
}
func (s *ReportExportService) Cancel(id uint64, reason string, actor Actor) error {
	if strings.TrimSpace(reason) == "" {
		return NewError("REASON_REQUIRED", "reason is required")
	}
	result := exportTaskDB(s.db, actor.Scope).Where("id=? AND status IN ('PENDING','PROCESSING') AND deleted_at IS NULL", id).Updates(map[string]any{"status": "CANCELLED", "failure_reason": reason, "completed_at": time.Now(), "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return NewError("INVALID_EXPORT_STATUS", "export cannot be cancelled")
	}
	return reportAudit(s.db, actor, "report:export:cancel", "report.export.cancel", "", id, nil, map[string]any{"status": "CANCELLED"}, reason)
}
