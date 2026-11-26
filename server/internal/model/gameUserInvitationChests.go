package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserInvitationChests Member invitation chests
type GameUserInvitationChests struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID          uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	ChestConfigID   uint64     `gorm:"column:chest_config_id;type:bigint(20) unsigned;not null" json:"chestConfigID"`
	ProgressInvites uint       `gorm:"column:progress_invites;type:int(11) unsigned;default:0;not null" json:"progressInvites"`
	Status          int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 locked, 2 available, 3 opened, 4 expired
	OpenedAt        *time.Time `gorm:"column:opened_at;type:datetime(3)" json:"openedAt"`
}

// GameUserInvitationChestsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserInvitationChestsColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"user_id":          true,
	"chest_config_id":  true,
	"progress_invites": true,
	"status":           true,
	"opened_at":        true,
}
