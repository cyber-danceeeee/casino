package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameDepositVoucherRewards Deposit voucher reward rules by amount option.
type GameDepositVoucherRewards struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID       uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID         uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	VoucherID      uint64           `gorm:"column:voucher_id;type:bigint(20) unsigned;not null;uniqueIndex:uk_voucher_amount" json:"voucherID"`
	AmountOptionID uint64           `gorm:"column:amount_option_id;type:bigint(20) unsigned;not null" json:"amountOptionID"`
	Amount         *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null;uniqueIndex:uk_voucher_amount" json:"amount"`
	BonusReward    *decimal.Decimal `gorm:"column:bonus_reward;type:decimal(18,2);default:0.00;not null" json:"bonusReward"`
	CashReward     *decimal.Decimal `gorm:"column:cash_reward;type:decimal(18,2);default:0.00;not null" json:"cashReward"`
	RewardAmount   *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	Sort           int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status         int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameDepositVoucherRewardsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameDepositVoucherRewardsColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"tenant_id":        true,
	"site_id":          true,
	"voucher_id":       true,
	"amount_option_id": true,
	"amount":           true,
	"bonus_reward":     true,
	"cash_reward":      true,
	"reward_amount":    true,
	"sort":             true,
	"status":           true,
}
