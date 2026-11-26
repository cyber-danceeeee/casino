package operationadmin

import (
	"strings"

	"github.com/shopspring/decimal"
)

var promotionTriggers = map[string]bool{
	"REGISTRATION_COMPLETED": true,
	"DEPOSIT_SUCCEEDED":      true,
	"GAME_BET_SETTLED":       true,
	"MANUAL_CLAIM":           true,
	"SCHEDULED":              true,
}

var promotionRuleTypes = map[string]bool{
	"FIRST_DEPOSIT":    true,
	"DEPOSIT_SEQUENCE": true,
	"DEPOSIT_AMOUNT":   true,
	"COUNTRY":          true,
	"CURRENCY":         true,
	"USER_VIP_LEVEL":   true,
}

var promotionComparisons = map[string]bool{
	"EQ": true, "NE": true, "GT": true, "GTE": true, "LT": true, "LTE": true, "IN": true,
}

func normalizePromotionInput(in *PromotionInput) {
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Type = strings.ToUpper(strings.TrimSpace(in.Type))
	in.TriggerType = strings.ToUpper(strings.TrimSpace(in.TriggerType))
	in.ClaimMode = strings.ToUpper(strings.TrimSpace(in.ClaimMode))
	in.StackingGroup = strings.ToUpper(strings.TrimSpace(in.StackingGroup))
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	in.Countries = normalizeCSV(in.Countries)
	in.Currencies = normalizeCSV(in.Currencies)
	in.Languages = strings.TrimSpace(in.Languages)
	for index := range in.ParticipationRules {
		in.ParticipationRules[index].RuleType = strings.ToUpper(strings.TrimSpace(in.ParticipationRules[index].RuleType))
		in.ParticipationRules[index].ComparisonOperator = strings.ToUpper(strings.TrimSpace(in.ParticipationRules[index].ComparisonOperator))
		in.ParticipationRules[index].TextValue = strings.TrimSpace(in.ParticipationRules[index].TextValue)
	}
	for index := range in.RewardRules {
		in.RewardRules[index].RewardType = strings.ToUpper(strings.TrimSpace(in.RewardRules[index].RewardType))
		in.RewardRules[index].CalculationType = strings.ToUpper(strings.TrimSpace(in.RewardRules[index].CalculationType))
		in.RewardRules[index].Currency = strings.ToUpper(strings.TrimSpace(in.RewardRules[index].Currency))
	}
}

func normalizeCSV(value string) string {
	result := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		if item = strings.ToUpper(strings.TrimSpace(item)); item != "" {
			result = append(result, item)
		}
	}
	return strings.Join(result, ",")
}

func validatePromotionInput(in PromotionInput) error {
	if !promotionTriggers[in.TriggerType] {
		return newError("PROMOTION_TRIGGER_INVALID", "unsupported promotion trigger")
	}
	if in.ClaimMode != "AUTO" && in.ClaimMode != "MANUAL" {
		return newError("PROMOTION_CLAIM_MODE_INVALID", "unsupported promotion claim mode")
	}
	if in.Status == "ACTIVE" && !promotionExecutable(in.TriggerType, in.ClaimMode) {
		return newError("PROMOTION_TRIGGER_NOT_CONNECTED", "this promotion trigger is not connected to an executable business event")
	}
	if len(in.RewardRules) == 0 {
		return newError("PROMOTION_REWARD_REQUIRED", "at least one reward rule is required")
	}
	hasEnabledReward := false
	for _, rule := range in.ParticipationRules {
		if !promotionRuleTypes[rule.RuleType] || !promotionComparisons[rule.ComparisonOperator] {
			return newError("PROMOTION_CONDITION_INVALID", "unsupported promotion condition")
		}
	}
	for _, rule := range in.RewardRules {
		hasEnabledReward = hasEnabledReward || rule.Enabled
		if rule.RewardType != "CASH" && rule.RewardType != "BONUS" {
			return newError("PROMOTION_REWARD_TYPE_INVALID", "only cash and bonus wallet rewards are currently executable")
		}
		if rule.CalculationType != "FIXED" && rule.CalculationType != "PERCENTAGE" {
			return newError("PROMOTION_CALCULATION_INVALID", "unsupported reward calculation type")
		}
		amount, err := decimal.NewFromString(strings.TrimSpace(rule.RewardAmount))
		if err != nil || amount.IsNegative() || (rule.CalculationType == "FIXED" && !amount.IsPositive()) {
			return newError("PROMOTION_REWARD_INVALID", "fixed rewards must have a positive amount")
		}
		percentage := decimal.Zero
		if strings.TrimSpace(rule.PercentageRate) != "" {
			percentage, err = decimal.NewFromString(strings.TrimSpace(rule.PercentageRate))
		}
		if err != nil || percentage.IsNegative() || (rule.CalculationType == "PERCENTAGE" && !percentage.IsPositive()) {
			return newError("PROMOTION_PERCENTAGE_INVALID", "percentage rewards must have a positive rate")
		}
	}
	if !hasEnabledReward {
		return newError("PROMOTION_REWARD_REQUIRED", "at least one enabled reward rule is required")
	}
	return nil
}

func promotionExecutable(triggerType string, claimMode string) bool {
	return strings.EqualFold(triggerType, "DEPOSIT_SUCCEEDED") && strings.EqualFold(claimMode, "AUTO")
}
