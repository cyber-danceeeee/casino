package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSocialLinks H5 share platform links and copy
type GameSocialLinks struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID           uint64 `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	PlatformCode     string `gorm:"column:platform_code;type:varchar(64);not null" json:"platformCode"`
	Label            string `gorm:"column:label;type:varchar(64);not null" json:"label"`
	IconURL          string `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	ShareURL         string `gorm:"column:share_url;type:varchar(500);not null" json:"shareURL"`
	ShareTitle       string `gorm:"column:share_title;type:varchar(255);not null" json:"shareTitle"`
	ShareText        string `gorm:"column:share_text;type:varchar(500);not null" json:"shareText"`
	ShareDescription string `gorm:"column:share_description;type:varchar(500);not null" json:"shareDescription"`
	Sort             int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status           int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameSocialLinksColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSocialLinksColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":           true,
	"platform_code":     true,
	"label":             true,
	"icon_url":          true,
	"share_url":         true,
	"share_title":       true,
	"share_text":        true,
	"share_description": true,
	"sort":              true,
	"status":            true,
}
