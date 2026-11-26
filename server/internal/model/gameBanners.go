package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameBanners Home banners and download bars
type GameBanners struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID     uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID       uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	BannerCode   string     `gorm:"column:banner_code;type:varchar(64);not null" json:"bannerCode"`
	Position     string     `gorm:"column:position;type:varchar(64);default:home;not null" json:"position"`
	Title        string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	ImageURL     string     `gorm:"column:image_url;type:varchar(500);not null" json:"imageURL"`
	LinkType     string     `gorm:"column:link_type;type:varchar(32);default:none;not null" json:"linkType"` // none, route, activity, external
	LinkURL      string     `gorm:"column:link_url;type:varchar(500);not null" json:"linkURL"`
	ActivitySlug string     `gorm:"column:activity_slug;type:varchar(128);not null" json:"activitySlug"`
	Sort         int        `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status       int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
	StartAt      *time.Time `gorm:"column:start_at;type:datetime(3)" json:"startAt"`
	EndAt        *time.Time `gorm:"column:end_at;type:datetime(3)" json:"endAt"`
}

// GameBannersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameBannersColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"tenant_id":     true,
	"site_id":       true,
	"banner_code":   true,
	"position":      true,
	"title":         true,
	"image_url":     true,
	"link_type":     true,
	"link_url":      true,
	"activity_slug": true,
	"sort":          true,
	"status":        true,
	"start_at":      true,
	"end_at":        true,
}
