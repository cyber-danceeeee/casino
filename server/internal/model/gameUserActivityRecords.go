package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameUserActivityRecords Member activity claims
type GameUserActivityRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID      uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	ActivityID  uint64           `gorm:"column:activity_id;type:bigint(20) unsigned;not null" json:"activityID"`
	ActionType  string           `gorm:"column:action_type;type:varchar(64);default:claim;not null" json:"actionType"`
	OrderNo     string           `gorm:"column:order_no;type:varchar(64);not null" json:"orderNo"`
	RewardCash  *decimal.Decimal `gorm:"column:reward_cash;type:decimal(18,2);default:0.00;not null" json:"rewardCash"`
	RewardBonus *decimal.Decimal `gorm:"column:reward_bonus;type:decimal(18,2);default:0.00;not null" json:"rewardBonus"`
	Status      int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 claimed, 3 rejected, 4 expired
	ClaimedAt   *time.Time       `gorm:"column:claimed_at;type:datetime(3)" json:"claimedAt"`
	Remark      string           `gorm:"column:remark;type:varchar(255);not null" json:"remark"`
}

// GameUserActivityRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserActivityRecordsColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"user_id":      true,
	"activity_id":  true,
	"action_type":  true,
	"order_no":     true,
	"reward_cash":  true,
	"reward_bonus": true,
	"status":       true,
	"claimed_at":   true,
	"remark":       true,
}
