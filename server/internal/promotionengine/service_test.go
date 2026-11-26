package promotionengine

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/model"
)

func TestDepositPromotionIsIdempotentAndH5Scoped(t *testing.T) {
	db := promotionTestDB(t)
	seedPromotionUser(t, db, 3, 101, "MX", "USD")
	promotion := seedDepositPromotion(t, db, 3, "WELCOME", "", 100)
	require.NoError(t, db.Create(&model.GamePromotionParticipationRule{
		SiteID: 3, PromotionID: promotion.ID, RuleType: "FIRST_DEPOSIT", ComparisonOperator: "EQ", Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&model.GamePromotionRewardRule{
		SiteID: 3, PromotionID: promotion.ID, RewardType: "BONUS", CalculationType: "FIXED",
		RewardAmount: money("50"), PercentageRate: money("0"), Currency: "USD", WageringMultiplier: money("2"),
		MinimumAmount: money("0"), MaximumReward: money("0"), Enabled: true,
	}).Error)
	service := New(db)
	trigger := DepositTrigger{
		SiteID: 3, UserID: 101, DepositOrderID: 501, OrderNo: "D501", EventID: "event-501",
		Amount: decimal.NewFromInt(100), Currency: "USD", Country: "MX", DepositSequence: 1,
	}
	require.NoError(t, service.ProcessDepositSucceeded(context.Background(), trigger))
	require.NoError(t, service.ProcessDepositSucceeded(context.Background(), trigger))

	assertPromotionCounts(t, db, 1, 1, 1)
	var wallet model.GameUserWallets
	require.NoError(t, db.Where("user_id=?", 101).First(&wallet).Error)
	require.Equal(t, "50.00", wallet.BonusBalance.StringFixed(2))
	var bonus model.GameBonusRecords
	require.NoError(t, db.First(&bonus).Error)
	require.Equal(t, "100.00", bonus.RequiredValidBet.StringFixed(2))

	trigger.EventID = "event-502"
	trigger.OrderNo = "D502"
	trigger.DepositSequence = 2
	require.NoError(t, service.ProcessDepositSucceeded(context.Background(), trigger))
	assertPromotionCounts(t, db, 1, 1, 1)

	trigger.EventID = "event-cross-h5"
	trigger.SiteID = 8
	require.NoError(t, service.ProcessDepositSucceeded(context.Background(), trigger))
	assertPromotionCounts(t, db, 1, 1, 1)
}

func TestPercentageRewardUsesCap(t *testing.T) {
	db := promotionTestDB(t)
	seedPromotionUser(t, db, 3, 101, "MX", "USD")
	promotion := seedDepositPromotion(t, db, 3, "RELOAD", "RELOAD", 100)
	require.NoError(t, db.Create(&model.GamePromotionRewardRule{
		SiteID: 3, PromotionID: promotion.ID, RewardType: "CASH", CalculationType: "PERCENTAGE",
		RewardAmount: money("0"), PercentageRate: money("50"), Currency: "USD", WageringMultiplier: money("0"),
		MinimumAmount: money("20"), MaximumReward: money("30"), Enabled: true,
	}).Error)
	require.NoError(t, New(db).ProcessDepositSucceeded(context.Background(), DepositTrigger{
		SiteID: 3, UserID: 101, EventID: "event-cap", OrderNo: "D-CAP", Amount: decimal.NewFromInt(100),
		Currency: "USD", Country: "MX", DepositSequence: 2,
	}))
	var wallet model.GameUserWallets
	require.NoError(t, db.Where("user_id=?", 101).First(&wallet).Error)
	require.Equal(t, "30.00", wallet.CashBalance.StringFixed(2))
}

func TestRewardFailureRollsBackWalletAndClaim(t *testing.T) {
	db := promotionTestDB(t)
	seedPromotionUser(t, db, 3, 101, "MX", "USD")
	promotion := seedDepositPromotion(t, db, 3, "ROLLBACK", "", 100)
	require.NoError(t, db.Create(&model.GamePromotionRewardRule{
		SiteID: 3, PromotionID: promotion.ID, RewardType: "CASH", CalculationType: "FIXED",
		RewardAmount: money("10"), PercentageRate: money("0"), Currency: "USD", WageringMultiplier: money("0"),
		MinimumAmount: money("0"), MaximumReward: money("0"), Enabled: true,
	}).Error)
	require.NoError(t, db.Migrator().DropTable(&model.GamePromotionClaimReward{}))
	err := New(db).ProcessDepositSucceeded(context.Background(), DepositTrigger{
		SiteID: 3, UserID: 101, EventID: "event-rollback", OrderNo: "D-RB", Amount: decimal.NewFromInt(100),
		Currency: "USD", Country: "MX", DepositSequence: 2,
	})
	require.Error(t, err)
	var wallet model.GameUserWallets
	require.NoError(t, db.Where("user_id=?", 101).First(&wallet).Error)
	require.Equal(t, "0.00", wallet.CashBalance.StringFixed(2))
	var claims int64
	require.NoError(t, db.Model(&model.GamePromotionClaim{}).Count(&claims).Error)
	require.Zero(t, claims)
}

func promotionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&promotionRecord{}, &model.GamePromotionParticipationRule{}, &model.GamePromotionRewardRule{},
		&model.GamePromotionClaim{}, &model.GamePromotionClaimReward{},
	))
	statements := []string{
		`CREATE TABLE game_users (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, user_no TEXT, username TEXT, phone_country_code TEXT, phone TEXT, vip_level INTEGER DEFAULT 0, country_code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_user_wallets (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER NOT NULL, user_id INTEGER NOT NULL, currency TEXT NOT NULL, cash_balance REAL NOT NULL DEFAULT 0, bonus_balance REAL NOT NULL DEFAULT 0, withdrawable_balance REAL NOT NULL DEFAULT 0, frozen_balance REAL NOT NULL DEFAULT 0, status INTEGER NOT NULL DEFAULT 1, last_transaction_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_wallet_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, txn_no TEXT, external_txn_no TEXT, user_id INTEGER, wallet_id INTEGER, wallet_type TEXT, direction TEXT, category TEXT, source_type TEXT, source_id INTEGER, provider_code TEXT, idempotency_key TEXT UNIQUE, order_no TEXT, amount REAL, currency TEXT, status INTEGER, balance_before REAL, balance_after REAL, frozen_balance_before REAL, frozen_balance_after REAL, remark TEXT, operator_id INTEGER, operator_username TEXT, completed_at DATETIME, raw_metadata TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_bonus_records (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, order_no TEXT, bonus_id TEXT, user_id INTEGER, source_type TEXT, source_id INTEGER, promotion_id INTEGER, bonus_type TEXT, amount REAL, currency TEXT, requirement_multiplier REAL, required_valid_bet REAL, completed_valid_bet REAL, current_wagering REAL, remaining_wagering REAL, withdrawable_amount REAL, status INTEGER, status_code TEXT, issued_at DATETIME, unlocked_at DATETIME, expires_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

func seedPromotionUser(t *testing.T, db *gorm.DB, siteID uint64, userID uint64, country string, currency string) {
	t.Helper()
	require.NoError(t, db.Table("game_users").Create(map[string]any{
		"id": userID, "site_id": siteID, "user_no": "U101", "username": "user",
		"phone_country_code": "+52", "phone": "10001", "country_code": country, "vip_level": 0,
	}).Error)
	require.NoError(t, db.Table("game_user_wallets").Create(map[string]any{
		"site_id": siteID, "user_id": userID, "currency": currency,
		"cash_balance": 0, "bonus_balance": 0, "withdrawable_balance": 0, "frozen_balance": 0, "status": 1,
	}).Error)
}

func seedDepositPromotion(t *testing.T, db *gorm.DB, siteID uint64, code string, stackingGroup string, budget int64) promotionRecord {
	t.Helper()
	promotion := promotionRecord{
		SiteID: siteID, ActivityCode: code, TriggerType: TriggerDepositSucceeded, ClaimMode: "AUTO",
		StackingGroup: stackingGroup, Countries: "MX", Currencies: "USD", StatusCode: "ACTIVE", TotalBudget: moneyInt(budget),
		PerUserLimit: 1, RewardAmount: money("0"),
	}
	require.NoError(t, db.Create(&promotion).Error)
	return promotion
}

func assertPromotionCounts(t *testing.T, db *gorm.DB, claims int64, rewards int64, walletTransactions int64) {
	t.Helper()
	var value int64
	require.NoError(t, db.Model(&model.GamePromotionClaim{}).Count(&value).Error)
	require.Equal(t, claims, value)
	require.NoError(t, db.Model(&model.GamePromotionClaimReward{}).Count(&value).Error)
	require.Equal(t, rewards, value)
	require.NoError(t, db.Model(&model.GameWalletTransactions{}).Count(&value).Error)
	require.Equal(t, walletTransactions, value)
}

func money(value string) *decimal.Decimal {
	result := decimal.RequireFromString(value)
	return &result
}

func moneyInt(value int64) *decimal.Decimal {
	result := decimal.NewFromInt(value)
	return &result
}
