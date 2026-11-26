package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

type GameUserTags struct {
	sgorm.Model `gorm:"embedded"`

	SiteID  uint64 `gorm:"column:site_id;type:bigint unsigned;not null;default:0" json:"siteID"`
	Name    string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	TagType string `gorm:"column:tag_type;type:varchar(16);not null" json:"tagType"`
}

func (GameUserTags) TableName() string { return "game_user_tags" }

type GameUserTagRelations struct {
	sgorm.Model `gorm:"embedded"`

	SiteID uint64 `gorm:"column:site_id;type:bigint unsigned;not null;default:0" json:"siteID"`
	UserID uint64 `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	TagID  uint64 `gorm:"column:tag_id;type:bigint unsigned;not null" json:"tagID"`
}

func (GameUserTagRelations) TableName() string { return "game_user_tag_relations" }
