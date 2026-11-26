package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameDepositOrders Deposit orders
type GameDepositOrders struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID                 uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	OrderNo                string           `gorm:"column:order_no;type:varchar(64);not null" json:"orderNo"`
	ChannelOrderNo         string           `gorm:"column:channel_order_no;type:varchar(128);not null" json:"channelOrderNo"`
	UserID                 uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	Amount                 *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	FeeAmount              *decimal.Decimal `gorm:"column:fee_amount;type:decimal(18,2);default:0.00;not null" json:"feeAmount"`
	PaidAmount             *decimal.Decimal `gorm:"column:paid_amount;type:decimal(18,2);default:0.00;not null" json:"paidAmount"`
	Currency               string           `gorm:"column:currency;type:varchar(8);default:PKR;not null" json:"currency"`
	MethodCode             string           `gorm:"column:method_code;type:varchar(64);not null" json:"methodCode"`
	ChannelCode            string           `gorm:"column:channel_code;type:varchar(64);not null" json:"channelCode"`
	VoucherID              uint64           `gorm:"column:voucher_id;type:bigint(20) unsigned;default:0;not null" json:"voucherID"`
	VoucherBonusReward     *decimal.Decimal `gorm:"column:voucher_bonus_reward;type:decimal(18,2);default:0.00;not null" json:"voucherBonusReward"`
	VoucherCashReward      *decimal.Decimal `gorm:"column:voucher_cash_reward;type:decimal(18,2);default:0.00;not null" json:"voucherCashReward"`
	VoucherRewardAmount    *decimal.Decimal `gorm:"column:voucher_reward_amount;type:decimal(18,2);default:0.00;not null" json:"voucherRewardAmount"`
	WalletTransactionID    uint64           `gorm:"column:wallet_transaction_id;type:bigint unsigned;default:0;not null" json:"walletTransactionID"`
	Status                 int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"`
	ReviewerID             uint64           `gorm:"column:reviewer_id;type:bigint unsigned;default:0;not null" json:"reviewerID"`
	ReviewerUsername       string           `gorm:"column:reviewer_username;type:varchar(64);not null" json:"reviewerUsername"`
	ReviewReason           string           `gorm:"column:review_reason;type:varchar(500);not null" json:"reviewReason"`
	EvidenceReference      string           `gorm:"column:evidence_reference;type:varchar(255);not null" json:"evidenceReference"`
	ChannelResponseCode    string           `gorm:"column:channel_response_code;type:varchar(128);not null" json:"channelResponseCode"`
	ChannelResponseMessage string           `gorm:"column:channel_response_message;type:varchar(500);not null" json:"channelResponseMessage"`
	PaidAt                 *time.Time       `gorm:"column:paid_at;type:datetime(3)" json:"paidAt"`
	CallbackAt             *time.Time       `gorm:"column:callback_at;type:datetime(3)" json:"callbackAt"`
	CompletedAt            *time.Time       `gorm:"column:completed_at;type:datetime(3)" json:"completedAt"`
	CallbackCount          uint             `gorm:"column:callback_count;type:int unsigned;not null" json:"callbackCount"`
	ExpiredAt              *time.Time       `gorm:"column:expired_at;type:datetime(3)" json:"expiredAt"`
	FailReason             string           `gorm:"column:fail_reason;type:varchar(255);not null" json:"failReason"`
}

// GameDepositOrdersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameDepositOrdersColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":                  true,
	"order_no":                 true,
	"channel_order_no":         true,
	"user_id":                  true,
	"amount":                   true,
	"fee_amount":               true,
	"paid_amount":              true,
	"currency":                 true,
	"method_code":              true,
	"channel_code":             true,
	"voucher_id":               true,
	"voucher_bonus_reward":     true,
	"voucher_cash_reward":      true,
	"voucher_reward_amount":    true,
	"wallet_transaction_id":    true,
	"status":                   true,
	"reviewer_id":              true,
	"reviewer_username":        true,
	"review_reason":            true,
	"evidence_reference":       true,
	"channel_response_code":    true,
	"channel_response_message": true,
	"paid_at":                  true,
	"callback_at":              true,
	"completed_at":             true,
	"callback_count":           true,
	"expired_at":               true,
	"fail_reason":              true,
}
