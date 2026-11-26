package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameRankingRecords Ranking records
type GameRankingRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	RankingPeriodID uint64           `gorm:"column:ranking_period_id;type:bigint(20) unsigned;not null" json:"rankingPeriodID"`
	UserID          uint64           `gorm:"column:user_id;type:bigint(20) unsigned;default:0;not null" json:"userID"`
	RankNo          uint             `gorm:"column:rank_no;type:int(11) unsigned;default:0;not null" json:"rankNo"`
	DisplayName     string           `gorm:"column:display_name;type:varchar(64);not null" json:"displayName"`
	AvatarURL       string           `gorm:"column:avatar_url;type:varchar(500);not null" json:"avatarURL"`
	TotalAmount     *decimal.Decimal `gorm:"column:total_amount;type:decimal(18,2);default:0.00;not null" json:"totalAmount"`
	RewardAmount    *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);default:0.00;not null" json:"rewardAmount"`
	Status          int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 pending, 2 rewarded, 3 invalid
}

// GameRankingRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameRankingRecordsColumnNames = map[string]bool{
	"id":                true,
	"created_at":        true,
	"updated_at":        true,
	"deleted_at":        true,
	"ranking_period_id": true,
	"user_id":           true,
	"rank_no":           true,
	"display_name":      true,
	"avatar_url":        true,
	"total_amount":      true,
	"reward_amount":     true,
	"status":            true,
}
