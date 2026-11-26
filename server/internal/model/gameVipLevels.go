package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameVipLevels VIP levels
type GameVipLevels struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID           uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID             uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	LevelNo            uint             `gorm:"column:level_no;type:int(11) unsigned;not null" json:"levelNo"`
	Name               string           `gorm:"column:name;type:varchar(64);not null" json:"name"`
	IconURL            string           `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	LevelUpReward      *decimal.Decimal `gorm:"column:level_up_reward;type:decimal(18,2);default:0.00;not null" json:"levelUpReward"`
	WeeklyReward       *decimal.Decimal `gorm:"column:weekly_reward;type:decimal(18,2);default:0.00;not null" json:"weeklyReward"`
	MonthlyReward      *decimal.Decimal `gorm:"column:monthly_reward;type:decimal(18,2);default:0.00;not null" json:"monthlyReward"`
	DailyWithdrawCount uint             `gorm:"column:daily_withdraw_count;type:int(11) unsigned;default:0;not null" json:"dailyWithdrawCount"`
	Status             int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameVipLevelsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameVipLevelsColumnNames = map[string]bool{
	"id":                   true,
	"created_at":           true,
	"updated_at":           true,
	"deleted_at":           true,
	"tenant_id":            true,
	"site_id":              true,
	"level_no":             true,
	"name":                 true,
	"icon_url":             true,
	"level_up_reward":      true,
	"weekly_reward":        true,
	"monthly_reward":       true,
	"daily_withdraw_count": true,
	"status":               true,
}
