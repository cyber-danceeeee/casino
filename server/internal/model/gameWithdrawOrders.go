package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
	"time"
)

// GameWithdrawOrders Withdraw orders
type GameWithdrawOrders struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID                uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID                  uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	OrderNo                 string           `gorm:"column:order_no;type:varchar(64);not null" json:"orderNo"`
	ChannelOrderNo          string           `gorm:"column:channel_order_no;type:varchar(128);not null" json:"channelOrderNo"`
	UserID                  uint64           `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	WithdrawAccountID       uint64           `gorm:"column:withdraw_account_id;type:bigint(20) unsigned;default:0;not null" json:"withdrawAccountID"`
	PaymentMethod           string           `gorm:"column:payment_method;type:varchar(64);not null" json:"paymentMethod"`
	Amount                  *decimal.Decimal `gorm:"column:amount;type:decimal(18,2);default:0.00;not null" json:"amount"`
	FeeAmount               *decimal.Decimal `gorm:"column:fee_amount;type:decimal(18,2);default:0.00;not null" json:"feeAmount"`
	ActualAmount            *decimal.Decimal `gorm:"column:actual_amount;type:decimal(18,2);default:0.00;not null" json:"actualAmount"`
	Currency                string           `gorm:"column:currency;type:varchar(8);default:PKR;not null" json:"currency"`
	Status                  int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"`            // 1 pending, 2 succeeded, 3 failed, 4 returned, 5 cancelled, 6 reviewing
	AuditStatus             int              `gorm:"column:audit_status;type:tinyint(4);default:1;not null" json:"auditStatus"` // 1 pending, 2 passed, 3 rejected
	RiskStatus              string           `gorm:"column:risk_status;type:varchar(32);not null" json:"riskStatus"`
	KYCStatus               string           `gorm:"column:kyc_status;type:varchar(32);not null" json:"kycStatus"`
	ReviewerID              uint64           `gorm:"column:reviewer_id;type:bigint unsigned;not null" json:"reviewerID"`
	ReviewerUsername        string           `gorm:"column:reviewer_username;type:varchar(64);not null" json:"reviewerUsername"`
	ReviewRemark            string           `gorm:"column:review_remark;type:varchar(255);not null" json:"reviewRemark"`
	FreezeRecordID          uint64           `gorm:"column:freeze_record_id;type:bigint unsigned;not null" json:"freezeRecordID"`
	FinalDebitTransactionID uint64           `gorm:"column:final_debit_transaction_id;type:bigint unsigned;not null" json:"finalDebitTransactionID"`
	RefundTransactionID     uint64           `gorm:"column:refund_transaction_id;type:bigint unsigned;not null" json:"refundTransactionID"`
	FailureReason           string           `gorm:"column:failure_reason;type:varchar(500);not null" json:"failureReason"`
	SubmittedAt             *time.Time       `gorm:"column:submitted_at;type:datetime(3)" json:"submittedAt"`
	ReviewedAt              *time.Time       `gorm:"column:reviewed_at;type:datetime(3)" json:"reviewedAt"`
	ApprovedAt              *time.Time       `gorm:"column:approved_at;type:datetime(3)" json:"approvedAt"`
	CompletedAt             *time.Time       `gorm:"column:completed_at;type:datetime(3)" json:"completedAt"`
	PayoutRetryCount        uint             `gorm:"column:payout_retry_count;type:int unsigned;not null" json:"payoutRetryCount"`
}

// GameWithdrawOrdersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameWithdrawOrdersColumnNames = map[string]bool{
	"id":                         true,
	"created_at":                 true,
	"updated_at":                 true,
	"deleted_at":                 true,
	"tenant_id":                  true,
	"site_id":                    true,
	"order_no":                   true,
	"channel_order_no":           true,
	"user_id":                    true,
	"withdraw_account_id":        true,
	"payment_method":             true,
	"amount":                     true,
	"fee_amount":                 true,
	"actual_amount":              true,
	"currency":                   true,
	"status":                     true,
	"audit_status":               true,
	"risk_status":                true,
	"kyc_status":                 true,
	"reviewer_id":                true,
	"reviewer_username":          true,
	"review_remark":              true,
	"freeze_record_id":           true,
	"final_debit_transaction_id": true,
	"refund_transaction_id":      true,
	"failure_reason":             true,
	"submitted_at":               true,
	"reviewed_at":                true,
	"approved_at":                true,
	"completed_at":               true,
	"payout_retry_count":         true,
}
