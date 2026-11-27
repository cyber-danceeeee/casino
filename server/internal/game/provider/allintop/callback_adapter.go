package allintop

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"casino/internal/game/domain"
	"github.com/shopspring/decimal"
)

type CallbackAdapter struct{ cfg Config }

func NewCallbackAdapter(c Config) *CallbackAdapter { return &CallbackAdapter{cfg: c} }
func (a *CallbackAdapter) Code() string            { return Code }
func (a *CallbackAdapter) ValidateRequest(_ context.Context, r domain.ProviderCallbackRequest) error {
	if !a.cfg.Enabled || a.cfg.ClientID == "" || a.cfg.ClientSecret == "" {
		return domain.ErrProviderDisabled
	}
	req := &http.Request{Header: r.Headers}
	user, pass, ok := req.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(a.cfg.ClientID))&subtle.ConstantTimeCompare([]byte(pass), []byte(a.cfg.ClientSecret)) != 1 {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "invalid callback authentication", nil)
	}
	if r.Method != http.MethodPost || len(r.Body) == 0 || len(r.Body) >= 1<<20 {
		return invalid("invalid callback method or size")
	}
	if r.Operation != "balance" && r.Operation != "transaction" && r.Operation != "batch-transactions" {
		return invalid("invalid callback operation")
	}
	path := a.cfg.CallbackPath
	if path == "" {
		path = "/api/v1/game/providers/" + Code + "/callback"
	}
	if r.Path != strings.TrimRight(path, "/")+"/api/"+r.Operation {
		return invalid("invalid callback path")
	}
	if len(a.cfg.CallbackIPAllowlist) > 0 {
		ip := net.ParseIP(r.ClientIP)
		allowed := false
		for _, s := range a.cfg.CallbackIPAllowlist {
			if ip != nil && ip.Equal(net.ParseIP(s)) {
				allowed = true
			}
			if _, n, e := net.ParseCIDR(s); e == nil && n.Contains(ip) {
				allowed = true
			}
		}
		if !allowed {
			return domain.NewWalletError(domain.WalletErrorInvalidSignature, "callback source is not allowed", nil)
		}
	}
	return nil
}
func invalid(s string) error { return domain.NewWalletError(domain.WalletErrorInvalidRequest, s, nil) }

type transaction struct {
	UserCode        string          `json:"userCode"`
	VendorCode      string          `json:"vendorCode"`
	GameCode        string          `json:"gameCode"`
	HistoryID       json.Number     `json:"historyId"`
	RoundID         string          `json:"roundId"`
	TransactionCode string          `json:"transactionCode"`
	IsFinished      *bool           `json:"isFinished"`
	IsCanceled      *bool           `json:"isCanceled"`
	Amount          json.RawMessage `json:"amount"`
}

func (a *CallbackAdapter) command(t transaction, raw []byte) (*domain.WalletCommand, error) {
	history, e := strconv.ParseInt(t.HistoryID.String(), 10, 64)
	if e != nil || history <= 0 || t.UserCode == "" || len(t.UserCode) > 128 || t.TransactionCode == "" || len(t.TransactionCode) > 128 || t.RoundID == "" || len(t.RoundID) > 128 || t.VendorCode == "" || strings.Contains(t.VendorCode, "/") || t.GameCode == "" || len(t.VendorCode+"/"+t.GameCode) > 128 || t.IsFinished == nil || t.IsCanceled == nil {
		return nil, invalid("invalid transaction identifiers or flags")
	}
	c := &domain.WalletCommand{ProviderCode: Code, ProviderAction: "transaction", ProviderPlayerID: t.UserCode, ProviderTransactionID: t.TransactionCode, ProviderRoundID: t.RoundID, ProviderGameID: t.VendorCode + "/" + t.GameCode, ProviderHistoryID: strconv.FormatInt(history, 10), TrackHistory: true, RoundFinished: *t.IsFinished, Currency: "IDR", RawPayload: raw}
	if *t.IsCanceled {
		c.Operation = domain.WalletOperationCancel
		c.RoundFinished = true
		return c, nil
	}
	if len(t.Amount) == 0 || string(t.Amount) == "null" {
		return nil, invalid("missing transaction amount")
	}
	var d decimal.Decimal
	if json.Unmarshal(t.Amount, &d) != nil {
		return nil, invalid("invalid transaction amount")
	}
	minor, err := moneyMinor(d)
	if err != nil {
		return nil, invalid("invalid transaction amount precision or range")
	}
	c.Operation = domain.WalletOperationCredit
	if minor < 0 {
		c.Operation = domain.WalletOperationDebit
		minor = -minor
	}
	if a.cfg.MaxTransactionAmountMinor > 0 && minor > a.cfg.MaxTransactionAmountMinor {
		return nil, invalid("transaction exceeds configured amount limit")
	}
	c.AmountMinor = minor
	return c, nil
}
func (a *CallbackAdapter) ParseCommand(_ context.Context, r domain.ProviderCallbackRequest) (*domain.WalletCommand, error) {
	if r.Operation == "balance" {
		var b struct {
			UserCode string `json:"userCode"`
		}
		if json.Unmarshal(r.Body, &b) != nil || b.UserCode == "" || len(b.UserCode) > 128 {
			return nil, invalid("invalid balance request")
		}
		return &domain.WalletCommand{ProviderCode: Code, ProviderPlayerID: b.UserCode, Currency: "IDR", Operation: domain.WalletOperationBalance}, nil
	}
	if r.Operation != "transaction" {
		return nil, invalid("invalid transaction request")
	}
	var t transaction
	if json.Unmarshal(r.Body, &t) != nil {
		return nil, invalid("invalid transaction JSON")
	}
	return a.command(t, r.Body)
}
func (a *CallbackAdapter) ParseBatchCommands(_ context.Context, r domain.ProviderCallbackRequest) ([]domain.WalletCommand, error) {
	var b struct {
		UserCode     string            `json:"userCode"`
		Transactions []json.RawMessage `json:"transactions"`
	}
	if r.Operation != "batch-transactions" || json.Unmarshal(r.Body, &b) != nil || b.UserCode == "" || len(b.UserCode) > 128 || len(b.Transactions) == 0 || len(b.Transactions) > 1000 {
		return nil, invalid("invalid transaction batch")
	}
	out := make([]domain.WalletCommand, 0, len(b.Transactions))
	for _, raw := range b.Transactions {
		var t transaction
		if json.Unmarshal(raw, &t) != nil {
			return nil, invalid("invalid batch transaction")
		}
		if t.UserCode == "" {
			t.UserCode = b.UserCode
		}
		if t.UserCode != b.UserCode {
			return nil, invalid("batch transactions must belong to the same player")
		}
		c, err := a.command(t, raw)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, nil
}
func (a *CallbackAdapter) BuildSuccessResponse(_ context.Context, _ domain.WalletCommand, r domain.WalletResult) (*domain.ProviderCallbackResponse, error) {
	return &domain.ProviderCallbackResponse{StatusCode: http.StatusOK, Body: map[string]any{"success": true, "message": json.Number(decimal.NewFromInt(r.BalanceMinor).Shift(-2).StringFixed(2)), "errorCode": 0}}, nil
}
func (a *CallbackAdapter) BuildErrorResponse(_ context.Context, _ domain.ProviderCallbackRequest, err error) *domain.ProviderCallbackResponse {
	code := 500
	var e *domain.WalletError
	if errors.As(err, &e) {
		switch e.Code {
		case domain.WalletErrorInvalidSignature:
			code = 401
		case domain.WalletErrorPlayerNotFound:
			code = 2
		case domain.WalletErrorInsufficientBalance:
			code = 4
		case domain.WalletErrorDuplicateTransaction, domain.WalletErrorTransactionConflict:
			code = 6
		case domain.WalletErrorRoundClosed:
			code = 7
		case domain.WalletErrorInvalidRequest, domain.WalletErrorCurrencyMismatch:
			code = 400
		}
	}
	if errors.Is(err, domain.ErrProviderDisabled) {
		code = 9
	}
	status := http.StatusOK
	if code == 401 {
		status = http.StatusUnauthorized
	}
	// Never reflect credentials, raw SQL, request bodies, or upstream errors.
	return &domain.ProviderCallbackResponse{StatusCode: status, Body: map[string]any{"success": false, "message": "wallet request rejected", "errorCode": code}}
}
