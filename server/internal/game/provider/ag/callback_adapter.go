package ag

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"casino/internal/game/domain"
)

type CallbackAdapter struct {
	cfg Config
	now func() time.Time
}

type callbackPayload struct {
	ReqTraceID    string `json:"reqTraceId"`
	PlayerID      string `json:"playerId"`
	CurrencyCode  string `json:"currencyCode"`
	GameCode      string `json:"gameCode"`
	RoundID       string `json:"roundId"`
	BetID         string `json:"betId"`
	MerchantBetID string `json:"merchantBetId"`
	BetAmount     string `json:"betAmount"`
	WinAmount     string `json:"winAmount"`
	IsFree        bool   `json:"isFree"`
	IsEnd         bool   `json:"isEnd"`
	BetTime       any    `json:"betTime"`
	SettledTime   any    `json:"settledTime"`
	Type          string `json:"type"`
	Timestamp     any    `json:"timestamp"`
	TransactionID string `json:"transactionId"`
	CancelType    string `json:"cancelType"`
}

func NewCallbackAdapter(cfg Config) *CallbackAdapter {
	return &CallbackAdapter{cfg: cfg, now: time.Now}
}

func (a *CallbackAdapter) Code() string {
	return Code
}

func (a *CallbackAdapter) ValidateRequest(ctx context.Context, request domain.ProviderCallbackRequest) error {
	if !a.cfg.CallbackReady() {
		return domain.NewWalletError(domain.WalletErrorProviderDisabled, "provider disabled", domain.ErrProviderDisabled)
	}
	if request.Method != http.MethodPost {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid callback method", nil)
	}
	if !a.ipAllowed(request.ClientIP) {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "callback source ip is not allowed", nil)
	}
	if len(request.Body) == 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "missing callback body", nil)
	}

	headerMerchant := strings.TrimSpace(request.Headers.Get("X-MERCHANT-CODE"))
	timestamp := strings.TrimSpace(request.Headers.Get("X-TIMESTAMP"))
	nonceValue := strings.TrimSpace(request.Headers.Get("X-NONCE"))
	signature := strings.TrimSpace(request.Headers.Get("X-SIGN"))
	signAlgorithm := strings.TrimSpace(request.Headers.Get("X-CONTENT-PROCESSING-TYPE"))
	if headerMerchant == "" || timestamp == "" || nonceValue == "" || signature == "" {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "missing callback signature headers", domain.ErrInvalidSignature)
	}
	if !strings.EqualFold(headerMerchant, a.cfg.MerchantCode) {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "merchant code mismatch", domain.ErrInvalidSignature)
	}
	if !strings.EqualFold(signAlgorithm, signType) {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "invalid signature type", domain.ErrInvalidSignature)
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return domain.NewWalletError(domain.WalletErrorInvalidTimestamp, "invalid callback timestamp", err)
	}
	receivedAt := request.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = a.now()
	}
	if a.cfg.MaxTimestampDrift > 0 && absDuration(receivedAt.Sub(time.UnixMilli(ts))) > a.cfg.MaxTimestampDrift {
		return domain.NewWalletError(domain.WalletErrorInvalidTimestamp, "expired callback timestamp", domain.ErrInvalidTimestamp)
	}
	if !validSignature(a.cfg.SecretKey, a.cfg.MerchantCode, timestamp, nonceValue, request.Body, signature) {
		return domain.NewWalletError(domain.WalletErrorInvalidSignature, "invalid callback signature", domain.ErrInvalidSignature)
	}
	return nil
}

func (a *CallbackAdapter) ParseCommand(ctx context.Context, request domain.ProviderCallbackRequest) (*domain.WalletCommand, error) {
	payload, err := parseCallbackPayload(request)
	if err != nil {
		return nil, err
	}
	operationName := normalizeOperationName(request.Operation)
	operation := domain.WalletOperationBalance
	amountMinor := int64(0)
	relatedTransactionID := strings.TrimSpace(payload.MerchantBetID)

	switch operationName {
	case "balance", "player-info":
		operation = domain.WalletOperationBalance
	case "bet":
		operation = domain.WalletOperationDebit
		amountMinor, err = amountToMinor(payload.BetAmount)
	case "win":
		operation, amountMinor, err = winOperationAndAmount(payload)
	case "cancel":
		operation = domain.WalletOperationRollback
		if relatedTransactionID == "" {
			relatedTransactionID = strings.TrimSpace(payload.BetID)
		}
	default:
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "unsupported ag callback operation", nil)
	}
	if err != nil {
		return nil, err
	}
	if amountMinor < 0 {
		amountMinor = -amountMinor
	}
	if a.cfg.MaxTransactionAmountMinor > 0 && amountMinor > a.cfg.MaxTransactionAmountMinor {
		return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "amount exceeds provider limit", nil)
	}

	transactionID := firstNonEmpty(payload.TransactionID, payload.BetID, payload.ReqTraceID)
	if operation == domain.WalletOperationBalance {
		transactionID = firstNonEmpty(payload.ReqTraceID, "balance:"+payload.PlayerID+":"+payload.CurrencyCode)
	}

	rawPayload, _ := json.Marshal(payload)
	return &domain.WalletCommand{
		ProviderCode:          Code,
		Operation:             operation,
		ProviderAction:        strings.ToUpper(operationName),
		ProviderPlayerID:      strings.TrimSpace(payload.PlayerID),
		ProviderTransactionID: transactionID,
		ProviderRoundID:       strings.TrimSpace(payload.RoundID),
		ProviderGameID:        strings.TrimSpace(payload.GameCode),
		ProviderCurrencyCode:  strings.ToUpper(strings.TrimSpace(payload.CurrencyCode)),
		AmountMinor:           amountMinor,
		Currency:              (currencyMapper{}).ToPlatformCurrency(payload.CurrencyCode),
		RelatedTransactionID:  relatedTransactionID,
		Timestamp:             callbackTimestamp(payload),
		RawPayload:            rawPayload,
	}, nil
}

func (a *CallbackAdapter) BuildSuccessResponse(ctx context.Context, command domain.WalletCommand, result domain.WalletResult) (*domain.ProviderCallbackResponse, error) {
	data := map[string]any{}
	switch command.ProviderAction {
	case "PLAYER-INFO", "PLAYER_INFO":
		data["nickName"] = command.ProviderPlayerID
		data["avatarUrl"] = ""
	case "BALANCE":
		data["balance"] = minorToAmount(result.BalanceMinor)
	case "BET", "WIN":
		data["merchantBetId"] = command.ProviderTransactionID
		data["balance"] = minorToAmount(result.BalanceMinor)
	case "CANCEL":
		data["merchantBetId"] = firstNonEmpty(command.RelatedTransactionID, command.ProviderTransactionID)
		data["betId"] = command.RelatedTransactionID
		data["balance"] = minorToAmount(result.BalanceMinor)
		currencyCode := strings.TrimSpace(command.ProviderCurrencyCode)
		if currencyCode == "" {
			currencyCode = strings.ToUpper(strings.TrimSpace(result.Currency))
		}
		data["currencyCode"] = currencyCode
	default:
		data["balance"] = minorToAmount(result.BalanceMinor)
	}
	return &domain.ProviderCallbackResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body: callbackEnvelope{
			Success: true,
			Code:    "C10000",
			Msg:     "Request succeeded",
			Data:    data,
		},
	}, nil
}

func (a *CallbackAdapter) BuildErrorResponse(ctx context.Context, request domain.ProviderCallbackRequest, err error) *domain.ProviderCallbackResponse {
	code := "C40000"
	msg := "Request failed"
	var walletErr *domain.WalletError
	if errors.As(err, &walletErr) {
		msg = walletErr.Error()
		switch walletErr.Code {
		case domain.WalletErrorInsufficientBalance:
			code = "G30003"
		case domain.WalletErrorInvalidSignature:
			code = "C10005"
		case domain.WalletErrorInvalidTimestamp:
			code = "C10002"
		}
	} else if err != nil {
		msg = err.Error()
	}
	return &domain.ProviderCallbackResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body: callbackEnvelope{
			Success: false,
			Code:    code,
			Msg:     msg,
			Data:    nil,
		},
	}
}

func parseCallbackPayload(request domain.ProviderCallbackRequest) (callbackPayload, error) {
	var payload callbackPayload
	if err := json.Unmarshal(request.Body, &payload); err != nil {
		return payload, domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid callback json", err)
	}
	if strings.TrimSpace(payload.PlayerID) == "" {
		return payload, domain.NewWalletError(domain.WalletErrorPlayerNotFound, "missing player id", nil)
	}
	if strings.TrimSpace(payload.CurrencyCode) == "" {
		return payload, domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "missing currency", nil)
	}
	return payload, nil
}

func winOperationAndAmount(payload callbackPayload) (domain.WalletOperation, int64, error) {
	winMinor, err := amountToMinor(payload.WinAmount)
	if err != nil {
		return "", 0, err
	}
	if strings.EqualFold(strings.TrimSpace(payload.Type), "bet_win") {
		betMinor, err := amountToMinor(payload.BetAmount)
		if err != nil {
			return "", 0, err
		}
		net := winMinor - betMinor
		if net < 0 {
			return domain.WalletOperationDebit, -net, nil
		}
		return domain.WalletOperationCredit, net, nil
	}
	return domain.WalletOperationCredit, winMinor, nil
}

func amountToMinor(value string) (int64, error) {
	amount, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return 0, domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid amount", err)
	}
	if amount.IsNegative() {
		return 0, domain.NewWalletError(domain.WalletErrorInvalidRequest, "negative amount", nil)
	}
	return amount.Mul(decimal.NewFromInt(100)).IntPart(), nil
}

func minorToAmount(value int64) string {
	return decimal.NewFromInt(value).Div(decimal.NewFromInt(100)).StringFixed(2)
}

func callbackTimestamp(payload callbackPayload) time.Time {
	for _, value := range []any{payload.Timestamp, payload.SettledTime, payload.BetTime} {
		if t, ok := parseTimeValue(value); ok {
			return t
		}
	}
	return time.Now()
}

func parseTimeValue(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case float64:
		if typed > 0 {
			return time.UnixMilli(int64(typed)), true
		}
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return time.Time{}, false
		}
		if millis, err := strconv.ParseInt(typed, 10, 64); err == nil {
			if millis > 1_000_000_000_000 {
				return time.UnixMilli(millis), true
			}
			return time.Unix(millis, 0), true
		}
		if parsed, err := time.Parse(time.RFC3339, typed); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func normalizeOperationName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	switch value {
	case "player", "info", "player-info", "player/info":
		return "player-info"
	case "wallet-balance", "wallet/balance":
		return "balance"
	case "wallet-bet", "wallet/bet":
		return "bet"
	case "wallet-win", "wallet/win":
		return "win"
	case "wallet-cancel", "wallet/cancel":
		return "cancel"
	default:
		return value
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

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
