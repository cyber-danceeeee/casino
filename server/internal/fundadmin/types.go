package fundadmin

import (
	"github.com/shopspring/decimal"

	"casino/internal/content"
)

type Actor struct {
	ID            uint64
	Username      string
	Permissions   []string
	Scope         content.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type Page[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type ListQuery struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    content.Scope
}

type WalletRow struct {
	ID                  uint64 `json:"id" gorm:"column:id"`
	UserID              uint64 `json:"userID" gorm:"column:user_id"`
	UserNo              string `json:"userNo" gorm:"column:user_no"`
	Username            string `json:"username" gorm:"column:username"`
	PhoneCountryCode    string `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone               string `json:"phone" gorm:"column:phone"`
	CountryCode         string `json:"countryCode" gorm:"column:country_code"`
	VIPLevel            uint   `json:"vipLevel" gorm:"column:vip_level"`
	Currency            string `json:"currency" gorm:"column:currency"`
	CashBalance         string `json:"availableBalance" gorm:"column:cash_balance"`
	FrozenBalance       string `json:"frozenBalance" gorm:"column:frozen_balance"`
	BonusBalance        string `json:"bonusBalance" gorm:"column:bonus_balance"`
	WithdrawableBalance string `json:"withdrawableBalance" gorm:"column:withdrawable_balance"`
	Status              string `json:"status" gorm:"-"`
	StatusValue         int    `json:"-" gorm:"column:status"`
	AccountStatus       string `json:"accountStatus" gorm:"-"`
	AccountStatusValue  int    `json:"-" gorm:"column:account_status"`
	KYCStatus           string `json:"kycStatus" gorm:"-"`
	RiskStatus          string `json:"riskStatus" gorm:"column:risk_status"`
	LastTransactionAt   string `json:"lastTransactionAt" gorm:"column:last_transaction_at"`
	CreatedAt           string `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt           string `json:"updatedAt" gorm:"column:updated_at"`
}

type WalletDetail struct {
	WalletRow
	PendingWithdrawalAmount string `json:"pendingWithdrawalAmount"`
	PendingDepositAmount    string `json:"pendingDepositAmount"`
	TotalBalance            string `json:"totalBalance"`
}

type TransactionRow struct {
	ID                  uint64 `json:"id" gorm:"column:id"`
	TxnNo               string `json:"transactionID" gorm:"column:txn_no"`
	ExternalTxnNo       string `json:"externalTransactionID" gorm:"column:external_txn_no"`
	UserID              uint64 `json:"userID" gorm:"column:user_id"`
	UserNo              string `json:"userNo" gorm:"column:user_no"`
	Username            string `json:"username" gorm:"column:username"`
	PhoneCountryCode    string `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone               string `json:"phone" gorm:"column:phone"`
	WalletID            uint64 `json:"walletID" gorm:"column:wallet_id"`
	TransactionType     string `json:"transactionType" gorm:"column:category"`
	Direction           string `json:"direction" gorm:"column:direction"`
	Amount              string `json:"amount" gorm:"column:amount"`
	Currency            string `json:"currency" gorm:"column:currency"`
	BalanceBefore       string `json:"balanceBefore" gorm:"column:balance_before"`
	BalanceAfter        string `json:"balanceAfter" gorm:"column:balance_after"`
	FrozenBalanceBefore string `json:"frozenBalanceBefore" gorm:"column:frozen_balance_before"`
	FrozenBalanceAfter  string `json:"frozenBalanceAfter" gorm:"column:frozen_balance_after"`
	ReferenceType       string `json:"referenceType" gorm:"column:source_type"`
	ReferenceID         uint64 `json:"referenceID" gorm:"column:source_id"`
	ReferenceNo         string `json:"referenceNo" gorm:"column:order_no"`
	ProviderCode        string `json:"providerCode" gorm:"column:provider_code"`
	Status              string `json:"status" gorm:"-"`
	StatusValue         int    `json:"-" gorm:"column:status"`
	Operator            string `json:"operator" gorm:"column:operator_username"`
	IdempotencyKey      string `json:"idempotencyKey" gorm:"column:idempotency_key"`
	FailureReason       string `json:"failureReason" gorm:"column:remark"`
	CreatedAt           string `json:"createdAt" gorm:"column:created_at"`
	CompletedAt         string `json:"completedAt" gorm:"column:completed_at"`
	RawMetadata         string `json:"rawMetadata,omitempty" gorm:"column:raw_metadata"`
}

type AdjustmentRow struct {
	ID                  uint64 `json:"id" gorm:"column:id"`
	AdjustmentNo        string `json:"adjustmentID" gorm:"column:adjustment_no"`
	UserID              uint64 `json:"userID" gorm:"column:user_id"`
	UserNo              string `json:"userNo" gorm:"column:user_no"`
	Username            string `json:"username" gorm:"column:username"`
	PhoneCountryCode    string `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone               string `json:"phone" gorm:"column:phone"`
	WalletID            uint64 `json:"walletID" gorm:"column:wallet_id"`
	AdjustmentType      string `json:"adjustmentType" gorm:"column:adjustment_type"`
	Amount              string `json:"amount" gorm:"column:amount"`
	Currency            string `json:"currency" gorm:"column:currency"`
	ReasonCode          string `json:"reasonCode" gorm:"column:reason_code"`
	DetailedReason      string `json:"detailedReason" gorm:"column:detailed_reason"`
	SupportingReference string `json:"supportingReference" gorm:"column:supporting_reference"`
	InternalRemark      string `json:"internalRemark" gorm:"column:internal_remark"`
	Status              string `json:"status" gorm:"column:status"`
	RequestedBalance    string `json:"requestedBalance" gorm:"column:requested_balance"`
	BalanceBefore       string `json:"balanceBefore" gorm:"column:balance_before"`
	BalanceAfter        string `json:"balanceAfter" gorm:"column:balance_after"`
	Applicant           string `json:"applicant" gorm:"column:applicant_username"`
	ApplicantID         uint64 `json:"applicantID" gorm:"column:applicant_id"`
	Approver            string `json:"approver" gorm:"column:approver_username"`
	ApprovalReason      string `json:"approvalReason" gorm:"column:approval_reason"`
	WalletTransactionID uint64 `json:"walletTransactionID" gorm:"column:wallet_transaction_id"`
	FailureReason       string `json:"failureReason" gorm:"column:failure_reason"`
	CreatedAt           string `json:"createdAt" gorm:"column:created_at"`
	ApprovedAt          string `json:"approvedAt" gorm:"column:approved_at"`
	CompletedAt         string `json:"completedAt" gorm:"column:completed_at"`
}

type DepositRow struct {
	ID                     uint64 `json:"id" gorm:"column:id"`
	OrderNo                string `json:"platformOrderID" gorm:"column:order_no"`
	ChannelOrderNo         string `json:"channelOrderID" gorm:"column:channel_order_no"`
	UserID                 uint64 `json:"userID" gorm:"column:user_id"`
	UserNo                 string `json:"userNo" gorm:"column:user_no"`
	Username               string `json:"username" gorm:"column:username"`
	ChannelCode            string `json:"paymentChannel" gorm:"column:channel_code"`
	MethodCode             string `json:"paymentMethod" gorm:"column:method_code"`
	Amount                 string `json:"requestedAmount" gorm:"column:amount"`
	FeeAmount              string `json:"fee" gorm:"column:fee_amount"`
	PaidAmount             string `json:"actualReceivedAmount" gorm:"column:paid_amount"`
	Currency               string `json:"currency" gorm:"column:currency"`
	Status                 string `json:"status" gorm:"-"`
	StatusValue            int    `json:"-" gorm:"column:status"`
	WalletTransactionID    uint64 `json:"walletTransactionID" gorm:"column:wallet_transaction_id"`
	Reviewer               string `json:"reviewer" gorm:"column:reviewer_username"`
	ReviewReason           string `json:"reviewReason" gorm:"column:review_reason"`
	EvidenceReference      string `json:"evidenceReference" gorm:"column:evidence_reference"`
	ChannelResponseCode    string `json:"channelResponseCode" gorm:"column:channel_response_code"`
	ChannelResponseMessage string `json:"channelResponseMessage" gorm:"column:channel_response_message"`
	FailureReason          string `json:"failureReason" gorm:"column:fail_reason"`
	CreatedAt              string `json:"createdAt" gorm:"column:created_at"`
	ExpiredAt              string `json:"expiredAt" gorm:"column:expired_at"`
	PaidAt                 string `json:"paidAt" gorm:"column:paid_at"`
	CallbackAt             string `json:"callbackAt" gorm:"column:callback_at"`
	CompletedAt            string `json:"completedAt" gorm:"column:completed_at"`
}

type WithdrawalRow struct {
	ID                      uint64 `json:"id" gorm:"column:id"`
	OrderNo                 string `json:"platformOrderID" gorm:"column:order_no"`
	ChannelOrderNo          string `json:"channelOrderID" gorm:"column:channel_order_no"`
	UserID                  uint64 `json:"userID" gorm:"column:user_id"`
	UserNo                  string `json:"userNo" gorm:"column:user_no"`
	Username                string `json:"username" gorm:"column:username"`
	Amount                  string `json:"requestedAmount" gorm:"column:amount"`
	FeeAmount               string `json:"fee" gorm:"column:fee_amount"`
	ActualAmount            string `json:"actualPayoutAmount" gorm:"column:actual_amount"`
	Currency                string `json:"currency" gorm:"column:currency"`
	PaymentMethod           string `json:"paymentMethod" gorm:"column:payment_method"`
	RecipientAccountMasked  string `json:"recipientAccountMasked" gorm:"column:account_masked"`
	RecipientNameMasked     string `json:"recipientNameMasked" gorm:"column:account_name_masked"`
	BankName                string `json:"bankName" gorm:"column:bank_name"`
	KYCStatus               string `json:"kycStatus" gorm:"column:kyc_status"`
	RiskStatus              string `json:"riskStatus" gorm:"column:risk_status"`
	Status                  string `json:"status" gorm:"-"`
	StatusValue             int    `json:"-" gorm:"column:status"`
	AuditStatus             string `json:"auditStatus" gorm:"-"`
	AuditStatusValue        int    `json:"-" gorm:"column:audit_status"`
	Reviewer                string `json:"reviewer" gorm:"column:reviewer_username"`
	ReviewRemark            string `json:"reviewRemark" gorm:"column:review_remark"`
	FreezeRecordID          uint64 `json:"freezeRecordID" gorm:"column:freeze_record_id"`
	FinalDebitTransactionID uint64 `json:"finalDebitTransactionID" gorm:"column:final_debit_transaction_id"`
	RefundTransactionID     uint64 `json:"refundTransactionID" gorm:"column:refund_transaction_id"`
	FailureReason           string `json:"failureReason" gorm:"column:failure_reason"`
	CreatedAt               string `json:"createdAt" gorm:"column:created_at"`
	SubmittedAt             string `json:"submittedAt" gorm:"column:submitted_at"`
	ReviewedAt              string `json:"reviewedAt" gorm:"column:reviewed_at"`
	ApprovedAt              string `json:"approvedAt" gorm:"column:approved_at"`
	CompletedAt             string `json:"completedAt" gorm:"column:completed_at"`
}

type ReconciliationRow struct {
	ID                 uint64 `json:"id" gorm:"column:id"`
	ReconciliationNo   string `json:"reconciliationID" gorm:"column:reconciliation_no"`
	BusinessDate       string `json:"businessDate" gorm:"column:business_date"`
	BusinessType       string `json:"businessType" gorm:"column:business_type"`
	ChannelCode        string `json:"channel" gorm:"column:channel_code"`
	Currency           string `json:"currency" gorm:"column:currency"`
	Mode               string `json:"mode" gorm:"column:reconciliation_mode"`
	PlatformOrderCount uint   `json:"platformOrderCount" gorm:"column:platform_order_count"`
	PlatformAmount     string `json:"platformAmount" gorm:"column:platform_amount"`
	ChannelOrderCount  uint   `json:"channelOrderCount" gorm:"column:channel_order_count"`
	ChannelAmount      string `json:"channelAmount" gorm:"column:channel_amount"`
	DifferenceAmount   string `json:"differenceAmount" gorm:"column:difference_amount"`
	DifferenceCount    uint   `json:"differenceCount" gorm:"column:difference_count"`
	Status             string `json:"status" gorm:"column:status"`
	GeneratedBy        string `json:"generatedBy" gorm:"column:generated_by"`
	GeneratedAt        string `json:"generatedAt" gorm:"column:generated_at"`
	CompletedAt        string `json:"completedAt" gorm:"column:completed_at"`
}

type CreateAdjustmentCommand struct {
	UserID              uint64 `json:"userID"`
	WalletID            uint64 `json:"walletID"`
	AdjustmentType      string `json:"adjustmentType"`
	Amount              string `json:"amount"`
	Currency            string `json:"currency"`
	ReasonCode          string `json:"reasonCode"`
	DetailedReason      string `json:"detailedReason"`
	SupportingReference string `json:"supportingReference"`
	InternalRemark      string `json:"internalRemark"`
	IdempotencyKey      string `json:"-"`
}

type ReasonCommand struct {
	Reason         string `json:"reason"`
	InternalRemark string `json:"internalRemark"`
}

type DepositConfirmCommand struct {
	Reason            string `json:"reason"`
	EvidenceReference string `json:"evidenceReference"`
	ChannelOrderNo    string `json:"channelOrderID"`
	ReceivedAmount    string `json:"receivedAmount"`
}

type GenerateReconciliationCommand struct {
	BusinessDate string `json:"businessDate"`
	BusinessType string `json:"businessType"`
	ChannelCode  string `json:"channel"`
	Currency     string `json:"currency"`
}

func ParseMoney(value string) (decimal.Decimal, error) {
	return decimal.NewFromString(value)
}
