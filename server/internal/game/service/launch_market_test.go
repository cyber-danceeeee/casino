package service

import (
	"casino/internal/game/domain"
	"casino/internal/game/provider"
	"casino/internal/game/repository"
	"casino/internal/model"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type marketLaunchProvider struct {
	calls int
	last  domain.LaunchDemoRequest
}

func (p *marketLaunchProvider) Code() string                           { return "test-brand" }
func (p *marketLaunchProvider) CurrencyMapping() domain.CurrencyMapper { return nil }
func (p *marketLaunchProvider) GetGames(context.Context, domain.GetGamesRequest) ([]domain.Game, error) {
	p.calls++
	return nil, nil
}
func (p *marketLaunchProvider) CreatePlayer(context.Context, domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	p.calls++
	return &domain.ProviderPlayer{}, nil
}
func (p *marketLaunchProvider) LaunchGame(context.Context, domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	p.calls++
	return &domain.LaunchGameResult{LaunchURL: "https://play.example.com"}, nil
}
func (p *marketLaunchProvider) LaunchDemo(_ context.Context, r domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	p.calls++
	p.last = r
	return &domain.LaunchGameResult{LaunchURL: "https://play.example.com/demo"}, nil
}
func (p *marketLaunchProvider) GetProviderBalance(context.Context) (*domain.ProviderBalance, error) {
	return nil, nil
}
func launchMarketDB(t *testing.T) (*gorm.DB, *LaunchService, *marketLaunchProvider) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	for _, modelValue := range []any{&model.GameGameProviders{}, &model.GameGames{}, &model.GameUsers{}, &model.GameUserWallets{}, &model.GameGameLaunchLogs{}} {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(modelValue))
		columns := []string{}
		for _, field := range stmt.Schema.Fields {
			if field.DBName == "" {
				continue
			}
			kind := "TEXT"
			if field.GORMDataType == schema.Int || field.GORMDataType == schema.Uint || field.GORMDataType == schema.Bool {
				kind = "INTEGER"
			} else if field.GORMDataType == schema.Float {
				kind = "NUMERIC"
			} else if field.GORMDataType == schema.Time {
				kind = "DATETIME"
			}
			if field.DBName == "id" {
				kind = "INTEGER PRIMARY KEY"
			}
			columns = append(columns, "`"+field.DBName+"` "+kind)
		}
		require.NoError(t, db.Exec("CREATE TABLE `"+stmt.Schema.Table+"` ("+strings.Join(columns, ",")+")").Error)
	}
	for _, sql := range []string{
		`CREATE TABLE game_site_config (id INTEGER PRIMARY KEY,country_code TEXT,currency_code TEXT,deleted_at DATETIME)`,
		`INSERT INTO game_site_config VALUES(1,'PK','PKR',NULL)`,
		`CREATE TABLE game_country_configs (id INTEGER PRIMARY KEY,site_id INTEGER,country_code TEXT,currency_code TEXT,currency_symbol TEXT,currency_decimal_places INTEGER,default_language TEXT,supported_languages TEXT,timezone TEXT,status INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_country_configs VALUES(1,0,'PK','PKR','Rs',2,'en','en','Asia/Karachi',1,NULL)`,
	} {
		require.NoError(t, db.Exec(sql).Error)
	}
	supplier := model.GameGameProviders{SiteID: 0, ProviderCode: "test-brand", Name: "Test supplier", Status: 1, SupportedCountries: "PK", SupportedCurrencies: "PKR"}
	supplier.ID = 100
	require.NoError(t, db.Create(&supplier).Error)
	game := model.GameGames{SiteID: 1, ProviderID: 100, GameCode: "test-game", ProviderGameCode: "upstream-game", Name: "Test game", Slug: "test-game", Status: 1, ProviderEnabled: true, PlatformEnabled: true, SupportsDemo: true}
	game.ID = 101
	require.NoError(t, db.Create(&game).Error)
	user := model.GameUsers{SiteID: 1, UserNo: "P100", Phone: "3001234567", CountryCode: "PK", Language: "en", InviteCode: "INV100"}
	user.ID = 1
	require.NoError(t, db.Create(&user).Error)
	wallet := model.GameUserWallets{SiteID: 1, UserID: 1, Currency: "PKR"}
	wallet.ID = 1
	require.NoError(t, db.Create(&wallet).Error)
	p := &marketLaunchProvider{}
	registry := provider.NewRegistry()
	require.NoError(t, registry.Register(p))
	return db, NewLaunchService(registry, repository.New(db)), p
}
func TestCrossMarketAndUnpublishedLaunchesNeverCallSupplier(t *testing.T) {
	for _, tc := range []struct{ name, sql, country, currency string }{
		{name: "country override", country: "CA"}, {name: "currency override", currency: "CAD"},
		{name: "supplier region", sql: "UPDATE game_game_providers SET supported_countries='CA'"},
		{name: "supplier currency", sql: "UPDATE game_game_providers SET supported_currencies='CAD'"},
		{name: "game region", sql: "UPDATE game_games SET supported_countries='CA'"},
		{name: "game currency", sql: "UPDATE game_games SET supported_currencies='CAD'"},
		{name: "unpublished", sql: "UPDATE game_games SET platform_enabled=0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, service, p := launchMarketDB(t)
			if tc.sql != "" {
				require.NoError(t, db.Exec(tc.sql).Error)
			}
			_, err := service.Launch(context.Background(), LaunchRequest{ProviderCode: p.Code(), GameID: "upstream-game", UserID: 1, Country: tc.country, Currency: tc.currency, Demo: true})
			require.Error(t, err)
			require.Zero(t, p.calls)
		})
	}
}
func TestLaunchDefaultsToPublishedDeploymentMarket(t *testing.T) {
	_, service, p := launchMarketDB(t)
	result, err := service.Launch(context.Background(), LaunchRequest{ProviderCode: p.Code(), GameID: "upstream-game", UserID: 1, Demo: true})
	require.NoError(t, err)
	require.Equal(t, "https://play.example.com/demo", result.LaunchURL)
	require.Equal(t, "PK", p.last.Country)
	require.Equal(t, "PKR", p.last.Currency)
	require.Equal(t, 1, p.calls)
}

func TestManualGameMaintenanceBlocksDemoAndRealLaunchBeforeSupplierCalls(t *testing.T) {
	for _, demo := range []bool{false, true} {
		db, service, supplier := launchMarketDB(t)
		require.NoError(t, db.Exec("UPDATE game_games SET status=2 WHERE id=101").Error)
		_, err := service.Launch(context.Background(), LaunchRequest{ProviderCode: supplier.Code(), GameID: "upstream-game", UserID: 1, Demo: demo})
		var walletError *domain.WalletError
		require.ErrorAs(t, err, &walletError)
		require.Equal(t, domain.WalletErrorGameMaintenance, walletError.Code)
		require.Zero(t, supplier.calls)
	}
}
