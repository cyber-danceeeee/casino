package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameHomeFeatureCards Home feature cards
type GameHomeFeatureCards struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID        uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID          uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	FeatureCode     string     `gorm:"column:feature_code;type:varchar(64);not null" json:"featureCode"`
	Title           string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	ImageURL        string     `gorm:"column:image_url;type:varchar(500);not null" json:"imageURL"`
	BackgroundStyle string     `gorm:"column:background_style;type:varchar(500);not null" json:"backgroundStyle"`
	LinkType        string     `gorm:"column:link_type;type:varchar(32);default:activity;not null" json:"linkType"`
	LinkURL         string     `gorm:"column:link_url;type:varchar(500);not null" json:"linkURL"`
	CountdownEndAt  *time.Time `gorm:"column:countdown_end_at;type:datetime(3)" json:"countdownEndAt"`
	Sort            int        `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status          int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameHomeFeatureCardsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameHomeFeatureCardsColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"tenant_id":        true,
	"site_id":          true,
	"feature_code":     true,
	"title":            true,
	"image_url":        true,
	"background_style": true,
	"link_type":        true,
	"link_url":         true,
	"countdown_end_at": true,
	"sort":             true,
	"status":           true,
}
