package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameWelfareCards Welfare cards
type GameWelfareCards struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID     uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID       uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	CardCode     string           `gorm:"column:card_code;type:varchar(64);not null" json:"cardCode"`
	Title        string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	PeriodType   string           `gorm:"column:period_type;type:varchar(32);default:day;not null" json:"periodType"`
	PriceAmount  *decimal.Decimal `gorm:"column:price_amount;type:decimal(18,2);default:0.00;not null" json:"priceAmount"`
	RewardAmount *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	RewardWallet string           `gorm:"column:reward_wallet;type:varchar(32);default:bonus;not null" json:"rewardWallet"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameWelfareCardsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameWelfareCardsColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"tenant_id":     true,
	"site_id":       true,
	"card_code":     true,
	"title":         true,
	"period_type":   true,
	"price_amount":  true,
	"reward_amount": true,
	"reward_wallet": true,
	"status":        true,
}
