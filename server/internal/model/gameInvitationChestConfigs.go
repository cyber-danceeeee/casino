package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameInvitationChestConfigs Invitation chest tiers
type GameInvitationChestConfigs struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	ChestCode       string           `gorm:"column:chest_code;type:varchar(64);not null" json:"chestCode"`
	Title           string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	RequiredInvites uint             `gorm:"column:required_invites;type:int(11) unsigned;default:0;not null" json:"requiredInvites"`
	RewardCash      *decimal.Decimal `gorm:"column:reward_cash;type:decimal(18,2);default:0.00;not null" json:"rewardCash"`
	RewardBonus     *decimal.Decimal `gorm:"column:reward_bonus;type:decimal(18,2);default:0.00;not null" json:"rewardBonus"`
	IconURL         string           `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	Sort            int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status          int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameInvitationChestConfigsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameInvitationChestConfigsColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"chest_code":       true,
	"title":            true,
	"required_invites": true,
	"reward_cash":      true,
	"reward_bonus":     true,
	"icon_url":         true,
	"sort":             true,
	"status":           true,
}
