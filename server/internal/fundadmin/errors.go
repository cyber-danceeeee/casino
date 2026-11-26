package fundadmin

import "fmt"

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func newError(code, message string) error {
	return &Error{Code: code, Message: message}
}

func wrapInternal(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*Error); ok {
		return err
	}
	return fmt.Errorf("internal funds error: %w", err)
}

var (
	ErrWalletNotFound       = newError("WALLET_NOT_FOUND", "wallet not found")
	ErrWalletFrozen         = newError("WALLET_FROZEN", "wallet is frozen")
	ErrInsufficientBalance  = newError("INSUFFICIENT_BALANCE", "insufficient wallet balance")
	ErrInvalidAmount        = newError("INVALID_AMOUNT", "amount must be greater than zero")
	ErrCurrencyMismatch     = newError("CURRENCY_MISMATCH", "wallet currency does not match")
	ErrAdjustmentNotFound   = newError("ADJUSTMENT_NOT_FOUND", "manual adjustment not found")
	ErrMakerChecker         = newError("MAKER_CHECKER_VIOLATION", "applicant cannot approve this adjustment")
	ErrInvalidState         = newError("INVALID_STATE_TRANSITION", "invalid state transition")
	ErrDepositNotFound      = newError("DEPOSIT_NOT_FOUND", "deposit order not found")
	ErrWithdrawalNotFound   = newError("WITHDRAWAL_NOT_FOUND", "withdrawal order not found")
	ErrWithdrawalAudit      = newError("WITHDRAWAL_AUDIT_INCOMPLETE", "cash audit requirement is not completed")
	ErrReconciliationAbsent = newError("RECONCILIATION_NOT_FOUND", "reconciliation record not found")
	ErrDifferenceNotFound   = newError("RECONCILIATION_DIFFERENCE_NOT_FOUND", "reconciliation difference not found")
	ErrReasonRequired       = newError("REASON_REQUIRED", "operator reason is required")
	ErrIdempotencyRequired  = newError("IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required")
)
