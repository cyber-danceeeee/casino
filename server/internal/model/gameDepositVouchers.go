package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameDepositVouchers Deposit voucher configuration
type GameDepositVouchers struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID           uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	VoucherCode      string           `gorm:"column:voucher_code;type:varchar(64);not null" json:"voucherCode"`
	Title            string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	TitleZh          string           `gorm:"column:title_zh;type:varchar(128);not null" json:"titleZh"`
	TitleEn          string           `gorm:"column:title_en;type:varchar(128);not null" json:"titleEn"`
	DescriptionZh    string           `gorm:"column:description_zh;type:text" json:"descriptionZh"`
	DescriptionEn    string           `gorm:"column:description_en;type:text" json:"descriptionEn"`
	VoucherType      string           `gorm:"column:voucher_type;type:varchar(32);not null" json:"voucherType"`
	IconURL          string           `gorm:"column:icon_url;type:varchar(500);not null" json:"iconURL"`
	MinAmount        *decimal.Decimal `gorm:"column:min_amount;type:decimal(18,2);default:0.00;not null" json:"minAmount"`
	MaxAmount        *decimal.Decimal `gorm:"column:max_amount;type:decimal(18,2);default:0.00;not null" json:"maxAmount"`
	MaxReward        *decimal.Decimal `gorm:"column:max_reward;type:decimal(18,2);default:0.00;not null" json:"maxReward"`
	BonusReward      *decimal.Decimal `gorm:"column:bonus_reward;type:decimal(18,2);default:0.00;not null" json:"bonusReward"`
	CashReward       *decimal.Decimal `gorm:"column:cash_reward;type:decimal(18,2);default:0.00;not null" json:"cashReward"`
	Sort             int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status           int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
	StartAt          *time.Time       `gorm:"column:start_at" json:"startTime"`
	EndAt            *time.Time       `gorm:"column:end_at" json:"endTime"`
	NewUsersOnly     bool             `gorm:"column:new_users_only;not null" json:"newUsersOnly"`
	FirstDepositOnly bool             `gorm:"column:first_deposit_only;not null" json:"firstDepositOnly"`
	DailyLimit       uint             `gorm:"column:daily_limit;not null" json:"dailyLimit"`
	WeeklyLimit      uint             `gorm:"column:weekly_limit;not null" json:"weeklyLimit"`
	PerUserLimit     uint             `gorm:"column:per_user_limit;not null" json:"perUserLimit"`
	Countries        string           `gorm:"column:countries;type:varchar(500);not null" json:"countries"`
	Currencies       string           `gorm:"column:currencies;type:varchar(128);not null" json:"currencies"`
	MethodCodes      string           `gorm:"column:method_codes;type:varchar(500);not null" json:"methodCodes"`
}

// GameDepositVouchersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameDepositVouchersColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":            true,
	"voucher_code":       true,
	"title":              true,
	"title_zh":           true,
	"title_en":           true,
	"description_zh":     true,
	"description_en":     true,
	"voucher_type":       true,
	"icon_url":           true,
	"min_amount":         true,
	"max_amount":         true,
	"max_reward":         true,
	"bonus_reward":       true,
	"cash_reward":        true,
	"sort":               true,
	"status":             true,
	"start_at":           true,
	"end_at":             true,
	"new_users_only":     true,
	"first_deposit_only": true,
	"daily_limit":        true,
	"weekly_limit":       true,
	"per_user_limit":     true,
	"countries":          true,
	"currencies":         true,
	"method_codes":       true,
}
