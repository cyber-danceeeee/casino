package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"time"
)

// GameRedPacketRounds Red packet rain rounds
type GameRedPacketRounds struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	RoundNo     string           `gorm:"column:round_no;type:varchar(64);not null" json:"roundNo"`
	ActivityID  uint64           `gorm:"column:activity_id;type:bigint(20) unsigned;default:0;not null" json:"activityID"`
	Title       string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	StartAt     *time.Time       `gorm:"column:start_at;type:datetime(3);not null" json:"startAt"`
	EndAt       *time.Time       `gorm:"column:end_at;type:datetime(3);not null" json:"endAt"`
	TotalAmount *decimal.Decimal `gorm:"column:total_amount;type:decimal(18,2);default:0.00;not null" json:"totalAmount"`
	PacketCount uint             `gorm:"column:packet_count;type:int(11) unsigned;default:0;not null" json:"packetCount"`
	Status      int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 active, 3 ended
}

// GameRedPacketRoundsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameRedPacketRoundsColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"round_no":     true,
	"activity_id":  true,
	"title":        true,
	"start_at":     true,
	"end_at":       true,
	"total_amount": true,
	"packet_count": true,
	"status":       true,
}


