package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameUserCheckins Daily check-in records
type GameUserCheckins struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID       uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	ActivityID   uint64           `gorm:"column:activity_id;type:bigint(20) unsigned;default:0;not null" json:"activityID"`
	CheckinDate  *time.Time       `gorm:"column:checkin_date;type:date;not null" json:"checkinDate"`
	DayNo        uint             `gorm:"column:day_no;type:int(11) unsigned;default:1;not null" json:"dayNo"`
	StreakCount  uint             `gorm:"column:streak_count;type:int(11) unsigned;default:1;not null" json:"streakCount"`
	RewardAmount *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	RewardWallet string           `gorm:"column:reward_wallet;type:varchar(32);default:bonus;not null" json:"rewardWallet"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 checked, 2 claimed, 3 expired
	ClaimedAt    *time.Time       `gorm:"column:claimed_at;type:datetime(3)" json:"claimedAt"`
}

// GameUserCheckinsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserCheckinsColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"user_id":       true,
	"activity_id":   true,
	"checkin_date":  true,
	"day_no":        true,
	"streak_count":  true,
	"reward_amount": true,
	"reward_wallet": true,
	"status":        true,
	"claimed_at":    true,
}
