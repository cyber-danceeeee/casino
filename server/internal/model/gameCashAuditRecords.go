package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameCashAuditRecords Cash audit requirements
type GameCashAuditRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID                uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	SourceType            string           `gorm:"column:source_type;type:varchar(64);not null" json:"sourceType"`
	SourceID              uint64           `gorm:"column:source_id;type:bigint(20) unsigned;default:0;not null" json:"sourceID"`
	Amount                *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	RequirementMultiplier *decimal.Decimal `gorm:"column:requirement_multiplier;type:decimal(8,2);default:1.00;not null" json:"requirementMultiplier"`
	RequiredValidBet      *decimal.Decimal `gorm:"column:required_valid_bet;type:decimal(18,2);default:0.00;not null" json:"requiredValidBet"`
	CompletedValidBet     *decimal.Decimal `gorm:"column:completed_valid_bet;type:decimal(18,2);default:0.00;not null" json:"completedValidBet"`
	Status                int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 completed, 3 cancelled
	StartedAt             *time.Time       `gorm:"column:started_at;type:datetime(3)" json:"startedAt"`
	CompletedAt           *time.Time       `gorm:"column:completed_at;type:datetime(3)" json:"completedAt"`
}

// GameCashAuditRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameCashAuditRecordsColumnNames = map[string]bool{
	"id":                     true,
	"created_at":             true,
	"updated_at":             true,
	"deleted_at":             true,
	"user_id":                true,
	"source_type":            true,
	"source_id":              true,
	"amount":                 true,
	"requirement_multiplier": true,
	"required_valid_bet":     true,
	"completed_valid_bet":    true,
	"status":                 true,
	"started_at":             true,
	"completed_at":           true,
}
