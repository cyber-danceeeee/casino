package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameProviderCategoryMappings maps provider category codes into platform category codes.
type GameProviderCategoryMappings struct {
	sgorm.Model `gorm:"embedded"`

	SiteID               uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	ProviderCode         string `gorm:"column:provider_code;type:varchar(64);not null;uniqueIndex:uk_game_provider_category_mappings_provider_category" json:"providerCode"`
	ProviderCategoryCode string `gorm:"column:provider_category_code;type:varchar(128);not null;uniqueIndex:uk_game_provider_category_mappings_provider_category" json:"providerCategoryCode"`
	PlatformCategoryID   uint64 `gorm:"column:platform_category_id;type:bigint(20) unsigned;default:0;not null" json:"platformCategoryID"`
	Status               int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

func (GameProviderCategoryMappings) TableName() string {
	return "game_provider_category_mapping"
}

// GameProviderCategoryMappingsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameProviderCategoryMappingsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":                true,
	"provider_code":          true,
	"provider_category_code": true,
	"platform_category_id":   true,
	"status":                 true,
}
