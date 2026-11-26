package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

type GameDepositCallbackLogs struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64    `gorm:"column:site_id" json:"siteID"`
	DepositOrderID   uint64    `gorm:"column:deposit_order_id" json:"depositOrderID"`
	CallbackTime     time.Time `gorm:"column:callback_time" json:"callbackTime"`
	SourceIP         string    `gorm:"column:source_ip" json:"sourceIP"`
	SignatureValid   int       `gorm:"column:signature_valid" json:"signatureValid"`
	ChannelStatus    string    `gorm:"column:channel_status" json:"channelStatus"`
	ProcessingResult string    `gorm:"column:processing_result" json:"processingResult"`
	HTTPResponse     string    `gorm:"column:http_response" json:"httpResponse"`
	RetryCount       uint      `gorm:"column:retry_count" json:"retryCount"`
	RawPayload       string    `gorm:"column:raw_payload;type:json" json:"rawPayload"`
}

func (GameDepositCallbackLogs) TableName() string { return "game_deposit_callback_logs" }

type GameWithdrawalApprovals struct {
	sgorm.Model `gorm:"embedded"`

	SiteID            uint64 `gorm:"column:site_id" json:"siteID"`
	WithdrawalOrderID uint64 `gorm:"column:withdrawal_order_id" json:"withdrawalOrderID"`
	OperatorID        uint64 `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername  string `gorm:"column:operator_username" json:"operatorUsername"`
	Stage             string `gorm:"column:stage" json:"stage"`
	Decision          string `gorm:"column:decision" json:"decision"`
	BeforeStatus      string `gorm:"column:before_status" json:"beforeStatus"`
	AfterStatus       string `gorm:"column:after_status" json:"afterStatus"`
	Reason            string `gorm:"column:reason" json:"reason"`
	InternalRemark    string `gorm:"column:internal_remark" json:"internalRemark"`
}

func (GameWithdrawalApprovals) TableName() string { return "game_withdrawal_approvals" }

type GameWithdrawalPayoutLogs struct {
	sgorm.Model `gorm:"embedded"`

	SiteID            uint64    `gorm:"column:site_id" json:"siteID"`
	WithdrawalOrderID uint64    `gorm:"column:withdrawal_order_id" json:"withdrawalOrderID"`
	ChannelCode       string    `gorm:"column:channel_code" json:"channelCode"`
	RequestResult     string    `gorm:"column:request_result" json:"requestResult"`
	ChannelOrderNo    string    `gorm:"column:channel_order_no" json:"channelOrderNo"`
	QueryResult       string    `gorm:"column:query_result" json:"queryResult"`
	CallbackResult    string    `gorm:"column:callback_result" json:"callbackResult"`
	RetryCount        uint      `gorm:"column:retry_count" json:"retryCount"`
	FailureReason     string    `gorm:"column:failure_reason" json:"failureReason"`
	RawMetadata       string    `gorm:"column:raw_metadata;type:json" json:"rawMetadata"`
	RequestedAt       time.Time `gorm:"column:requested_at" json:"requestedAt"`
}

func (GameWithdrawalPayoutLogs) TableName() string { return "game_withdrawal_payout_logs" }

type GameReconciliationRecords struct {
	sgorm.Model `gorm:"embedded"`

	SiteID             uint64           `gorm:"column:site_id" json:"siteID"`
	ReconciliationNo   string           `gorm:"column:reconciliation_no" json:"reconciliationNo"`
	BusinessDate       time.Time        `gorm:"column:business_date;type:date" json:"businessDate"`
	BusinessType       string           `gorm:"column:business_type" json:"businessType"`
	ChannelCode        string           `gorm:"column:channel_code" json:"channelCode"`
	Currency           string           `gorm:"column:currency" json:"currency"`
	ReconciliationMode string           `gorm:"column:reconciliation_mode" json:"reconciliationMode"`
	PlatformOrderCount uint             `gorm:"column:platform_order_count" json:"platformOrderCount"`
	PlatformAmount     *decimal.Decimal `gorm:"column:platform_amount" json:"platformAmount"`
	ChannelOrderCount  uint             `gorm:"column:channel_order_count" json:"channelOrderCount"`
	ChannelAmount      *decimal.Decimal `gorm:"column:channel_amount" json:"channelAmount"`
	DifferenceCount    uint             `gorm:"column:difference_count" json:"differenceCount"`
	DifferenceAmount   *decimal.Decimal `gorm:"column:difference_amount" json:"differenceAmount"`
	Status             string           `gorm:"column:status" json:"status"`
	GeneratedByID      uint64           `gorm:"column:generated_by_id" json:"generatedByID"`
	GeneratedBy        string           `gorm:"column:generated_by" json:"generatedBy"`
	GeneratedAt        time.Time        `gorm:"column:generated_at" json:"generatedAt"`
	CompletedAt        *time.Time       `gorm:"column:completed_at" json:"completedAt"`
}

func (GameReconciliationRecords) TableName() string { return "game_reconciliation_records" }

type GameReconciliationDifferences struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64           `gorm:"column:site_id" json:"siteID"`
	ReconciliationID uint64           `gorm:"column:reconciliation_id" json:"reconciliationID"`
	PlatformOrderID  string           `gorm:"column:platform_order_id" json:"platformOrderID"`
	ChannelOrderID   string           `gorm:"column:channel_order_id" json:"channelOrderID"`
	PlatformStatus   string           `gorm:"column:platform_status" json:"platformStatus"`
	ChannelStatus    string           `gorm:"column:channel_status" json:"channelStatus"`
	PlatformAmount   *decimal.Decimal `gorm:"column:platform_amount" json:"platformAmount"`
	ChannelAmount    *decimal.Decimal `gorm:"column:channel_amount" json:"channelAmount"`
	DifferenceType   string           `gorm:"column:difference_type" json:"differenceType"`
	DifferenceAmount *decimal.Decimal `gorm:"column:difference_amount" json:"differenceAmount"`
	ResolutionStatus string           `gorm:"column:resolution_status" json:"resolutionStatus"`
	ResolutionReason string           `gorm:"column:resolution_reason" json:"resolutionReason"`
	ResolvedByID     uint64           `gorm:"column:resolved_by_id" json:"resolvedByID"`
	ResolvedBy       string           `gorm:"column:resolved_by" json:"resolvedBy"`
	DetectedAt       time.Time        `gorm:"column:detected_at" json:"detectedAt"`
	ResolvedAt       *time.Time       `gorm:"column:resolved_at" json:"resolvedAt"`
}

func (GameReconciliationDifferences) TableName() string {
	return "game_reconciliation_differences"
}

type GameReconciliationLogs struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64 `gorm:"column:site_id" json:"siteID"`
	ReconciliationID uint64 `gorm:"column:reconciliation_id" json:"reconciliationID"`
	Level            string `gorm:"column:level" json:"level"`
	Message          string `gorm:"column:message" json:"message"`
}

func (GameReconciliationLogs) TableName() string { return "game_reconciliation_logs" }

type GameReconciliationResolutions struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64 `gorm:"column:site_id" json:"siteID"`
	ReconciliationID uint64 `gorm:"column:reconciliation_id" json:"reconciliationID"`
	DifferenceID     uint64 `gorm:"column:difference_id" json:"differenceID"`
	OperatorID       uint64 `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username" json:"operatorUsername"`
	Action           string `gorm:"column:action" json:"action"`
	Reason           string `gorm:"column:reason" json:"reason"`
}

func (GameReconciliationResolutions) TableName() string {
	return "game_reconciliation_resolutions"
}
