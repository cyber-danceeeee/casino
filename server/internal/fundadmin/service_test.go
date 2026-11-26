package fundadmin

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMoneyAndStatusNormalization(t *testing.T) {
	if walletStatus(1) != "ACTIVE" || walletStatus(2) != "FROZEN" {
		t.Fatal("wallet status mapping changed")
	}
	if depositStatus(6) != "MANUAL_REVIEW" || withdrawalStatus(9) != "PROCESSING" {
		t.Fatal("order status mapping changed")
	}
	if transactionStatus(3) != "REVERSED" {
		t.Fatal("transaction status mapping changed")
	}
	value, err := ParseMoney("123.45")
	if err != nil || value.StringFixed(2) != "123.45" {
		t.Fatal("decimal parsing failed")
	}
}

func TestMetadataRedaction(t *testing.T) {
	redacted := redactMetadata(`{"token":"secret","signature":"sig","amount":"10.00","provider":{"api_key":"key","result":"ok"}}`)
	if redacted != `{"amount":"10.00","provider":{"api_key":"[REDACTED]","result":"ok"},"signature":"[REDACTED]","token":"[REDACTED]"}` {
		t.Fatalf("unexpected redaction: %s", redacted)
	}
	if redactMetadata("not-json") != "{}" {
		t.Fatal("invalid metadata must not be exposed")
	}
}

func TestAdjustmentApprovalStateMachine(t *testing.T) {
	if err := validateAdjustmentApproval("PENDING_APPROVAL", 1, 2, false); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(validateAdjustmentApproval("COMPLETED", 1, 2, false), ErrInvalidState) {
		t.Fatal("duplicate approval must be blocked")
	}
	if !errors.Is(validateAdjustmentApproval("PENDING_APPROVAL", 1, 1, false), ErrMakerChecker) {
		t.Fatal("self approval must be blocked")
	}
	if err := validateAdjustmentApproval("PENDING_APPROVAL", 1, 1, true); err != nil {
		t.Fatal("super admin self approval must be allowed")
	}
}

func TestAdjustedBalance(t *testing.T) {
	balance := decimal.RequireFromString("100.00")
	after, direction, err := adjustedBalance("CREDIT", balance, decimal.RequireFromString("25.50"))
	if err != nil || direction != "IN" || after.StringFixed(2) != "125.50" {
		t.Fatal("credit calculation failed")
	}
	after, direction, err = adjustedBalance("DEBIT", balance, decimal.RequireFromString("25.50"))
	if err != nil || direction != "OUT" || after.StringFixed(2) != "74.50" {
		t.Fatal("debit calculation failed")
	}
	if _, _, err = adjustedBalance("DEBIT", balance, decimal.RequireFromString("100.01")); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatal("insufficient debit must fail")
	}
}

func TestDepositConfirmationStateMachine(t *testing.T) {
	for _, status := range []int{1, 2, 6} {
		if err := validateDepositConfirmation(status); err != nil {
			t.Fatalf("status %d should be accepted: %v", status, err)
		}
	}
	for _, status := range []int{3, 4, 5} {
		if !errors.Is(validateDepositConfirmation(status), ErrInvalidState) {
			t.Fatalf("status %d should be rejected", status)
		}
	}
}

func TestWithdrawalReviewStateMachine(t *testing.T) {
	if next, err := withdrawalReviewTarget(1, "RISK"); err != nil || next != 6 {
		t.Fatal("pending must enter risk review")
	}
	if next, err := withdrawalReviewTarget(6, "FINANCE"); err != nil || next != 7 {
		t.Fatal("risk review must enter finance review")
	}
	if _, err := withdrawalReviewTarget(7, "RISK"); !errors.Is(err, ErrInvalidState) {
		t.Fatal("invalid review regression must fail")
	}
}

func TestInputValidationDoesNotTouchDatabase(t *testing.T) {
	service := &Service{}
	if _, err := service.CreateAdjustment(CreateAdjustmentCommand{}, Actor{}); !errors.Is(err, ErrIdempotencyRequired) {
		t.Fatal("missing idempotency key must fail")
	}
	if _, err := service.CreateWithdrawal(1, 1, decimal.Zero, Actor{}); !errors.Is(err, ErrInvalidAmount) {
		t.Fatal("zero withdrawal must fail")
	}
	if err := service.SetWalletStatus(1, true, "", Actor{}); !errors.Is(err, ErrReasonRequired) {
		t.Fatal("wallet freeze reason must be required")
	}
}
