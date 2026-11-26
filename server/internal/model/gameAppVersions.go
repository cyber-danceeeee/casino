package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameAppVersions App download versions
type GameAppVersions struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	Platform    string `gorm:"column:platform;type:varchar(32);default:android;not null" json:"platform"`
	VersionName string `gorm:"column:version_name;type:varchar(32);not null" json:"versionName"`
	VersionCode uint   `gorm:"column:version_code;type:int(11) unsigned;default:0;not null" json:"versionCode"`
	DownloadURL string `gorm:"column:download_url;type:varchar(500);not null" json:"downloadURL"`
	ForceUpdate int    `gorm:"column:force_update;type:tinyint(4);default:0;not null" json:"forceUpdate"`
	ReleaseNote string `gorm:"column:release_note;type:varchar(500);not null" json:"releaseNote"`
	Status      int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameAppVersionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameAppVersionsColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"platform":     true,
	"version_name": true,
	"version_code": true,
	"download_url": true,
	"force_update": true,
	"release_note": true,
	"status":       true,
}
