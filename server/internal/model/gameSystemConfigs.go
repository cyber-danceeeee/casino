package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSystemConfigs System configuration
type GameSystemConfigs struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID    uint64 `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID      uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	ConfigKey   string `gorm:"column:config_key;type:varchar(128);not null" json:"configKey"`
	ConfigValue string `gorm:"column:config_value;type:mediumtext" json:"configValue"`
	Description string `gorm:"column:description;type:varchar(255);not null" json:"description"`
	Status      int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameSystemConfigsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSystemConfigsColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"tenant_id":    true,
	"site_id":      true,
	"config_key":   true,
	"config_value": true,
	"description":  true,
	"status":       true,
}
