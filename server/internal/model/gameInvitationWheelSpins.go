package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameInvitationWheelSpins Invitation wheel spin records
type GameInvitationWheelSpins struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SpinNo       string           `gorm:"column:spin_no;type:varchar(64);not null" json:"spinNo"`
	UserID       uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	RoundID      uint64           `gorm:"column:round_id;type:bigint(20) unsigned;default:0;not null" json:"roundID"`
	PrizeID      uint64           `gorm:"column:prize_id;type:bigint(20) unsigned;default:0;not null" json:"prizeID"`
	RewardType   string           `gorm:"column:reward_type;type:varchar(32);not null" json:"rewardType"`
	RewardAmount *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 won, 3 failed
	SpunAt       *time.Time       `gorm:"column:spun_at;type:datetime(3)" json:"spunAt"`
}

// GameInvitationWheelSpinsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameInvitationWheelSpinsColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"spin_no":       true,
	"user_id":       true,
	"round_id":      true,
	"prize_id":      true,
	"reward_type":   true,
	"reward_amount": true,
	"status":        true,
	"spun_at":       true,
}
