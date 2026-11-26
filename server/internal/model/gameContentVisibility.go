package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

// GameContentVisibility stores site visibility choices for platform-owned operation defaults.
type GameContentVisibility struct {
	sgorm.Model `gorm:"embedded"`

	SiteID       uint64 `gorm:"column:site_id;type:bigint unsigned;not null;uniqueIndex:uk_content_visibility,priority:2" json:"siteID"`
	ResourceType string `gorm:"column:resource_type;type:varchar(32);not null;uniqueIndex:uk_content_visibility,priority:3" json:"resourceType"`
	ResourceID   uint64 `gorm:"column:resource_id;type:bigint unsigned;not null;uniqueIndex:uk_content_visibility,priority:4" json:"resourceID"`
	Enabled      bool   `gorm:"column:enabled;type:tinyint(1);not null;default:1" json:"enabled"`
	CreatedBy    uint64 `gorm:"column:created_by;type:bigint unsigned;not null;default:0" json:"createdBy"`
	UpdatedBy    uint64 `gorm:"column:updated_by;type:bigint unsigned;not null;default:0" json:"updatedBy"`
}

func (GameContentVisibility) TableName() string {
	return "game_content_visibility"
}
