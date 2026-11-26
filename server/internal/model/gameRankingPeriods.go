package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameRankingPeriods Ranking periods
type GameRankingPeriods struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	RankingType string     `gorm:"column:ranking_type;type:varchar(64);default:commission;not null" json:"rankingType"`
	Title       string     `gorm:"column:title;type:varchar(128);not null" json:"title"`
	RankingDate *time.Time `gorm:"column:ranking_date;type:date" json:"rankingDate"`
	PeriodStart *time.Time `gorm:"column:period_start;type:datetime(3);not null" json:"periodStart"`
	PeriodEnd   *time.Time `gorm:"column:period_end;type:datetime(3);not null" json:"periodEnd"`
	Status      int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 settled, 3 disabled
}

// GameRankingPeriodsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameRankingPeriodsColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"ranking_type": true,
	"title":        true,
	"ranking_date": true,
	"period_start": true,
	"period_end":   true,
	"status":       true,
}


