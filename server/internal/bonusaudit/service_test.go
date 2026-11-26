package bonusaudit

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/tenancy"
)

func TestRegisterRewardProgressUsesSettledValidBetsAndScope(t *testing.T) {
	db := bonusAuditTestDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	claimedAt := now.Add(time.Minute)
	amount := decimal.RequireFromString("10")
	claim := model.GameRegisterRewardClaims{
		TenantID: 7, SiteID: 3, UserID: 101, Amount: &amount, Currency: "USD", WalletType: "CASH", Status: 2,
		ClaimedAt: &claimedAt,
	}
	require.NoError(t, db.Create(&claim).Error)
	require.NoError(t, db.Table("game_system_configs").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "config_key": registerRewardAuditMultiplierKey, "config_value": "3", "status": 1,
	}).Error)

	insertBet(t, db, 7, 3, 101, "USD", "SETTLED", "NONE", "12", claimedAt.Add(time.Minute))
	insertBet(t, db, 7, 3, 101, "USD", "SETTLED", "NONE", "20", claimedAt.Add(2*time.Minute))
	insertBet(t, db, 7, 3, 101, "USD", "BET_ACCEPTED", "NONE", "100", claimedAt.Add(3*time.Minute))
	insertBet(t, db, 7, 3, 101, "USD", "SETTLED", "ROLLED_BACK", "50", claimedAt.Add(4*time.Minute))
	insertBet(t, db, 8, 3, 101, "USD", "SETTLED", "NONE", "100", claimedAt.Add(5*time.Minute))

	progress, err := RegisterRewardProgress(db, tenancy.Scope{TenantID: 7, SiteID: 3}, claim)
	require.NoError(t, err)
	require.Equal(t, "3.00", progress.Multiplier.StringFixed(2))
	require.Equal(t, "30.00", progress.Required.StringFixed(2))
	require.Equal(t, "30.00", progress.Completed.StringFixed(2))
	require.Equal(t, "0.00", progress.Remaining.StringFixed(2))
}

func TestRemainingForUserBlocksOnlyUnfinishedScopedReward(t *testing.T) {
	db := bonusAuditTestDB(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	amount := decimal.RequireFromString("100")
	claim := model.GameRegisterRewardClaims{
		TenantID: 7, SiteID: 3, UserID: 101, Amount: &amount, Currency: "USD", WalletType: "CASH", Status: 2,
		ClaimedAt: &now,
	}
	require.NoError(t, db.Create(&claim).Error)
	require.NoError(t, db.Table("game_system_configs").Create(map[string]any{
		"tenant_id": 7, "site_id": 3, "config_key": registerRewardAuditMultiplierKey, "config_value": "2", "status": 1,
	}).Error)
	insertBet(t, db, 7, 3, 101, "USD", "SETTLED", "NONE", "50", now.Add(time.Minute))
	insertBet(t, db, 8, 3, 101, "USD", "SETTLED", "NONE", "500", now.Add(time.Minute))

	remaining, err := RemainingForUser(db, tenancy.Scope{TenantID: 7, SiteID: 3}, 101, "USD")
	require.NoError(t, err)
	require.Equal(t, "150.00", remaining.StringFixed(2))
}

func bonusAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE game_register_reward_claims (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME NULL,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		user_id INTEGER NOT NULL,
		amount NUMERIC NOT NULL DEFAULT 0,
		currency TEXT NOT NULL DEFAULT 'PKR',
		wallet_type TEXT NOT NULL DEFAULT 'CASH',
		status INTEGER NOT NULL DEFAULT 1,
		cash_before NUMERIC NOT NULL DEFAULT 0,
		cash_after NUMERIC NOT NULL DEFAULT 0,
		wallet_transaction_id INTEGER NOT NULL DEFAULT 0,
		idempotency_key TEXT NOT NULL,
		claimed_at DATETIME NULL,
		raw_metadata TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_system_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL,
		site_id INTEGER NOT NULL,
		config_key TEXT NOT NULL,
		config_value TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		deleted_at DATETIME NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_bet_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL,
		site_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		currency TEXT NOT NULL,
		valid_bet_amount NUMERIC NOT NULL DEFAULT 0,
		bet_status TEXT NOT NULL,
		rollback_status TEXT NOT NULL DEFAULT '',
		bet_at DATETIME NULL,
		created_at DATETIME NOT NULL,
		deleted_at DATETIME NULL
	)`).Error)
	return db
}

func insertBet(t *testing.T, db *gorm.DB, tenantID, siteID, userID uint64, currency, betStatus, rollbackStatus, validBet string, betAt time.Time) {
	t.Helper()
	require.NoError(t, db.Table("game_bet_records").Create(map[string]any{
		"tenant_id": tenantID, "site_id": siteID, "user_id": userID, "currency": currency,
		"valid_bet_amount": validBet, "bet_status": betStatus, "rollback_status": rollbackStatus,
		"bet_at": betAt, "created_at": betAt,
	}).Error)
}
