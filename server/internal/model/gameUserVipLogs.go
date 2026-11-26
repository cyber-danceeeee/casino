package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameUserVipLogs Member VIP change logs
type GameUserVipLogs struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID         uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID         uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	OldLevelNo     uint             `gorm:"column:old_level_no;type:int(11) unsigned;default:0;not null" json:"oldLevelNo"`
	NewLevelNo     uint             `gorm:"column:new_level_no;type:int(11) unsigned;default:0;not null" json:"newLevelNo"`
	Reason         string           `gorm:"column:reason;type:varchar(64);not null" json:"reason"`
	DepositAmount  *decimal.Decimal `gorm:"column:deposit_amount;type:decimal(18,2);default:0.00;not null" json:"depositAmount"`
	ValidBetAmount *decimal.Decimal `gorm:"column:valid_bet_amount;type:decimal(18,2);default:0.00;not null" json:"validBetAmount"`
	RewardAmount   *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	ChangedAt      *time.Time       `gorm:"column:changed_at;type:datetime(3)" json:"changedAt"`
}

// GameUserVipLogsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserVipLogsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":          true,
	"user_id":          true,
	"old_level_no":     true,
	"new_level_no":     true,
	"reason":           true,
	"deposit_amount":   true,
	"valid_bet_amount": true,
	"reward_amount":    true,
	"changed_at":       true,
}
