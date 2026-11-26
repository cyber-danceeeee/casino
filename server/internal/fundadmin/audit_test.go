package fundadmin

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/tenancy"
)

func TestEnsureWithdrawalAuditComplete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	createWithdrawalAuditTables(t, db)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, db.Table("game_register_reward_claims").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "user_id": 101, "amount": "100", "currency": "USD",
		"wallet_type": "CASH", "status": 2, "idempotency_key": "register_reward:101", "claimed_at": now, "created_at": now,
	}).Error)
	require.NoError(t, db.Table("game_system_configs").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "config_key": "operation.register_reward_audit_multiplier", "config_value": "2", "status": 1,
	}).Error)
	require.NoError(t, db.Table("game_bet_records").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "user_id": 101, "currency": "USD", "valid_bet_amount": decimal.RequireFromString("50"),
		"bet_status": "SETTLED", "rollback_status": "NONE", "bet_at": now.Add(time.Minute), "created_at": now.Add(time.Minute),
	}).Error)

	scope := tenancy.Scope{TenantID: 7, SiteID: 3}
	require.ErrorIs(t, ensureWithdrawalAuditComplete(db, scope, 101, "USD"), ErrWithdrawalAudit)

	require.NoError(t, db.Table("game_bet_records").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "user_id": 101, "currency": "USD", "valid_bet_amount": decimal.RequireFromString("150"),
		"bet_status": "SETTLED", "rollback_status": "NONE", "bet_at": now.Add(2 * time.Minute), "created_at": now.Add(2 * time.Minute),
	}).Error)
	require.NoError(t, ensureWithdrawalAuditComplete(db, scope, 101, "USD"))
	// The same calculation is safe to repeat because it aggregates canonical bet records.
	require.NoError(t, ensureWithdrawalAuditComplete(db, scope, 101, "USD"))
}

func createWithdrawalAuditTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE game_register_reward_claims (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME NULL,
		tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, user_id INTEGER NOT NULL, amount NUMERIC NOT NULL,
		currency TEXT NOT NULL, wallet_type TEXT NOT NULL, status INTEGER NOT NULL, cash_before NUMERIC DEFAULT 0,
		cash_after NUMERIC DEFAULT 0, wallet_transaction_id INTEGER DEFAULT 0, idempotency_key TEXT NOT NULL,
		claimed_at DATETIME NULL, raw_metadata TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_system_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL,
		config_key TEXT NOT NULL, config_value TEXT NOT NULL, status INTEGER NOT NULL DEFAULT 1, deleted_at DATETIME NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_bet_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL, currency TEXT NOT NULL, valid_bet_amount NUMERIC NOT NULL DEFAULT 0,
		bet_status TEXT NOT NULL, rollback_status TEXT NOT NULL DEFAULT '', bet_at DATETIME NULL,
		created_at DATETIME NOT NULL, deleted_at DATETIME NULL
	)`).Error)
}
