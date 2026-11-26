package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactMapHidesCallbackSecrets(t *testing.T) {
	result := redactMap(map[string][]string{"action": {"debit"}, "key": {"signature"}, "token": {"launch-token"}})
	require.Equal(t, []string{"debit"}, result["action"])
	require.Equal(t, []string{"REDACTED"}, result["key"])
	require.Equal(t, []string{"REDACTED"}, result["token"])
}

func TestRedactBodyRecursivelyHidesSecrets(t *testing.T) {
	result := redactBody([]byte(`{"username":"player","nested":{"salt_key":"secret","amount":100}}`))
	require.Contains(t, result, `"username":"player"`)
	require.Contains(t, result, `"salt_key":"REDACTED"`)
	require.NotContains(t, result, "secret")
}
