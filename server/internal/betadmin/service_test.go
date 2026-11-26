package betadmin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoneyUsesExactMinorUnits(t *testing.T) {
	require.Equal(t, "0.01", money(1))
	require.Equal(t, "123456789012.34", money(12345678901234))
	require.Equal(t, "-10.05", money(-1005))
}

func TestParseMinorUsesDecimalArithmetic(t *testing.T) {
	value, ok := parseMinor("100.29")
	require.True(t, ok)
	require.Equal(t, int64(10029), value)
	_, ok = parseMinor("invalid")
	require.False(t, ok)
}

func TestSafeMapDoesNotReturnUnapprovedFields(t *testing.T) {
	row := map[string]any{"provider_code": "softaggregator", "request_headers": "safe", "api_password": "secret", "salt_key": "secret"}
	result := safeMap(row, []string{"provider_code", "request_headers"})
	require.Equal(t, "softaggregator", result["provider_code"])
	require.NotContains(t, result, "api_password")
	require.NotContains(t, result, "salt_key")
}
