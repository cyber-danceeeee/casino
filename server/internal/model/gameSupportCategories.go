package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSupportCategories FAQ categories
type GameSupportCategories struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	CategoryCode string `gorm:"column:category_code;type:varchar(64);not null" json:"categoryCode"`
	Name         string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Sort         int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status       int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameSupportCategoriesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSupportCategoriesColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"category_code": true,
	"name":          true,
	"sort":          true,
	"status":        true,
}
