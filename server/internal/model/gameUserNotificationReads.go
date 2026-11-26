package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserNotificationReads stores H5 notification read state per user.
type GameUserNotificationReads struct {
	sgorm.Model `gorm:"embedded"`

	UserID         uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	NotificationID uint64     `gorm:"column:notification_id;type:bigint(20) unsigned;not null" json:"notificationID"`
	ReadAt         *time.Time `gorm:"column:read_at;type:datetime(3)" json:"readAt"`
}

func (GameUserNotificationReads) TableName() string { return "game_user_notification_reads" }
