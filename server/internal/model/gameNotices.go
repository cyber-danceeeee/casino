package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameNotices Notices and messages
type GameNotices struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID    uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID      uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	NoticeCode  string     `gorm:"column:notice_code;type:varchar(64);not null" json:"noticeCode"`
	NoticeType  string     `gorm:"column:notice_type;type:varchar(32);default:system;not null" json:"noticeType"`
	Title       string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	ContentHTML string     `gorm:"column:content_html;type:mediumtext" json:"contentHTML"`
	ImageURL    string     `gorm:"column:image_url;type:varchar(500);not null" json:"imageURL"`
	Sort        int        `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status      int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
	StartAt     *time.Time `gorm:"column:start_at;type:datetime(3)" json:"startAt"`
	EndAt       *time.Time `gorm:"column:end_at;type:datetime(3)" json:"endAt"`
}

// GameNoticesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameNoticesColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"tenant_id":    true,
	"site_id":      true,
	"notice_code":  true,
	"notice_type":  true,
	"title":        true,
	"content_html": true,
	"image_url":    true,
	"sort":         true,
	"status":       true,
	"start_at":     true,
	"end_at":       true,
}
