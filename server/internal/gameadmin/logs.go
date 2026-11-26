package gameadmin

import (
	"strings"

	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

const syncRawPreviewLimit = 20_000

func (s *Service) ListSyncLogs(q ListQuery) (PageResult, error) {
	var rows []syncLogRecord
	db := s.db.Model(&syncLogRecord{}).Where("deleted_at IS NULL")
	db = content.ApplySite(db, "", q.Scope)
	db = equal(db, "provider_code", q.Values["providerCode"])
	db = equal(db, "status", q.Values["status"])
	db = contains(db, "triggered_by", q.Values["triggeredBy"])
	db = timeRange(db, "started_at", q.Values["startedFrom"], q.Values["startedTo"])
	db = timeRange(db, "completed_at", q.Values["completedFrom"], q.Values["completedTo"])
	return paginate(db.Order("id DESC"), q, &rows)
}

type syncItemRecord struct {
	ID             uint64 `json:"id"`
	ProviderGameID string `json:"providerGameId"`
	GameID         uint64 `json:"gameId"`
	GameName       string `json:"gameName"`
	ChangeType     string `json:"changeType"`
	BeforeState    string `json:"beforeState"`
	AfterState     string `json:"afterState"`
	ErrorMessage   string `json:"errorMessage"`
}

func (s *Service) SyncLogDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row syncLogRecord
	if err := content.ApplySite(s.db.Model(&syncLogRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, notFound(err, "SYNC_LOG_NOT_FOUND")
	}
	var failedTotal, changedTotal int64
	_ = content.ApplySite(s.db.Table("game_provider_sync_log_items"), "", scope).
		Where("sync_log_id=? AND change_type IN ? AND deleted_at IS NULL", id, []string{"FAILED", "THUMBNAIL_FAILED"}).
		Count(&failedTotal).Error
	_ = content.ApplySite(s.db.Table("game_provider_sync_log_items"), "", scope).
		Where("sync_log_id=? AND change_type IN ? AND deleted_at IS NULL", id, []string{"NEW", "UPDATED", "DISABLED"}).
		Count(&changedTotal).Error
	return map[string]any{
		"summary":            row,
		"failedRecordsTotal": failedTotal,
		"changedGamesTotal":  changedTotal,
	}, nil
}

func (s *Service) SyncLogItems(id uint64, itemType string, q ListQuery) (PageResult, error) {
	if err := s.ensureSyncLogExists(id, q.Scope); err != nil {
		return PageResult{}, err
	}
	types, err := syncItemTypes(itemType)
	if err != nil {
		return PageResult{}, err
	}
	var rows []syncItemRecord
	db := content.ApplySite(s.db.Table("game_provider_sync_log_items"), "", q.Scope).
		Where("sync_log_id=? AND change_type IN ? AND deleted_at IS NULL", id, types).
		Order("id ASC")
	return paginate(db, q, &rows)
}

func (s *Service) SyncLogRawResponse(id uint64, scope content.Scope) (map[string]any, error) {
	var row syncLogRecord
	if err := content.ApplySite(s.db.Model(&syncLogRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, notFound(err, "SYNC_LOG_NOT_FOUND")
	}
	return syncRawProviderPreview(row.RawResponse), nil
}

func (s *Service) SyncLogOperationLogs(id uint64, q ListQuery) (PageResult, error) {
	var row syncLogRecord
	if err := content.ApplySite(s.db.Model(&syncLogRecord{}), "", q.Scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return PageResult{}, notFound(err, "SYNC_LOG_NOT_FOUND")
	}
	var rows []model.GameAdminOperationLogs
	db := content.ApplySite(s.db.Model(&model.GameAdminOperationLogs{}), "", q.Scope).Where(
		"target_type='game_provider' AND action='synchronize' AND JSON_UNQUOTE(JSON_EXTRACT(after_state, '$.syncID'))=?",
		row.SyncNo,
	).Order("id DESC")
	return paginate(db, q, &rows)
}

func (s *Service) ensureSyncLogExists(id uint64, scope content.Scope) error {
	var count int64
	if err := content.ApplySite(s.db.Model(&syncLogRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return &Error{Code: "SYNC_LOG_NOT_FOUND", Message: "sync log not found"}
	}
	return nil
}

func syncItemTypes(itemType string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(itemType)) {
	case "failed", "failedrecords", "failed_records":
		return []string{"FAILED", "THUMBNAIL_FAILED"}, nil
	case "changed", "changedgames", "changed_games":
		return []string{"NEW", "UPDATED", "DISABLED"}, nil
	default:
		return nil, &Error{Code: "INVALID_REQUEST", Message: "invalid sync log item type"}
	}
}

func syncRawProviderPreview(raw string) map[string]any {
	size := len(raw)
	if strings.TrimSpace(raw) == "" {
		return map[string]any{"sizeBytes": 0, "truncated": false, "body": map[string]any{}}
	}
	if size > syncRawPreviewLimit {
		return map[string]any{
			"sizeBytes": size,
			"truncated": true,
			"preview":   raw[:syncRawPreviewLimit],
		}
	}
	return map[string]any{"sizeBytes": size, "truncated": false, "body": redactJSON(raw)}
}

func (s *Service) syncOperationLogs(syncNo string) []model.GameAdminOperationLogs {
	var rows []model.GameAdminOperationLogs
	_ = s.db.Where("target_type='game_provider' AND action='synchronize' AND JSON_UNQUOTE(JSON_EXTRACT(after_state, '$.syncID'))=?", syncNo).Order("id DESC").Limit(100).Find(&rows).Error
	return rows
}

type launchLogRecord struct {
	ID                uint64 `json:"id"`
	LaunchID          string `gorm:"column:launch_no" json:"launchID"`
	SessionID         string `json:"sessionID"`
	UserID            uint64 `json:"userID"`
	UserNo            string `json:"userNo"`
	Username          string `json:"username"`
	GameID            uint64 `json:"gameID"`
	GameCode          string `json:"gameCode"`
	GameName          string `json:"gameName"`
	ProviderCode      string `json:"providerCode"`
	ProviderGameID    string `json:"providerGameId"`
	LaunchType        string `json:"launchType"`
	Currency          string `json:"currency"`
	Language          string `json:"language"`
	Country           string `json:"country"`
	CountryID         uint64 `json:"countryID"`
	CountryName       string `json:"countryName"`
	Device            string `json:"device"`
	Status            string `json:"status"`
	FailureReason     string `gorm:"column:error_message" json:"failureReason"`
	ClientIP          string `json:"clientIP"`
	ProviderRequest   string `gorm:"column:provider_request" json:"-"`
	ProviderResponse  string `gorm:"column:provider_response" json:"-"`
	CorrelationID     string `json:"correlationID"`
	LaunchTime        string `json:"launchTime"`
	LaunchSuccessTime string `json:"launchSuccessTime"`
	ExitTime          string `json:"exitTime"`
	CloseTime         string `json:"closeTime"`
	LastHeartbeatTime string `json:"lastHeartbeatTime"`
	DurationSeconds   uint64 `json:"durationSeconds"`
	ExitReason        string `json:"exitReason"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

func (s *Service) ListLaunchLogs(q ListQuery) (PageResult, error) {
	var rows []launchLogRecord
	db := s.db.Table("game_game_launch_logs l").Select(`l.id,l.launch_no,l.session_id,l.user_id,u.user_no,CASE WHEN NULLIF(u.phone,'') IS NOT NULL THEN TRIM(CONCAT_WS(' ',NULLIF(u.phone_country_code,''),u.phone)) ELSE u.username END username,l.game_id,g.game_code,g.name game_name,p.provider_code,g.provider_game_code provider_game_id,l.launch_type,
		COALESCE(NULLIF(l.currency_code_snapshot,''),l.currency) currency,COALESCE(NULLIF(l.language_code_snapshot,''),l.language) language,
		COALESCE(NULLIF(l.country_code_snapshot,''),l.country) country,l.country_id,COALESCE(NULLIF(cc.country_name_en,''),cc.country_name_zh) country_name,l.device,
		COALESCE(NULLIF(l.session_status,''),CASE l.launch_status WHEN 1 THEN 'LAUNCH_REQUESTED' WHEN 2 THEN 'LAUNCH_SUCCESS' ELSE 'LAUNCH_FAILED' END) status,
		l.error_message,l.client_ip,l.correlation_id,l.launch_time,l.launch_success_time,l.exit_time,l.close_time,l.last_heartbeat_time,l.duration_seconds,l.exit_reason,l.created_at,l.updated_at`).
		Joins("LEFT JOIN game_users u ON u.id=l.user_id AND u.deleted_at IS NULL").
		Joins("JOIN game_games g ON g.id=l.game_id").
		Joins("JOIN game_game_providers p ON p.id=l.provider_id").
		Joins("LEFT JOIN game_country_configs cc ON cc.id=l.country_id AND cc.deleted_at IS NULL").
		Where("l.deleted_at IS NULL")
	db = content.ApplySite(db, "l", q.Scope)
	db = content.ApplySite(db, "u", q.Scope)
	db = applyGameCatalogReadableScope(db, "g", q.Scope)
	db = applyGameCatalogReadableScope(db, "p", q.Scope)
	db = contains(db, "l.session_id", q.Values["sessionID"])
	if user := strings.TrimSpace(q.Values["userID"]); user != "" {
		db = db.Where("u.user_no=? OR u.id=?", user, parseID(user))
	}
	if username := strings.TrimSpace(q.Values["username"]); username != "" {
		pattern := "%" + username + "%"
		db = db.Where("u.username LIKE ? OR u.phone LIKE ? OR TRIM(CONCAT_WS(' ',NULLIF(u.phone_country_code,''),u.phone)) LIKE ?", pattern, pattern, pattern)
	}
	db = equal(db, "p.provider_code", q.Values["providerCode"])
	if game := strings.TrimSpace(q.Values["gameID"]); game != "" {
		db = db.Where("g.game_code=? OR g.id=?", game, parseID(game))
	}
	db = contains(db, "g.provider_game_code", q.Values["providerGameID"])
	db = equal(db, "l.launch_type", q.Values["launchType"])
	db = equal(db, "COALESCE(NULLIF(l.currency_code_snapshot,''),l.currency)", q.Values["currency"])
	db = equal(db, "COALESCE(NULLIF(l.language_code_snapshot,''),l.language)", q.Values["language"])
	db = equal(db, "COALESCE(NULLIF(l.country_code_snapshot,''),l.country)", q.Values["country"])
	db = equal(db, "l.device", q.Values["device"])
	if status := q.Values["status"]; status != "" {
		mapping := map[string]int{"PENDING": 1, "SUCCESS": 2, "FAILED": 3}
		if value, ok := mapping[status]; ok {
			db = db.Where("l.launch_status=?", value)
		} else {
			db = db.Where("l.session_status=?", status)
		}
	}
	db = timeRange(db, "l.created_at", q.Values["createdFrom"], q.Values["createdTo"])
	return paginate(db.Order("l.id DESC"), q, &rows)
}

func (s *Service) LaunchLogDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row launchLogRecord
	db := s.db.Table("game_game_launch_logs l").Select(`l.id,l.launch_no,l.session_id,l.user_id,u.user_no,CASE WHEN NULLIF(u.phone,'') IS NOT NULL THEN TRIM(CONCAT_WS(' ',NULLIF(u.phone_country_code,''),u.phone)) ELSE u.username END username,l.game_id,g.game_code,g.name game_name,p.provider_code,g.provider_game_code provider_game_id,l.launch_type,
		COALESCE(NULLIF(l.currency_code_snapshot,''),l.currency) currency,COALESCE(NULLIF(l.language_code_snapshot,''),l.language) language,
		COALESCE(NULLIF(l.country_code_snapshot,''),l.country) country,l.country_id,COALESCE(NULLIF(cc.country_name_en,''),cc.country_name_zh) country_name,l.device,
		COALESCE(NULLIF(l.session_status,''),CASE l.launch_status WHEN 1 THEN 'LAUNCH_REQUESTED' WHEN 2 THEN 'LAUNCH_SUCCESS' ELSE 'LAUNCH_FAILED' END) status,
			l.error_message,l.client_ip,COALESCE(NULLIF(l.provider_launch_request,''),l.provider_request) provider_request,COALESCE(NULLIF(l.provider_launch_response,''),l.provider_response) provider_response,l.correlation_id,
		l.launch_time,l.launch_success_time,l.exit_time,l.close_time,l.last_heartbeat_time,l.duration_seconds,l.exit_reason,l.created_at,l.updated_at`).
		Joins("LEFT JOIN game_users u ON u.id=l.user_id").
		Joins("JOIN game_games g ON g.id=l.game_id").
		Joins("JOIN game_game_providers p ON p.id=l.provider_id").
		Joins("LEFT JOIN game_country_configs cc ON cc.id=l.country_id AND cc.deleted_at IS NULL").
		Where("l.id=? AND l.deleted_at IS NULL", id)
	db = content.ApplySite(db, "l", scope)
	db = content.ApplySite(db, "u", scope)
	db = applyGameCatalogReadableScope(db, "g", scope)
	db = applyGameCatalogReadableScope(db, "p", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "LAUNCH_LOG_NOT_FOUND")
	}
	timeline := []map[string]any{{"status": "LAUNCH_REQUESTED", "time": row.LaunchTime}}
	if row.LaunchSuccessTime != "" {
		timeline = append(timeline, map[string]any{"status": "LAUNCH_SUCCESS", "time": row.LaunchSuccessTime})
	}
	if row.LaunchSuccessTime != "" {
		playingTime := row.LastHeartbeatTime
		if playingTime == "" {
			playingTime = row.LaunchSuccessTime
		}
		timeline = append(timeline, map[string]any{"status": "PLAYING", "time": playingTime})
	}
	if row.ExitTime != "" {
		timeline = append(timeline, map[string]any{"status": "EXITED", "time": row.ExitTime, "exitReason": row.ExitReason})
	}
	if row.CloseTime != "" {
		timeline = append(timeline, map[string]any{"status": "CLOSED", "time": row.CloseTime, "durationSeconds": row.DurationSeconds})
	}
	country := map[string]any{"countryID": row.CountryID, "countryCode": row.Country, "countryName": row.CountryName, "currency": row.Currency, "language": row.Language}
	return map[string]any{"overview": row, "sessionTimeline": timeline, "providerRequest": row.ProviderRequest, "providerResponse": row.ProviderResponse, "countryConfiguration": country, "operationLogs": s.operationLogs("game_launch", id)}, nil
}

// Launch records belong to the site, while games and providers are normally
// platform-owned catalog records. The joins must therefore allow the platform
// catalog without widening the H5 launch record scope.
func applyGameCatalogReadableScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = strings.TrimSpace(alias) + "."
	}
	if scope.Platform {
		return db.Where(prefix + "site_id = 0")
	}
	if !scope.SiteScoped() {
		return db.Where("1 = 0")
	}
	return db.Where(prefix+"site_id IN (0, ?)", scope.SiteID)
}
