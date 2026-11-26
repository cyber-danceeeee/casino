package service

import (
	"time"

	"gorm.io/gorm"

	"casino/internal/config"
	"casino/internal/database"
	gameprovider "casino/internal/game/provider"
	"casino/internal/game/provider/ag"
	"casino/internal/game/provider/softaggregator"
	"casino/internal/game/repository"
	"casino/internal/hotranking"
)

type Runtime struct {
	ProviderRegistry *gameprovider.Registry
	CallbackRegistry *gameprovider.CallbackRegistry
	Repository       *repository.Repository
	GameService      *GameService
	LaunchService    *LaunchService
	WalletService    *WalletCallbackService
}

func NewRuntime(db *gorm.DB) *Runtime {
	providerRegistry := gameprovider.NewRegistry()
	callbackRegistry := gameprovider.NewCallbackRegistry()
	repo := repository.New(db)

	softAggregatorConfig := softAggregatorConfigFromYAML()
	providerAPIBaseURLs := map[string]string{}
	if softAggregatorConfig.Enabled {
		_ = providerRegistry.Register(softaggregator.NewProvider(softAggregatorConfig, nil))
		_ = callbackRegistry.Register(softaggregator.NewCallbackAdapter(softAggregatorConfig))
		providerAPIBaseURLs[softaggregator.Code] = softAggregatorConfig.EndpointURL()
	}
	agConfig := agConfigFromYAML()
	if agConfig.Enabled {
		_ = providerRegistry.Register(ag.NewProvider(agConfig, nil))
		_ = callbackRegistry.Register(ag.NewCallbackAdapter(agConfig))
		providerAPIBaseURLs[ag.Code] = agConfig.EndpointURL()
	}

	return &Runtime{
		ProviderRegistry: providerRegistry,
		CallbackRegistry: callbackRegistry,
		Repository:       repo,
		GameService:      NewGameService(providerRegistry, repo, providerAPIBaseURLs),
		LaunchService:    newLaunchService(providerRegistry, repo),
		WalletService:    NewWalletCallbackService(repo),
	}
}

func newLaunchService(registry *gameprovider.Registry, repo *repository.Repository) *LaunchService {
	service := NewLaunchService(registry, repo)
	service.SetHotRanker(hotranking.New(database.GetRedisCli()))
	return service
}

func agConfigFromYAML() ag.Config {
	source := config.Get().Game.AG
	return ag.Config{
		Enabled:                   source.Enabled,
		APIBaseURL:                source.APIBaseURL,
		MerchantCode:              source.MerchantCode,
		SecretKey:                 source.SecretKey,
		SubMerchantCode:           source.SubMerchantCode,
		DefaultCurrency:           source.DefaultCurrency,
		CallbackIPAllowlist:       source.CallbackIPAllowlist,
		MaxTimestampDrift:         time.Duration(intDefault(source.MaxTimestampDriftSeconds, 30)) * time.Second,
		RequestTimeout:            time.Duration(intDefault(source.RequestTimeoutSeconds, 10)) * time.Second,
		ConnectTimeout:            time.Duration(intDefault(source.ConnectTimeoutSeconds, 3)) * time.Second,
		MaxTransactionAmountMinor: int64Default(source.MaxTransactionAmountMinor, 100000000),
	}
}

func softAggregatorConfigFromYAML() softaggregator.Config {
	source := config.Get().Game.SoftAggregator
	return softaggregator.Config{
		Enabled:                   source.Enabled,
		APIBaseURL:                source.APIBaseURL,
		APILogin:                  source.APILogin,
		APIPassword:               source.APIPassword,
		SaltKey:                   source.SaltKey,
		CallbackPath:              stringDefault(source.CallbackPath, "/api/v1/game/providers/softaggregator/callback"),
		CallbackIPAllowlist:       source.CallbackIPAllowlist,
		MaxTimestampDrift:         time.Duration(intDefault(source.MaxTimestampDriftSeconds, 30)) * time.Second,
		RequestTimeout:            time.Duration(intDefault(source.RequestTimeoutSeconds, 10)) * time.Second,
		ConnectTimeout:            time.Duration(intDefault(source.ConnectTimeoutSeconds, 3)) * time.Second,
		MaxTransactionAmountMinor: int64Default(source.MaxTransactionAmountMinor, 100000000),
	}
}

func intDefault(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func int64Default(value int64, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}

func stringDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
