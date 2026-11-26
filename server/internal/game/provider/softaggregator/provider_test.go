package softaggregator

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"casino/internal/game/domain"
)

func TestGameListMapping(t *testing.T) {
	game := mapGame(gameListItem{
		Name:                "Game Name",
		Type:                "video-slots",
		Category:            "pragmaticplay",
		IDHash:              "provider/GameName",
		FreeroundsSupported: true,
		PlayForFunSupported: true,
		Image:               "https://cdn/game.webp",
	})
	require.Equal(t, Code, game.ProviderCode)
	require.Equal(t, "provider/GameName", game.ProviderGameID)
	require.Equal(t, "slot", game.GameType)
	require.True(t, game.SupportsDemo)
	require.True(t, game.SupportsFreeSpins)
	require.JSONEq(t, `{"id":0,"name":"Game Name","type":"video-slots","category":"pragmaticplay","id_hash":"provider/GameName","freerounds_supported":true,"play_for_fun_supported":true,"image":"https://cdn/game.webp"}`, string(game.RawMetadata))
}

func TestProviderGetGamesUsesDocumentedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "getGameList", payload["method"])
		require.Equal(t, "login", payload["api_login"])
		require.Equal(t, "password", payload["api_password"])
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": 0,
			"response": []map[string]any{{
				"id": 1, "name": "Game", "type": "video-slots", "category": "studio", "id_hash": "studio/Game",
			}},
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	games, err := provider.GetGames(context.Background(), domain.GetGamesRequest{})
	require.NoError(t, err)
	require.Len(t, games, 1)
	require.Equal(t, "studio/Game", games[0].ProviderGameID)
}

func TestProviderConnectionUsesGameListOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "getGameList", payload["method"])
		_ = json.NewEncoder(w).Encode(map[string]any{"error": 0, "response": []any{}})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	require.NoError(t, provider.TestConnection(context.Background()))
}

func TestProviderGetGamesReadsFullLargeCatalogResponse(t *testing.T) {
	items := make([]map[string]any, 0, 9000)
	for i := 0; i < 9000; i++ {
		items = append(items, map[string]any{
			"id":                     i + 1,
			"name":                   "Large Catalog Game",
			"type":                   "video-slots",
			"category":               "studio",
			"id_hash":                "studio/game",
			"image":                  "https://images.example.com/statics/production/publics/images/games/large-catalog-game-with-long-path-and-query.png",
			"freerounds_supported":   false,
			"play_for_fun_supported": true,
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":    0,
			"response": items,
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	games, err := provider.GetGames(context.Background(), domain.GetGamesRequest{})
	require.NoError(t, err)
	require.Len(t, games, len(items))
}

func TestProviderCreatePlayer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "createPlayer", payload["method"])
		require.Equal(t, "sa_1", payload["user_username"])
		require.Equal(t, "MXN", payload["currency"])
		_ = json.NewEncoder(w).Encode(map[string]any{"error": 0, "response": map[string]any{}})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	player, err := provider.CreatePlayer(context.Background(), domain.CreatePlayerRequest{
		UserID: "1", ExternalPlayerID: "sa_1", ProviderPlayerPassword: "secret", Currency: "MXN",
	})
	require.NoError(t, err)
	require.Equal(t, "sa_1", player.ProviderPlayerID)
}

func TestProviderLaunchURLMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "getGame", payload["method"])
		require.Equal(t, "studio/Game", payload["gameid"])
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": 0,
			"response": map[string]any{
				"gameurl": "https://launch.example/game",
			},
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	result, err := provider.LaunchGame(context.Background(), domain.LaunchGameRequest{
		ProviderGameID: "studio/Game", ProviderPlayerID: "sa_1", ProviderPlayerPassword: "secret", Currency: "MXN", Language: "es",
	})
	require.NoError(t, err)
	require.Equal(t, "https://launch.example/game", result.LaunchURL)
}

func TestProviderDemoLaunchURLMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "getGameDemo", payload["method"])
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": 0,
			"response": map[string]any{
				"gameurl": "https://launch.example/demo",
			},
		})
	}))
	defer server.Close()

	provider := NewProvider(Config{Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password"}, nil)
	result, err := provider.LaunchDemo(context.Background(), domain.LaunchDemoRequest{ProviderGameID: "studio/Game", Currency: "MXN", Language: "es"})
	require.NoError(t, err)
	require.Equal(t, "https://launch.example/demo", result.LaunchURL)
}

func TestProviderDisabledConfiguration(t *testing.T) {
	provider := NewProvider(Config{Enabled: false, APILogin: "login", APIPassword: "password"}, nil)
	_, err := provider.GetGames(context.Background(), domain.GetGamesRequest{})
	require.ErrorIs(t, err, domain.ErrProviderDisabled)
}

func TestProviderTimeoutErrorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	provider := NewProvider(Config{
		Enabled: true, APIBaseURL: server.URL, APILogin: "login", APIPassword: "password", RequestTimeout: time.Millisecond,
	}, nil)
	_, err := provider.GetGames(context.Background(), domain.GetGamesRequest{})
	require.Error(t, err)
}

func TestRedactSensitive(t *testing.T) {
	redacted := RedactSensitive(`{"api_login":"api-user","api_password":"api-pass","salt_key":"salt-secret","url":"https://x?key=abc&token=def"}`)
	require.NotContains(t, redacted, "api-user")
	require.NotContains(t, redacted, "api-pass")
	require.NotContains(t, redacted, "salt-secret")
	require.NotContains(t, redacted, "abc")
	require.NotContains(t, redacted, "def")
}

func TestProviderErrorIsNotSupportedForBalance(t *testing.T) {
	provider := NewProvider(Config{Enabled: true}, nil)
	_, err := provider.GetProviderBalance(context.Background())
	require.True(t, errors.Is(err, domain.ErrNotSupported))
}
