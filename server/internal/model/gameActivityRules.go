package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameActivityRules Activity rule texts
type GameActivityRules struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	ActivityID  uint64 `gorm:"column:activity_id;type:bigint(20) unsigned;not null" json:"activityID"`
	Title       string `gorm:"column:title;type:varchar(128);not null" json:"title"`
	ContentHTML string `gorm:"column:content_html;type:mediumtext" json:"contentHTML"`
	Sort        int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status      int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameActivityRulesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameActivityRulesColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"activity_id":  true,
	"title":        true,
	"content_html": true,
	"sort":         true,
	"status":       true,
}
