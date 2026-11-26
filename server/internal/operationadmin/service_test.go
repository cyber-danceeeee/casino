package operationadmin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDecimalUsesExactArithmetic(t *testing.T) {
	value, err := parseDecimal("1000000000000.29", "amount")
	require.NoError(t, err)
	require.Equal(t, "1000000000000.29", value.StringFixed(2))
	_, err = parseDecimal("-0.01", "amount")
	require.Error(t, err)
}

func TestValidateExternalURLRejectsUnsafeProtocols(t *testing.T) {
	require.NoError(t, validateExternalURL("EXTERNAL_URL", "https://example.com/promotion"))
	require.Error(t, validateExternalURL("EXTERNAL_URL", "javascript:alert(1)"))
	require.NoError(t, validateExternalURL("INTERNAL_ROUTE", "/offers?tab=event"))
}

func TestValidateTimeRange(t *testing.T) {
	start := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(-time.Second)
	require.Error(t, validateTime(&start, &end))
	end = start.Add(time.Hour)
	require.NoError(t, validateTime(&start, &end))
}

func TestContainsCode(t *testing.T) {
	require.True(t, containsCode("PK,MY", "pk"))
	require.True(t, containsCode("ALL", "US"))
	require.True(t, containsCode("", "US"))
	require.False(t, containsCode("PK,MY", "US"))
}

func TestBonusAndPromotionStatusMappings(t *testing.T) {
	require.Equal(t, 1, bonusLegacyStatus("ACTIVE"))
	require.Equal(t, 4, bonusLegacyStatus("FORFEITED"))
	require.Equal(t, 1, promotionLegacyStatus("SCHEDULED"))
	require.Equal(t, 3, promotionLegacyStatus("CANCELLED"))
}

func TestAudienceValuesTrimsAndDropsEmptyEntries(t *testing.T) {
	require.Equal(t, []string{"PK", "MY", "US"}, audienceValues(" PK,MY\n\nUS, "))
}

func TestPromotionEngineValidationOnlyActivatesConnectedTrigger(t *testing.T) {
	input := PromotionInput{
		Code: "welcome", Type: "FIRST_DEPOSIT_BONUS", TriggerType: "deposit_succeeded", ClaimMode: "auto",
		Status: "active", Countries: "mx", Currencies: "usd", Languages: "zh-CN,en-US", TotalBudget: "1000",
		RewardRules: []RewardRuleInput{{
			RewardType: "bonus", CalculationType: "fixed", RewardAmount: "50", PercentageRate: "0", Currency: "usd",
			WageringMultiplier: "2", MinimumAmount: "0", MaximumReward: "100", Enabled: true,
		}},
	}
	normalizePromotionInput(&input)
	require.NoError(t, validatePromotionInput(input))
	require.Equal(t, "WELCOME", input.Code)
	require.Equal(t, "DEPOSIT_SUCCEEDED", input.TriggerType)

	input.TriggerType = "REGISTRATION_COMPLETED"
	require.Error(t, validatePromotionInput(input))
	input.Status = "DRAFT"
	require.NoError(t, validatePromotionInput(input))
	input.RewardRules[0].RewardType = "FREE_SPINS"
	require.Error(t, validatePromotionInput(input))
	input.RewardRules[0].RewardType = "BONUS"
	input.RewardRules[0].Enabled = false
	require.Error(t, validatePromotionInput(input))
}
