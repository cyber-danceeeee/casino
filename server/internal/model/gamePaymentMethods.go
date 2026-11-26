package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GamePaymentMethods Payment methods
type GamePaymentMethods struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID   uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID     uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	MethodCode string           `gorm:"column:method_code;type:varchar(64);not null" json:"methodCode"`
	Name       string           `gorm:"column:name;type:varchar(128);not null" json:"name"`
	MethodType string           `gorm:"column:method_type;type:varchar(32);not null" json:"methodType"` // deposit, withdraw, both
	IconURL    string           `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	MinAmount  *decimal.Decimal `gorm:"column:min_amount;type:decimal(18,2);default:0.00;not null" json:"minAmount"`
	MaxAmount  *decimal.Decimal `gorm:"column:max_amount;type:decimal(18,2);default:0.00;not null" json:"maxAmount"`
	Sort       int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status     int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GamePaymentMethodsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GamePaymentMethodsColumnNames = map[string]bool{
	"id":          true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
	"tenant_id":   true,
	"site_id":     true,
	"method_code": true,
	"name":        true,
	"method_type": true,
	"icon_url":    true,
	"min_amount":  true,
	"max_amount":  true,
	"sort":        true,
	"status":      true,
}
