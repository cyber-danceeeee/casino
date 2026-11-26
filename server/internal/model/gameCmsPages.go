package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameCmsPages Static pages such as about and agreement
type GameCmsPages struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID    uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID      uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	PageCode    string     `gorm:"column:page_code;type:varchar(64);not null" json:"pageCode"`
	Title       string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	ContentHTML string     `gorm:"column:content_html;type:mediumtext" json:"contentHTML"`
	VersionNo   uint       `gorm:"column:version_no;type:int(11) unsigned;default:1;not null" json:"versionNo"`
	Status      int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 published, 2 draft, 3 disabled
	EffectiveAt *time.Time `gorm:"column:effective_at;type:datetime(3)" json:"effectiveAt"`
}

// GameCmsPagesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameCmsPagesColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"tenant_id":    true,
	"site_id":      true,
	"page_code":    true,
	"title":        true,
	"content_html": true,
	"version_no":   true,
	"status":       true,
	"effective_at": true,
}
