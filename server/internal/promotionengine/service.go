package promotionengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/eventbus"
	"casino/internal/model"
)

const (
	TriggerDepositSucceeded = "DEPOSIT_SUCCEEDED"
	ClaimStatusCompleted    = "COMPLETED"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type DepositTrigger struct {
	TenantID        uint64
	SiteID          uint64
	UserID          uint64
	DepositOrderID  uint64
	OrderNo         string
	EventID         string
	Amount          decimal.Decimal
	Currency        string
	Country         string
	DepositSequence uint64
	TraceID         string
	CorrelationID   string
	OccurredAt      time.Time
}

type promotionRecord struct {
	ID              uint64           `gorm:"primaryKey"`
	TenantID        uint64           `gorm:"column:tenant_id"`
	SiteID          uint64           `gorm:"column:site_id"`
	ActivityCode    string           `gorm:"column:activity_code"`
	TriggerType     string           `gorm:"column:trigger_type"`
	ClaimMode       string           `gorm:"column:claim_mode"`
	StackingGroup   string           `gorm:"column:stacking_group"`
	Priority        int              `gorm:"column:priority"`
	Countries       string           `gorm:"column:countries"`
	Currencies      string           `gorm:"column:currencies"`
	StatusCode      string           `gorm:"column:status_code"`
	TotalBudget     *decimal.Decimal `gorm:"column:total_budget"`
	PerUserLimit    uint             `gorm:"column:per_user_limit"`
	TotalClaimLimit uint             `gorm:"column:total_claim_limit"`
	ClaimCount      uint             `gorm:"column:claim_count"`
	RewardAmount    *decimal.Decimal `gorm:"column:reward_amount"`
	StartAt         *time.Time       `gorm:"column:start_at"`
	EndAt           *time.Time       `gorm:"column:end_at"`
	DeletedAt       *time.Time       `gorm:"column:deleted_at"`
}

func (promotionRecord) TableName() string { return "game_activity_events" }

type calculatedReward struct {
	Rule   model.GamePromotionRewardRule
	Amount decimal.Decimal
}

func (s *Service) HandleDepositEvent(ctx eventbus.Context) error {
	event := ctx.Event
	if event.EventType != eventbus.EventPlayerDepositSucceeded {
		return nil
	}
	amount, err := decimal.NewFromString(payloadString(event.Payload, "amount"))
	if err != nil {
		return fmt.Errorf("invalid deposit event amount: %w", err)
	}
	trigger := DepositTrigger{
		TenantID: event.TenantID, SiteID: event.SiteID, UserID: payloadUint(event.Payload, "userID"),
		DepositOrderID: payloadUint(event.Payload, "depositOrderID"), OrderNo: payloadString(event.Payload, "orderNo"),
		EventID: event.EventID, Amount: amount, Currency: strings.ToUpper(payloadString(event.Payload, "currency")),
		Country: strings.ToUpper(payloadString(event.Payload, "country")), DepositSequence: payloadUint(event.Payload, "depositSequence"),
		TraceID: event.TraceID, CorrelationID: event.CorrelationID, OccurredAt: event.OccurredAt,
	}
	return s.ProcessDepositSucceeded(context.Background(), trigger)
}

func (s *Service) ProcessDepositSucceeded(ctx context.Context, trigger DepositTrigger) error {
	if trigger.TenantID == 0 || trigger.SiteID == 0 || trigger.UserID == 0 || trigger.EventID == "" || !trigger.Amount.IsPositive() {
		return fmt.Errorf("invalid trusted promotion trigger scope or payload")
	}
	if trigger.OccurredAt.IsZero() {
		trigger.OccurredAt = time.Now()
	}
	var promotions []promotionRecord
	err := s.db.WithContext(ctx).Where(
		"tenant_id=? AND site_id=? AND trigger_type=? AND claim_mode='AUTO' AND status_code='ACTIVE' AND deleted_at IS NULL",
		trigger.TenantID, trigger.SiteID, TriggerDepositSucceeded,
	).Where("start_at IS NULL OR start_at<=?", trigger.OccurredAt).
		Where("end_at IS NULL OR end_at>=?", trigger.OccurredAt).
		Order("priority DESC,id ASC").Find(&promotions).Error
	if err != nil {
		return err
	}
	usedGroups := map[string]bool{}
	for _, promotion := range promotions {
		if !csvContains(promotion.Countries, trigger.Country) || !csvContains(promotion.Currencies, trigger.Currency) {
			continue
		}
		group := strings.ToUpper(strings.TrimSpace(promotion.StackingGroup))
		if group != "" && usedGroups[group] {
			continue
		}
		granted, processErr := s.processPromotion(ctx, promotion.ID, trigger)
		if processErr != nil {
			return processErr
		}
		if granted && group != "" {
			usedGroups[group] = true
		}
	}
	return nil
}

func (s *Service) processPromotion(ctx context.Context, promotionID uint64, trigger DepositTrigger) (bool, error) {
	granted := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var promotion promotionRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
			"id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", promotionID, trigger.TenantID, trigger.SiteID,
		).First(&promotion).Error; err != nil {
			return err
		}
		claimKey := fmt.Sprintf("promotion:%s:%d", trigger.EventID, promotion.ID)
		var existing model.GamePromotionClaim
		if err := tx.Where("idempotency_key=? AND deleted_at IS NULL", claimKey).First(&existing).Error; err == nil {
			granted = existing.Status == ClaimStatusCompleted
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		if promotion.StatusCode != "ACTIVE" || promotion.TriggerType != TriggerDepositSucceeded || promotion.ClaimMode != "AUTO" {
			return nil
		}
		var user model.GameUsers
		if err := tx.Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", trigger.UserID, trigger.TenantID, trigger.SiteID).First(&user).Error; err != nil {
			return err
		}
		if !csvContains(promotion.Countries, user.CountryCode) || !csvContains(promotion.Currencies, trigger.Currency) {
			return nil
		}
		depositSequence := trigger.DepositSequence
		if depositSequence == 0 {
			var count int64
			if err := tx.Model(&model.GameDepositOrders{}).Where(
				"tenant_id=? AND site_id=? AND user_id=? AND status=2 AND deleted_at IS NULL", trigger.TenantID, trigger.SiteID, trigger.UserID,
			).Count(&count).Error; err != nil {
				return err
			}
			depositSequence = uint64(count)
		}
		var rules []model.GamePromotionParticipationRule
		if err := tx.Where("tenant_id=? AND site_id=? AND promotion_id=? AND enabled=1 AND deleted_at IS NULL", trigger.TenantID, trigger.SiteID, promotion.ID).
			Order("sort_order,id").Find(&rules).Error; err != nil {
			return err
		}
		for _, rule := range rules {
			if !conditionMatches(rule, trigger, user.VipLevel, depositSequence) {
				return nil
			}
		}
		var userClaims int64
		if err := tx.Model(&model.GamePromotionClaim{}).Where(
			"tenant_id=? AND site_id=? AND promotion_id=? AND user_id=? AND status=? AND deleted_at IS NULL",
			trigger.TenantID, trigger.SiteID, promotion.ID, trigger.UserID, ClaimStatusCompleted,
		).Count(&userClaims).Error; err != nil {
			return err
		}
		if promotion.PerUserLimit > 0 && uint(userClaims) >= promotion.PerUserLimit {
			return nil
		}
		if promotion.TotalClaimLimit > 0 && promotion.ClaimCount >= promotion.TotalClaimLimit {
			return nil
		}
		var rulesReward []model.GamePromotionRewardRule
		if err := tx.Where("tenant_id=? AND site_id=? AND promotion_id=? AND enabled=1 AND deleted_at IS NULL", trigger.TenantID, trigger.SiteID, promotion.ID).
			Order("sort_order,id").Find(&rulesReward).Error; err != nil {
			return err
		}
		rewards := calculateRewards(rulesReward, trigger.Amount, trigger.Currency)
		if len(rewards) == 0 {
			return nil
		}
		total := decimal.Zero
		for _, reward := range rewards {
			total = total.Add(reward.Amount)
		}
		currentReward := decimalValue(promotion.RewardAmount)
		if budget := decimalValue(promotion.TotalBudget); budget.IsPositive() && currentReward.Add(total).GreaterThan(budget) {
			return nil
		}
		var wallet model.GameUserWallets
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
			"tenant_id=? AND site_id=? AND user_id=? AND currency=? AND deleted_at IS NULL",
			trigger.TenantID, trigger.SiteID, trigger.UserID, trigger.Currency,
		).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.Status != 1 {
			return fmt.Errorf("promotion wallet is not active")
		}
		now := time.Now()
		claim := model.GamePromotionClaim{
			TenantID: trigger.TenantID, SiteID: trigger.SiteID, PromotionID: promotion.ID, UserID: trigger.UserID,
			TriggerType: TriggerDepositSucceeded, TriggerEventID: trigger.EventID, TriggerOrderNo: trigger.OrderNo,
			IdempotencyKey: claimKey, Currency: trigger.Currency, BaseAmount: decimalPtr(trigger.Amount), RewardAmount: decimalPtr(decimal.Zero),
			Status: "PROCESSING", TraceID: trigger.TraceID, CorrelationID: trigger.CorrelationID,
		}
		create := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&claim)
		if create.Error != nil {
			return create.Error
		}
		if create.RowsAffected == 0 {
			if err := tx.Where("idempotency_key=? AND deleted_at IS NULL", claimKey).First(&existing).Error; err != nil {
				return err
			}
			granted = existing.Status == ClaimStatusCompleted
			return nil
		}
		cashBalance := decimalValue(wallet.CashBalance)
		bonusBalance := decimalValue(wallet.BonusBalance)
		for _, reward := range rewards {
			balanceBefore := bonusBalance
			balanceAfter := bonusBalance.Add(reward.Amount)
			walletType := "BONUS"
			if reward.Rule.RewardType == "CASH" {
				walletType = "CASH"
				balanceBefore = cashBalance
				balanceAfter = cashBalance.Add(reward.Amount)
				cashBalance = balanceAfter
			} else {
				bonusBalance = balanceAfter
			}
			txnKey := fmt.Sprintf("promotion:%d:reward:%d", claim.ID, reward.Rule.ID)
			txn := model.GameWalletTransactions{
				TenantID: trigger.TenantID, SiteID: trigger.SiteID, TxnNo: promotionNo("PMT"), UserID: trigger.UserID,
				WalletID: wallet.ID, WalletType: walletType, Direction: "IN", Category: "PROMOTION_REWARD", SourceType: "promotion",
				SourceID: promotion.ID, IdempotencyKey: &txnKey, OrderNo: trigger.OrderNo, Amount: decimalPtr(reward.Amount), Currency: trigger.Currency,
				Status: 1, BalanceBefore: decimalPtr(balanceBefore), BalanceAfter: decimalPtr(balanceAfter),
				FrozenBalanceBefore: wallet.FrozenBalance, FrozenBalanceAfter: wallet.FrozenBalance, Remark: "promotion reward",
				CompletedAt: &now, RawMetadata: promotionMetadata(claim.ID, reward.Rule.ID, trigger.EventID),
			}
			if err := tx.Create(&txn).Error; err != nil {
				return err
			}
			bonusRecordID := uint64(0)
			if reward.Rule.RewardType == "BONUS" {
				bonusRecord, err := createBonusRecord(tx, promotion, reward, trigger, now)
				if err != nil {
					return err
				}
				bonusRecordID = bonusRecord.ID
			}
			entry := model.GamePromotionClaimReward{
				TenantID: trigger.TenantID, SiteID: trigger.SiteID, ClaimID: claim.ID, PromotionID: promotion.ID,
				RewardRuleID: reward.Rule.ID, RewardType: reward.Rule.RewardType, CalculationType: reward.Rule.CalculationType,
				Currency: trigger.Currency, BaseAmount: decimalPtr(trigger.Amount), RewardAmount: decimalPtr(reward.Amount),
				WageringMultiplier: reward.Rule.WageringMultiplier, WalletTransactionID: txn.ID, BonusRecordID: bonusRecordID, Status: ClaimStatusCompleted,
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&wallet).Updates(map[string]any{"cash_balance": cashBalance, "bonus_balance": bonusBalance, "last_transaction_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&claim).Updates(map[string]any{"reward_amount": total, "status": ClaimStatusCompleted, "processed_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&promotion).Updates(map[string]any{"claim_count": gorm.Expr("claim_count + 1"), "reward_amount": gorm.Expr("reward_amount + ?", total)}).Error; err != nil {
			return err
		}
		granted = true
		return nil
	})
	return granted, err
}

func calculateRewards(rules []model.GamePromotionRewardRule, base decimal.Decimal, currency string) []calculatedReward {
	result := make([]calculatedReward, 0, len(rules))
	for _, rule := range rules {
		if !strings.EqualFold(rule.Currency, currency) || !base.GreaterThanOrEqual(decimalValue(rule.MinimumAmount)) {
			continue
		}
		amount := decimalValue(rule.RewardAmount)
		if strings.EqualFold(rule.CalculationType, "PERCENTAGE") {
			amount = base.Mul(decimalValue(rule.PercentageRate)).Div(decimal.NewFromInt(100))
		}
		if maximum := decimalValue(rule.MaximumReward); maximum.IsPositive() && amount.GreaterThan(maximum) {
			amount = maximum
		}
		amount = amount.Round(2)
		if amount.IsPositive() && (rule.RewardType == "CASH" || rule.RewardType == "BONUS") {
			result = append(result, calculatedReward{Rule: rule, Amount: amount})
		}
	}
	return result
}

func conditionMatches(rule model.GamePromotionParticipationRule, trigger DepositTrigger, vipLevel uint, depositSequence uint64) bool {
	switch strings.ToUpper(rule.RuleType) {
	case "FIRST_DEPOSIT":
		return depositSequence == 1
	case "DEPOSIT_SEQUENCE":
		return compareDecimal(decimal.NewFromInt(int64(depositSequence)), decimalValue(rule.NumericValue), rule.ComparisonOperator)
	case "DEPOSIT_AMOUNT":
		return compareDecimal(trigger.Amount, decimalValue(rule.NumericValue), rule.ComparisonOperator)
	case "USER_VIP_LEVEL":
		return compareDecimal(decimal.NewFromInt(int64(vipLevel)), decimalValue(rule.NumericValue), rule.ComparisonOperator)
	case "COUNTRY":
		return compareText(trigger.Country, rule.TextValue, rule.ComparisonOperator)
	case "CURRENCY":
		return compareText(trigger.Currency, rule.TextValue, rule.ComparisonOperator)
	default:
		return false
	}
}

func compareDecimal(actual, expected decimal.Decimal, operator string) bool {
	switch strings.ToUpper(operator) {
	case "EQ":
		return actual.Equal(expected)
	case "NE":
		return !actual.Equal(expected)
	case "GT":
		return actual.GreaterThan(expected)
	case "GTE":
		return actual.GreaterThanOrEqual(expected)
	case "LT":
		return actual.LessThan(expected)
	case "LTE":
		return actual.LessThanOrEqual(expected)
	default:
		return false
	}
}

func compareText(actual, expected, operator string) bool {
	switch strings.ToUpper(operator) {
	case "EQ":
		return strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
	case "NE":
		return !strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
	case "IN":
		return csvContains(expected, actual)
	default:
		return false
	}
}

func createBonusRecord(tx *gorm.DB, promotion promotionRecord, reward calculatedReward, trigger DepositTrigger, now time.Time) (model.GameBonusRecords, error) {
	multiplier := decimalValue(reward.Rule.WageringMultiplier)
	required := reward.Amount.Mul(multiplier).Round(2)
	status, statusCode := 1, "PENDING"
	var unlockedAt *time.Time
	withdrawable := decimal.Zero
	if !required.IsPositive() {
		status, statusCode, unlockedAt, withdrawable = 2, "COMPLETED", &now, reward.Amount
	}
	var expiresAt *time.Time
	if reward.Rule.ExpirationDays > 0 {
		value := now.AddDate(0, 0, int(reward.Rule.ExpirationDays))
		expiresAt = &value
	}
	bonusNo := promotionNo("BONUS")
	record := model.GameBonusRecords{
		TenantID: trigger.TenantID, SiteID: trigger.SiteID, OrderNo: bonusNo, BonusID: bonusNo, UserID: trigger.UserID,
		SourceType: "PROMOTION", SourceID: promotion.ID, PromotionID: promotion.ID, BonusType: "PROMOTION",
		Amount: decimalPtr(reward.Amount), Currency: trigger.Currency, RequirementMultiplier: decimalPtr(multiplier),
		RequiredValidBet: decimalPtr(required), CompletedValidBet: decimalPtr(decimal.Zero), CurrentWagering: decimalPtr(decimal.Zero),
		RemainingWagering: decimalPtr(required), WithdrawableAmount: decimalPtr(withdrawable), Status: status, StatusCode: statusCode,
		IssuedAt: &now, UnlockedAt: unlockedAt, ExpiresAt: expiresAt,
	}
	return record, tx.Create(&record).Error
}

func promotionNo(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func promotionMetadata(claimID, rewardRuleID uint64, eventID string) string {
	data, _ := json.Marshal(map[string]any{"claimID": claimID, "rewardRuleID": rewardRuleID, "eventID": eventID})
	return string(data)
}

func csvContains(csv, value string) bool {
	if strings.TrimSpace(csv) == "" {
		return true
	}
	for _, item := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(item), "ALL") || strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func decimalValue(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

func decimalPtr(value decimal.Decimal) *decimal.Decimal { return &value }

func payloadString(payload map[string]any, key string) string {
	if payload[key] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(payload[key]))
}

func payloadUint(payload map[string]any, key string) uint64 {
	value, err := decimal.NewFromString(payloadString(payload, key))
	if err != nil || !value.IsPositive() {
		return 0
	}
	return value.BigInt().Uint64()
}
