package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

// GameProviderTransactions stores normalized provider wallet callback audit records.
type GameProviderTransactions struct {
	sgorm.Model `gorm:"embedded"`

	TenantID              uint64 `gorm:"column:tenant_id" json:"tenantID"`
	SiteID                uint64 `gorm:"column:site_id" json:"siteID"`
	ProviderCode          string `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	ProviderTransactionID string `gorm:"column:provider_transaction_id;type:varchar(128);not null" json:"providerTransactionID"`
	ProviderRoundID       string `gorm:"column:provider_round_id;type:varchar(128);not null" json:"providerRoundID"`
	ProviderGameID        string `gorm:"column:provider_game_id;type:varchar(128);not null" json:"providerGameID"`
	UserID                uint64 `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	WalletTransactionID   uint64 `gorm:"column:wallet_transaction_id;type:bigint(20) unsigned;default:0;not null" json:"walletTransactionID"`
	Operation             string `gorm:"column:operation;type:varchar(32);not null" json:"operation"`
	AmountMinor           int64  `gorm:"column:amount_minor;type:bigint(20);default:0;not null" json:"amountMinor"`
	Currency              string `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	Status                string `gorm:"column:status;type:varchar(32);not null" json:"status"` // PENDING, SUCCESS, REJECTED, ROLLED_BACK, FAILED
	RelatedTransactionID  string `gorm:"column:related_transaction_id;type:varchar(128);not null" json:"relatedTransactionID"`
	RequestHash           string `gorm:"column:request_hash;type:char(64);not null" json:"requestHash"`
	RawRequest            string `gorm:"column:raw_request;type:json" json:"rawRequest"`
	RawResponse           string `gorm:"column:raw_response;type:json" json:"rawResponse"`
	BalanceBeforeMinor    int64  `gorm:"column:balance_before_minor;type:bigint(20);default:0;not null" json:"balanceBeforeMinor"`
	BalanceAfterMinor     int64  `gorm:"column:balance_after_minor;type:bigint(20);default:0;not null" json:"balanceAfterMinor"`
}

func (GameProviderTransactions) TableName() string {
	return "game_provider_transactions"
}

const (
	GameProviderTransactionPending    = "PENDING"
	GameProviderTransactionSuccess    = "SUCCESS"
	GameProviderTransactionRejected   = "REJECTED"
	GameProviderTransactionRolledBack = "ROLLED_BACK"
	GameProviderTransactionFailed     = "FAILED"
)
