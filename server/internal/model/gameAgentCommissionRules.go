package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameAgentCommissionRules Agent commission rules
type GameAgentCommissionRules struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	LevelNo        uint             `gorm:"column:level_no;type:int(11) unsigned;default:1;not null" json:"levelNo"`
	MinReferrals   uint             `gorm:"column:min_referrals;type:int(11) unsigned;default:0;not null" json:"minReferrals"`
	MinValidBet    *decimal.Decimal `gorm:"column:min_valid_bet;type:decimal(18,2);default:0.00;not null" json:"minValidBet"`
	CommissionRate *decimal.Decimal `gorm:"column:commission_rate;type:decimal(8,4);default:0.0000;not null" json:"commissionRate"`
	BonusRate      *decimal.Decimal `gorm:"column:bonus_rate;type:decimal(8,4);default:0.0000;not null" json:"bonusRate"`
	Status         int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameAgentCommissionRulesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameAgentCommissionRulesColumnNames = map[string]bool{
	"id":              true,
	"created_at":      true,
	"updated_at":      true,
	"deleted_at":      true,
	"level_no":        true,
	"min_referrals":   true,
	"min_valid_bet":   true,
	"commission_rate": true,
	"bonus_rate":      true,
	"status":          true,
}


