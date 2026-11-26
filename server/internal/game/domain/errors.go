package domain

import "errors"

var (
	ErrNotSupported        = errors.New("game provider operation is not supported")
	ErrProviderNotFound    = errors.New("game provider not found")
	ErrProviderDisabled    = errors.New("game provider disabled")
	ErrDuplicateProvider   = errors.New("game provider already registered")
	ErrCallbackNotFound    = errors.New("game provider callback adapter not found")
	ErrDuplicateCallback   = errors.New("game provider callback adapter already registered")
	ErrInvalidCallback     = errors.New("invalid provider callback")
	ErrInvalidSignature    = errors.New("invalid provider callback signature")
	ErrInvalidTimestamp    = errors.New("invalid provider callback timestamp")
	ErrTransactionConflict = errors.New("provider transaction conflict")
)

type WalletErrorCode string

const (
	WalletErrorInvalidSignature     WalletErrorCode = "INVALID_SIGNATURE"
	WalletErrorInvalidTimestamp     WalletErrorCode = "INVALID_TIMESTAMP"
	WalletErrorInvalidRequest       WalletErrorCode = "INVALID_REQUEST"
	WalletErrorPlayerNotFound       WalletErrorCode = "PLAYER_NOT_FOUND"
	WalletErrorGameNotFound         WalletErrorCode = "GAME_NOT_FOUND"
	WalletErrorProviderDisabled     WalletErrorCode = "PROVIDER_DISABLED"
	WalletErrorProviderMaintenance  WalletErrorCode = "PROVIDER_MAINTENANCE"
	WalletErrorGameMaintenance      WalletErrorCode = "GAME_MAINTENANCE"
	WalletErrorCurrencyMismatch     WalletErrorCode = "CURRENCY_MISMATCH"
	WalletErrorInsufficientBalance  WalletErrorCode = "INSUFFICIENT_BALANCE"
	WalletErrorDuplicateTransaction WalletErrorCode = "DUPLICATE_TRANSACTION"
	WalletErrorTransactionConflict  WalletErrorCode = "TRANSACTION_CONFLICT"
	WalletErrorProviderUnavailable  WalletErrorCode = "PROVIDER_UNAVAILABLE"
	WalletErrorMarketUnavailable    WalletErrorCode = "MARKET_UNAVAILABLE"
	WalletErrorInternal             WalletErrorCode = "INTERNAL_ERROR"
)

type WalletError struct {
	Code         WalletErrorCode
	Message      string
	BalanceMinor int64
	Currency     string
	Err          error
}

func (e *WalletError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *WalletError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewWalletError(code WalletErrorCode, message string, err error) *WalletError {
	return &WalletError{Code: code, Message: message, Err: err}
}
