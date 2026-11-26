package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type GamePromotionClaim struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID         uint64           `gorm:"column:site_id" json:"siteID"`
	PromotionID    uint64           `gorm:"column:promotion_id" json:"promotionID"`
	UserID         uint64           `gorm:"column:user_id" json:"userID"`
	TriggerType    string           `gorm:"column:trigger_type" json:"triggerType"`
	TriggerEventID string           `gorm:"column:trigger_event_id" json:"triggerEventID"`
	TriggerOrderNo string           `gorm:"column:trigger_order_no" json:"triggerOrderNo"`
	IdempotencyKey string           `gorm:"column:idempotency_key;uniqueIndex:uk_promotion_claim_idempotency" json:"idempotencyKey"`
	Currency       string           `gorm:"column:currency" json:"currency"`
	BaseAmount     *decimal.Decimal `gorm:"column:base_amount" json:"baseAmount"`
	RewardAmount   *decimal.Decimal `gorm:"column:reward_amount" json:"rewardAmount"`
	Status         string           `gorm:"column:status" json:"status"`
	FailureCode    string           `gorm:"column:failure_code" json:"failureCode"`
	FailureReason  string           `gorm:"column:failure_reason" json:"failureReason"`
	TraceID        string           `gorm:"column:trace_id" json:"traceID"`
	CorrelationID  string           `gorm:"column:correlation_id" json:"correlationID"`
	ProcessedAt    *time.Time       `gorm:"column:processed_at" json:"processedAt"`
	CreatedAt      time.Time        `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time        `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      *time.Time       `gorm:"column:deleted_at" json:"-"`
}

func (GamePromotionClaim) TableName() string { return "game_promotion_claims" }

type GamePromotionClaimReward struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID              uint64           `gorm:"column:site_id" json:"siteID"`
	ClaimID             uint64           `gorm:"column:claim_id;uniqueIndex:uk_promotion_claim_reward" json:"claimID"`
	PromotionID         uint64           `gorm:"column:promotion_id" json:"promotionID"`
	RewardRuleID        uint64           `gorm:"column:reward_rule_id;uniqueIndex:uk_promotion_claim_reward" json:"rewardRuleID"`
	RewardType          string           `gorm:"column:reward_type" json:"rewardType"`
	CalculationType     string           `gorm:"column:calculation_type" json:"calculationType"`
	Currency            string           `gorm:"column:currency" json:"currency"`
	BaseAmount          *decimal.Decimal `gorm:"column:base_amount" json:"baseAmount"`
	RewardAmount        *decimal.Decimal `gorm:"column:reward_amount" json:"rewardAmount"`
	WageringMultiplier  *decimal.Decimal `gorm:"column:wagering_multiplier" json:"wageringMultiplier"`
	WalletTransactionID uint64           `gorm:"column:wallet_transaction_id" json:"walletTransactionID"`
	BonusRecordID       uint64           `gorm:"column:bonus_record_id" json:"bonusRecordID"`
	Status              string           `gorm:"column:status" json:"status"`
	CreatedAt           time.Time        `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt           time.Time        `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt           *time.Time       `gorm:"column:deleted_at" json:"-"`
}

func (GamePromotionClaimReward) TableName() string { return "game_promotion_claim_rewards" }
