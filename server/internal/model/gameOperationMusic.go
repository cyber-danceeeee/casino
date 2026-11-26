package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameOperationMusic stores platform-owned audio used by lobby and activity scenes.
type GameOperationMusic struct {
	sgorm.Model `gorm:"embedded"`

	SiteID        uint64     `gorm:"column:site_id;type:bigint unsigned;not null;default:0" json:"siteID"`
	Code          string     `gorm:"column:code;type:varchar(64);not null" json:"code"`
	NameZh        string     `gorm:"column:name_zh;type:varchar(128);not null" json:"nameZh"`
	NameEn        string     `gorm:"column:name_en;type:varchar(128);not null" json:"nameEn"`
	AudioURL      string     `gorm:"column:audio_url;type:varchar(500);not null" json:"audioUrl"`
	Scene         string     `gorm:"column:scene;type:varchar(32);not null;index:idx_game_operation_music_scene" json:"scene"`
	Loop          bool       `gorm:"column:loop_play;type:tinyint(1);not null;default:1" json:"loop"`
	VolumePercent uint       `gorm:"column:volume_percent;type:tinyint unsigned;not null;default:70" json:"volumePercent"`
	SortOrder     int        `gorm:"column:sort_order;type:int;not null;default:0" json:"sortOrder"`
	Enabled       bool       `gorm:"column:enabled;type:tinyint(1);not null;default:1;index:idx_game_operation_music_enabled" json:"enabled"`
	StartAt       *time.Time `gorm:"column:start_at;type:datetime(3)" json:"startTime"`
	EndAt         *time.Time `gorm:"column:end_at;type:datetime(3)" json:"endTime"`
}

func (GameOperationMusic) TableName() string { return "game_operation_music" }
