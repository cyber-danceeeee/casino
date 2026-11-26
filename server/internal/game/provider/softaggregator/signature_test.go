package softaggregator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCallbackSignatureSuccess(t *testing.T) {
	key := CallbackKey("1720000000", "salt")
	require.True(t, ValidCallbackKey("1720000000", "salt", key))
}

func TestCallbackSignatureInvalid(t *testing.T) {
	require.False(t, ValidCallbackKey("1720000000", "salt", "bad-signature"))
}

func TestEndpointURLAppendsDocumentedPath(t *testing.T) {
	cfg := Config{APIBaseURL: "https://api.softaggregator.com"}
	require.Equal(t, "https://api.softaggregator.com/api/v1", cfg.EndpointURL())
}

func TestEndpointURLKeepsExistingDocumentedPath(t *testing.T) {
	cfg := Config{APIBaseURL: "https://api.softaggregator.com/api/v1"}
	require.Equal(t, "https://api.softaggregator.com/api/v1", cfg.EndpointURL())
}
