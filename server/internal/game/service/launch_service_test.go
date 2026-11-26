package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeLaunchRequestPlatformDefaults(t *testing.T) {
	req := normalizeLaunchRequest(LaunchRequest{ProviderCode: " SoftAggregator ", GameID: "game", UserID: 1})
	require.Equal(t, "softaggregator", req.ProviderCode)
	require.Empty(t, req.Currency)
	require.Empty(t, req.Language)
	require.Empty(t, req.Country)
	require.Equal(t, "web", req.Device)
	require.NotEmpty(t, req.SessionID)
}

func TestValidateLaunchRequestRejectsInvalidCurrency(t *testing.T) {
	err := validateLaunchRequest(LaunchRequest{ProviderCode: "softaggregator", GameID: "game", UserID: 1, Currency: "MX"})
	require.Error(t, err)
}

func TestProviderPlayerIDUsesSoftAggregatorPrefix(t *testing.T) {
	require.Equal(t, "sa_99", providerPlayerID("softaggregator", 99))
}
