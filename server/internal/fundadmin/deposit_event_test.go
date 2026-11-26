package fundadmin

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/eventbus"
)

func TestConfirmDepositPublishesPromotionEventAtomically(t *testing.T) {
	db := depositEventTestDB(t, true)
	depositID := seedDepositEventRows(t, db)
	service := New(db)
	actor := Actor{ID: 9, Username: "finance", Scope: content.Scope{SiteID: 3}, CorrelationID: "trace-deposit-1"}
	require.NoError(t, service.ConfirmDeposit(depositID, DepositConfirmCommand{Reason: "paid", ReceivedAmount: "100"}, actor))
	require.NoError(t, service.ConfirmDeposit(depositID, DepositConfirmCommand{Reason: "paid", ReceivedAmount: "100"}, actor))

	var outbox struct {
		SiteID          uint64
		EventType       string
		TargetConsumers string
		Payload         string
	}
	require.NoError(t, db.Table("event_outbox").First(&outbox).Error)
	require.Equal(t, uint64(3), outbox.SiteID)
	require.Equal(t, eventbus.EventPlayerDepositSucceeded, outbox.EventType)
	require.Contains(t, outbox.TargetConsumers, eventbus.ConsumerPromotionDeposit)
	require.Contains(t, outbox.Payload, `"depositSequence":1`)
	var count int64
	require.NoError(t, db.Table("event_outbox").Count(&count).Error)
	require.Equal(t, int64(1), count)
	var balance decimal.Decimal
	require.NoError(t, db.Table("game_user_wallets").Select("cash_balance").Where("user_id=101").Scan(&balance).Error)
	require.Equal(t, "100.00", balance.StringFixed(2))
}

func TestConfirmDepositRollsBackWhenOutboxIsUnavailable(t *testing.T) {
	db := depositEventTestDB(t, false)
	depositID := seedDepositEventRows(t, db)
	err := New(db).ConfirmDeposit(depositID, DepositConfirmCommand{Reason: "paid", ReceivedAmount: "100"}, Actor{
		ID: 9, Username: "finance", Scope: content.Scope{SiteID: 3}, CorrelationID: "trace-deposit-fail",
	})
	require.Error(t, err)
	var balance decimal.Decimal
	require.NoError(t, db.Table("game_user_wallets").Select("cash_balance").Where("user_id=101").Scan(&balance).Error)
	require.Equal(t, "0.00", balance.StringFixed(2))
	var status int
	require.NoError(t, db.Table("game_deposit_orders").Select("status").Where("id=?", depositID).Scan(&status).Error)
	require.Equal(t, 1, status)
	var transactions int64
	require.NoError(t, db.Table("game_wallet_transactions").Count(&transactions).Error)
	require.Zero(t, transactions)
}

func depositEventTestDB(t *testing.T, withOutbox bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	statements := []string{
		`CREATE TABLE game_deposit_orders (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, order_no TEXT, channel_order_no TEXT DEFAULT '', user_id INTEGER, amount REAL DEFAULT 0, paid_amount REAL DEFAULT 0, currency TEXT, wallet_transaction_id INTEGER DEFAULT 0, status INTEGER DEFAULT 1, reviewer_id INTEGER DEFAULT 0, reviewer_username TEXT DEFAULT '', review_reason TEXT DEFAULT '', evidence_reference TEXT DEFAULT '', paid_at DATETIME, completed_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_user_wallets (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, user_id INTEGER, currency TEXT, cash_balance REAL DEFAULT 0, bonus_balance REAL DEFAULT 0, withdrawable_balance REAL DEFAULT 0, frozen_balance REAL DEFAULT 0, status INTEGER DEFAULT 1, last_transaction_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_users (id INTEGER PRIMARY KEY, site_id INTEGER, country_code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_wallet_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, txn_no TEXT, external_txn_no TEXT, user_id INTEGER, wallet_id INTEGER, wallet_type TEXT, direction TEXT, category TEXT, source_type TEXT, source_id INTEGER, provider_code TEXT, idempotency_key TEXT UNIQUE, order_no TEXT, amount REAL, currency TEXT, status INTEGER, balance_before REAL, balance_after REAL, frozen_balance_before REAL, frozen_balance_after REAL, remark TEXT, operator_id INTEGER, operator_username TEXT, completed_at DATETIME, raw_metadata TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, operator_id INTEGER, operator_username TEXT, permission TEXT, action TEXT, target_type TEXT, target_id INTEGER, target_user_id INTEGER, target_device_id INTEGER DEFAULT 0, before_state TEXT, after_state TEXT, amount REAL, currency TEXT, reason TEXT, ip_address TEXT, user_agent TEXT, correlation_id TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
	if withOutbox {
		require.NoError(t, eventbus.Migrate(db))
	}
	return db
}

func seedDepositEventRows(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	require.NoError(t, db.Table("game_users").Create(map[string]any{"id": 101, "site_id": 3, "country_code": "MX"}).Error)
	require.NoError(t, db.Table("game_user_wallets").Create(map[string]any{
		"site_id": 3, "user_id": 101, "currency": "USD", "cash_balance": 0, "bonus_balance": 0,
		"withdrawable_balance": 0, "frozen_balance": 0, "status": 1,
	}).Error)
	row := map[string]any{
		"site_id": 3, "order_no": "D1001", "user_id": 101, "amount": 100,
		"paid_amount": 0, "currency": "USD", "status": 1,
	}
	require.NoError(t, db.Table("game_deposit_orders").Create(row).Error)
	var id uint64
	require.NoError(t, db.Table("game_deposit_orders").Select("id").Where("order_no='D1001'").Scan(&id).Error)
	return id
}
