package bonusaudit

import (
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/tenancy"
)

const registerRewardAuditMultiplierKey = "operation.register_reward_audit_multiplier"

// Progress is the wagering progress for one cash reward claim.
type Progress struct {
	Multiplier decimal.Decimal
	Required   decimal.Decimal
	Completed  decimal.Decimal
	Remaining  decimal.Decimal
}

// RegisterRewardProgress calculates progress from settled bet records, which is
// the same source used by the admin betting center.
func RegisterRewardProgress(db *gorm.DB, scope tenancy.Scope, claim model.GameRegisterRewardClaims) (Progress, error) {
	multiplier := registerRewardAuditMultiplier(db, scope)
	amount := decimalValue(claim.Amount)
	required := amount.Mul(multiplier).Round(2)
	if !required.GreaterThan(decimal.Zero) {
		return Progress{Multiplier: multiplier, Required: required}, nil
	}

	startAt := claim.ClaimedAt
	if startAt == nil {
		startAt = &claim.CreatedAt
	}

	var completed decimal.Decimal
	query := tenancy.ApplySite(db.Table("game_bet_records AS b"), "b", scope).
		Where("b.user_id = ? AND b.currency = ? AND b.deleted_at IS NULL", claim.UserID, strings.ToUpper(strings.TrimSpace(claim.Currency))).
		Where("b.bet_status = ? AND COALESCE(b.rollback_status, '') <> ?", "SETTLED", "ROLLED_BACK").
		Where("b.valid_bet_amount > 0")
	if startAt != nil {
		query = query.Where("COALESCE(b.bet_at, b.created_at) >= ?", *startAt)
	}
	if err := query.Select("COALESCE(SUM(b.valid_bet_amount), 0)").Scan(&completed).Error; err != nil {
		return Progress{}, err
	}
	if completed.LessThan(decimal.Zero) {
		completed = decimal.Zero
	}
	if completed.GreaterThan(required) {
		completed = required
	}
	remaining := required.Sub(completed)
	if remaining.LessThan(decimal.Zero) {
		remaining = decimal.Zero
	}
	return Progress{Multiplier: multiplier, Required: required, Completed: completed, Remaining: remaining}, nil
}

// RemainingForUser returns the total unfinished register-reward wagering for a
// player. All queries are scoped by tenant and site before any user condition.
func RemainingForUser(db *gorm.DB, scope tenancy.Scope, userID uint64, currency string) (decimal.Decimal, error) {
	var claims []model.GameRegisterRewardClaims
	query := tenancy.ApplySite(db.Model(&model.GameRegisterRewardClaims{}), "", scope).
		Where("user_id = ? AND status = ? AND wallet_type = ? AND currency = ? AND deleted_at IS NULL", userID, 2, "CASH", strings.ToUpper(strings.TrimSpace(currency)))
	if err := query.Order("claimed_at ASC, id ASC").Find(&claims).Error; err != nil {
		return decimal.Zero, err
	}

	remaining := decimal.Zero
	for _, claim := range claims {
		progress, err := RegisterRewardProgress(db, scope, claim)
		if err != nil {
			return decimal.Zero, err
		}
		remaining = remaining.Add(progress.Remaining)
	}
	return remaining, nil
}

func registerRewardAuditMultiplier(db *gorm.DB, scope tenancy.Scope) decimal.Decimal {
	var value string
	_ = tenancy.ApplySite(db.Table("game_system_configs"), "", scope).
		Select("config_value").
		Where("config_key = ? AND status = 1 AND deleted_at IS NULL", registerRewardAuditMultiplierKey).
		Scan(&value).Error
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.NewFromInt(1)
	}
	multiplier, err := decimal.NewFromString(value)
	if err != nil || multiplier.LessThan(decimal.Zero) {
		return decimal.NewFromInt(1)
	}
	return multiplier
}

func decimalValue(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}
