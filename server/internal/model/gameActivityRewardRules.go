package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameActivityRewardRules Activity reward tiers
type GameActivityRewardRules struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	ActivityID            uint64           `gorm:"column:activity_id;type:bigint(20) unsigned;not null" json:"activityID"`
	Label                 string           `gorm:"column:label;type:varchar(128);not null" json:"label"`
	RewardValue           string           `gorm:"column:reward_value;type:varchar(128);not null" json:"rewardValue"`
	BonusText             string           `gorm:"column:bonus_text;type:varchar(128);not null" json:"bonusText"`
	MinAmount             *decimal.Decimal `gorm:"column:min_amount;type:decimal(18,2);default:0.00;not null" json:"minAmount"`
	MaxAmount             *decimal.Decimal `gorm:"column:max_amount;type:decimal(18,2);default:0.00;not null" json:"maxAmount"`
	RewardCash            *decimal.Decimal `gorm:"column:reward_cash;type:decimal(18,2);default:0.00;not null" json:"rewardCash"`
	RewardBonus           *decimal.Decimal `gorm:"column:reward_bonus;type:decimal(18,2);default:0.00;not null" json:"rewardBonus"`
	RequirementMultiplier *decimal.Decimal `gorm:"column:requirement_multiplier;type:decimal(8,2);default:1.00;not null" json:"requirementMultiplier"`
	Sort                  int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status                int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameActivityRewardRulesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameActivityRewardRulesColumnNames = map[string]bool{
	"id":                     true,
	"created_at":             true,
	"updated_at":             true,
	"deleted_at":             true,
	"activity_id":            true,
	"label":                  true,
	"reward_value":           true,
	"bonus_text":             true,
	"min_amount":             true,
	"max_amount":             true,
	"reward_cash":            true,
	"reward_bonus":           true,
	"requirement_multiplier": true,
	"sort":                   true,
	"status":                 true,
}
