package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGameCategoryRelations Game and category relations
type GameGameCategoryRelations struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID           uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	GameID           uint64 `gorm:"column:game_id;type:bigint(20) unsigned;not null;uniqueIndex:uk_game_game_category_relations_game_category" json:"gameID"`
	CategoryID       uint64 `gorm:"column:category_id;type:bigint(20) unsigned;not null;uniqueIndex:uk_game_game_category_relations_game_category" json:"categoryID"`
	CategoryPriority int    `gorm:"column:category_priority;default:0;not null" json:"categoryPriority"`
	Sort             int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
}

// GameGameCategoryRelationsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGameCategoryRelationsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":     true,
	"game_id":     true,
	"category_id": true,
	"sort":        true,
}
