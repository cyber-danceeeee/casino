package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameInvitationWheelRounds Invitation wheel rounds
type GameInvitationWheelRounds struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	RoundNo       string           `gorm:"column:round_no;type:varchar(64);not null" json:"roundNo"`
	Title         string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	StartAt       *time.Time       `gorm:"column:start_at;type:datetime(3);not null" json:"startAt"`
	EndAt         *time.Time       `gorm:"column:end_at;type:datetime(3);not null" json:"endAt"`
	TargetAmount  *decimal.Decimal `gorm:"column:target_amount;type:decimal(18,2);default:0.00;not null" json:"targetAmount"`
	CurrentAmount *decimal.Decimal `gorm:"column:current_amount;type:decimal(18,2);default:0.00;not null" json:"currentAmount"`
	Status        int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 ended, 3 settled
}

// GameInvitationWheelRoundsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameInvitationWheelRoundsColumnNames = map[string]bool{
	"id":             true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
	"round_no":       true,
	"title":          true,
	"start_at":       true,
	"end_at":         true,
	"target_amount":  true,
	"current_amount": true,
	"status":         true,
}
