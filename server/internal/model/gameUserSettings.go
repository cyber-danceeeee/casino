package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserSettings Member settings
type GameUserSettings struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID       uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID       uint64 `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	Language     string `gorm:"column:language;type:varchar(16);default:en;not null" json:"language"`
	MusicEnabled int    `gorm:"column:music_enabled;type:tinyint(4);default:1;not null" json:"musicEnabled"`
	PushEnabled  int    `gorm:"column:push_enabled;type:tinyint(4);default:1;not null" json:"pushEnabled"`
}

// GameUserSettingsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserSettingsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":       true,
	"user_id":       true,
	"language":      true,
	"music_enabled": true,
	"push_enabled":  true,
}
