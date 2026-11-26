package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"time"
)

// GameBonusRecords Bonus center records
type GameBonusRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID              uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID                uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	OrderNo               string           `gorm:"column:order_no;type:varchar(64);not null" json:"orderNo"`
	BonusID               string           `gorm:"column:bonus_id;type:varchar(64);not null" json:"bonusID"`
	UserID                uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	SourceType            string           `gorm:"column:source_type;type:varchar(64);not null" json:"sourceType"`
	SourceID              uint64           `gorm:"column:source_id;type:bigint(20) unsigned;default:0;not null" json:"sourceID"`
	PromotionID           uint64           `gorm:"column:promotion_id;type:bigint(20) unsigned;default:0;not null" json:"promotionID"`
	BonusType             string           `gorm:"column:bonus_type;type:varchar(32);not null" json:"bonusType"`
	Amount                *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	Currency              string           `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	RequirementMultiplier *decimal.Decimal `gorm:"column:requirement_multiplier;type:decimal(8,2);default:1.00;not null" json:"requirementMultiplier"`
	RequiredValidBet      *decimal.Decimal `gorm:"column:required_valid_bet;type:decimal(18,2);default:0.00;not null" json:"requiredValidBet"`
	CompletedValidBet     *decimal.Decimal `gorm:"column:completed_valid_bet;type:decimal(18,2);default:0.00;not null" json:"completedValidBet"`
	CurrentWagering       *decimal.Decimal `gorm:"column:current_wagering;type:decimal(18,2);default:0.00;not null" json:"currentWagering"`
	RemainingWagering     *decimal.Decimal `gorm:"column:remaining_wagering;type:decimal(18,2);default:0.00;not null" json:"remainingWagering"`
	WithdrawableAmount    *decimal.Decimal `gorm:"column:withdrawable_amount;type:decimal(18,2);default:0.00;not null" json:"withdrawableAmount"`
	Status                int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 locked, 2 unlocked, 3 expired, 4 cancelled
	StatusCode            string           `gorm:"column:status_code;type:varchar(32);not null" json:"statusCode"`
	IssuedAt              *time.Time       `gorm:"column:issued_at;type:datetime(3)" json:"issuedAt"`
	UnlockedAt            *time.Time       `gorm:"column:unlocked_at;type:datetime(3)" json:"unlockedAt"`
	ExpiresAt             *time.Time       `gorm:"column:expires_at;type:datetime(3)" json:"expiresAt"`
}

// GameBonusRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameBonusRecordsColumnNames = map[string]bool{
	"id":                     true,
	"created_at":             true,
	"updated_at":             true,
	"deleted_at":             true,
	"tenant_id":              true,
	"site_id":                true,
	"order_no":               true,
	"bonus_id":               true,
	"user_id":                true,
	"source_type":            true,
	"source_id":              true,
	"promotion_id":           true,
	"bonus_type":             true,
	"amount":                 true,
	"currency":               true,
	"requirement_multiplier": true,
	"required_valid_bet":     true,
	"completed_valid_bet":    true,
	"current_wagering":       true,
	"remaining_wagering":     true,
	"withdrawable_amount":    true,
	"status":                 true,
	"status_code":            true,
	"issued_at":              true,
	"unlocked_at":            true,
	"expires_at":             true,
}
