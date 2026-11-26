package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameBetReports Bet report aggregates
type GameBetReports struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID         uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID         uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	ReportDate     *time.Time       `gorm:"column:report_date;type:date;not null" json:"reportDate"`
	GameType       string           `gorm:"column:game_type;type:varchar(32);not null" json:"gameType"`
	BetCount       uint             `gorm:"column:bet_count;type:int(11) unsigned;default:0;not null" json:"betCount"`
	BetAmount      *decimal.Decimal `gorm:"column:bet_amount;type:decimal(18,2);default:0.00;not null" json:"betAmount"`
	ValidBetAmount *decimal.Decimal `gorm:"column:valid_bet_amount;type:decimal(18,2);default:0.00;not null" json:"validBetAmount"`
	PayoutAmount   *decimal.Decimal `gorm:"column:payout_amount;type:decimal(18,2);default:0.00;not null" json:"payoutAmount"`
	ProfitAmount   *decimal.Decimal `gorm:"column:profit_amount;type:decimal(18,2);default:0.00;not null" json:"profitAmount"`
}

// GameBetReportsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameBetReportsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":          true,
	"user_id":          true,
	"report_date":      true,
	"game_type":        true,
	"bet_count":        true,
	"bet_amount":       true,
	"valid_bet_amount": true,
	"payout_amount":    true,
	"profit_amount":    true,
}
