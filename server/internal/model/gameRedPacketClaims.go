package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"time"
)

// GameRedPacketClaims Red packet rain claim records
type GameRedPacketClaims struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	ClaimNo      string           `gorm:"column:claim_no;type:varchar(64);not null" json:"claimNo"`
	UserID       uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	RoundID      uint64           `gorm:"column:round_id;type:bigint(20) unsigned;not null" json:"roundID"`
	RewardAmount *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 claimed, 3 failed
	ClaimedAt    *time.Time       `gorm:"column:claimed_at;type:datetime(3)" json:"claimedAt"`
}

// GameRedPacketClaimsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameRedPacketClaimsColumnNames = map[string]bool{
	"id":            true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
	"claim_no":      true,
	"user_id":       true,
	"round_id":      true,
	"reward_amount": true,
	"status":        true,
	"claimed_at":    true,
}


