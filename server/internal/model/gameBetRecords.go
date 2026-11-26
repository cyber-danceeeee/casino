package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameBetRecords Bet records
type GameBetRecords struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID         uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	BetNo          string           `gorm:"column:bet_no;type:varchar(64);not null" json:"betNo"`
	UserID         uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	GameID         uint64           `gorm:"column:game_id;type:bigint(20) unsigned;default:0;not null" json:"gameID"`
	ProviderID     uint64           `gorm:"column:provider_id;type:bigint(20) unsigned;default:0;not null" json:"providerID"`
	GameType       string           `gorm:"column:game_type;type:varchar(32);not null" json:"gameType"`
	BetAmount      *decimal.Decimal `gorm:"column:bet_amount;type:decimal(18,2);default:0.00;not null" json:"betAmount"`
	ValidBetAmount *decimal.Decimal `gorm:"column:valid_bet_amount;type:decimal(18,2);default:0.00;not null" json:"validBetAmount"`
	PayoutAmount   *decimal.Decimal `gorm:"column:payout_amount;type:decimal(18,2);default:0.00;not null" json:"payoutAmount"`
	ProfitAmount   *decimal.Decimal `gorm:"column:profit_amount;type:decimal(18,2);default:0.00;not null" json:"profitAmount"`
	Status         int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 not settled, 2 settled, 3 cancelled
	BetAt          *time.Time       `gorm:"column:bet_at;type:datetime(3);not null" json:"betAt"`
	SettledAt      *time.Time       `gorm:"column:settled_at;type:datetime(3)" json:"settledAt"`
}

// GameBetRecordsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameBetRecordsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":          true,
	"bet_no":           true,
	"user_id":          true,
	"game_id":          true,
	"provider_id":      true,
	"game_type":        true,
	"bet_amount":       true,
	"valid_bet_amount": true,
	"payout_amount":    true,
	"profit_amount":    true,
	"status":           true,
	"bet_at":           true,
	"settled_at":       true,
}
