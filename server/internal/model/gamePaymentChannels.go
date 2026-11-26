package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GamePaymentChannels Payment channels
type GamePaymentChannels struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID       uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	MethodID     uint64           `gorm:"column:method_id;type:bigint(20) unsigned;not null" json:"methodID"`
	ChannelCode  string           `gorm:"column:channel_code;type:varchar(64);not null" json:"channelCode"`
	Name         string           `gorm:"column:name;type:varchar(128);not null" json:"name"`
	CountryCode  string           `gorm:"column:country_code;type:varchar(8);default:'';not null" json:"countryCode"`
	CurrencyCode string           `gorm:"column:currency_code;type:varchar(8);default:'';not null" json:"currencyCode"`
	FeeRate      *decimal.Decimal `gorm:"column:fee_rate;type:decimal(8,4);default:0.0000;not null" json:"feeRate"`
	MinAmount    *decimal.Decimal `gorm:"column:min_amount;type:decimal(18,2);default:0.00;not null" json:"minAmount"`
	MaxAmount    *decimal.Decimal `gorm:"column:max_amount;type:decimal(18,2);default:0.00;not null" json:"maxAmount"`
	Sort         int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GamePaymentChannelsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GamePaymentChannelsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":       true,
	"method_id":     true,
	"channel_code":  true,
	"name":          true,
	"country_code":  true,
	"currency_code": true,
	"fee_rate":      true,
	"min_amount":    true,
	"max_amount":    true,
	"sort":          true,
	"status":        true,
}
