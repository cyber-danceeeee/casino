package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameInvitationWheelPrizes Invitation wheel prizes
type GameInvitationWheelPrizes struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	PrizeCode  string           `gorm:"column:prize_code;type:varchar(64);not null" json:"prizeCode"`
	Label      string           `gorm:"column:label;type:varchar(64);not null" json:"label"`
	RewardType string           `gorm:"column:reward_type;type:varchar(32);not null" json:"rewardType"` // cash, bonus, spin, chest, empty
	MinAmount  *decimal.Decimal `gorm:"column:min_amount;type:decimal(18,2);default:0.00;not null" json:"minAmount"`
	MaxAmount  *decimal.Decimal `gorm:"column:max_amount;type:decimal(18,2);default:0.00;not null" json:"maxAmount"`
	Weight     uint             `gorm:"column:weight;type:int(11) unsigned;default:0;not null" json:"weight"`
	IconURL    string           `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	Sort       int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status     int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameInvitationWheelPrizesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameInvitationWheelPrizesColumnNames = map[string]bool{
	"id":          true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
	"prize_code":  true,
	"label":       true,
	"reward_type": true,
	"min_amount":  true,
	"max_amount":  true,
	"weight":      true,
	"icon_url":    true,
	"sort":        true,
	"status":      true,
}


