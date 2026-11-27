package gameadmin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/eventbus"
	"casino/internal/game/connections"
	"casino/internal/game/domain"
	"casino/internal/model"
)

func (s *Service) ListProviders(q ListQuery) (PageResult, error) {
	if err := s.recoverStaleProviderSyncs(context.Background(), q.Scope, ""); err != nil {
		return PageResult{}, err
	}
	var rows []providerRecord
	db := s.db.Table("game_game_providers p").
		Select("p.id, p.provider_code, p.name, p.logo_url, p.api_base_url, p.environment, p.api_status, p.status, p.wallet_mode, p.callback_url, p.supported_countries, p.supported_currencies, p.maintenance_message, p.sort, p.last_tested_at, p.last_sync_at, p.created_at, p.updated_at, COUNT(g.id) AS game_count, COALESCE((SELECT sl.id FROM game_provider_sync_logs AS sl WHERE sl.site_id = p.site_id AND sl.provider_code = p.provider_code AND sl.deleted_at IS NULL ORDER BY sl.id DESC LIMIT 1), 0) AS latest_sync_id, COALESCE((SELECT sl.status FROM game_provider_sync_logs AS sl WHERE sl.site_id = p.site_id AND sl.provider_code = p.provider_code AND sl.deleted_at IS NULL ORDER BY sl.id DESC LIMIT 1), '') AS latest_sync_status, (SELECT sl.started_at FROM game_provider_sync_logs AS sl WHERE sl.site_id = p.site_id AND sl.provider_code = p.provider_code AND sl.deleted_at IS NULL ORDER BY sl.id DESC LIMIT 1) AS latest_sync_started_at").
		Joins("LEFT JOIN game_games g ON g.provider_id=p.id AND g.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").Group("p.id")
	db = providerScope(db, "p", q.Scope)
	db = contains(db, "p.provider_code", q.Values["providerCode"])
	db = contains(db, "p.name", q.Values["providerName"])
	db = equal(db, "p.environment", q.Values["environment"])
	if status := strings.ToUpper(strings.TrimSpace(q.Values["status"])); status != "" {
		if status == "ENABLED" {
			db = db.Where("p.status=1")
		} else if status == "MAINTENANCE" {
			db = db.Where("p.status=2")
		} else if status == "DISABLED" {
			db = db.Where("p.status=3")
		}
	}
	db = timeRange(db, "p.last_sync_at", q.Values["lastSyncFrom"], q.Values["lastSyncTo"])
	return paginate(db.Order("p.sort ASC,p.id ASC"), q, &rows)
}

func (s *Service) ProviderDetail(code string, scope content.Scope) (map[string]any, error) {
	var row providerRecord
	db := s.db.Table("game_game_providers p").Select(`p.*,COUNT(g.id) game_count`).
		Joins("LEFT JOIN game_games g ON g.provider_id=p.id AND g.deleted_at IS NULL").
		Where("p.provider_code=? AND p.deleted_at IS NULL", strings.ToLower(strings.TrimSpace(code))).Group("p.id")
	db = providerScope(db, "p", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "PROVIDER_NOT_FOUND")
	}
	registered := false
	if _, err := s.runtime.ProviderRegistry.Get(row.ProviderCode); err == nil {
		registered = true
	}
	var syncs, launches int64
	_ = providerScope(s.db.Table("game_provider_sync_logs"), "", scope).Where("provider_code=? AND deleted_at IS NULL", row.ProviderCode).Count(&syncs).Error
	_ = providerScope(s.db.Table("game_game_launch_logs"), "", scope).Where("provider_id=? AND deleted_at IS NULL", row.ID).Count(&launches).Error
	return map[string]any{
		"overview": row,
		"apiConfiguration": map[string]any{
			"apiBaseUrl":            row.APIBaseURL,
			"environment":           row.Environment,
			"walletMode":            row.WalletMode,
			"supportedCountries":    row.SupportedCountries,
			"supportedCurrencies":   row.SupportedCurrencies,
			"maintenance":           row.Status == 2,
			"maintenanceMessage":    row.MaintenanceMessage,
			"credentialsConfigured": registered,
			"apiPassword":           "***",
			"saltKey":               "***",
		},
		"callbackConfiguration": map[string]any{"providerCode": row.ProviderCode, "callbackUrl": row.CallbackURL, "configured": registered, "credentials": "***"},
		"statistics":            map[string]any{"gameCount": row.GameCount, "syncCount": syncs, "launchCount": launches},
		"operationLogs":         s.operationLogs("game_provider", row.ID),
	}, nil
}

func (s *Service) SetProviderStatus(code string, command ReasonCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row providerRecord
		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("provider_code=? AND deleted_at IS NULL", strings.ToLower(strings.TrimSpace(code))).Take(&row).Error; err != nil {
			return notFound(err, "PROVIDER_NOT_FOUND")
		}
		before := row.Status
		after := 3
		if command.Enabled {
			after = 1
		}
		if before == after {
			return newError("PROVIDER_STATUS_UNCHANGED", "provider status is unchanged")
		}
		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("id=?", row.ID).Updates(map[string]any{"status": after}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "game:provider:update", "update_status", "game_provider", row.ID, map[string]any{"status": before}, map[string]any{"status": after}, command.Reason)
	})
}

func (s *Service) SetProviderMaintenance(code string, command ProviderMaintenanceCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row providerRecord
		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("provider_code=? AND deleted_at IS NULL", strings.ToLower(strings.TrimSpace(code))).Take(&row).Error; err != nil {
			return notFound(err, "PROVIDER_NOT_FOUND")
		}
		after := 1
		if command.Maintenance {
			after = 2
		}
		updates := map[string]any{
			"status":              after,
			"maintenance_message": strings.TrimSpace(command.MaintenanceMessage),
		}
		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("id=?", row.ID).Updates(updates).Error; err != nil {
			return err
		}
		before := map[string]any{"status": row.Status, "maintenanceMessage": row.MaintenanceMessage}
		afterState := map[string]any{"status": after, "maintenanceMessage": updates["maintenance_message"]}
		return audit(tx, actor, "game:provider:update", "maintenance", "game_provider", row.ID, before, afterState, command.Reason)
	})
}

func (s *Service) TestProvider(ctx context.Context, code string, actor Actor) (map[string]any, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	provider, err := s.runtime.ProviderRegistry.Get(code)
	if err != nil {
		return nil, newError("PROVIDER_NOT_REGISTERED", "provider adapter is not registered")
	}
	started := time.Now()
	tester, ok := provider.(domain.ProviderConnectionTester)
	if !ok {
		return nil, newError("PROVIDER_TEST_UNSUPPORTED", "provider connection test is not implemented")
	}
	err = tester.TestConnection(ctx)
	status := "HEALTHY"
	message := ""
	if err != nil {
		status = "UNHEALTHY"
		message = connections.New(s.db).SanitizeError(code, err)
	}
	var row providerRecord
	if dbErr := providerScope(s.db.Table("game_game_providers"), "", actor.Scope).Where("provider_code=? AND deleted_at IS NULL", code).Take(&row).Error; dbErr != nil {
		return nil, notFound(dbErr, "PROVIDER_NOT_FOUND")
	}
	now := time.Now()
	if dbErr := s.db.Transaction(func(tx *gorm.DB) error {
		if updateErr := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("id=?", row.ID).Updates(map[string]any{"api_status": status, "last_tested_at": now}).Error; updateErr != nil {
			return updateErr
		}
		return audit(tx, actor, "game:provider:test", "test_connection", "game_provider", row.ID, map[string]any{"apiStatus": row.APIStatus}, map[string]any{"apiStatus": status}, "provider connection test")
	}); dbErr != nil {
		return nil, dbErr
	}
	result := map[string]any{"status": status, "durationMs": time.Since(started).Milliseconds(), "message": message}
	if err != nil {
		return result, newError("PROVIDER_TEST_FAILED", "provider connection test failed "+message)
	}
	return result, nil
}

func (s *Service) DeleteProvider(ctx context.Context, code string, command ReasonCommand, actor Actor) (map[string]any, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return nil, newError("INVALID_PROVIDER", "provider code is required")
	}

	result := map[string]any{"providerCode": code, "deletedCount": int64(0), "providerDeleted": false}
	if err := s.recoverStaleProviderSyncs(ctx, actor.Scope, code); err != nil {
		return nil, err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider providerRecord
		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("provider_code=? AND deleted_at IS NULL", code).Take(&provider).Error; err != nil {
			return notFound(err, "PROVIDER_NOT_FOUND")
		}

		var running int64
		if err := providerScope(tx.Table("game_provider_sync_logs"), "", actor.Scope).Where("provider_code=? AND status=? AND deleted_at IS NULL", code, "RUNNING").Count(&running).Error; err != nil {
			return err
		}
		if running > 0 {
			return newError("PROVIDER_SYNC_RUNNING", "provider synchronization is running")
		}

		var gameIDs []uint64
		if err := providerScope(tx.Table("game_games"), "", actor.Scope).Where("provider_id=?", provider.ID).Pluck("id", &gameIDs).Error; err != nil {
			return err
		}
		if len(gameIDs) > 0 {
			for _, table := range []string{
				"game_game_category_relations",
				"game_game_tag_relations",
				"game_game_collection_relations",
				"game_promotion_eligible_games",
				"game_user_favorite_games",
			} {
				if tx.Migrator().HasTable(table) {
					if err := tx.Exec("DELETE FROM "+table+" WHERE game_id IN ?", gameIDs).Error; err != nil {
						return err
					}
				}
			}
			if err := providerScope(tx.Unscoped().Model(&model.GameGames{}), "", actor.Scope).Where("provider_id=?", provider.ID).Delete(&model.GameGames{}).Error; err != nil {
				return err
			}
		}

		mappingTable := (model.GameProviderCategoryMappings{}).TableName()
		if tx.Migrator().HasTable(mappingTable) {
			if err := providerScope(tx.Unscoped().Model(&model.GameProviderCategoryMappings{}), "", actor.Scope).Where("provider_code = ?", code).Delete(&model.GameProviderCategoryMappings{}).Error; err != nil {
				return err
			}
		}

		if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", provider.ID).Updates(map[string]any{
			"status":       3,
			"last_sync_at": nil,
			"deleted_at":   time.Now(),
		}).Error; err != nil {
			return err
		}
		result["deletedCount"] = int64(len(gameIDs))
		result["providerDeleted"] = true
		return audit(tx, actor, "game:provider:update", "delete", "game_provider", provider.ID,
			map[string]any{"providerCode": code, "status": provider.Status, "gameCount": len(gameIDs)},
			map[string]any{"providerCode": code, "deleted": true, "gameCount": 0}, command.Reason)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type syncLogRecord struct {
	SiteID         uint64     `gorm:"column:site_id" json:"siteID"`
	ID             uint64     `json:"id"`
	SyncNo         string     `json:"syncID"`
	ProviderCode   string     `json:"providerCode"`
	Status         string     `json:"status"`
	NewCount       int        `json:"newCount"`
	UpdatedCount   int        `json:"updatedCount"`
	UnchangedCount int        `json:"unchangedCount"`
	DisabledCount  int        `json:"disabledCount"`
	FailedCount    int        `json:"failedCount"`
	DurationMS     int64      `json:"durationMs"`
	TriggeredByID  uint64     `json:"triggeredByID"`
	TriggeredBy    string     `json:"triggeredBy"`
	ErrorMessage   string     `json:"errorMessage"`
	RawResponse    string     `gorm:"column:raw_provider_response" json:"-"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func (syncLogRecord) TableName() string { return "game_provider_sync_logs" }

func (s *Service) SyncProvider(ctx context.Context, code string, reason string, actor Actor) (syncLogRecord, error) {
	if !actor.Scope.Platform {
		return syncLogRecord{}, newError("PLATFORM_SCOPE_REQUIRED", "provider synchronization is platform-only")
	}
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return syncLogRecord{}, newError("INVALID_PROVIDER", "provider code is required")
	}
	if _, err := s.runtime.ProviderRegistry.Get(code); err != nil {
		return syncLogRecord{}, newError("PROVIDER_NOT_REGISTERED", "provider adapter is not registered")
	}
	if _, err := s.runtime.Repository.ProviderByCode(ctx, code); err != nil {
		return syncLogRecord{}, notFound(err, "PROVIDER_NOT_FOUND")
	}
	if err := s.recoverStaleProviderSyncs(ctx, actor.Scope, code); err != nil {
		return syncLogRecord{}, err
	}
	var running syncLogRecord
	if err := providerScope(s.db.WithContext(ctx).Model(&syncLogRecord{}), "", actor.Scope).Where("provider_code=? AND status IN ('PENDING','RUNNING') AND deleted_at IS NULL", code).Order("id DESC").Take(&running).Error; err == nil {
		return running, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return syncLogRecord{}, err
	}
	start := time.Now()
	row := syncLogRecord{SiteID: actor.Scope.SiteID, SyncNo: generatedCode("SYNC"), ProviderCode: code, Status: "PENDING", TriggeredByID: actor.ID, TriggeredBy: actor.Username, StartedAt: start}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: actor.Scope, SiteID: actor.Scope.SiteID,
			EventType: eventbus.EventProviderSyncRequested, AggregateType: "provider_sync", AggregateID: fmt.Sprint(row.ID),
			Source: "gameadmin", Producer: "gameadmin", IdempotencyKey: "provider-sync:" + row.SyncNo,
			RequestID: actor.CorrelationID, TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID,
			TargetConsumers: []string{eventbus.ConsumerProviderSync},
			Payload:         map[string]any{"syncLogID": row.ID, "providerCode": code, "reason": reason, "operatorID": actor.ID, "operatorUsername": actor.Username},
		})
		return err
	}); err != nil {
		return row, err
	}
	eventbus.SignalDispatch()
	return row, nil
}

func (s *Service) executeProviderSync(row syncLogRecord, code string, reason string, actor Actor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	summary, syncErr := s.runtime.GameService.SyncProviderGamesScoped(ctx, actor.Scope, code)
	var currentStatus string
	if err := s.db.Model(&syncLogRecord{}).Select("status").Where("id=?", row.ID).Scan(&currentStatus).Error; err != nil {
		return err
	}
	if currentStatus == "CANCELLED" {
		return nil
	}
	completed := time.Now()
	row.CompletedAt = &completed
	row.DurationMS = completed.Sub(row.StartedAt).Milliseconds()
	row.NewCount, row.UpdatedCount, row.UnchangedCount, row.DisabledCount, row.FailedCount = summary.Inserted, summary.Updated, summary.Unchanged, summary.Disabled, summary.Failed
	row.RawResponse = summary.RawResponse
	row.Status = "SUCCESS"
	if syncErr != nil || summary.Failed > 0 {
		row.Status = "FAILED"
		if syncErr != nil {
			row.ErrorMessage = truncateSyncError(syncErr.Error())
		}
	}
	var provider providerRecord
	providerErr := providerScope(s.db.Table("game_game_providers"), "", actor.Scope).Where("provider_code=? AND deleted_at IS NULL", code).Take(&provider).Error
	if providerErr != nil && !errors.Is(providerErr, gorm.ErrRecordNotFound) {
		row.Status = "FAILED"
		row.ErrorMessage = truncateSyncError(providerErr.Error())
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&row).Updates(map[string]any{"status": row.Status, "new_count": row.NewCount, "updated_count": row.UpdatedCount, "unchanged_count": row.UnchangedCount, "disabled_count": row.DisabledCount, "failed_count": row.FailedCount, "duration_ms": row.DurationMS, "error_message": row.ErrorMessage, "raw_provider_response": row.RawResponse, "completed_at": completed}).Error; err != nil {
			return err
		}
		for _, item := range summary.Items {
			errorMessage := item.ErrorMessage
			if len(errorMessage) > 500 {
				errorMessage = errorMessage[:500]
			}
			if err := tx.Table("game_provider_sync_log_items").Create(map[string]any{
				"site_id":     actor.Scope.SiteID,
				"sync_log_id": row.ID, "provider_game_id": item.ProviderGameID, "game_name": item.GameName,
				"change_type": item.ChangeType, "error_message": errorMessage, "created_at": completed, "updated_at": completed,
			}).Error; err != nil {
				return err
			}
		}
		if provider.ID > 0 {
			if err := providerScope(tx.Table("game_game_providers"), "", actor.Scope).Where("id=?", provider.ID).Updates(map[string]any{"last_sync_at": completed}).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, "game:provider:sync", "synchronize", "game_provider", provider.ID, nil, map[string]any{"syncID": row.SyncNo, "new": row.NewCount, "updated": row.UpdatedCount, "disabled": row.DisabledCount, "failed": row.FailedCount}, reason)
	}); err != nil {
		_ = s.db.Model(&row).Updates(map[string]any{"status": "FAILED", "error_message": truncateSyncError(err.Error()), "completed_at": completed}).Error
		return err
	}
	if syncErr != nil {
		return syncErr
	}
	if summary.Failed > 0 {
		return newError("PROVIDER_SYNC_PARTIAL_FAILED", "provider synchronization contains failed games")
	}
	return nil
}

func providerScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	if !scope.Platform {
		return content.ApplySite(db, alias, scope)
	}
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = strings.TrimSpace(alias) + "."
	}
	return db.Where(prefix + "site_id=0")
}

func (s *Service) recoverStaleProviderSyncs(ctx context.Context, scope content.Scope, providerCode string) error {
	cutoff := time.Now().Add(-35 * time.Minute)
	db := providerScope(s.db.WithContext(ctx).Model(&syncLogRecord{}), "", scope).
		Where("status IN ('PENDING','RUNNING') AND started_at<? AND deleted_at IS NULL", cutoff)
	if providerCode = strings.ToLower(strings.TrimSpace(providerCode)); providerCode != "" {
		db = db.Where("provider_code=?", providerCode)
	}
	completed := time.Now()
	return db.Updates(map[string]any{"status": "FAILED", "completed_at": completed,
		"error_message": "provider synchronization timed out and was recovered"}).Error
}

func (s *Service) CancelProviderSync(ctx context.Context, code string, reason string, actor Actor) (syncLogRecord, error) {
	if !actor.Scope.Platform {
		return syncLogRecord{}, newError("PLATFORM_SCOPE_REQUIRED", "provider synchronization is platform-only")
	}
	code = strings.ToLower(strings.TrimSpace(code))
	reason = strings.TrimSpace(reason)
	if code == "" || reason == "" {
		return syncLogRecord{}, newError("INVALID_REQUEST", "provider code and reason are required")
	}
	if err := s.recoverStaleProviderSyncs(ctx, actor.Scope, code); err != nil {
		return syncLogRecord{}, err
	}
	var row syncLogRecord
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := providerScope(tx.Model(&syncLogRecord{}), "", actor.Scope).
			Where("provider_code=? AND status IN ('PENDING','RUNNING') AND deleted_at IS NULL", code).
			Order("id DESC").Take(&row).Error; err != nil {
			return notFound(err, "PROVIDER_SYNC_NOT_RUNNING")
		}
		completed := time.Now()
		updates := map[string]any{"status": "CANCELLED", "completed_at": completed, "duration_ms": completed.Sub(row.StartedAt).Milliseconds(), "error_message": "cancelled: " + reason}
		result := tx.Model(&syncLogRecord{}).Where("id=? AND status IN ('PENDING','RUNNING')", row.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return newError("PROVIDER_SYNC_NOT_RUNNING", "provider synchronization is not running")
		}
		row.Status, row.CompletedAt, row.ErrorMessage = "CANCELLED", &completed, "cancelled: "+reason
		return audit(tx, actor, "game:provider:sync", "cancel_sync", "provider_sync", row.ID,
			map[string]any{"status": "RUNNING", "providerCode": code}, updates, reason)
	})
	return row, err
}

func (s *Service) HandleProviderSyncEvent(ctx eventbus.Context) error {
	id := payloadUint64(ctx.Event.Payload["syncLogID"])
	if id == 0 {
		return newError("PROVIDER_SYNC_LOG_REQUIRED", "provider sync log is required")
	}
	var row syncLogRecord
	if err := s.db.Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return err
	}
	if row.SiteID != ctx.Event.SiteID {
		return newError("PROVIDER_SYNC_SCOPE_MISMATCH", "provider sync is outside event scope")
	}
	if row.Status == "SUCCESS" {
		return nil
	}
	claimedAt := time.Now()
	claimed := s.db.Model(&syncLogRecord{}).Where("id=? AND (status IN ('PENDING','FAILED') OR (status='RUNNING' AND started_at<?))", row.ID, claimedAt.Add(-35*time.Minute)).Updates(map[string]any{"status": "RUNNING", "error_message": "", "started_at": claimedAt})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected == 0 {
		return newError("PROVIDER_SYNC_ALREADY_RUNNING", "provider synchronization is already running")
	}
	row.Status = "RUNNING"
	row.StartedAt = claimedAt
	actor := Actor{
		ID: payloadUint64(ctx.Event.Payload["operatorID"]), Username: strings.TrimSpace(fmt.Sprint(ctx.Event.Payload["operatorUsername"])),
		Scope: content.Scope{SiteID: row.SiteID, Platform: row.SiteID == 0}, CorrelationID: ctx.Event.CorrelationID,
	}
	return s.executeProviderSync(row, row.ProviderCode, strings.TrimSpace(fmt.Sprint(ctx.Event.Payload["reason"])), actor)
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
	}
	return 0
}

func truncateSyncError(message string) string {
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func (s *Service) SyncAll(ctx context.Context, reason string, actor Actor) ([]syncLogRecord, error) {
	if !actor.Scope.Platform {
		return nil, newError("PLATFORM_SCOPE_REQUIRED", "provider synchronization is platform-only")
	}
	eligible := map[string]bool{}
	for _, code := range connections.New(s.db).PublicCodes(ctx) {
		eligible[code] = true
	}
	providers := s.runtime.ProviderRegistry.List()
	results := make([]syncLogRecord, 0, len(providers))
	var failed []string
	for _, provider := range providers {
		if !eligible[provider.Code()] {
			continue
		}
		result, err := s.SyncProvider(ctx, provider.Code(), reason, actor)
		results = append(results, result)
		if err != nil {
			failed = append(failed, provider.Code())
		}
	}
	if len(failed) > 0 {
		return results, newError("PROVIDER_SYNC_PARTIAL_FAILED", fmt.Sprintf("failed providers: %s", strings.Join(failed, ",")))
	}
	return results, nil
}

func providerGames(ctx context.Context, provider domain.GameProvider) ([]domain.Game, error) {
	return provider.GetGames(ctx, domain.GetGamesRequest{})
}

func (s *Service) ensureRegisteredProviders(ctx context.Context) error {
	for _, provider := range s.runtime.ProviderRegistry.List() {
		if err := s.ensureRegisteredProvider(ctx, provider.Code()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureRegisteredProvider(ctx context.Context, code string) error {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return nil
	}
	if _, err := s.runtime.ProviderRegistry.Get(code); err != nil {
		return nil
	}
	var deletedCount int64
	if err := s.db.WithContext(ctx).Unscoped().Model(&model.GameGameProviders{}).
		Where("provider_code=? AND deleted_at IS NOT NULL", code).Count(&deletedCount).Error; err != nil {
		return err
	}
	if deletedCount > 0 {
		return nil
	}
	name, endpoint := connections.New(s.db).Description(code)
	row, err := s.runtime.Repository.EnsureProvider(ctx, code, name, endpoint, true)
	if err == nil && code == "allintop" && row.SupportedCountries == "" && row.SupportedCurrencies == "" {
		// The Nacos account is IDR-only; never infer its market from the deployed site.
		err = s.db.WithContext(ctx).Model(row).Updates(map[string]any{"supported_countries": "ID", "supported_currencies": "IDR", "environment": "sandbox"}).Error
	}
	return err
}
