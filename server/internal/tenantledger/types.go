package tenantledger

import (
	"errors"
	"strings"
	"time"

	"casino/internal/tenancy"

	"github.com/shopspring/decimal"
)

const (
	AccountTenantDeposit      = "TENANT_DEPOSIT"
	AccountTenantCredit       = "TENANT_CREDIT"
	AccountTenantAvailable    = "TENANT_AVAILABLE"
	AccountTenantFrozen       = "TENANT_FROZEN"
	AccountPlatformCommission = "PLATFORM_COMMISSION"
	AccountProviderPayable    = "PROVIDER_PAYABLE"
	AccountAdjustment         = "ADJUSTMENT"

	DirectionDebit  = "DEBIT"
	DirectionCredit = "CREDIT"

	EventTenantDeposit  = "TENANT_DEPOSIT"
	EventTenantCredit   = "TENANT_CREDIT"
	EventTenantFreeze   = "TENANT_FREEZE"
	EventTenantUnfreeze = "TENANT_UNFREEZE"
	EventReversal       = "REVERSAL"

	StatusActive = "ACTIVE"
	StatusPosted = "POSTED"
)

var accountNormalBalance = map[string]string{
	AccountTenantDeposit:      DirectionCredit,
	AccountTenantCredit:       DirectionCredit,
	AccountTenantAvailable:    DirectionDebit,
	AccountTenantFrozen:       DirectionDebit,
	AccountPlatformCommission: DirectionCredit,
	AccountProviderPayable:    DirectionCredit,
	AccountAdjustment:         DirectionCredit,
}

var (
	ErrInvalidScope          = NewError("INVALID_SCOPE", "invalid tenant ledger scope")
	ErrPermissionDenied      = NewError("PERMISSION_DENIED", "permission denied")
	ErrInvalidAccountType    = NewError("INVALID_ACCOUNT_TYPE", "invalid ledger account type")
	ErrInvalidDirection      = NewError("INVALID_DIRECTION", "invalid ledger entry direction")
	ErrInvalidAmount         = NewError("INVALID_AMOUNT", "amount must be greater than zero")
	ErrCurrencyRequired      = NewError("CURRENCY_REQUIRED", "currency is required")
	ErrCurrencyMismatch      = NewError("CURRENCY_MISMATCH", "all ledger entries must use the transaction currency")
	ErrIdempotencyRequired   = NewError("IDEMPOTENCY_KEY_REQUIRED", "idempotency key is required")
	ErrUnbalancedTransaction = NewError("UNBALANCED_TRANSACTION", "ledger debit total must equal credit total")
	ErrAccountScopeMismatch  = NewError("ACCOUNT_SCOPE_MISMATCH", "ledger account scope does not match transaction scope")
	ErrInsufficientBalance   = NewError("INSUFFICIENT_BALANCE", "insufficient tenant available balance")
	ErrInsufficientFrozen    = NewError("INSUFFICIENT_FROZEN_BALANCE", "insufficient tenant frozen balance")
	ErrTransactionNotFound   = NewError("TRANSACTION_NOT_FOUND", "ledger transaction not found")
	ErrTransactionReversed   = NewError("TRANSACTION_ALREADY_REVERSED", "ledger transaction was already reversed")
	ErrReasonRequired        = NewError("REASON_REQUIRED", "reason is required")
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func NewError(code, message string) error {
	return &Error{Code: code, Message: message}
}

func ErrorCode(err error) string {
	var target *Error
	if errors.As(err, &target) {
		return target.Code
	}
	return "INTERNAL_ERROR"
}

type Actor struct {
	ID            uint64
	Username      string
	TenantID      uint64
	SiteID        uint64
	Platform      bool
	IP            string
	UserAgent     string
	CorrelationID string
}

func (a Actor) Scope() tenancy.Scope {
	return tenancy.Scope{TenantID: a.TenantID, SiteID: a.SiteID, Platform: a.Platform}
}

type EntryCommand struct {
	AccountID   uint64
	AccountType string
	Direction   string
	Amount      decimal.Decimal
	Memo        string
}

type PostCommand struct {
	Scope                   tenancy.Scope
	TenantID                uint64
	SiteID                  uint64
	Currency                string
	EventType               string
	BusinessType            string
	BusinessOrderNo         string
	IdempotencyKey          string
	Reason                  string
	OperatorID              uint64
	OperatorUsername        string
	CorrelationID           string
	ApprovalID              uint64
	Metadata                map[string]any
	ReversalOfTransactionID *uint64
	Entries                 []EntryCommand
}

type TenantMoneyCommand struct {
	Scope          tenancy.Scope
	TenantID       uint64
	SiteID         uint64
	Currency       string
	Amount         decimal.Decimal
	IdempotencyKey string
	Reason         string
	Actor          Actor
}

type CreditLimitCommand struct {
	Scope          tenancy.Scope
	TenantID       uint64
	SiteID         uint64
	Currency       string
	CreditLimit    decimal.Decimal
	IdempotencyKey string
	Reason         string
	Actor          Actor
}

type TransactionResult struct {
	ID            uint64           `json:"id"`
	TransactionNo string           `json:"transactionNo"`
	TenantID      uint64           `json:"tenantID"`
	SiteID        uint64           `json:"siteID"`
	Currency      string           `json:"currency"`
	EventType     string           `json:"eventType"`
	DebitAmount   decimal.Decimal  `json:"debitAmount"`
	CreditAmount  decimal.Decimal  `json:"creditAmount"`
	Idempotent    bool             `json:"idempotent"`
	ReversalOf    *uint64          `json:"reversalOfTransactionID,omitempty"`
	Entries       []EntryResult    `json:"entries"`
	CreditLimit   *decimal.Decimal `json:"creditLimit,omitempty"`
	CreatedAt     time.Time        `json:"createdAt"`
}

type EntryResult struct {
	ID            uint64          `json:"id"`
	AccountID     uint64          `json:"accountID"`
	AccountType   string          `json:"accountType"`
	Direction     string          `json:"direction"`
	Amount        decimal.Decimal `json:"amount"`
	BalanceBefore decimal.Decimal `json:"balanceBefore"`
	BalanceAfter  decimal.Decimal `json:"balanceAfter"`
}

type AccountBalance struct {
	ID               uint64          `json:"id"`
	AccountNo        string          `json:"accountNo"`
	TenantID         uint64          `json:"tenantID"`
	SiteID           uint64          `json:"siteID"`
	Currency         string          `json:"currency"`
	AccountType      string          `json:"accountType"`
	NormalBalance    string          `json:"normalBalance"`
	Balance          decimal.Decimal `json:"balance"`
	AvailableBalance decimal.Decimal `json:"availableBalance"`
	FrozenBalance    decimal.Decimal `json:"frozenBalance"`
	Status           string          `json:"status"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

func NormalizeCurrency(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
