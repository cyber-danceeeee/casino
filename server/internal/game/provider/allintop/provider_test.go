package allintop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"casino/internal/game/domain"
	"github.com/stretchr/testify/require"
)

func TestCatalogLaunchAndSharedTokenCache(t *testing.T) {
	var tokens atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/auth/createtoken" {
			tokens.Add(1)
			require.Equal(t, http.MethodPost, r.Method)
			json.NewEncoder(w).Encode(map[string]any{"token": "test-token", "expiration": time.Now().Add(time.Hour).Unix()})
			return
		}
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		var message any
		switch r.URL.Path {
		case "/vendors/list":
			message = []Vendor{{VendorCode: "slot-one", Type: 2}, {VendorCode: "live-two", Type: 1}}
		case "/games/list":
			var in map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&in))
			message = []map[string]any{{"vendorCode": in["vendorCode"], "gameCode": "same-code", "gameName": "Game", "thumbnail": "https://example.com/game.webp", "underMaintenance": in["vendorCode"] == "live-two"}}
		case "/game/launch-url":
			var in map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&in))
			require.Equal(t, "slot-one", in["vendorCode"])
			require.Equal(t, "same-code", in["gameCode"])
			require.Equal(t, "player1", in["userCode"])
			message = "https://games.example.com/play"
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "message": message, "errorCode": 0})
	}))
	defer upstream.Close()
	cfg := Config{Enabled: true, APIBaseURL: upstream.URL, ClientID: "id", ClientSecret: "private-secret"}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := NewProvider(cfg, upstream.Client()).TestConnection(context.Background()); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	require.EqualValues(t, 1, tokens.Load())
	p := NewProvider(cfg, upstream.Client())
	games, err := p.GetGames(context.Background(), domain.GetGamesRequest{})
	require.NoError(t, err)
	require.Len(t, games, 2)
	require.Equal(t, "slot-one/same-code", games[0].ProviderGameID)
	require.NotEqual(t, games[0].ProviderGameID, games[1].ProviderGameID)
	require.True(t, games[0].IsEnabled)
	require.False(t, games[1].IsEnabled)
	_, err = p.CreatePlayer(context.Background(), domain.CreatePlayerRequest{ExternalPlayerID: "player1", Currency: "IDR"})
	require.NoError(t, err)
	result, err := p.LaunchGame(context.Background(), domain.LaunchGameRequest{ProviderGameID: games[0].ProviderGameID, ProviderPlayerID: "player1", Country: "ID", Currency: "IDR"})
	require.NoError(t, err)
	require.Equal(t, "https://games.example.com/play", result.LaunchURL)
	_, err = p.LaunchGame(context.Background(), domain.LaunchGameRequest{Country: "PK", Currency: "PKR"})
	require.Error(t, err)
	require.Empty(t, p.CurrencyMapping().ToPlatformCurrency("IDRK"))
	require.EqualValues(t, 1, tokens.Load())
}
func TestAuthFailureBacksOffAndRedactsResponse(t *testing.T) {
	var requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(403)
		w.Write([]byte("sensitive credential upstream response"))
	}))
	defer s.Close()
	c := Config{Enabled: true, APIBaseURL: s.URL, ClientID: "id", ClientSecret: "secret"}
	err := NewProvider(c, s.Client()).TestConnection(context.Background())
	require.ErrorContains(t, err, "HTTP 403")
	require.NotContains(t, err.Error(), "sensitive")
	err = NewProvider(c, s.Client()).TestConnection(context.Background())
	require.ErrorContains(t, err, "cooling down")
	require.EqualValues(t, 1, requests.Load())
}
func TestIncompleteCatalogDoesNotReturnPartialGames(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/createtoken":
			json.NewEncoder(w).Encode(map[string]any{"token": "t", "expiration": time.Now().Add(time.Hour).Unix()})
		case "/vendors/list":
			json.NewEncoder(w).Encode(map[string]any{"success": true, "errorCode": 0, "message": []Vendor{{VendorCode: "one"}, {VendorCode: "two"}}})
		default:
			var in map[string]string
			json.NewDecoder(r.Body).Decode(&in)
			if in["vendorCode"] == "two" {
				w.WriteHeader(503)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"success": true, "errorCode": 0, "message": []map[string]any{{"gameCode": "1", "gameName": "Game"}}})
		}
	}))
	defer s.Close()
	p := NewProvider(Config{Enabled: true, APIBaseURL: s.URL, ClientID: "id", ClientSecret: "secret"}, s.Client())
	games, err := p.GetGames(context.Background(), domain.GetGamesRequest{})
	require.Error(t, err)
	require.Nil(t, games)
}
func TestCallbackAuthenticationPrecisionCancellationAndBatch(t *testing.T) {
	a := NewCallbackAdapter(Config{Enabled: true, ClientID: "client", ClientSecret: "secret", MaxTransactionAmountMinor: 1000000})
	body := `{"userCode":"player","vendorCode":"slot-test","gameCode":"a","historyId":9007199254740993,"roundId":"r1","transactionCode":"t1","isFinished":false,"isCanceled":false,"amount":-12.34}`
	r := domain.ProviderCallbackRequest{Method: "POST", Operation: "transaction", Path: "/api/v1/game/providers/allintop/callback/api/transaction", Body: []byte(body), Headers: http.Header{}}
	require.Error(t, a.ValidateRequest(context.Background(), r))
	httpReq := httptest.NewRequest("POST", "/", nil)
	httpReq.SetBasicAuth("client", "secret")
	r.Headers = httpReq.Header
	require.NoError(t, a.ValidateRequest(context.Background(), r))
	c, err := a.ParseCommand(context.Background(), r)
	require.NoError(t, err)
	require.Equal(t, int64(1234), c.AmountMinor)
	require.Equal(t, "9007199254740993", c.ProviderHistoryID)
	require.Equal(t, domain.WalletOperationDebit, c.Operation)
	r.Body = []byte(`{"userCode":"player","vendorCode":"slot-test","gameCode":"a","historyId":1,"roundId":"r1","transactionCode":"c1","isFinished":true,"isCanceled":true,"amount":9999999999999999999999}`)
	c, err = a.ParseCommand(context.Background(), r)
	require.NoError(t, err)
	require.Equal(t, domain.WalletOperationCancel, c.Operation)
	require.Zero(t, c.AmountMinor)
	r.Body = []byte(`{"userCode":"player","vendorCode":"v","gameCode":"a","historyId":1,"roundId":"r","transactionCode":"t","isFinished":true,"isCanceled":false,"amount":0.001}`)
	_, err = a.ParseCommand(context.Background(), r)
	require.Error(t, err)
	r.Operation = "batch-transactions"
	r.Body = []byte(`{"userCode":"player","transactions":[` + body + `]}`)
	commands, err := a.ParseBatchCommands(context.Background(), r)
	require.NoError(t, err)
	require.Len(t, commands, 1)
	r.Body = []byte(`{"userCode":"other","transactions":[` + body + `]}`)
	_, err = a.ParseBatchCommands(context.Background(), r)
	require.Error(t, err)
	response, err := a.BuildSuccessResponse(context.Background(), domain.WalletCommand{}, domain.WalletResult{BalanceMinor: 12345})
	require.NoError(t, err)
	encoded, err := json.Marshal(response.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true,"message":123.45,"errorCode":0}`, string(encoded))
	for code, want := range map[domain.WalletErrorCode]int{domain.WalletErrorDuplicateTransaction: 6, domain.WalletErrorRoundClosed: 7, domain.WalletErrorInsufficientBalance: 4, domain.WalletErrorPlayerNotFound: 2} {
		resp := a.BuildErrorResponse(context.Background(), r, domain.NewWalletError(code, "private internal info", nil))
		require.Equal(t, want, resp.Body.(map[string]any)["errorCode"])
		require.NotContains(t, resp.Body.(map[string]any)["message"], "private")
	}
}

func TestExcludedVendorsAndDocumentedLanguageFallback(t *testing.T) {
	var bad atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/createtoken":
			json.NewEncoder(w).Encode(map[string]any{"token": "t", "expiration": time.Now().Add(time.Hour).Unix()})
		case "/vendors/list":
			json.NewEncoder(w).Encode(map[string]any{"success": true, "errorCode": 0, "message": []Vendor{{VendorCode: "slot-pgsoft", Type: 2}, {VendorCode: "board-poker", Type: 6}}})
		default:
			var in map[string]string
			json.NewDecoder(r.Body).Decode(&in)
			if in["vendorCode"] == "board-poker" {
				bad.Add(1)
				w.WriteHeader(500)
				return
			}
			require.Equal(t, "en", in["language"])
			json.NewEncoder(w).Encode(map[string]any{"success": true, "errorCode": 0, "message": []map[string]any{{"gameCode": "1", "gameName": "Game"}}})
		}
	}))
	defer s.Close()
	p := NewProvider(Config{Enabled: true, APIBaseURL: s.URL, ClientID: "id", ClientSecret: "secret", ExcludedVendorCodes: []string{"board-poker"}}, s.Client())
	games, err := p.GetGames(context.Background(), domain.GetGamesRequest{})
	require.NoError(t, err)
	require.Len(t, games, 1)
	require.Zero(t, bad.Load())
	require.Equal(t, "id", providerLanguage("slot-pgsoft", "id"))
	require.Equal(t, "en", providerLanguage("casino-pragmatic", "id"))
}
