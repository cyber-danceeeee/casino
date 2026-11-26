package betadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"casino/internal/adminauth"
	"casino/internal/content"
	"casino/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	reconciliationSourceInternal = "INTERNAL"

	reconciliationRunRunning   = "RUNNING"
	reconciliationRunCompleted = "COMPLETED"
	reconciliationRunFailed    = "FAILED"

	reconciliationStatusOpen          = "OPEN"
	reconciliationStatusConfirmed     = "CONFIRMED"
	reconciliationStatusAutoFixed     = "AUTO_FIXED"
	reconciliationStatusManuallyFixed = "MANUALLY_FIXED"
	reconciliationStatusIgnored       = "IGNORED"

	discrepancyMissingPlatformOrder = "MISSING_PLATFORM_ORDER"
	discrepancyMissingProviderOrder = "MISSING_PROVIDER_ORDER"
	discrepancyBetAmountMismatch    = "BET_AMOUNT_MISMATCH"
	discrepancyWinAmountMismatch    = "WIN_AMOUNT_MISMATCH"
	discrepancyStatusMismatch       = "STATUS_MISMATCH"
	discrepancyDuplicateCallback    = "DUPLICATE_CALLBACK"
	discrepancyMissingWalletEntry   = "MISSING_WALLET_ENTRY"
	discrepancyRollbackNotApplied   = "ROLLBACK_NOT_APPLIED"
)

var allowedReconciliationStatuses = map[string]bool{
	reconciliationStatusOpen:          true,
	reconciliationStatusConfirmed:     true,
	reconciliationStatusAutoFixed:     true,
	reconciliationStatusManuallyFixed: true,
	reconciliationStatusIgnored:       true,
}

type reconciliationStats struct {
	ProviderTransactions int64
	Callbacks            int64
	WalletTransactions   int64
	BetRecords           int64
}

type reconciliationBuild struct {
	Facts []model.GameTransactionFacts
	Items []model.ProviderReconciliationItems
	Stats reconciliationStats
}

type reconcileTransactionRow struct {
	ProviderTransactionRowID uint64 `gorm:"column:provider_transaction_row_id"`

	SiteID                  uint64    `gorm:"column:site_id"`
	ProviderID              uint64    `gorm:"column:provider_id"`
	ProviderCode            string    `gorm:"column:provider_code"`
	ProviderTransactionID   string    `gorm:"column:provider_transaction_id"`
	ProviderRoundID         string    `gorm:"column:provider_round_id"`
	ProviderGameID          string    `gorm:"column:provider_game_id"`
	UserID                  uint64    `gorm:"column:user_id"`
	WalletTransactionID     uint64    `gorm:"column:wallet_transaction_id"`
	Operation               string    `gorm:"column:operation"`
	ProviderStatus          string    `gorm:"column:provider_status"`
	ProviderAmountMinor     int64     `gorm:"column:provider_amount_minor"`
	Currency                string    `gorm:"column:currency"`
	RelatedTransactionID    string    `gorm:"column:related_transaction_id"`
	BalanceBeforeMinor      int64     `gorm:"column:balance_before_minor"`
	BalanceAfterMinor       int64     `gorm:"column:balance_after_minor"`
	SourceCreatedAt         time.Time `gorm:"column:source_created_at"`
	WalletExists            uint64    `gorm:"column:wallet_exists"`
	WalletAmountMinor       int64     `gorm:"column:wallet_amount_minor"`
	WalletStatus            int       `gorm:"column:wallet_status"`
	WalletDirection         string    `gorm:"column:wallet_direction"`
	WalletCurrency          string    `gorm:"column:wallet_currency"`
	WalletOrderNo           string    `gorm:"column:wallet_order_no"`
	WalletTxnNo             string    `gorm:"column:wallet_txn_no"`
	BetRecordID             uint64    `gorm:"column:bet_record_id"`
	BetAmountMinor          int64     `gorm:"column:bet_amount_minor"`
	WinAmountMinor          int64     `gorm:"column:win_amount_minor"`
	BetStatus               string    `gorm:"column:bet_status"`
	RollbackStatus          string    `gorm:"column:rollback_status"`
	BetProviderWinTx        string    `gorm:"column:bet_provider_win_tx"`
	RollbackRecordID        uint64    `gorm:"column:rollback_record_id"`
	RollbackAmountMinor     int64     `gorm:"column:rollback_amount_minor"`
	RollbackRecordStatus    string    `gorm:"column:rollback_record_status"`
	CallbackLogID           uint64    `gorm:"column:callback_log_id"`
	SuccessfulCallbackCount int64     `gorm:"column:successful_callback_count"`
}

type unmatchedCallbackRow struct {
	CallbackLogID uint64 `gorm:"column:callback_log_id"`

	SiteID                  uint64    `gorm:"column:site_id"`
	ProviderID              uint64    `gorm:"column:provider_id"`
	ProviderCode            string    `gorm:"column:provider_code"`
	ProviderTransactionID   string    `gorm:"column:provider_transaction_id"`
	ProviderRoundID         string    `gorm:"column:provider_round_id"`
	ProviderGameID          string    `gorm:"column:provider_game_id"`
	UserID                  uint64    `gorm:"column:user_id"`
	Operation               string    `gorm:"column:operation"`
	AmountMinor             int64     `gorm:"column:amount_minor"`
	Currency                string    `gorm:"column:currency"`
	CreatedAt               time.Time `gorm:"column:created_at"`
	SuccessfulCallbackCount int64     `gorm:"column:successful_callback_count"`
}

func (s *Service) GenerateProviderReconciliation(command GenerateReconciliationCommand, actor Actor) (ReconciliationRunRow, error) {
	providerCode := strings.ToLower(strings.TrimSpace(command.ProviderCode))
	currency := strings.ToUpper(strings.TrimSpace(command.Currency))
	if providerCode == "" || currency == "" {
		return ReconciliationRunRow{}, newError("INVALID_REQUEST", "providerCode and currency are required")
	}
	businessDate, err := parseBusinessDate(command.BusinessDate)
	if err != nil {
		return ReconciliationRunRow{}, err
	}
	scope, err := reconciliationScope(command, actor.Scope)
	if err != nil {
		return ReconciliationRunRow{}, err
	}
	providerID := s.providerID(providerCode, scope)
	startedAt := time.Now()
	run, err := s.startReconciliationRun(scope, providerID, providerCode, businessDate, currency, actor, startedAt)
	if err != nil {
		return ReconciliationRunRow{}, err
	}
	build, err := s.buildReconciliation(run, scope)
	if err != nil {
		_ = s.markReconciliationFailed(run.ID, err)
		return ReconciliationRunRow{}, err
	}
	if err := s.persistReconciliation(run, build, actor); err != nil {
		_ = s.markReconciliationFailed(run.ID, err)
		return ReconciliationRunRow{}, err
	}
	return s.reconciliationRunByID(run.ID, actor.Scope)
}

func (s *Service) ListReconciliationRuns(query ListQuery) (PageResult, error) {
	rows := []ReconciliationRunRow{}
	v := query.Values
	db := s.db.Table("provider_reconciliation_runs r").Select("r.id, r.run_no, r.site_id, r.provider_id, r.provider_code, r.business_date, r.currency, r.source_mode, r.status, r.provider_transaction_count, r.callback_count, r.wallet_transaction_count, r.bet_record_count, r.discrepancy_count, r.open_discrepancy_count, r.started_at, r.completed_at, r.failure_reason, r.created_by_username, r.created_at").
		Where("r.deleted_at IS NULL")
	db = content.ApplySite(db, "r", query.Scope)
	db = equals(db, "r.provider_code", strings.ToLower(v["providerCode"]))
	db = equals(db, "r.currency", strings.ToUpper(v["currency"]))
	db = equals(db, "r.status", strings.ToUpper(v["status"]))
	db = timeRange(db, "r.business_date", v["businessDateFrom"], v["businessDateTo"])
	db = timeRange(db, "r.created_at", v["createdFrom"], v["createdTo"])
	db = db.Order("r.business_date DESC,r.id DESC")
	return paginate(db, query, &rows)
}

func (s *Service) ReconciliationRunDetail(id uint64, scope content.Scope) (map[string]any, error) {
	run, err := s.reconciliationRunByID(id, scope)
	if err != nil {
		return nil, err
	}
	items := []ReconciliationItemRow{}
	itemDB := s.db.Table("provider_reconciliation_items i").Select(reconciliationItemSelect()).
		Joins("LEFT JOIN game_users u ON u.id=i.user_id AND u.deleted_at IS NULL").
		Where("i.run_id=? AND i.deleted_at IS NULL", id)
	if err := content.ApplySite(itemDB, "i", scope).Order("FIELD(i.status,'OPEN','CONFIRMED','MANUALLY_FIXED','AUTO_FIXED','IGNORED'),i.id DESC").Limit(200).Scan(&items).Error; err != nil {
		return nil, err
	}
	fillReconciliationItemMoney(items)
	return map[string]any{"run": run, "items": items}, nil
}

func (s *Service) ListReconciliationItems(query ListQuery) (PageResult, error) {
	rows := []ReconciliationItemRow{}
	v := query.Values
	db := s.db.Table("provider_reconciliation_items i").Select(reconciliationItemSelect()).
		Joins("LEFT JOIN game_users u ON u.id=i.user_id AND u.deleted_at IS NULL").
		Where("i.deleted_at IS NULL")
	db = content.ApplySite(db, "i", query.Scope)
	db = equals(db, "i.provider_code", strings.ToLower(v["providerCode"]))
	db = equals(db, "i.currency", strings.ToUpper(v["currency"]))
	db = equals(db, "i.discrepancy_type", strings.ToUpper(v["discrepancyType"]))
	db = equals(db, "i.status", strings.ToUpper(v["status"]))
	db = contains(db, "i.provider_transaction_id", v["providerTransactionID"])
	db = contains(db, "i.provider_round_id", v["providerRoundID"])
	db = contains(db, "u.user_no", v["userNo"])
	if runID := parseUint(v["runID"]); runID > 0 {
		db = db.Where("i.run_id=?", runID)
	}
	db = timeRange(db, "i.business_date", v["businessDateFrom"], v["businessDateTo"])
	db = timeRange(db, "i.created_at", v["createdFrom"], v["createdTo"])
	db = db.Order("FIELD(i.status,'OPEN','CONFIRMED','MANUALLY_FIXED','AUTO_FIXED','IGNORED'),i.created_at DESC,i.id DESC")
	result, err := paginate(db, query, &rows)
	fillReconciliationItemMoney(rows)
	result.List = rows
	return result, err
}

func (s *Service) UpdateReconciliationItemStatus(id uint64, command ReconciliationStatusCommand, actor Actor) error {
	status := strings.ToUpper(strings.TrimSpace(command.Status))
	if !allowedReconciliationStatuses[status] {
		return newError("INVALID_RECONCILIATION_STATUS", "invalid reconciliation status")
	}
	reason := strings.TrimSpace(command.Reason)
	if reason == "" {
		return newError("INVALID_REQUEST", "reason is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item model.ProviderReconciliationItems
		db := content.ApplySite(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.ProviderReconciliationItems{}), "", actor.Scope).
			Where("id=? AND deleted_at IS NULL", id)
		if err := db.First(&item).Error; err != nil {
			return notFound(err)
		}
		before := map[string]any{"status": item.Status, "reason": item.ResolutionReason}
		now := time.Now()
		updates := map[string]any{"status": status, "resolution_reason": reason, "updated_at": now}
		if status == reconciliationStatusOpen {
			updates["resolved_by"] = 0
			updates["resolved_by_username"] = ""
			updates["resolved_at"] = nil
		} else {
			updates["resolved_by"] = actor.ID
			updates["resolved_by_username"] = actor.Username
			updates["resolved_at"] = now
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		if err := s.refreshRunOpenCount(tx, item.RunID); err != nil {
			return err
		}
		return audit(tx, actor, adminauth.PermissionBetReconciliationResolve, "update_reconciliation_item_status", "provider_reconciliation_item", item.ID, item.UserID, before, map[string]any{"status": status, "reason": reason}, reason)
	})
}

func (s *Service) startReconciliationRun(scope content.Scope, providerID uint64, providerCode string, businessDate time.Time, currency string, actor Actor, startedAt time.Time) (model.ProviderReconciliationRuns, error) {
	var run model.ProviderReconciliationRuns
	err := s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("site_id = ? AND provider_code = ? AND business_date = ? AND currency = ? AND deleted_at IS NULL", scope.SiteID, providerCode, businessDate, currency)
		err := db.First(&run).Error
		if err == nil {
			updates := map[string]any{
				"run_no": runNo(), "provider_id": providerID, "source_mode": reconciliationSourceInternal, "status": reconciliationRunRunning,
				"failure_reason": "", "started_at": startedAt, "completed_at": nil, "created_by": actor.ID, "created_by_username": actor.Username,
				"provider_transaction_count": 0, "callback_count": 0, "wallet_transaction_count": 0, "bet_record_count": 0,
				"discrepancy_count": 0, "open_discrepancy_count": 0, "updated_at": startedAt,
			}
			if err := tx.Model(&run).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.ProviderReconciliationItems{}).Where("run_id=? AND deleted_at IS NULL", run.ID).Update("deleted_at", startedAt).Error; err != nil {
				return err
			}
			return tx.Where("id=?", run.ID).First(&run).Error
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		run = model.ProviderReconciliationRuns{
			RunNo: runNo(), SiteID: scope.SiteID, ProviderID: providerID, ProviderCode: providerCode,
			BusinessDate: businessDate, Currency: currency, SourceMode: reconciliationSourceInternal, Status: reconciliationRunRunning,
			StartedAt: startedAt, CreatedBy: actor.ID, CreatedByUsername: actor.Username,
		}
		return tx.Create(&run).Error
	})
	return run, err
}

func (s *Service) buildReconciliation(run model.ProviderReconciliationRuns, scope content.Scope) (reconciliationBuild, error) {
	start := run.BusinessDate
	end := start.AddDate(0, 0, 1)
	rows := []reconcileTransactionRow{}
	db := s.baseReconciliationQuery(scope).
		Where("pt.created_at>=? AND pt.created_at<? AND pt.provider_code=? AND pt.currency=? AND pt.deleted_at IS NULL", start, end, run.ProviderCode, run.Currency).
		Order("pt.id ASC")
	if err := db.Scan(&rows).Error; err != nil {
		return reconciliationBuild{}, err
	}
	unmatchedCallbacks, err := s.unmatchedCallbacks(scope, run.ProviderCode, run.Currency, start, end)
	if err != nil {
		return reconciliationBuild{}, err
	}
	stats := reconciliationStats{ProviderTransactions: int64(len(rows))}
	stats.Callbacks = s.countScoped("game_provider_callback_logs", scope, run.ProviderCode, run.Currency, start, end)
	stats.WalletTransactions = s.countScoped("game_wallet_transactions", scope, run.ProviderCode, run.Currency, start, end)
	stats.BetRecords = s.countScoped("game_bet_records", scope, run.ProviderCode, run.Currency, start, end)

	build := reconciliationBuild{Facts: make([]model.GameTransactionFacts, 0, len(rows)), Items: []model.ProviderReconciliationItems{}, Stats: stats}
	for _, row := range rows {
		fact, items := s.reconcileProviderTransaction(run, row)
		build.Facts = append(build.Facts, fact)
		build.Items = append(build.Items, items...)
	}
	for _, row := range unmatchedCallbacks {
		build.Items = append(build.Items, s.callbackOnlyItem(run, row, discrepancyMissingPlatformOrder, "HIGH", row.AmountMinor, 0, map[string]any{"reason": "successful provider callback has no platform provider transaction"}))
		if row.SuccessfulCallbackCount > 1 {
			build.Items = append(build.Items, s.callbackOnlyItem(run, row, discrepancyDuplicateCallback, "LOW", row.AmountMinor, row.AmountMinor, map[string]any{"callbackCount": row.SuccessfulCallbackCount}))
		}
	}
	return build, nil
}

func (s *Service) baseReconciliationQuery(scope content.Scope) *gorm.DB {
	db := s.db.Table("game_provider_transactions pt").Select("pt.id AS provider_transaction_row_id, pt.site_id, COALESCE(gp.id, 0) AS provider_id, pt.provider_code, pt.provider_transaction_id, pt.provider_round_id, pt.provider_game_id, pt.user_id, pt.wallet_transaction_id, pt.operation, pt.status AS provider_status, pt.amount_minor AS provider_amount_minor, pt.currency, pt.related_transaction_id, pt.balance_before_minor, pt.balance_after_minor, pt.created_at AS source_created_at, COALESCE(wt.id, 0) AS wallet_exists, CAST(ROUND(COALESCE(wt.amount, 0) * 100) AS SIGNED) AS wallet_amount_minor, COALESCE(wt.status, 0) AS wallet_status, COALESCE(wt.direction, '') AS wallet_direction, COALESCE(wt.currency, '') AS wallet_currency, COALESCE(wt.order_no, '') AS wallet_order_no, COALESCE(wt.txn_no, '') AS wallet_txn_no, COALESCE(br.id, 0) AS bet_record_id, COALESCE(br.bet_amount_minor, 0) AS bet_amount_minor, COALESCE(br.win_amount_minor, 0) AS win_amount_minor, COALESCE(br.bet_status, '') AS bet_status, COALESCE(br.rollback_status, '') AS rollback_status, COALESCE(br.provider_win_transaction_id, '') AS bet_provider_win_tx, COALESCE(rb.id, 0) AS rollback_record_id, COALESCE(rb.rollback_amount_minor, 0) AS rollback_amount_minor, COALESCE(rb.status, '') AS rollback_record_status, COALESCE(cb.id, 0) AS callback_log_id, COALESCE(cbc.successful_callback_count, 0) AS successful_callback_count").
		Joins("LEFT JOIN game_game_providers gp ON gp.provider_code=pt.provider_code AND gp.deleted_at IS NULL").
		Joins("LEFT JOIN game_wallet_transactions wt ON wt.id=pt.wallet_transaction_id AND wt.deleted_at IS NULL").
		Joins("LEFT JOIN game_bet_records AS br ON br.id = (SELECT br2.id FROM game_bet_records AS br2 WHERE (pt.site_id = 0 OR br2.site_id IN (0, pt.site_id)) AND br2.provider_code = pt.provider_code AND br2.deleted_at IS NULL AND (br2.provider_bet_transaction_id = pt.provider_transaction_id OR br2.provider_win_transaction_id = pt.provider_transaction_id OR (pt.provider_round_id <> '' AND br2.provider_round_id = pt.provider_round_id)) ORDER BY br2.id DESC LIMIT 1)").
		Joins("LEFT JOIN game_rollbacks AS rb ON rb.id = (SELECT rb2.id FROM game_rollbacks AS rb2 WHERE (pt.site_id = 0 OR rb2.site_id IN (0, pt.site_id)) AND rb2.provider_code = pt.provider_code AND rb2.provider_transaction_id = pt.provider_transaction_id AND rb2.deleted_at IS NULL ORDER BY rb2.id DESC LIMIT 1)").
		Joins("LEFT JOIN game_provider_callback_logs AS cb ON cb.id = (SELECT cb2.id FROM game_provider_callback_logs AS cb2 WHERE (pt.site_id = 0 OR cb2.site_id IN (0, pt.site_id)) AND cb2.provider_code = pt.provider_code AND cb2.provider_transaction_id = pt.provider_transaction_id AND cb2.operation = pt.operation AND cb2.processing_result = 'SUCCESS' AND cb2.deleted_at IS NULL ORDER BY cb2.id DESC LIMIT 1)").
		Joins("LEFT JOIN (SELECT site_id, provider_code, provider_transaction_id, operation, COUNT(*) AS successful_callback_count FROM game_provider_callback_logs WHERE processing_result = 'SUCCESS' AND deleted_at IS NULL GROUP BY site_id, provider_code, provider_transaction_id, operation) AS cbc ON cbc.site_id = pt.site_id AND cbc.provider_code = pt.provider_code AND cbc.provider_transaction_id = pt.provider_transaction_id AND cbc.operation = pt.operation")
	return content.ApplySite(db, "pt", scope)
}

func (s *Service) unmatchedCallbacks(scope content.Scope, providerCode string, currency string, start time.Time, end time.Time) ([]unmatchedCallbackRow, error) {
	rows := []unmatchedCallbackRow{}
	db := s.db.Table("game_provider_callback_logs c").Select("c.id AS callback_log_id, c.site_id, COALESCE(gp.id, 0) AS provider_id, c.provider_code, c.provider_transaction_id, c.provider_round_id, c.provider_game_id, c.user_id, c.operation, c.amount_minor, c.currency, c.created_at, COALESCE(cbc.successful_callback_count, 0) AS successful_callback_count").
		Joins("LEFT JOIN game_game_providers gp ON gp.provider_code=c.provider_code AND gp.deleted_at IS NULL").
		Joins("LEFT JOIN game_provider_transactions AS pt ON pt.site_id = c.site_id AND pt.provider_code = c.provider_code AND pt.provider_transaction_id = c.provider_transaction_id AND pt.operation = c.operation AND pt.deleted_at IS NULL").
		Joins("LEFT JOIN (SELECT site_id, provider_code, provider_transaction_id, operation, COUNT(*) AS successful_callback_count FROM game_provider_callback_logs WHERE processing_result = 'SUCCESS' AND deleted_at IS NULL GROUP BY site_id, provider_code, provider_transaction_id, operation) AS cbc ON cbc.site_id = c.site_id AND cbc.provider_code = c.provider_code AND cbc.provider_transaction_id = c.provider_transaction_id AND cbc.operation = c.operation").
		Where("c.created_at>=? AND c.created_at<? AND c.provider_code=? AND c.currency=? AND c.processing_result='SUCCESS' AND c.operation IN ? AND c.deleted_at IS NULL AND pt.id IS NULL",
			start, end, providerCode, currency, []string{"DEBIT", "CREDIT", "REFUND", "ROLLBACK"}).
		Order("c.id ASC")
	err := content.ApplySite(db, "c", scope).Scan(&rows).Error
	return rows, err
}

func (s *Service) reconcileProviderTransaction(run model.ProviderReconciliationRuns, row reconcileTransactionRow) (model.GameTransactionFacts, []model.ProviderReconciliationItems) {
	fact := model.GameTransactionFacts{
		SiteID: row.SiteID, ProviderID: row.ProviderID, ProviderCode: row.ProviderCode,
		ProviderTransactionRowID: row.ProviderTransactionRowID, ProviderTransactionID: row.ProviderTransactionID,
		ProviderRoundID: row.ProviderRoundID, ProviderGameID: row.ProviderGameID, UserID: row.UserID,
		WalletTransactionID: row.WalletTransactionID, BetRecordID: row.BetRecordID, RollbackRecordID: row.RollbackRecordID,
		Operation: row.Operation, ProviderStatus: row.ProviderStatus,
		BusinessDate: run.BusinessDate, Currency: row.Currency, ProviderAmountMinor: row.ProviderAmountMinor,
		WalletAmountMinor: row.WalletAmountMinor, BetAmountMinor: row.BetAmountMinor, WinAmountMinor: row.WinAmountMinor,
		SourceType: "game_provider_transaction", SourceID: row.ProviderTransactionRowID,
		SourceCreatedAt: row.SourceCreatedAt, LastReconciliationRunID: run.ID, LastReconciliationStatus: reconciliationRunCompleted,
		LastReconciliationCheckedAt: time.Now(),
	}
	items := []model.ProviderReconciliationItems{}
	successful := row.ProviderStatus == "SUCCESS" || row.ProviderStatus == "ROLLED_BACK"
	if row.SuccessfulCallbackCount == 0 {
		items = append(items, s.transactionItem(run, row, discrepancyMissingProviderOrder, "HIGH", row.ProviderAmountMinor, 0, map[string]any{"reason": "platform provider transaction has no successful callback record"}))
	}
	if row.SuccessfulCallbackCount > 1 {
		items = append(items, s.transactionItem(run, row, discrepancyDuplicateCallback, "LOW", row.ProviderAmountMinor, row.ProviderAmountMinor, map[string]any{"callbackCount": row.SuccessfulCallbackCount}))
	}
	if !successful {
		return fact, items
	}
	if row.WalletTransactionID == 0 || row.WalletExists == 0 {
		items = append(items, s.transactionItem(run, row, discrepancyMissingWalletEntry, "CRITICAL", row.ProviderAmountMinor, 0, map[string]any{"reason": "successful provider transaction has no wallet transaction"}))
	} else {
		if strings.ToUpper(row.WalletCurrency) != strings.ToUpper(row.Currency) || row.WalletStatus != 1 || !expectedWalletDirection(row.Operation, row.WalletDirection) {
			items = append(items, s.transactionItem(run, row, discrepancyStatusMismatch, "HIGH", row.ProviderAmountMinor, row.WalletAmountMinor, map[string]any{"walletStatus": row.WalletStatus, "walletDirection": row.WalletDirection, "walletCurrency": row.WalletCurrency}))
		}
		if abs64(row.WalletAmountMinor) != abs64(row.ProviderAmountMinor) {
			items = append(items, s.transactionItem(run, row, amountDiscrepancyType(row.Operation), "HIGH", row.ProviderAmountMinor, row.WalletAmountMinor, map[string]any{"source": "wallet_transaction"}))
		}
	}
	switch strings.ToUpper(row.Operation) {
	case "DEBIT":
		if row.BetRecordID == 0 {
			items = append(items, s.transactionItem(run, row, discrepancyMissingPlatformOrder, "HIGH", row.ProviderAmountMinor, 0, map[string]any{"reason": "debit has no bet record"}))
		} else if row.BetAmountMinor != row.ProviderAmountMinor {
			items = append(items, s.transactionItem(run, row, discrepancyBetAmountMismatch, "HIGH", row.ProviderAmountMinor, row.BetAmountMinor, map[string]any{"source": "game_bet_records.bet_amount_minor"}))
		}
	case "CREDIT", "REFUND":
		if row.BetRecordID == 0 {
			items = append(items, s.transactionItem(run, row, discrepancyMissingPlatformOrder, "HIGH", row.ProviderAmountMinor, 0, map[string]any{"reason": "credit/refund has no bet record"}))
		} else if row.BetProviderWinTx == row.ProviderTransactionID && row.WinAmountMinor != row.ProviderAmountMinor {
			items = append(items, s.transactionItem(run, row, discrepancyWinAmountMismatch, "HIGH", row.ProviderAmountMinor, row.WinAmountMinor, map[string]any{"source": "game_bet_records.win_amount_minor"}))
		}
	case "ROLLBACK":
		if row.RollbackRecordID == 0 || row.RollbackRecordStatus != "COMPLETED" || row.RollbackAmountMinor != row.ProviderAmountMinor {
			items = append(items, s.transactionItem(run, row, discrepancyRollbackNotApplied, "CRITICAL", row.ProviderAmountMinor, row.RollbackAmountMinor, map[string]any{"rollbackRecordStatus": row.RollbackRecordStatus, "relatedTransactionID": row.RelatedTransactionID}))
		}
	}
	return fact, items
}

func (s *Service) persistReconciliation(run model.ProviderReconciliationRuns, build reconciliationBuild, actor Actor) error {
	completedAt := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, fact := range build.Facts {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "source_type"}, {Name: "source_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"site_id", "provider_id", "provider_code", "provider_transaction_id", "provider_round_id", "provider_game_id",
					"user_id", "wallet_transaction_id", "bet_record_id", "rollback_record_id", "operation",
					"provider_status", "business_date", "currency", "provider_amount_minor", "wallet_amount_minor", "bet_amount_minor",
					"win_amount_minor", "source_created_at", "last_reconciliation_run_id",
					"last_reconciliation_status", "last_reconciliation_checked_at", "updated_at",
				}),
			}).Create(&fact).Error; err != nil {
				return err
			}
		}
		for index := range build.Items {
			build.Items[index].RunID = run.ID
			build.Items[index].ItemNo = itemNo()
			build.Items[index].Status = reconciliationStatusOpen
			if strings.TrimSpace(build.Items[index].Detail) == "" {
				build.Items[index].Detail = "{}"
			}
			if err := tx.Create(&build.Items[index]).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{
			"status": reconciliationRunCompleted, "completed_at": completedAt, "failure_reason": "",
			"provider_transaction_count": build.Stats.ProviderTransactions, "callback_count": build.Stats.Callbacks,
			"wallet_transaction_count": build.Stats.WalletTransactions, "bet_record_count": build.Stats.BetRecords,
			"discrepancy_count":      int64(len(build.Items)),
			"open_discrepancy_count": int64(len(build.Items)), "updated_at": completedAt,
		}
		if err := tx.Model(&model.ProviderReconciliationRuns{}).Where("id=?", run.ID).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, adminauth.PermissionBetReconciliationGenerate, "generate_provider_reconciliation", "provider_reconciliation_run", run.ID, 0, nil, map[string]any{"runID": run.ID, "discrepancyCount": len(build.Items)}, "provider reconciliation generated")
	})
}

func (s *Service) transactionItem(run model.ProviderReconciliationRuns, row reconcileTransactionRow, discrepancyType string, severity string, expected int64, actual int64, detail map[string]any) model.ProviderReconciliationItems {
	return model.ProviderReconciliationItems{
		SiteID: run.SiteID, ProviderID: row.ProviderID, ProviderCode: row.ProviderCode, BusinessDate: run.BusinessDate,
		Currency: row.Currency, DiscrepancyType: discrepancyType, Severity: severity, UserID: row.UserID,
		ProviderTransactionRowID: row.ProviderTransactionRowID, ProviderTransactionID: row.ProviderTransactionID, ProviderRoundID: row.ProviderRoundID,
		ProviderGameID: row.ProviderGameID, CallbackLogID: row.CallbackLogID, WalletTransactionID: row.WalletTransactionID,
		BetRecordID: row.BetRecordID, RollbackRecordID: row.RollbackRecordID, ExpectedAmountMinor: expected, ActualAmountMinor: actual, DifferenceAmountMinor: actual - expected, Detail: jsonText(detail),
	}
}

func (s *Service) callbackOnlyItem(run model.ProviderReconciliationRuns, row unmatchedCallbackRow, discrepancyType string, severity string, expected int64, actual int64, detail map[string]any) model.ProviderReconciliationItems {
	return model.ProviderReconciliationItems{
		SiteID: run.SiteID, ProviderID: row.ProviderID, ProviderCode: row.ProviderCode, BusinessDate: run.BusinessDate,
		Currency: row.Currency, DiscrepancyType: discrepancyType, Severity: severity, UserID: row.UserID, ProviderTransactionID: row.ProviderTransactionID,
		ProviderRoundID: row.ProviderRoundID, ProviderGameID: row.ProviderGameID, CallbackLogID: row.CallbackLogID,
		ExpectedAmountMinor: expected, ActualAmountMinor: actual, DifferenceAmountMinor: actual - expected, Detail: jsonText(detail),
	}
}

func (s *Service) reconciliationRunByID(id uint64, scope content.Scope) (ReconciliationRunRow, error) {
	var row ReconciliationRunRow
	db := s.db.Table("provider_reconciliation_runs r").Select("r.id, r.run_no, r.site_id, r.provider_id, r.provider_code, r.business_date, r.currency, r.source_mode, r.status, r.provider_transaction_count, r.callback_count, r.wallet_transaction_count, r.bet_record_count, r.discrepancy_count, r.open_discrepancy_count, r.started_at, r.completed_at, r.failure_reason, r.created_by_username, r.created_at").
		Where("r.id=? AND r.deleted_at IS NULL", id)
	if err := content.ApplySite(db, "r", scope).Take(&row).Error; err != nil {
		return ReconciliationRunRow{}, notFound(err)
	}
	return row, nil
}

func reconciliationItemSelect() string {
	return fmt.Sprintf(`i.id, i.run_id, i.item_no, i.site_id, i.provider_id, i.provider_code, i.business_date, i.currency, i.discrepancy_type, i.severity, i.status, i.user_id, u.user_no, %s AS username, i.provider_transaction_row_id, i.provider_transaction_id, i.provider_round_id, i.provider_game_id, i.callback_log_id, i.wallet_transaction_id, i.bet_record_id, i.rollback_record_id, i.expected_amount_minor, i.actual_amount_minor, i.difference_amount_minor, i.detail, i.resolved_by_username, i.resolved_at, i.resolution_reason, i.created_at, i.updated_at`,

		userPhoneDisplay("u"))
}

func fillReconciliationItemMoney(rows []ReconciliationItemRow) {
	for i := range rows {
		rows[i].ExpectedAmount = money(rows[i].ExpectedAmountMinor)
		rows[i].ActualAmount = money(rows[i].ActualAmountMinor)
		rows[i].DifferenceAmount = money(rows[i].DifferenceAmountMinor)
	}
}

func reconciliationScope(command GenerateReconciliationCommand, actorScope content.Scope) (content.Scope, error) {
	if actorScope.Platform {
		if command.SiteID == 0 {
			return content.Scope{}, newError("SITE_REQUIRED", "siteID is required for platform reconciliation")
		}
		return content.Scope{SiteID: command.SiteID}, nil
	}
	if actorScope.SiteID == 0 {
		return content.Scope{}, newError("INVALID_SCOPE", "H5 content scope is required")
	}
	siteID := actorScope.SiteID
	if siteID == 0 {
		siteID = command.SiteID
	}
	return content.Scope{SiteID: siteID}, nil
}

func parseBusinessDate(value string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.UTC)
	if err != nil {
		return time.Time{}, newError("INVALID_BUSINESS_DATE", "businessDate must be YYYY-MM-DD")
	}
	return parsed, nil
}

func (s *Service) providerID(providerCode string, scope content.Scope) uint64 {
	var id uint64
	db := s.db.Table("game_game_providers").Select("id").Where("provider_code=? AND deleted_at IS NULL", providerCode)
	_ = content.ApplySite(db, "", scope).Order("id DESC").Scan(&id).Error
	return id
}

func (s *Service) countScoped(table string, scope content.Scope, providerCode string, currency string, start time.Time, end time.Time) int64 {
	var count int64
	db := s.db.Table(table).Where("created_at>=? AND created_at<? AND provider_code=? AND currency=? AND deleted_at IS NULL", start, end, providerCode, currency)
	_ = content.ApplySite(db, "", scope).Count(&count).Error
	return count
}

func (s *Service) markReconciliationFailed(id uint64, err error) error {
	return s.db.Model(&model.ProviderReconciliationRuns{}).Where("id=?", id).Updates(map[string]any{
		"status": reconciliationRunFailed, "failure_reason": truncateError(err), "completed_at": time.Now(), "updated_at": time.Now(),
	}).Error
}

func (s *Service) refreshRunOpenCount(tx *gorm.DB, runID uint64) error {
	var total, open int64
	if err := tx.Model(&model.ProviderReconciliationItems{}).Where("run_id=? AND deleted_at IS NULL", runID).Count(&total).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.ProviderReconciliationItems{}).Where("run_id=? AND status=? AND deleted_at IS NULL", runID, reconciliationStatusOpen).Count(&open).Error; err != nil {
		return err
	}
	return tx.Model(&model.ProviderReconciliationRuns{}).Where("id=?", runID).Updates(map[string]any{"discrepancy_count": total, "open_discrepancy_count": open, "updated_at": time.Now()}).Error
}

func expectedWalletDirection(operation string, direction string) bool {
	operation = strings.ToUpper(strings.TrimSpace(operation))
	direction = strings.ToLower(strings.TrimSpace(direction))
	switch operation {
	case "DEBIT":
		return direction == "debit"
	case "CREDIT", "REFUND":
		return direction == "credit"
	case "ROLLBACK":
		return direction == "credit" || direction == "debit"
	default:
		return true
	}
}

func amountDiscrepancyType(operation string) string {
	switch strings.ToUpper(strings.TrimSpace(operation)) {
	case "CREDIT", "REFUND":
		return discrepancyWinAmountMismatch
	default:
		return discrepancyBetAmountMismatch
	}
}

func abs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func runNo() string  { return randomID("RECON") }
func itemNo() string { return randomID("RCI") }

func jsonText(value any) string {
	if value == nil {
		return "{}"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	value := err.Error()
	if len(value) > 500 {
		return value[:500]
	}
	return value
}
