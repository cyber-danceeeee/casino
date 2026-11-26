package domain

import (
	"context"
	"net/http"
	"time"
)

type WalletOperation string

const (
	WalletOperationBalance  WalletOperation = "BALANCE"
	WalletOperationDebit    WalletOperation = "DEBIT"
	WalletOperationCredit   WalletOperation = "CREDIT"
	WalletOperationRefund   WalletOperation = "REFUND"
	WalletOperationRollback WalletOperation = "ROLLBACK"
)

type ProviderCallbackRequest struct {
	ProviderCode string
	Operation    string
	Method       string
	Path         string
	Query        map[string][]string
	Headers      http.Header
	Body         []byte
	ClientIP     string
	ReceivedAt   time.Time
}

func (r ProviderCallbackRequest) QueryValue(key string) string {
	if len(r.Query) == 0 {
		return ""
	}
	values := r.Query[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

type WalletCommand struct {
	TenantID              uint64
	SiteID                uint64
	ProviderCode          string
	Operation             WalletOperation
	ProviderAction        string
	ProviderPlayerID      string
	ProviderTransactionID string
	ProviderRoundID       string
	ProviderGameID        string
	ProviderCurrencyCode  string
	AmountMinor           int64
	Currency              string
	RelatedTransactionID  string
	Timestamp             time.Time
	RequestHash           string
	RawPayload            []byte
}

type WalletResult struct {
	Success bool

	UserID string

	InternalTransactionID string
	ProviderTransactionID string

	BalanceMinor int64
	Currency     string

	Duplicate bool

	ErrorCode    WalletErrorCode
	ErrorMessage string
}

type ProviderCallbackResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       any
}

type ProviderCallbackAdapter interface {
	Code() string
	ValidateRequest(ctx context.Context, request ProviderCallbackRequest) error
	ParseCommand(ctx context.Context, request ProviderCallbackRequest) (*WalletCommand, error)
	BuildSuccessResponse(ctx context.Context, command WalletCommand, result WalletResult) (*ProviderCallbackResponse, error)
	BuildErrorResponse(ctx context.Context, request ProviderCallbackRequest, err error) *ProviderCallbackResponse
}

type ProviderCallbackRegistry interface {
	Register(adapter ProviderCallbackAdapter) error
	Get(providerCode string) (ProviderCallbackAdapter, error)
	List() []ProviderCallbackAdapter
}
