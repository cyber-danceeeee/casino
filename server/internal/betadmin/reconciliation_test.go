package betadmin

import (
	"testing"
	"time"

	"casino/internal/model"
	"casino/internal/tenancy"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGenerateProviderReconciliationDetectsMissingInternalRecords(t *testing.T) {
	db := newReconciliationTestDB(t)
	service := &Service{db: db}
	businessDate := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	seedProvider(t, db, 7, 3, "soft")
	seedProviderTransaction(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "provider_code": "soft", "provider_transaction_id": "tx-missing",
		"provider_round_id": "round-1", "provider_game_id": "slot-1", "user_id": 100, "wallet_transaction_id": 0,
		"operation": "DEBIT", "amount_minor": 1000, "currency": "PKR", "status": "SUCCESS", "created_at": businessDate.Add(time.Hour),
	})
	seedCallback(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "callback_id": "cb-missing", "provider_code": "soft", "operation": "DEBIT",
		"user_id": 100, "provider_transaction_id": "tx-missing", "provider_round_id": "round-1", "provider_game_id": "slot-1",
		"amount_minor": 1000, "currency": "PKR", "processing_result": "SUCCESS", "created_at": businessDate.Add(time.Hour),
	})

	run, err := service.GenerateProviderReconciliation(GenerateReconciliationCommand{
		TenantID: 7, SiteID: 3, ProviderCode: "soft", BusinessDate: "2026-07-18", Currency: "PKR",
	}, Actor{ID: 1, Username: "platform", Scope: tenancy.Scope{Platform: true}})
	require.NoError(t, err)
	require.Equal(t, int64(1), run.ProviderTransactionCount)
	require.Equal(t, int64(3), run.DiscrepancyCount)

	types := reconciliationItemTypes(t, db, run.ID)
	require.Contains(t, types, discrepancyMissingWalletEntry)
	require.Contains(t, types, discrepancyMissingPlatformOrder)
	require.Contains(t, types, discrepancyMissingLedgerEntry)

	var factCount int64
	require.NoError(t, db.Model(&model.GameTransactionFacts{}).Where("tenant_id=? AND site_id=?", 7, 3).Count(&factCount).Error)
	require.Equal(t, int64(1), factCount)
}

func TestGenerateProviderReconciliationDetectsAmountMismatchWithinScope(t *testing.T) {
	db := newReconciliationTestDB(t)
	service := &Service{db: db}
	businessDate := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	seedProvider(t, db, 7, 3, "soft")
	walletID := seedWalletTransaction(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "txn_no": "wtx-amount", "user_id": 101, "wallet_id": 11, "wallet_type": "cash",
		"direction": "debit", "category": "bet", "source_type": "game_provider_transaction", "provider_code": "soft",
		"order_no": "tx-amount", "amount": "9.00", "currency": "PKR", "status": 1, "created_at": businessDate.Add(time.Hour),
	})
	seedProviderTransaction(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "provider_code": "soft", "provider_transaction_id": "tx-amount",
		"provider_round_id": "round-amount", "provider_game_id": "slot-1", "user_id": 101, "wallet_transaction_id": walletID,
		"operation": "DEBIT", "amount_minor": 1000, "currency": "PKR", "status": "SUCCESS", "created_at": businessDate.Add(time.Hour),
	})
	seedBetRecord(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "platform_bet_id": "bet-amount", "user_id": 101, "wallet_id": 11,
		"provider_code": "soft", "provider_game_id": "slot-1", "provider_round_id": "round-amount", "provider_bet_transaction_id": "tx-amount",
		"bet_amount_minor": 1000, "valid_bet_amount_minor": 1000, "win_amount_minor": 0, "platform_profit_minor": 1000,
		"currency": "PKR", "bet_status": "BET_ACCEPTED", "rollback_status": "NONE", "created_at": businessDate.Add(time.Hour),
	})
	seedCallback(t, db, map[string]any{
		"tenant_id": 7, "site_id": 3, "callback_id": "cb-amount", "provider_code": "soft", "operation": "DEBIT",
		"user_id": 101, "provider_transaction_id": "tx-amount", "provider_round_id": "round-amount", "provider_game_id": "slot-1",
		"amount_minor": 1000, "currency": "PKR", "processing_result": "SUCCESS", "created_at": businessDate.Add(time.Hour),
	})
	seedLedgerTransaction(t, db, 7, 3, "PKR", "tx-amount", "10.00", businessDate.Add(time.Hour))
	seedProviderTransaction(t, db, map[string]any{
		"tenant_id": 8, "site_id": 3, "provider_code": "soft", "provider_transaction_id": "tx-other-tenant",
		"operation": "DEBIT", "amount_minor": 5000, "currency": "PKR", "status": "SUCCESS", "created_at": businessDate.Add(time.Hour),
	})

	run, err := service.GenerateProviderReconciliation(GenerateReconciliationCommand{
		TenantID: 7, SiteID: 3, ProviderCode: "soft", BusinessDate: "2026-07-18", Currency: "PKR",
	}, Actor{ID: 1, Username: "platform", Scope: tenancy.Scope{Platform: true}})
	require.NoError(t, err)
	require.Equal(t, int64(1), run.ProviderTransactionCount)
	require.Equal(t, int64(1), run.DiscrepancyCount)
	require.Equal(t, []string{discrepancyBetAmountMismatch}, reconciliationItemTypes(t, db, run.ID))
}

func TestUpdateReconciliationItemStatusRespectsTenantScope(t *testing.T) {
	db := newReconciliationTestDB(t)
	service := &Service{db: db}
	run7 := seedReconciliationRun(t, db, 7, 3)
	item7 := seedReconciliationItem(t, db, run7.ID, 7, 3)
	run8 := seedReconciliationRun(t, db, 8, 3)
	item8 := seedReconciliationItem(t, db, run8.ID, 8, 3)

	tenantActor := Actor{ID: 22, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 3}}
	err := service.UpdateReconciliationItemStatus(item8.ID, ReconciliationStatusCommand{Status: reconciliationStatusIgnored, Reason: "not mine"}, tenantActor)
	require.Error(t, err)

	require.NoError(t, service.UpdateReconciliationItemStatus(item7.ID, ReconciliationStatusCommand{Status: reconciliationStatusIgnored, Reason: "accepted"}, tenantActor))
	var item model.ProviderReconciliationItems
	require.NoError(t, db.Where("id=?", item7.ID).Take(&item).Error)
	require.Equal(t, reconciliationStatusIgnored, item.Status)
	require.Equal(t, uint64(22), item.ResolvedBy)
	var run model.ProviderReconciliationRuns
	require.NoError(t, db.Where("id=?", run7.ID).Take(&run).Error)
	require.Equal(t, int64(0), run.OpenDiscrepancyCount)
}

func newReconciliationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	for _, statement := range reconciliationTestTables {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

var reconciliationTestTables = []string{
	`CREATE TABLE game_game_providers (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, provider_code TEXT, name TEXT, logo_url TEXT,
		api_base_url TEXT, environment TEXT, api_status TEXT, status INTEGER, sort INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_provider_transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, provider_code TEXT, provider_transaction_id TEXT,
		provider_round_id TEXT, provider_game_id TEXT, user_id INTEGER, wallet_transaction_id INTEGER, operation TEXT, amount_minor INTEGER,
		currency TEXT, status TEXT, related_transaction_id TEXT, request_hash TEXT, raw_request TEXT, raw_response TEXT,
		balance_before_minor INTEGER DEFAULT 0, balance_after_minor INTEGER DEFAULT 0, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_wallet_transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, txn_no TEXT, external_txn_no TEXT DEFAULT '',
		user_id INTEGER, wallet_id INTEGER, wallet_type TEXT, direction TEXT, category TEXT, source_type TEXT, source_id INTEGER DEFAULT 0,
		provider_code TEXT, idempotency_key TEXT, order_no TEXT, amount DECIMAL(18,2), currency TEXT, status INTEGER,
		balance_before DECIMAL(18,2), balance_after DECIMAL(18,2), frozen_balance_before DECIMAL(18,2), frozen_balance_after DECIMAL(18,2),
		remark TEXT, operator_id INTEGER DEFAULT 0, operator_username TEXT DEFAULT '', completed_at DATETIME, raw_metadata TEXT DEFAULT '{}',
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_bet_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, bet_no TEXT, platform_bet_id TEXT,
		user_id INTEGER, wallet_id INTEGER, game_id INTEGER DEFAULT 0, provider_id INTEGER DEFAULT 0, provider_code TEXT,
		provider_game_id TEXT, provider_round_id TEXT, provider_bet_transaction_id TEXT, provider_win_transaction_id TEXT,
		game_type TEXT, bet_amount_minor INTEGER DEFAULT 0, valid_bet_amount_minor INTEGER DEFAULT 0, win_amount_minor INTEGER DEFAULT 0,
		platform_profit_minor INTEGER DEFAULT 0, currency TEXT, bet_status TEXT, rollback_status TEXT, failure_reason TEXT DEFAULT '',
		status INTEGER DEFAULT 1, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_rollbacks (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, rollback_id TEXT, provider_code TEXT,
		provider_transaction_id TEXT, provider_round_id TEXT, original_provider_transaction_id TEXT, original_wallet_transaction_id INTEGER DEFAULT 0,
		rollback_wallet_transaction_id INTEGER DEFAULT 0, bet_record_id INTEGER DEFAULT 0, user_id INTEGER DEFAULT 0, game_id INTEGER DEFAULT 0,
		provider_game_id TEXT, original_operation TEXT, original_amount_minor INTEGER DEFAULT 0, rollback_amount_minor INTEGER DEFAULT 0,
		currency TEXT, status TEXT, failure_reason TEXT DEFAULT '', completed_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_provider_callback_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, callback_id TEXT, provider_code TEXT, operation TEXT,
		user_id INTEGER DEFAULT 0, provider_player_id TEXT DEFAULT '', provider_transaction_id TEXT, provider_round_id TEXT, provider_game_id TEXT,
		amount_minor INTEGER DEFAULT 0, currency TEXT, http_method TEXT, request_path TEXT, request_headers TEXT, query_parameters TEXT, request_body TEXT,
		source_ip TEXT DEFAULT '', signature_result TEXT, timestamp_validation TEXT, processing_result TEXT, http_status INTEGER,
		normalized_command TEXT, wallet_result TEXT, provider_response TEXT, duration_ms INTEGER DEFAULT 0, correlation_id TEXT DEFAULT '',
		failure_reason TEXT DEFAULT '', marked_for_investigation INTEGER DEFAULT 0, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE ledger_transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER, site_id INTEGER, currency TEXT, business_order_no TEXT,
		debit_amount DECIMAL(20,2), created_at DATETIME, deleted_at DATETIME
	)`,
	`CREATE TABLE game_transaction_facts (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
		tenant_id INTEGER, site_id INTEGER, provider_id INTEGER, provider_code TEXT, provider_transaction_row_id INTEGER,
		provider_transaction_id TEXT, provider_round_id TEXT, provider_game_id TEXT, user_id INTEGER, wallet_transaction_id INTEGER,
		bet_record_id INTEGER, rollback_record_id INTEGER, ledger_transaction_id INTEGER, operation TEXT, provider_status TEXT,
		business_date DATE, currency TEXT, provider_amount_minor INTEGER, wallet_amount_minor INTEGER, bet_amount_minor INTEGER,
		win_amount_minor INTEGER, ledger_amount_minor INTEGER, source_type TEXT, source_id INTEGER, source_created_at DATETIME,
		last_reconciliation_run_id INTEGER, last_reconciliation_status TEXT, last_reconciliation_checked_at DATETIME,
		UNIQUE(source_type, source_id)
	)`,
	`CREATE TABLE provider_reconciliation_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, run_no TEXT UNIQUE,
		tenant_id INTEGER, site_id INTEGER, provider_id INTEGER, provider_code TEXT, business_date DATE, currency TEXT, source_mode TEXT,
		status TEXT, provider_transaction_count INTEGER DEFAULT 0, callback_count INTEGER DEFAULT 0, wallet_transaction_count INTEGER DEFAULT 0,
		bet_record_count INTEGER DEFAULT 0, ledger_transaction_count INTEGER DEFAULT 0, discrepancy_count INTEGER DEFAULT 0,
		open_discrepancy_count INTEGER DEFAULT 0, started_at DATETIME, completed_at DATETIME, failure_reason TEXT DEFAULT '',
		created_by INTEGER DEFAULT 0, created_by_username TEXT DEFAULT '', UNIQUE(tenant_id, site_id, provider_code, business_date, currency)
	)`,
	`CREATE TABLE provider_reconciliation_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, run_id INTEGER, item_no TEXT UNIQUE,
		tenant_id INTEGER, site_id INTEGER, provider_id INTEGER, provider_code TEXT, business_date DATE, currency TEXT, discrepancy_type TEXT,
		severity TEXT, status TEXT, user_id INTEGER DEFAULT 0, provider_transaction_row_id INTEGER DEFAULT 0, provider_transaction_id TEXT DEFAULT '',
		provider_round_id TEXT DEFAULT '', provider_game_id TEXT DEFAULT '', callback_log_id INTEGER DEFAULT 0, wallet_transaction_id INTEGER DEFAULT 0,
		bet_record_id INTEGER DEFAULT 0, rollback_record_id INTEGER DEFAULT 0, ledger_transaction_id INTEGER DEFAULT 0,
		expected_amount_minor INTEGER DEFAULT 0, actual_amount_minor INTEGER DEFAULT 0, difference_amount_minor INTEGER DEFAULT 0,
		detail TEXT, resolved_by INTEGER DEFAULT 0, resolved_by_username TEXT DEFAULT '', resolved_at DATETIME, resolution_reason TEXT DEFAULT ''
	)`,
	`CREATE TABLE game_admin_operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
		tenant_id INTEGER, site_id INTEGER, operator_id INTEGER, operator_username TEXT, permission TEXT, action TEXT,
		target_type TEXT, target_id INTEGER, target_user_id INTEGER, target_device_id INTEGER DEFAULT 0,
		before_state TEXT, after_state TEXT, amount DECIMAL(18,2), currency TEXT DEFAULT '', reason TEXT,
		ip_address TEXT DEFAULT '', user_agent TEXT DEFAULT '', correlation_id TEXT DEFAULT ''
	)`,
}

func seedProvider(t *testing.T, db *gorm.DB, tenantID, siteID uint64, providerCode string) {
	t.Helper()
	now := time.Now()
	require.NoError(t, db.Table("game_game_providers").Create(map[string]any{
		"tenant_id": tenantID, "site_id": siteID, "provider_code": providerCode, "name": providerCode,
		"logo_url": "", "api_base_url": "", "environment": "test", "api_status": "healthy", "status": 1,
		"created_at": now, "updated_at": now,
	}).Error)
}

func seedWalletTransaction(t *testing.T, db *gorm.DB, values map[string]any) uint64 {
	t.Helper()
	now := time.Now()
	values["created_at"] = defaultTime(values["created_at"], now)
	values["updated_at"] = now
	values["balance_before"] = "100.00"
	values["balance_after"] = "91.00"
	values["frozen_balance_before"] = "0.00"
	values["frozen_balance_after"] = "0.00"
	values["remark"] = ""
	require.NoError(t, db.Table("game_wallet_transactions").Create(values).Error)
	var id uint64
	require.NoError(t, db.Table("game_wallet_transactions").Select("id").Where("txn_no=?", values["txn_no"]).Scan(&id).Error)
	return id
}

func seedProviderTransaction(t *testing.T, db *gorm.DB, values map[string]any) {
	t.Helper()
	now := time.Now()
	values["created_at"] = defaultTime(values["created_at"], now)
	values["updated_at"] = now
	values["related_transaction_id"] = ""
	values["request_hash"] = values["provider_transaction_id"]
	values["raw_request"] = "{}"
	values["raw_response"] = "{}"
	require.NoError(t, db.Table("game_provider_transactions").Create(values).Error)
}

func seedBetRecord(t *testing.T, db *gorm.DB, values map[string]any) {
	t.Helper()
	now := time.Now()
	values["created_at"] = defaultTime(values["created_at"], now)
	values["updated_at"] = now
	values["bet_no"] = values["platform_bet_id"]
	values["provider_win_transaction_id"] = ""
	values["game_type"] = "slot"
	values["status"] = 1
	require.NoError(t, db.Table("game_bet_records").Create(values).Error)
}

func seedCallback(t *testing.T, db *gorm.DB, values map[string]any) {
	t.Helper()
	now := time.Now()
	values["created_at"] = defaultTime(values["created_at"], now)
	values["updated_at"] = now
	values["http_method"] = "POST"
	values["request_path"] = "/callback"
	values["signature_result"] = "VALID"
	values["timestamp_validation"] = "VALID"
	values["http_status"] = 200
	values["normalized_command"] = "{}"
	values["wallet_result"] = "{}"
	values["provider_response"] = "{}"
	require.NoError(t, db.Table("game_provider_callback_logs").Create(values).Error)
}

func seedLedgerTransaction(t *testing.T, db *gorm.DB, tenantID, siteID uint64, currency, orderNo, amount string, createdAt time.Time) {
	t.Helper()
	require.NoError(t, db.Table("ledger_transactions").Create(map[string]any{
		"tenant_id": tenantID, "site_id": siteID, "currency": currency, "business_order_no": orderNo, "debit_amount": amount, "created_at": createdAt,
	}).Error)
}

func seedReconciliationRun(t *testing.T, db *gorm.DB, tenantID, siteID uint64) model.ProviderReconciliationRuns {
	t.Helper()
	now := time.Now()
	run := model.ProviderReconciliationRuns{
		RunNo: runNo(), TenantID: tenantID, SiteID: siteID, ProviderCode: "soft", BusinessDate: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		Currency: "PKR", SourceMode: reconciliationSourceInternal, Status: reconciliationRunCompleted, DiscrepancyCount: 1, OpenDiscrepancyCount: 1, StartedAt: now,
	}
	require.NoError(t, db.Create(&run).Error)
	return run
}

func seedReconciliationItem(t *testing.T, db *gorm.DB, runID, tenantID, siteID uint64) model.ProviderReconciliationItems {
	t.Helper()
	item := model.ProviderReconciliationItems{
		RunID: runID, ItemNo: itemNo(), TenantID: tenantID, SiteID: siteID, ProviderCode: "soft",
		BusinessDate: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC), Currency: "PKR",
		DiscrepancyType: discrepancyMissingWalletEntry, Severity: "HIGH", Status: reconciliationStatusOpen, Detail: "{}",
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}

func reconciliationItemTypes(t *testing.T, db *gorm.DB, runID uint64) []string {
	t.Helper()
	var types []string
	require.NoError(t, db.Table("provider_reconciliation_items").Where("run_id=? AND deleted_at IS NULL", runID).Order("discrepancy_type").Pluck("discrepancy_type", &types).Error)
	return types
}

func defaultTime(value any, fallback time.Time) any {
	if value == nil {
		return fallback
	}
	return value
}
