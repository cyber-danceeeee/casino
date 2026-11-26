package domain

import "context"

type ProviderBalance struct {
	BalanceMinor int64
	Currency     string
}

// CurrencyMapper is implemented by each provider because upstream currency
// codes are not a platform-wide contract.
type CurrencyMapper interface {
	ToProviderCurrency(country, currency string) string
	ToPlatformCurrency(providerCurrency string) string
}

type GameProvider interface {
	Code() string
	CurrencyMapping() CurrencyMapper
	GetGames(ctx context.Context, req GetGamesRequest) ([]Game, error)
	CreatePlayer(ctx context.Context, req CreatePlayerRequest) (*ProviderPlayer, error)
	LaunchGame(ctx context.Context, req LaunchGameRequest) (*LaunchGameResult, error)
	LaunchDemo(ctx context.Context, req LaunchDemoRequest) (*LaunchGameResult, error)
	GetProviderBalance(ctx context.Context) (*ProviderBalance, error)
}

// ProviderConnectionTester verifies connectivity with an operation supported
// by the concrete provider instead of assuming a provider-balance endpoint.
type ProviderConnectionTester interface {
	TestConnection(ctx context.Context) error
}

type ProviderRegistry interface {
	Register(provider GameProvider) error
	Get(code string) (GameProvider, error)
	List() []GameProvider
}
