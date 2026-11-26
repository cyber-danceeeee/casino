package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameRebateJackpots Home rebate jackpot display
type GameRebateJackpots struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	PeriodType    string           `gorm:"column:period_type;type:varchar(32);default:weekly;not null" json:"periodType"`
	PeriodStart   *time.Time       `gorm:"column:period_start;type:datetime(3);not null" json:"periodStart"`
	PeriodEnd     *time.Time       `gorm:"column:period_end;type:datetime(3);not null" json:"periodEnd"`
	JackpotAmount *decimal.Decimal `gorm:"column:jackpot_amount;type:decimal(18,2);default:0.00;not null" json:"jackpotAmount"`
	DigitText     string           `gorm:"column:digit_text;type:varchar(32);not null" json:"digitText"`
	Status        int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 settled, 3 disabled
}

// GameRebateJackpotsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameRebateJackpotsColumnNames = map[string]bool{
	"id":             true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
	"period_type":    true,
	"period_start":   true,
	"period_end":     true,
	"jackpot_amount": true,
	"digit_text":     true,
	"status":         true,
}
