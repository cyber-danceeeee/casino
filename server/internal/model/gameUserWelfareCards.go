package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserWelfareCards Member welfare cards
type GameUserWelfareCards struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID        uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	WelfareCardID uint64     `gorm:"column:welfare_card_id;type:bigint(20) unsigned;not null" json:"welfareCardID"`
	Status        int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 expired, 3 cancelled
	ActivatedAt   *time.Time `gorm:"column:activated_at;type:datetime(3)" json:"activatedAt"`
	ExpiredAt     *time.Time `gorm:"column:expired_at;type:datetime(3)" json:"expiredAt"`
}

// GameUserWelfareCardsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserWelfareCardsColumnNames = map[string]bool{
	"id":              true,
	"created_at":      true,
	"updated_at":      true,
	"deleted_at":      true,
	"user_id":         true,
	"welfare_card_id": true,
	"status":          true,
	"activated_at":    true,
	"expired_at":      true,
}
