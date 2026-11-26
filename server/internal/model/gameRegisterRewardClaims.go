package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameRegisterRewardClaims stores H5 register reward claim lifecycle.
type GameRegisterRewardClaims struct {
	sgorm.Model `gorm:"embedded"`

	SiteID              uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID              uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	Amount              *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	Currency            string           `gorm:"column:currency;type:varchar(8);default:PKR;not null" json:"currency"`
	WalletType          string           `gorm:"column:wallet_type;type:varchar(32);default:CASH;not null" json:"walletType"`
	Status              int              `gorm:"column:status;type:tinyint;default:1;not null" json:"status"` // 1 pending, 2 claimed, 3 cancelled
	CashBefore          *decimal.Decimal `gorm:"column:cash_before;type:decimal(18,2);default:0.00;not null" json:"cashBefore"`
	CashAfter           *decimal.Decimal `gorm:"column:cash_after;type:decimal(18,2);default:0.00;not null" json:"cashAfter"`
	WalletTransactionID uint64           `gorm:"column:wallet_transaction_id;type:bigint(20) unsigned;default:0;not null" json:"walletTransactionID"`
	IdempotencyKey      string           `gorm:"column:idempotency_key;type:varchar(128);not null" json:"idempotencyKey"`
	ClaimedAt           *time.Time       `gorm:"column:claimed_at;type:datetime(3)" json:"claimedAt"`
	RawMetadata         string           `gorm:"column:raw_metadata;type:json" json:"rawMetadata"`
}

func (GameRegisterRewardClaims) TableName() string { return "game_register_reward_claims" }
