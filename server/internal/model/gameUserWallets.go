package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameUserWallets Member wallet balances
type GameUserWallets struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID            uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID              uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID              uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	Currency            string           `gorm:"column:currency;type:varchar(8);default:PKR;not null" json:"currency"`
	CashBalance         *decimal.Decimal `gorm:"column:cash_balance;type:decimal(18,2);default:0.00;not null" json:"cashBalance"`
	BonusBalance        *decimal.Decimal `gorm:"column:bonus_balance;type:decimal(18,2);default:0.00;not null" json:"bonusBalance"`
	WithdrawableBalance *decimal.Decimal `gorm:"column:withdrawable_balance;type:decimal(18,2);default:0.00;not null" json:"withdrawableBalance"`
	FrozenBalance       *decimal.Decimal `gorm:"column:frozen_balance;type:decimal(18,2);default:0.00;not null" json:"frozenBalance"`
	Status              int              `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	LastTransactionAt   *time.Time       `gorm:"column:last_transaction_at;type:datetime(3)" json:"lastTransactionAt"`
}

// GameUserWalletsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserWalletsColumnNames = map[string]bool{
	"id":                   true,
	"created_at":           true,
	"updated_at":           true,
	"deleted_at":           true,
	"tenant_id":            true,
	"site_id":              true,
	"user_id":              true,
	"currency":             true,
	"cash_balance":         true,
	"bonus_balance":        true,
	"withdrawable_balance": true,
	"frozen_balance":       true,
	"status":               true,
	"last_transaction_at":  true,
}
