package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGameProviders Game providers
type GameGameProviders struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID            uint64 `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID              uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	ProviderCode        string `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	Name                string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	LogoURL             string `gorm:"column:logo_url;type:varchar(500);not null" json:"logoURL"`
	ApiBaseURL          string `gorm:"column:api_base_url;type:varchar(500);not null" json:"apiBaseURL"`
	Environment         string `gorm:"column:environment;type:varchar(24);not null" json:"environment"`
	APIStatus           string `gorm:"column:api_status;type:varchar(24);not null" json:"apiStatus"`
	Status              int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 maintenance, 3 disabled
	WalletMode          string `gorm:"column:wallet_mode;type:varchar(32);default:SEAMLESS;not null" json:"walletMode"`
	CallbackURL         string `gorm:"column:callback_url;type:varchar(500);not null" json:"callbackURL"`
	SupportedCountries  string `gorm:"column:supported_countries;type:varchar(500);not null" json:"supportedCountries"`
	SupportedCurrencies string `gorm:"column:supported_currencies;type:varchar(500);not null" json:"supportedCurrencies"`
	MaintenanceMessage  string `gorm:"column:maintenance_message;type:varchar(500);not null" json:"maintenanceMessage"`
	Sort                int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
}

// GameGameProvidersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGameProvidersColumnNames = map[string]bool{
	"id":                   true,
	"created_at":           true,
	"updated_at":           true,
	"deleted_at":           true,
	"tenant_id":            true,
	"site_id":              true,
	"provider_code":        true,
	"name":                 true,
	"logo_url":             true,
	"api_base_url":         true,
	"environment":          true,
	"api_status":           true,
	"status":               true,
	"wallet_mode":          true,
	"callback_url":         true,
	"supported_countries":  true,
	"supported_currencies": true,
	"maintenance_message":  true,
	"sort":                 true,
}
