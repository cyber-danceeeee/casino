package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGameCategories Game categories
type GameGameCategories struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID    uint64 `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null;uniqueIndex:uk_game_game_categories_scope_key,priority:1" json:"tenantID"`
	SiteID      uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null;uniqueIndex:uk_game_game_categories_scope_key,priority:2" json:"siteID"`
	CategoryKey string `gorm:"column:category_key;type:varchar(32);not null;uniqueIndex:uk_game_game_categories_scope_key,priority:3" json:"categoryKey"`
	Label       string `gorm:"column:label;type:varchar(64);not null" json:"label"`
	NameZh      string `gorm:"column:name_zh;type:varchar(64);not null" json:"nameZh"`
	NameEn      string `gorm:"column:name_en;type:varchar(64);not null" json:"nameEn"`
	ShortLabel  string `gorm:"column:short_label;type:varchar(16);not null" json:"shortLabel"`
	IconURL     string `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	BannerURL   string `gorm:"column:banner_url;type:varchar(500);not null" json:"bannerURL"`
	Sort        int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status      int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameGameCategoriesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGameCategoriesColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"tenant_id":    true,
	"site_id":      true,
	"category_key": true,
	"label":        true,
	"name_zh":      true,
	"name_en":      true,
	"short_label":  true,
	"icon_url":     true,
	"banner_url":   true,
	"sort":         true,
	"status":       true,
}
