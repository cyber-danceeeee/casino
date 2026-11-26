package softaggregator

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"casino/internal/game/domain"
)

type CallbackAdapter struct {
	cfg Config
	now func() time.Time
}

func NewCallbackAdapter(cfg Config) *CallbackAdapter {
	return &CallbackAdapter{
		cfg: cfg,
		now: time.Now,
	}
}

func (a *CallbackAdapter) Code() string {
	return Code
}

func (a *CallbackAdapter) ValidateRequest(ctx context.Context, request domain.ProviderCallbackRequest) error {
	if !a.cfg.CallbackReady() {
		return domain.NewWalletError(domain.WalletErrorProviderDisabled, "provider disabled", domain.ErrProviderDisabled)
	}
	if request.Method != http.MethodGet {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid callback method", nil)
	}
	if !a.ipAllowed(request.ClientIP) {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "callback source ip is not allowed", nil)
	}

	timestampValue := strings.TrimSpace(request.QueryValue("timestamp"))
	key := strings.TrimSpace(request.QueryValue("key"))
	if timestampValue == "" || key == "" {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "missing callback signature", domain.ErrInvalidSignature)
	}
	timestamp, err := strconv.ParseInt(timestampValue, 10, 64)
	if err != nil {
		return domain.NewWalletError(domain.WalletErrorInvalidTimestamp, "invalid callback timestamp", err)
	}
	receivedAt := request.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = a.now()
	}
	if absDuration(receivedAt.Sub(time.Unix(timestamp, 0))) > a.cfg.MaxTimestampDrift {
		return domain.NewWalletError(domain.WalletErrorInvalidTimestamp, "expired callback timestamp", domain.ErrInvalidTimestamp)
	}
	if !ValidCallbackKey(timestampValue, a.cfg.SaltKey, key) {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "invalid callback signature", domain.ErrInvalidSignature)
	}

	if strings.TrimSpace(callbackAction(request)) == "" {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing callback action", nil)
	}
	if strings.TrimSpace(request.QueryValue("username")) == "" {
		return domain.NewWalletError(domain.WalletErrorPlayerNotFound, "missing provider player id", nil)
	}
	if strings.TrimSpace(callbackCurrency(request)) == "" {
		return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "missing currency", nil)
	}
	return nil
}

func (a *CallbackAdapter) ParseCommand(ctx context.Context, request domain.ProviderCallbackRequest) (*domain.WalletCommand, error) {
	action := strings.ToLower(strings.TrimSpace(callbackAction(request)))
	operation, err := callbackOperation(action)
	if err != nil {
		return nil, err
	}

	amountMinor := int64(0)
	if operation == domain.WalletOperationDebit || operation == domain.WalletOperationCredit || operation == domain.WalletOperationRefund {
		amountValue := strings.TrimSpace(request.QueryValue("amount"))
		if amountValue == "" {
			return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing amount", nil)
		}
		amountMinor, err = strconv.ParseInt(amountValue, 10, 64)
		if err != nil || amountMinor < 0 {
			return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid amount", err)
		}
		if a.cfg.MaxTransactionAmountMinor > 0 && amountMinor > a.cfg.MaxTransactionAmountMinor {
			return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "amount exceeds provider limit", nil)
		}
		if operation == domain.WalletOperationDebit && strings.EqualFold(request.QueryValue("type"), "bonus_fs") {
			amountMinor = 0
		}
	}

	timestamp, _ := strconv.ParseInt(strings.TrimSpace(request.QueryValue("timestamp")), 10, 64)
	rawPayload, _ := json.Marshal(redactedQuery(request.Query))
	callID := strings.TrimSpace(request.QueryValue("call_id"))
	if operation == domain.WalletOperationBalance && callID == "" {
		callID = "balance:" + strings.TrimSpace(request.QueryValue("timestamp"))
	}

	return &domain.WalletCommand{
		ProviderCode:          Code,
		Operation:             operation,
		ProviderAction:        strings.ToUpper(action),
		ProviderPlayerID:      strings.TrimSpace(request.QueryValue("username")),
		ProviderTransactionID: callID,
		ProviderRoundID:       strings.TrimSpace(request.QueryValue("round_id")),
		ProviderGameID:        strings.TrimSpace(request.QueryValue("game_id")),
		AmountMinor:           amountMinor,
		Currency:              strings.ToUpper(strings.TrimSpace(callbackCurrency(request))),
		RelatedTransactionID:  strings.TrimSpace(request.QueryValue("related_call_id")),
		Timestamp:             time.Unix(timestamp, 0),
		RawPayload:            rawPayload,
	}, nil
}

func (a *CallbackAdapter) BuildSuccessResponse(ctx context.Context, command domain.WalletCommand, result domain.WalletResult) (*domain.ProviderCallbackResponse, error) {
	return &domain.ProviderCallbackResponse{
		StatusCode: http.StatusOK,
		Body: callbackResponse{
			Error:   0,
			Balance: result.BalanceMinor,
		},
	}, nil
}

func (a *CallbackAdapter) BuildErrorResponse(ctx context.Context, request domain.ProviderCallbackRequest, err error) *domain.ProviderCallbackResponse {
	balance := int64(0)
	code := 2
	var walletErr *domain.WalletError
	if errors.As(err, &walletErr) {
		balance = walletErr.BalanceMinor
		if walletErr.Code == domain.WalletErrorInsufficientBalance {
			code = 1
		}
	}
	return &domain.ProviderCallbackResponse{
		StatusCode: http.StatusOK,
		Body: callbackResponse{
			Error:   code,
			Balance: balance,
		},
	}
}

func (a *CallbackAdapter) ipAllowed(clientIP string) bool {
	if len(a.cfg.CallbackIPAllowlist) == 0 {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return false
	}
	for _, allowed := range a.cfg.CallbackIPAllowlist {
		if allowedIP := net.ParseIP(allowed); allowedIP != nil && allowedIP.Equal(ip) {
			return true
		}
		if _, cidr, err := net.ParseCIDR(allowed); err == nil && cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func callbackOperation(action string) (domain.WalletOperation, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "balance":
		return domain.WalletOperationBalance, nil
	case "debit":
		return domain.WalletOperationDebit, nil
	case "credit":
		return domain.WalletOperationCredit, nil
	case "refund":
		return domain.WalletOperationRefund, nil
	case "rollback":
		return domain.WalletOperationRollback, nil
	default:
		return "", domain.NewWalletError(domain.WalletErrorInvalidRequest, "unsupported callback action", nil)
	}
}

func callbackAction(request domain.ProviderCallbackRequest) string {
	if action := request.QueryValue("action"); action != "" {
		return action
	}
	return request.Operation
}

func callbackCurrency(request domain.ProviderCallbackRequest) string {
	for _, key := range []string{"currency", "curr", "cur", "¤cy"} {
		if value := request.QueryValue(key); value != "" {
			return value
		}
	}
	return ""
}

func redactedQuery(query map[string][]string) map[string][]string {
	out := make(map[string][]string, len(query))
	for key, values := range query {
		if strings.EqualFold(key, "key") {
			out[key] = []string{"REDACTED"}
			continue
		}
		cp := make([]string, len(values))
		copy(cp, values)
		out[key] = cp
	}
	return out
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
