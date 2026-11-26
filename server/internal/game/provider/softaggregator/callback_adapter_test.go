package softaggregator

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"casino/internal/game/domain"
)

func callbackRequest(action string, timestamp time.Time, salt string) domain.ProviderCallbackRequest {
	ts := strconv.FormatInt(timestamp.Unix(), 10)
	return domain.ProviderCallbackRequest{
		ProviderCode: Code,
		Method:       http.MethodGet,
		ClientIP:     "127.0.0.1",
		ReceivedAt:   timestamp,
		Query: map[string][]string{
			"username":  {"sa_1"},
			"currency":  {"MXN"},
			"action":    {action},
			"call_id":   {"call-1"},
			"timestamp": {ts},
			"key":       {CallbackKey(ts, salt)},
		},
	}
}

func TestCallbackAdapterValidateSuccess(t *testing.T) {
	now := time.Unix(1720000000, 0)
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	require.NoError(t, adapter.ValidateRequest(context.Background(), callbackRequest("balance", now, "salt")))
}

func TestCallbackAdapterRejectsInvalidSignature(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("balance", now, "salt")
	req.Query["key"] = []string{"bad"}
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	err := adapter.ValidateRequest(context.Background(), req)
	require.Error(t, err)
	var walletErr *domain.WalletError
	require.True(t, errors.As(err, &walletErr))
	require.Equal(t, domain.WalletErrorInvalidSignature, walletErr.Code)
}

func TestCallbackAdapterRejectsExpiredTimestamp(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("balance", now.Add(-time.Minute), "salt")
	req.ReceivedAt = now
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	err := adapter.ValidateRequest(context.Background(), req)
	require.Error(t, err)
	var walletErr *domain.WalletError
	require.True(t, errors.As(err, &walletErr))
	require.Equal(t, domain.WalletErrorInvalidTimestamp, walletErr.Code)
}

func TestCallbackAdapterParseBalance(t *testing.T) {
	now := time.Unix(1720000000, 0)
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	cmd, err := adapter.ParseCommand(context.Background(), callbackRequest("balance", now, "salt"))
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationBalance, cmd.Operation)
	require.Equal(t, "sa_1", cmd.ProviderPlayerID)
}

func TestCallbackAdapterParseDebit(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("debit", now, "salt")
	req.Query["amount"] = []string{"125"}
	req.Query["round_id"] = []string{"round-1"}
	req.Query["game_id"] = []string{"provider/Game"}
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationDebit, cmd.Operation)
	require.Equal(t, int64(125), cmd.AmountMinor)
	require.NotContains(t, string(cmd.RawPayload), req.QueryValue("key"))
}

func TestCallbackAdapterParseCredit(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("credit", now, "salt")
	req.Query["amount"] = []string{"250"}
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationCredit, cmd.Operation)
	require.Equal(t, int64(250), cmd.AmountMinor)
}

func TestCallbackAdapterParseRefund(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("refund", now, "salt")
	req.Query["amount"] = []string{"250"}
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationRefund, cmd.Operation)
	require.Equal(t, "REFUND", cmd.ProviderAction)
	require.Equal(t, int64(250), cmd.AmountMinor)
}

func TestCallbackAdapterFreeSpinDebitDoesNotDeduct(t *testing.T) {
	now := time.Unix(1720000000, 0)
	req := callbackRequest("debit", now, "salt")
	req.Query["amount"] = []string{"250"}
	req.Query["type"] = []string{"bonus_fs"}
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, int64(0), cmd.AmountMinor)
}

func TestCallbackAdapterMapsInsufficientBalanceError(t *testing.T) {
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	resp := adapter.BuildErrorResponse(context.Background(), domain.ProviderCallbackRequest{}, &domain.WalletError{
		Code:         domain.WalletErrorInsufficientBalance,
		BalanceMinor: 99,
	})
	body := resp.Body.(callbackResponse)
	require.Equal(t, 1, body.Error)
	require.Equal(t, int64(99), body.Balance)
}

func TestCallbackAdapterMapsGenericError(t *testing.T) {
	adapter := NewCallbackAdapter(Config{Enabled: true, SaltKey: "salt", MaxTimestampDrift: 30 * time.Second})
	resp := adapter.BuildErrorResponse(context.Background(), domain.ProviderCallbackRequest{}, errors.New("bad"))
	body := resp.Body.(callbackResponse)
	require.Equal(t, 2, body.Error)
	require.Equal(t, int64(0), body.Balance)
}
