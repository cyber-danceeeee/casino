package operationadmin

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/model"
)

func applyVIPRewards(row *vipRecord, before vipRecord, in VIPLevelInput) error {
	zero := decimal.Zero
	row.LevelUpReward, row.WeeklyReward = before.LevelUpReward, before.WeeklyReward
	row.DailyWithdrawCount = before.DailyWithdrawCount
	if row.LevelUpReward == nil {
		row.LevelUpReward = &zero
	}
	if row.WeeklyReward == nil {
		row.WeeklyReward = &zero
	}
	for _, field := range []struct {
		name   string
		input  *string
		target **decimal.Decimal
	}{{"levelUpReward", in.LevelUpReward, &row.LevelUpReward}, {"weeklyReward", in.WeeklyReward, &row.WeeklyReward}} {
		if field.input == nil {
			continue
		}
		amount, err := parseDecimal(*field.input, field.name)
		if err != nil {
			return err
		}
		*field.target = &amount
	}
	if in.DailyWithdrawCount != nil {
		row.DailyWithdrawCount = *in.DailyWithdrawCount
	}
	return nil
}

func loadVIPRetention(db *gorm.DB, level *vipRecord) error {
	level.RetainPeriodDays, level.RetainDeposit, level.RetainValidBet = 0, "0.00", "0.00"
	var row model.GameVipRequirements
	err := db.Where("site_id=? AND vip_level_id=? AND requirement_type='retain'", level.SiteID, level.ID).Order("id DESC").Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	level.RetainPeriodDays = row.PeriodDays
	if row.DepositAmount != nil {
		level.RetainDeposit = row.DepositAmount.StringFixed(2)
	}
	if row.ValidBetAmount != nil {
		level.RetainValidBet = row.ValidBetAmount.StringFixed(2)
	}
	return nil
}

// Called inside SaveVIPLevel's transaction, so the admin and H5 catalogs cannot
// disagree after a successful save. Omitted optional fields preserve old clients.
func saveVIPRequirements(tx *gorm.DB, level vipRecord, in VIPLevelInput) error {
	if err := upsertVIPRequirement(tx, level, "level_up", 0, *level.RequiredDeposit, *level.RequiredValidBet); err != nil {
		return err
	}
	if in.RetainPeriodDays == nil && in.RetainDeposit == nil && in.RetainValidBet == nil {
		return nil
	}
	if in.RetainPeriodDays == nil || in.RetainDeposit == nil || in.RetainValidBet == nil {
		return newError("INVALID_VIP_RETENTION", "retention period, deposit and valid bets must be supplied together")
	}
	deposit, err := parseDecimal(*in.RetainDeposit, "retainDeposit")
	if err != nil {
		return err
	}
	bet, err := parseDecimal(*in.RetainValidBet, "retainValidBet")
	if err != nil {
		return err
	}
	if *in.RetainPeriodDays == 0 {
		if !deposit.IsZero() || !bet.IsZero() {
			return newError("INVALID_VIP_RETENTION", "a retention period is required for retention thresholds")
		}
		return tx.Where("site_id=? AND vip_level_id=? AND requirement_type='retain'", level.SiteID, level.ID).Delete(&model.GameVipRequirements{}).Error
	}
	return upsertVIPRequirement(tx, level, "retain", *in.RetainPeriodDays, deposit, bet)
}

func upsertVIPRequirement(tx *gorm.DB, level vipRecord, kind string, days uint, deposit, bet decimal.Decimal) error {
	var row model.GameVipRequirements
	err := tx.Where("site_id=? AND vip_level_id=? AND requirement_type=?", level.SiteID, level.ID, kind).Order("id DESC").Take(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = model.GameVipRequirements{SiteID: level.SiteID, VipLevelID: level.ID, RequirementType: kind, TargetLevelNo: level.LevelNo, PeriodDays: days, DepositAmount: &deposit, ValidBetAmount: &bet}
		return tx.Create(&row).Error
	}
	return tx.Model(&row).Updates(map[string]any{"target_level_no": level.LevelNo, "period_days": days, "deposit_amount": deposit, "valid_bet_amount": bet}).Error
}
