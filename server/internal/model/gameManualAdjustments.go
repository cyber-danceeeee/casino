package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

type GameManualAdjustments struct {
	sgorm.Model         `gorm:"embedded"`
	TenantID            uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID              uint64           `gorm:"column:site_id" json:"siteID"`
	AdjustmentNo        string           `gorm:"column:adjustment_no" json:"adjustmentNo"`
	IdempotencyKey      string           `gorm:"column:idempotency_key" json:"idempotencyKey"`
	UserID              uint64           `gorm:"column:user_id" json:"userID"`
	WalletID            uint64           `gorm:"column:wallet_id" json:"walletID"`
	AdjustmentType      string           `gorm:"column:adjustment_type" json:"adjustmentType"`
	Amount              *decimal.Decimal `gorm:"column:amount" json:"amount"`
	Currency            string           `gorm:"column:currency" json:"currency"`
	ReasonCode          string           `gorm:"column:reason_code" json:"reasonCode"`
	DetailedReason      string           `gorm:"column:detailed_reason" json:"detailedReason"`
	SupportingReference string           `gorm:"column:supporting_reference" json:"supportingReference"`
	InternalRemark      string           `gorm:"column:internal_remark" json:"internalRemark"`
	Status              string           `gorm:"column:status" json:"status"`
	RequestedBalance    *decimal.Decimal `gorm:"column:requested_balance" json:"requestedBalance"`
	BalanceBefore       *decimal.Decimal `gorm:"column:balance_before" json:"balanceBefore"`
	BalanceAfter        *decimal.Decimal `gorm:"column:balance_after" json:"balanceAfter"`
	ApplicantID         uint64           `gorm:"column:applicant_id" json:"applicantID"`
	ApplicantUsername   string           `gorm:"column:applicant_username" json:"applicantUsername"`
	ApproverID          uint64           `gorm:"column:approver_id" json:"approverID"`
	ApproverUsername    string           `gorm:"column:approver_username" json:"approverUsername"`
	ApprovalReason      string           `gorm:"column:approval_reason" json:"approvalReason"`
	WalletTransactionID uint64           `gorm:"column:wallet_transaction_id" json:"walletTransactionID"`
	FailureReason       string           `gorm:"column:failure_reason" json:"failureReason"`
	ApprovedAt          *time.Time       `gorm:"column:approved_at" json:"approvedAt"`
	CompletedAt         *time.Time       `gorm:"column:completed_at" json:"completedAt"`
}

func (GameManualAdjustments) TableName() string { return "game_manual_adjustments" }

type GameManualAdjustmentApprovals struct {
	sgorm.Model      `gorm:"embedded"`
	TenantID         uint64 `gorm:"column:tenant_id" json:"tenantID"`
	SiteID           uint64 `gorm:"column:site_id" json:"siteID"`
	AdjustmentID     uint64 `gorm:"column:adjustment_id" json:"adjustmentID"`
	OperatorID       uint64 `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username" json:"operatorUsername"`
	Decision         string `gorm:"column:decision" json:"decision"`
	BeforeStatus     string `gorm:"column:before_status" json:"beforeStatus"`
	AfterStatus      string `gorm:"column:after_status" json:"afterStatus"`
	Reason           string `gorm:"column:reason" json:"reason"`
}

func (GameManualAdjustmentApprovals) TableName() string {
	return "game_manual_adjustment_approvals"
}
