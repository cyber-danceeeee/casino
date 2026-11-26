package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameActivityAssets Activity visual assets
type GameActivityAssets struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	ActivityID uint64 `gorm:"column:activity_id;type:bigint(20) unsigned;not null" json:"activityID"`
	AssetKey   string `gorm:"column:asset_key;type:varchar(64);not null" json:"assetKey"`
	AssetURL   string `gorm:"column:asset_url;type:varchar(500);not null" json:"assetURL"`
	Sort       int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
}

// GameActivityAssetsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameActivityAssetsColumnNames = map[string]bool{
	"id":          true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
	"activity_id": true,
	"asset_key":   true,
	"asset_url":   true,
	"sort":        true,
}


