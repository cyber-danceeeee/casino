package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameWithdrawAccounts Withdrawal bank and e-wallet accounts
type GameWithdrawAccounts struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID         uint64 `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID           uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID           uint64 `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	AccountType      string `gorm:"column:account_type;type:varchar(16);not null" json:"accountType"` // bank, wallet
	BankName         string `gorm:"column:bank_name;type:varchar(128);not null" json:"bankName"`
	WalletType       string `gorm:"column:wallet_type;type:varchar(32);not null" json:"walletType"`
	AccountName      string `gorm:"column:account_name;type:varchar(128);not null" json:"accountName"`
	AccountNo        string `gorm:"column:account_no;type:varchar(128);not null" json:"accountNo"`
	PhoneCountryCode string `gorm:"column:phone_country_code;type:varchar(8);default:+92;not null" json:"phoneCountryCode"`
	Phone            string `gorm:"column:phone;type:varchar(32);not null" json:"phone"`
	IsDefault        int    `gorm:"column:is_default;type:tinyint(4);default:0;not null" json:"isDefault"`
	Status           int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 disabled, 3 reviewing
}

// GameWithdrawAccountsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameWithdrawAccountsColumnNames = map[string]bool{
	"id":                 true,
	"created_at":         true,
	"updated_at":         true,
	"deleted_at":         true,
	"tenant_id":          true,
	"site_id":            true,
	"user_id":            true,
	"account_type":       true,
	"bank_name":          true,
	"wallet_type":        true,
	"account_name":       true,
	"account_no":         true,
	"phone_country_code": true,
	"phone":              true,
	"is_default":         true,
	"status":             true,
}
