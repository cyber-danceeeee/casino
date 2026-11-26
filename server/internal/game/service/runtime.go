package service

import (
	"gorm.io/gorm"

	"casino/internal/database"
	"casino/internal/game/connections"
	"casino/internal/game/domain"
	"casino/internal/game/repository"
	"casino/internal/hotranking"
)

type Runtime struct {
	ProviderRegistry domain.ProviderRegistry
	CallbackRegistry domain.ProviderCallbackRegistry
	Repository       *repository.Repository
	GameService      *GameService
	LaunchService    *LaunchService
	WalletService    *WalletCallbackService
}

func NewRuntime(db *gorm.DB) *Runtime {
	store := connections.New(db)
	providerRegistry := connections.NewRegistry(store)
	callbackRegistry := connections.NewCallbackRegistry(store)
	repo := repository.New(db)
	games := NewGameService(providerRegistry, repo, nil)
	games.connectionMetadata = store.Description

	return &Runtime{
		ProviderRegistry: providerRegistry,
		CallbackRegistry: callbackRegistry,
		Repository:       repo,
		GameService:      games,
		LaunchService:    newLaunchService(providerRegistry, repo),
		WalletService:    NewWalletCallbackService(repo),
	}
}

func newLaunchService(registry domain.ProviderRegistry, repo *repository.Repository) *LaunchService {
	service := NewLaunchService(registry, repo)
	service.SetHotRanker(hotranking.New(database.GetRedisCli()))
	return service
}

func stringDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
