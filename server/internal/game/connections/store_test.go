package connections

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"casino/internal/config"
	"casino/internal/game/domain"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) { config.Set(&config.Config{}); os.Exit(m.Run()) }

func connectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	old := *config.Get()
	cfg := old
	cfg.AdminAuth.MFAEncryptionKey = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	config.Set(&cfg)
	t.Cleanup(func() { config.Set(&old) })
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	require.NoError(t, db.AutoMigrate(&Record{}))
	for _, sql := range []string{
		`CREATE TABLE game_site_config (id INTEGER PRIMARY KEY,country_code TEXT,currency_code TEXT,h5_domain TEXT,deleted_at DATETIME)`,
		`INSERT INTO game_site_config VALUES(1,'PK','PKR','https://h5.example.com',NULL)`,
		`CREATE TABLE game_country_configs (country_code TEXT,country_name_en TEXT,currency_code TEXT,status INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_country_configs VALUES('PK','Pakistan','PKR',1,NULL),('CA','Canada','CAD',1,NULL)`,
		`CREATE TABLE game_game_providers (id INTEGER PRIMARY KEY AUTOINCREMENT,provider_code TEXT,site_id INTEGER,status INTEGER,name TEXT,api_base_url TEXT,environment TEXT,supported_countries TEXT,supported_currencies TEXT,callback_url TEXT,wallet_mode TEXT,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME)`,
	} {
		require.NoError(t, db.Exec(sql).Error)
	}
	return db
}
func supplierInput(code, country, currency string) Input {
	return Input{Config: Config{ProviderCode: code, AdapterType: "softaggregator", Name: code, APIBaseURL: "https://example.com", Environment: "sandbox", Enabled: true, SupportedCountries: country, SupportedCurrencies: currency}, Credentials: map[string]string{"apiLogin": "merchant-login", "apiPassword": "merchant-password-distinctive", "saltKey": "merchant-signing-key-distinctive"}, Reason: "configure supplier"}
}
func TestSecretsRemainEncryptedAndBlankEditsPreserveThem(t *testing.T) {
	db := connectionTestDB(t)
	store := New(db)
	input := supplierInput("brand-pk", "pk", "pkr")
	view, err := store.Save(input)
	require.NoError(t, err)
	require.Equal(t, uint64(1), view.Version)
	require.True(t, view.CredentialsConfigured)
	var row Record
	require.NoError(t, db.First(&row).Error)
	encoded, _ := json.Marshal(view)
	for _, secret := range input.Credentials {
		require.NotContains(t, row.CredentialsCiphertext, secret)
		require.NotContains(t, string(encoded), secret)
	}
	_, err = store.Save(input)
	require.ErrorIs(t, err, ErrConflict)
	input.Version = view.Version
	input.Name = "Updated supplier"
	input.Credentials = map[string]string{"apiPassword": ""}
	view, err = store.Save(input)
	require.NoError(t, err)
	c, values, found, err := store.load(input.ProviderCode)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "merchant-password-distinctive", values["apiPassword"])
	require.Equal(t, view.Version, c.Version)
	// Ciphertext is bound to both the account code and protocol.
	_, err = open("other-account", c.AdapterType, row.CredentialsCiphertext)
	require.Error(t, err)
}
func TestPublishedMarketAndDisableChangesApplyWithoutRestart(t *testing.T) {
	db := connectionTestDB(t)
	store := New(db)
	registry := NewRegistry(store)
	callbacks := NewCallbackRegistry(store)
	pk := supplierInput("brand-pk", "PK", "PKR")
	_, err := store.Save(pk)
	require.NoError(t, err)
	ca := supplierInput("brand-ca", "CA", "CAD")
	_, err = store.Save(ca)
	require.NoError(t, err)
	require.Equal(t, []string{"brand-pk"}, store.PublicCodes(context.Background()))
	provider, err := registry.Get("brand-pk")
	require.NoError(t, err)
	require.Equal(t, "brand-pk", provider.Code())
	require.NoError(t, db.Exec("UPDATE game_site_config SET country_code='CA',currency_code='CAD' WHERE id=1").Error)
	require.Equal(t, []string{"brand-ca"}, store.PublicCodes(context.Background()))
	ca.Enabled = false
	ca.Version = 1
	ca.Credentials = nil
	_, err = store.Save(ca)
	require.NoError(t, err)
	_, err = registry.Get("brand-ca")
	require.ErrorIs(t, err, domain.ErrProviderDisabled)
	callback, err := callbacks.Get("brand-ca")
	require.NoError(t, err)
	require.Equal(t, "brand-ca", callback.Code())
	require.Empty(t, store.PublicCodes(context.Background()))
}
func TestExplicitRulesAndCredentialsAreRequiredBeforeEnabling(t *testing.T) {
	store := New(connectionTestDB(t))
	for _, change := range []func(*Input){
		func(i *Input) { i.SupportedCountries = "" }, func(i *Input) { i.SupportedCountries = "*" }, func(i *Input) { i.SupportedCurrencies = "PK" }, func(i *Input) { delete(i.Credentials, "saltKey") }, func(i *Input) { i.Environment = "production"; i.APIBaseURL = "http://example.com" },
	} {
		input := supplierInput("invalid-account", "PK", "PKR")
		change(&input)
		_, err := store.Save(input)
		require.Error(t, err)
	}
}

func TestIndependentAccountsReuseProtocolAndReloadCredentials(t *testing.T) {
	db := connectionTestDB(t)
	store := New(db)
	registry := NewRegistry(store)
	passwords := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			w.WriteHeader(400)
			return
		}
		passwords = append(passwords, body["api_password"].(string))
		if body["method"] == "getGameList" {
			_ = json.NewEncoder(w).Encode(map[string]any{"error": 0, "response": []map[string]any{{"name": "Demo", "type": "video-slots", "id_hash": "studio/demo"}}})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{"error": 0, "response": "https://play.example.com/demo"})
		}
	}))
	defer server.Close()
	input := supplierInput("independent-brand", "PK", "PKR")
	input.APIBaseURL = server.URL
	_, err := store.Save(input)
	require.NoError(t, err)
	p, err := registry.Get(input.ProviderCode)
	require.NoError(t, err)
	games, err := p.GetGames(context.Background(), domain.GetGamesRequest{Currency: "PKR"})
	require.NoError(t, err)
	require.Len(t, games, 1)
	require.Equal(t, input.ProviderCode, games[0].ProviderCode)
	input.Version = 1
	input.Credentials = map[string]string{"apiPassword": "rotated-merchant-password"}
	_, err = store.Save(input)
	require.NoError(t, err)
	p, err = registry.Get(input.ProviderCode)
	require.NoError(t, err)
	_, err = p.GetGames(context.Background(), domain.GetGamesRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{"merchant-password-distinctive", "rotated-merchant-password"}, passwords)
}

func TestReservedLegacyCodeCannotChangeProtocol(t *testing.T) {
	store := New(connectionTestDB(t))
	input := supplierInput("ag", "PK", "PKR")
	_, err := store.Save(input)
	require.Error(t, err)
}

func TestAllInTopNacosAccountAndBatchWrapper(t *testing.T) {
	db := connectionTestDB(t)
	cfg := *config.Get()
	cfg.Game.AllInTop = config.AllInTop{Enabled: true, APIBaseURL: "https://api.example.com/api/v2", ClientID: "client", ClientSecret: "secret"}
	config.Set(&cfg)
	store := New(db)
	list, err := store.List()
	require.NoError(t, err)
	rows := list["list"].([]View)
	var found *View
	for i := range rows {
		if rows[i].ProviderCode == "allintop" {
			found = &rows[i]
		}
	}
	require.NotNil(t, found)
	require.Equal(t, "ID", found.SupportedCountries)
	require.Equal(t, "IDR", found.SupportedCurrencies)
	require.True(t, found.CredentialsConfigured)
	require.False(t, found.MarketAllowed)
	cb, ok, err := store.Callback("allintop")
	require.NoError(t, err)
	require.True(t, ok)
	batch, ok := cb.(domain.ProviderBatchCallbackAdapter)
	require.True(t, ok)
	commands, err := batch.ParseBatchCommands(context.Background(), domain.ProviderCallbackRequest{Operation: "batch-transactions", Body: []byte(`{"userCode":"u","transactions":[{"vendorCode":"v","gameCode":"g","historyId":1,"roundId":"r","transactionCode":"t","isFinished":true,"isCanceled":false,"amount":-10}]}`)})
	require.NoError(t, err)
	require.Len(t, commands, 1)
	require.Equal(t, "allintop", commands[0].ProviderCode)
	encoded, err := json.Marshal(list)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "\"clientSecret\":\"secret\"")
}
