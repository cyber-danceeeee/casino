package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"casino/internal/game/domain"
)

type testProvider struct{ code string }

func (p testProvider) Code() string                           { return p.code }
func (p testProvider) CurrencyMapping() domain.CurrencyMapper { return testCurrencyMapper{} }
func (p testProvider) GetGames(context.Context, domain.GetGamesRequest) ([]domain.Game, error) {
	return nil, nil
}
func (p testProvider) CreatePlayer(context.Context, domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	return nil, nil
}
func (p testProvider) LaunchGame(context.Context, domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	return nil, nil
}
func (p testProvider) LaunchDemo(context.Context, domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	return nil, nil
}
func (p testProvider) GetProviderBalance(context.Context) (*domain.ProviderBalance, error) {
	return nil, domain.ErrNotSupported
}

type testCurrencyMapper struct{}

func (testCurrencyMapper) ToProviderCurrency(_, currency string) string { return currency }
func (testCurrencyMapper) ToPlatformCurrency(currency string) string    { return currency }

type testCallbackAdapter struct{ code string }

func (a testCallbackAdapter) Code() string { return a.code }
func (a testCallbackAdapter) ValidateRequest(context.Context, domain.ProviderCallbackRequest) error {
	return nil
}
func (a testCallbackAdapter) ParseCommand(context.Context, domain.ProviderCallbackRequest) (*domain.WalletCommand, error) {
	return nil, nil
}
func (a testCallbackAdapter) BuildSuccessResponse(context.Context, domain.WalletCommand, domain.WalletResult) (*domain.ProviderCallbackResponse, error) {
	return nil, nil
}
func (a testCallbackAdapter) BuildErrorResponse(context.Context, domain.ProviderCallbackRequest, error) *domain.ProviderCallbackResponse {
	return nil
}

func TestProviderRegistry(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, registry.Register(testProvider{code: "SoftAggregator"}))
	_, err := registry.Get("softaggregator")
	require.NoError(t, err)
	require.Len(t, registry.List(), 1)
	require.ErrorIs(t, registry.Register(testProvider{code: "softaggregator"}), domain.ErrDuplicateProvider)
	_, err = registry.Get("missing")
	require.ErrorIs(t, err, domain.ErrProviderNotFound)
}

func TestCallbackRegistry(t *testing.T) {
	registry := NewCallbackRegistry()
	require.NoError(t, registry.Register(testCallbackAdapter{code: "SoftAggregator"}))
	_, err := registry.Get("softaggregator")
	require.NoError(t, err)
	require.Len(t, registry.List(), 1)
	require.ErrorIs(t, registry.Register(testCallbackAdapter{code: "softaggregator"}), domain.ErrDuplicateCallback)
	_, err = registry.Get("missing")
	require.ErrorIs(t, err, domain.ErrCallbackNotFound)
}
