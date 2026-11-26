package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"time"
)

// GameAgentCommissionRecords Agent commission settlements
type GameAgentCommissionRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID           uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	PeriodStart      *time.Time       `gorm:"column:period_start;type:datetime(3);not null" json:"periodStart"`
	PeriodEnd        *time.Time       `gorm:"column:period_end;type:datetime(3);not null" json:"periodEnd"`
	ReferralsCount   uint             `gorm:"column:referrals_count;type:int(11) unsigned;default:0;not null" json:"referralsCount"`
	ValidBetAmount   *decimal.Decimal `gorm:"column:valid_bet_amount;type:decimal(18,2);default:0.00;not null" json:"validBetAmount"`
	CommissionAmount *decimal.Decimal `gorm:"column:commission_amount;type:decimal(18,2);default:0.00;not null" json:"commissionAmount"`
	BonusAmount      *decimal.Decimal `gorm:"column:bonus_amount;type:decimal(18,2);default:0.00;not null" json:"bonusAmount"`
	Status           int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 settled, 3 rejected
	SettledAt        *time.Time       `gorm:"column:settled_at;type:datetime(3)" json:"settledAt"`
}

// GameAgentCommissionRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameAgentCommissionRecordsColumnNames = map[string]bool{
	"id":                true,
	"created_at":        true,
	"updated_at":        true,
	"deleted_at":        true,
	"user_id":           true,
	"period_start":      true,
	"period_end":        true,
	"referrals_count":   true,
	"valid_bet_amount":  true,
	"commission_amount": true,
	"bonus_amount":      true,
	"status":            true,
	"settled_at":        true,
}


