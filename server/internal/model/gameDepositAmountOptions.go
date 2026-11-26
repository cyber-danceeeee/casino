package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameDepositAmountOptions Deposit amount presets and rewards
type GameDepositAmountOptions struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID     uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID       uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	OptionCode   string           `gorm:"column:option_code;type:varchar(64);not null" json:"optionCode"`
	CurrencyCode string           `gorm:"column:currency_code;type:varchar(8);default:PKR;not null" json:"currencyCode"`
	Amount       *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	BonusReward  *decimal.Decimal `gorm:"column:bonus_reward;type:decimal(18,2);default:0.00;not null" json:"bonusReward"`
	CashReward   *decimal.Decimal `gorm:"column:cash_reward;type:decimal(18,2);default:0.00;not null" json:"cashReward"`
	Sort         int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameDepositAmountOptionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameDepositAmountOptionsColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"tenant_id":     true,
	"site_id":       true,
	"option_code":   true,
	"currency_code": true,
	"amount":        true,
	"bonus_reward":  true,
	"cash_reward":   true,
	"sort":          true,
	"status":        true,
}
