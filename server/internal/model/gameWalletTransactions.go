package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameWalletTransactions Wallet statements
type GameWalletTransactions struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID            uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID              uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	TxnNo               string           `gorm:"column:txn_no;type:varchar(64);not null" json:"txnNo"`
	ExternalTxnNo       string           `gorm:"column:external_txn_no;type:varchar(128);not null" json:"externalTxnNo"`
	UserID              uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	WalletID            uint64           `gorm:"column:wallet_id;type:bigint(20) unsigned;not null" json:"walletID"`
	WalletType          string           `gorm:"column:wallet_type;type:varchar(32);not null" json:"walletType"`
	Direction           string           `gorm:"column:direction;type:varchar(16);not null" json:"direction"`
	Category            string           `gorm:"column:category;type:varchar(64);not null" json:"category"`
	SourceType          string           `gorm:"column:source_type;type:varchar(64);not null" json:"sourceType"`
	SourceID            uint64           `gorm:"column:source_id;type:bigint(20) unsigned;default:0;not null" json:"sourceID"`
	ProviderCode        string           `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	IdempotencyKey      *string          `gorm:"column:idempotency_key;type:varchar(128)" json:"idempotencyKey"`
	OrderNo             string           `gorm:"column:order_no;type:varchar(64);not null" json:"orderNo"`
	Amount              *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	Currency            string           `gorm:"column:currency;type:varchar(8);default:PKR;not null" json:"currency"`
	Status              int              `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	BalanceBefore       *decimal.Decimal `gorm:"column:balance_before;type:decimal(18,2);default:0.00;not null" json:"balanceBefore"`
	BalanceAfter        *decimal.Decimal `gorm:"column:balance_after;type:decimal(18,2);default:0.00;not null" json:"balanceAfter"`
	FrozenBalanceBefore *decimal.Decimal `gorm:"column:frozen_balance_before;type:decimal(18,2);default:0.00;not null" json:"frozenBalanceBefore"`
	FrozenBalanceAfter  *decimal.Decimal `gorm:"column:frozen_balance_after;type:decimal(18,2);default:0.00;not null" json:"frozenBalanceAfter"`
	Remark              string           `gorm:"column:remark;type:varchar(255);not null" json:"remark"`
	OperatorID          uint64           `gorm:"column:operator_id;type:bigint unsigned;not null" json:"operatorID"`
	OperatorUsername    string           `gorm:"column:operator_username;type:varchar(64);not null" json:"operatorUsername"`
	CompletedAt         *time.Time       `gorm:"column:completed_at;type:datetime(3)" json:"completedAt"`
	RawMetadata         string           `gorm:"column:raw_metadata;type:json" json:"rawMetadata"`
}

// GameWalletTransactionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameWalletTransactionsColumnNames = map[string]bool{
	"id":                    true,
	"created_at":            true,
	"updated_at":            true,
	"deleted_at":            true,
	"tenant_id":             true,
	"site_id":               true,
	"txn_no":                true,
	"external_txn_no":       true,
	"user_id":               true,
	"wallet_id":             true,
	"wallet_type":           true,
	"direction":             true,
	"category":              true,
	"source_type":           true,
	"source_id":             true,
	"provider_code":         true,
	"idempotency_key":       true,
	"order_no":              true,
	"amount":                true,
	"currency":              true,
	"status":                true,
	"balance_before":        true,
	"balance_after":         true,
	"frozen_balance_before": true,
	"frozen_balance_after":  true,
	"remark":                true,
	"operator_id":           true,
	"operator_username":     true,
	"completed_at":          true,
	"raw_metadata":          true,
}
