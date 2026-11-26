package operationadmin

import (
	"testing"

	"casino/internal/content"
	"casino/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestVIPAdminSaveKeepsRewardsAndH5RequirementsTogether(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	require.NoError(t, db.AutoMigrate(&vipRecord{}, &model.GameVIPBenefit{}, &model.GameAdminOperationLogs{}))
	require.NoError(t, db.Exec(`CREATE TABLE game_vip_requirements (
		id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER, vip_level_id INTEGER,
		requirement_type TEXT, target_level_no INTEGER, period_days INTEGER,
		deposit_amount DECIMAL(18,2), valid_bet_amount DECIMAL(18,2),
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`).Error)
	s := New(db)
	actor := Actor{ID: 1, Username: "test-admin", Scope: content.Scope{SiteID: 1}}
	upgrade, weekly, retainDeposit, retainBet := "2.25", "3.50", "50.00", "500.00"
	days, withdrawals := uint(7), uint(3)
	in := VIPLevelInput{Level: 1, Code: "vip1", RequiredDeposit: "100.00", RequiredValidBet: "1000.00", WithdrawalLimit: "1000", CashbackRate: "0", BirthdayReward: "0", MonthlyReward: "4.75", Enabled: true,
		LevelUpReward: &upgrade, WeeklyReward: &weekly, DailyWithdrawCount: &withdrawals, RetainPeriodDays: &days, RetainDeposit: &retainDeposit, RetainValidBet: &retainBet}
	id, err := s.SaveVIPLevel(0, in, actor)
	require.NoError(t, err)
	var saved vipRecord
	require.NoError(t, db.First(&saved, id).Error)
	require.Equal(t, "2.25", saved.LevelUpReward.StringFixed(2))
	require.Equal(t, "3.50", saved.WeeklyReward.StringFixed(2))
	require.Equal(t, uint(3), saved.DailyWithdrawCount)
	require.NoError(t, loadVIPRetention(db, &saved))
	require.Equal(t, uint(7), saved.RetainPeriodDays)
	require.Equal(t, "50.00", saved.RetainDeposit)
	var rules []model.GameVipRequirements
	require.NoError(t, db.Order("requirement_type").Find(&rules).Error)
	require.Len(t, rules, 2)
	require.Equal(t, "100.00", rules[0].DepositAmount.StringFixed(2))
	require.Equal(t, "1000.00", rules[0].ValidBetAmount.StringFixed(2))

	// Older admin clients omit the new fields; editing a threshold preserves them.
	in.LevelUpReward, in.WeeklyReward, in.DailyWithdrawCount = nil, nil, nil
	in.RetainPeriodDays, in.RetainDeposit, in.RetainValidBet = nil, nil, nil
	in.RequiredDeposit = "200.00"
	in.Level = 3
	_, err = s.SaveVIPLevel(id, in, actor)
	require.NoError(t, err)
	require.NoError(t, db.First(&saved, id).Error)
	require.Equal(t, "2.25", saved.LevelUpReward.StringFixed(2))
	require.NoError(t, loadVIPRetention(db, &saved))
	require.Equal(t, uint(7), saved.RetainPeriodDays)
	rules = nil
	require.NoError(t, db.Order("requirement_type").Find(&rules).Error)
	require.Len(t, rules, 2)
	require.Equal(t, "200.00", rules[0].DepositAmount.StringFixed(2))
	require.Equal(t, uint(3), rules[0].TargetLevelNo)
	require.Equal(t, uint(3), rules[1].TargetLevelNo)

	// Validation rolls back both the level and its display requirements.
	negative := "-1"
	in.RequiredDeposit, in.RetainDeposit, in.RetainValidBet, in.RetainPeriodDays = "999.00", &negative, &retainBet, &days
	_, err = s.SaveVIPLevel(id, in, actor)
	require.Error(t, err)
	require.NoError(t, db.First(&saved, id).Error)
	require.Equal(t, "200.00", saved.RequiredDeposit.StringFixed(2))

	zero, noDays := "0", uint(0)
	in.RequiredDeposit, in.RetainDeposit, in.RetainValidBet, in.RetainPeriodDays = "200.00", &zero, &zero, &noDays
	_, err = s.SaveVIPLevel(id, in, actor)
	require.NoError(t, err)
	rules = nil
	require.NoError(t, db.Find(&rules).Error)
	require.Len(t, rules, 1)
	require.Equal(t, "level_up", rules[0].RequirementType)
	require.NoError(t, loadVIPRetention(db, &saved))
	require.Zero(t, saved.RetainPeriodDays)
}
