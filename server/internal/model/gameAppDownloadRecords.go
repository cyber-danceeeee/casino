package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameAppDownloadRecords App download reward records
type GameAppDownloadRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID       uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	DeviceID     string           `gorm:"column:device_id;type:varchar(128);not null" json:"deviceID"`
	AppVersionID uint64           `gorm:"column:app_version_id;type:bigint(20) unsigned;default:0;not null" json:"appVersionID"`
	RewardAmount *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	RewardWallet string           `gorm:"column:reward_wallet;type:varchar(32);default:bonus;not null" json:"rewardWallet"`
	Status       int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 downloaded, 2 rewarded, 3 rejected
	RewardedAt   *time.Time       `gorm:"column:rewarded_at;type:datetime(3)" json:"rewardedAt"`
}

// GameAppDownloadRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameAppDownloadRecordsColumnNames = map[string]bool{
	"id":             true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
	"user_id":        true,
	"device_id":      true,
	"app_version_id": true,
	"reward_amount":  true,
	"reward_wallet":  true,
	"status":         true,
	"rewarded_at":    true,
}
