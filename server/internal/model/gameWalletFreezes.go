package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

type GameWalletFreezes struct {
	sgorm.Model `gorm:"embedded"`

	SiteID               uint64           `gorm:"column:site_id" json:"siteID"`
	FreezeNo             string           `gorm:"column:freeze_no" json:"freezeNo"`
	WalletID             uint64           `gorm:"column:wallet_id" json:"walletID"`
	UserID               uint64           `gorm:"column:user_id" json:"userID"`
	Amount               *decimal.Decimal `gorm:"column:amount" json:"amount"`
	Currency             string           `gorm:"column:currency" json:"currency"`
	Reason               string           `gorm:"column:reason" json:"reason"`
	ReferenceType        string           `gorm:"column:reference_type" json:"referenceType"`
	ReferenceID          uint64           `gorm:"column:reference_id" json:"referenceID"`
	ReferenceNo          string           `gorm:"column:reference_no" json:"referenceNo"`
	Status               int              `gorm:"column:status" json:"status"`
	FreezeTransactionID  uint64           `gorm:"column:freeze_transaction_id" json:"freezeTransactionID"`
	ReleaseTransactionID uint64           `gorm:"column:release_transaction_id" json:"releaseTransactionID"`
	OperatorID           uint64           `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername     string           `gorm:"column:operator_username" json:"operatorUsername"`
	FrozenAt             time.Time        `gorm:"column:frozen_at" json:"frozenAt"`
	ReleasedAt           *time.Time       `gorm:"column:released_at" json:"releasedAt"`
}

func (GameWalletFreezes) TableName() string { return "game_wallet_freezes" }
