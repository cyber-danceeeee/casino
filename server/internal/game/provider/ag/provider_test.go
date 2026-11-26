package ag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"casino/internal/game/domain"
)

func TestSignatureMatchesDocumentRule(t *testing.T) {
	body := []byte(`{"merchantSession":"9259c473c8a94f68a7ec1e766d33b297","currencyCode":"BR_BRL"}`)
	got := sign("xxxxxxxxxxxx", "123123", "1744897560043", "7c83a1c5f989498d8c73a2d54d66657f", body)
	require.Len(t, got, 64)
	require.True(t, validSignature("xxxxxxxxxxxx", "123123", "1744897560043", "7c83a1c5f989498d8c73a2d54d66657f", body, got))
}

func TestProviderCurrencyMapping(t *testing.T) {
	cfg := Config{DefaultCurrency: "PKR"}
	mapper := currencyMapper{}
	require.Equal(t, "US_USD", mapper.ToProviderCurrency("US", "USD"))
	require.Equal(t, "PK_PKR", mapper.ToProviderCurrency("PK", "PKR"))
	require.Equal(t, "PKR", cfg.SyncCurrency(""))
	require.Equal(t, "BR_BRL", mapper.ToProviderCurrency("BR", "BRL"))
	require.Equal(t, "PKR", mapper.ToPlatformCurrency("PK_PKR"))
}

func TestDeduplicateGamesPrefersNonPromotionalCategory(t *testing.T) {
	games := deduplicateGames([]domain.Game{
		{ProviderGameID: "same", Name: "Hot Copy", Category: "HOT_PG"},
		{ProviderGameID: "same", Name: "Catalog Copy", Category: "SLOT_PG"},
		{ProviderGameID: "other", Name: "Other", Category: "NEW_PG"},
	})
	require.Len(t, games, 2)
	require.Equal(t, "SLOT_PG", games[0].Category)
	require.Equal(t, "other", games[1].ProviderGameID)
}

func TestGetGamesReadsCategoriesAndGames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NotEmpty(t, r.Header.Get("X-SIGN"))
		switch r.URL.Path {
		case "/game/v5/categories":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true, "code": "C10000", "msg": "ok",
				"data": map[string]any{"categoryDataList": []map[string]any{{"categoryName": "Bingo", "categoryCode": "BINGO"}}, "totalSize": 1},
			})
		case "/game/v5/games":
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "BINGO", payload["categoryCode"])
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true, "code": "C10000", "msg": "ok",
				"data": map[string]any{"categoryCode": "BINGO", "totalSize": 1, "gameItemList": []map[string]any{{
					"gameName": "Bingo Game", "gameCode": "BINGO_001", "status": 1, "gameType": "BINGO",
				}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, MerchantCode: "merchant", SecretKey: "secret"}, nil)
	games, err := provider.GetGames(context.Background(), domain.GetGamesRequest{Currency: "PKR", Language: "en"})
	require.NoError(t, err)
	require.Len(t, games, 1)
	require.Equal(t, "BINGO_001", games[0].ProviderGameID)
	require.Equal(t, "bingo", games[0].GameType)
	require.Equal(t, "BINGO", games[0].Category)
}

func TestProviderConnectionUsesCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/game/v5/categories", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "PKR", payload["currencyCode"])
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true, "code": "C10000", "msg": "ok",
			"data": map[string]any{"categoryDataList": []any{}, "totalSize": 0},
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, MerchantCode: "merchant", SecretKey: "secret", DefaultCurrency: "PKR"}, nil)
	require.NoError(t, provider.TestConnection(context.Background()))
}

func TestLaunchGameMapsDocumentedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/game/v5/game/url", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "FORTUNE_DRAGON", payload["gameCode"])
		require.Equal(t, "ag_1", payload["playerId"])
		require.Equal(t, "US_USD", payload["currencyCode"])
		require.Equal(t, "h5.example", payload["domain"])
		_, hasSubMerchant := payload["subMerchantCode"]
		require.False(t, hasSubMerchant)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true, "code": "C10000", "msg": "ok",
			"data": map[string]any{"gameCode": "FORTUNE_DRAGON", "playerId": "ag_1", "gameUrl": "https://launch.example/ag"},
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, MerchantCode: "merchant", SecretKey: "secret"}, nil)
	result, err := provider.LaunchGame(context.Background(), domain.LaunchGameRequest{
		ProviderGameID: "FORTUNE_DRAGON", ProviderPlayerID: "ag_1", Currency: "USD", Country: "US", Language: "en", Device: "mobile", ReturnURL: "https://h5.example",
	})
	require.NoError(t, err)
	require.Equal(t, "https://launch.example/ag", result.LaunchURL)
	require.Contains(t, result.ProviderRequestTrace, "FORTUNE_DRAGON")
}

func TestLaunchGameRejectsMissingURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "code": "C10000", "msg": "ok", "data": map[string]any{}})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, MerchantCode: "merchant", SecretKey: "secret"}, nil)
	_, err := provider.LaunchGame(context.Background(), domain.LaunchGameRequest{ProviderGameID: "game", ProviderPlayerID: "ag_1", Currency: "PKR", Language: "en"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing data.gameUrl")
}

func TestCallbackAdapterParseBetWinNetCredit(t *testing.T) {
	now := time.Unix(1720000000, 0)
	body := []byte(`{"reqTraceId":"trace","playerId":"ag_1","currencyCode":"PK_PKR","gameCode":"FORTUNE_DRAGON","roundId":"round","transactionId":"txn","betId":"bet","betAmount":"100.00","winAmount":"140.50","type":"bet_win","settledTime":1720000000000}`)
	req := signedCallbackRequest("win", now, body)
	adapter := NewCallbackAdapter(Config{Enabled: true, MerchantCode: "merchant", SecretKey: "secret", MaxTimestampDrift: time.Minute})
	require.NoError(t, adapter.ValidateRequest(context.Background(), req))
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationCredit, cmd.Operation)
	require.Equal(t, int64(4050), cmd.AmountMinor)
	require.Equal(t, "PKR", cmd.Currency)
	require.Equal(t, "FORTUNE_DRAGON", cmd.ProviderGameID)
}

func TestCallbackAdapterParseCancelAsRollback(t *testing.T) {
	now := time.Unix(1720000000, 0)
	body := []byte(`{"reqTraceId":"trace","playerId":"ag_1","currencyCode":"PK_PKR","gameCode":"FORTUNE_DRAGON","roundId":"round","transactionId":"cancel-txn","betId":"bet","merchantBetId":"original-txn","betAmount":"100.00","cancelType":"BET"}`)
	req := signedCallbackRequest("cancel", now, body)
	adapter := NewCallbackAdapter(Config{Enabled: true, MerchantCode: "merchant", SecretKey: "secret", MaxTimestampDrift: time.Minute})
	cmd, err := adapter.ParseCommand(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationRollback, cmd.Operation)
	require.Equal(t, "cancel-txn", cmd.ProviderTransactionID)
	require.Equal(t, "original-txn", cmd.RelatedTransactionID)
}

func signedCallbackRequest(operation string, now time.Time, body []byte) domain.ProviderCallbackRequest {
	ts := timestampMillis(now)
	n := strings.Repeat("a", 32)
	return domain.ProviderCallbackRequest{
		ProviderCode: Code,
		Operation:    operation,
		Method:       http.MethodPost,
		ClientIP:     "127.0.0.1",
		ReceivedAt:   now,
		Body:         body,
		Headers: http.Header{
			"X-Merchant-Code":           []string{"merchant"},
			"X-Timestamp":               []string{ts},
			"X-Nonce":                   []string{n},
			"X-Content-Processing-Type": []string{signType},
			"X-Sign":                    []string{sign("secret", "merchant", ts, n, body)},
		},
	}
}
