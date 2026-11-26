package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const remoteFixture = `app:
  name: from-nacos
  version: remote-v1
  staticDir: ../static
http:
  port: 1639
database:
  driver: mysql
  mysql:
    dsn: remote-dsn
`

func isolatedNacosEnvironment(t *testing.T) {
	t.Helper()
	previous, previousSource := config, loadedSource
	t.Cleanup(func() { config, loadedSource = previous, previousSource })
	for _, key := range []string{"CASINO_CONFIG_SOURCE", "NACOS_SERVER_ADDR", "NACOS_NAMESPACE_ID", "NACOS_GROUP", "NACOS_DATA_ID", "NACOS_USERNAME", "NACOS_PASSWORD", "NACOS_TIMEOUT_SECONDS"} {
		t.Setenv(key, "")
	}
}

func TestNacosAuthenticatedStartupReplacesLocalYAMLAndAppliesEnvironment(t *testing.T) {
	isolatedNacosEnvironment(t)
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/nacos/v1/auth/login":
			require.Equal(t, "reader", r.FormValue("username"))
			require.Equal(t, "private+&% password", r.FormValue("password"))
			_, _ = w.Write([]byte(`{"accessToken":"private+&token"}`))
		case "/nacos/v1/cs/configs":
			require.Equal(t, "game", r.URL.Query().Get("dataId"))
			require.Equal(t, "DEFAULT_GROUP", r.URL.Query().Get("group"))
			require.Equal(t, "private+&token", r.URL.Query().Get("accessToken"))
			require.False(t, r.URL.Query().Has("tenant"), "public namespace must omit the HTTP tenant parameter")
			_, _ = w.Write([]byte(remoteFixture))
		default:
			t.Errorf("unexpected Nacos operation: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	t.Setenv("NACOS_SERVER_ADDR", server.URL)
	t.Setenv("NACOS_DATA_ID", "game")
	t.Setenv("NACOS_USERNAME", "reader")
	t.Setenv("NACOS_PASSWORD", "private+&% password")
	t.Setenv("CASINO_MYSQL_DSN", "explicit-local-connection")
	// No local YAML exists: the filename anchors relative resource paths only.
	filename := filepath.Join(t.TempDir(), "configs", "missing.yml")
	require.NoError(t, Init(filename))
	require.Equal(t, "nacos", Source())
	require.Equal(t, "from-nacos", Get().App.Name)
	require.Equal(t, "remote-v1", Get().App.Version)
	require.Equal(t, "explicit-local-connection", Get().Database.Mysql.Dsn)
	require.Equal(t, filepath.Join(filepath.Dir(filename), "../static"), Get().App.StaticDir)
	require.Equal(t, []string{"POST /nacos/v1/auth/login", "GET /nacos/v1/cs/configs"}, calls)
}

func TestNacosFailuresNeverFallBackOrExposeResponseSecrets(t *testing.T) {
	for _, test := range []struct {
		name, path string
		status     int
		body       string
	}{
		{"bad-login", "/nacos/v1/auth/login", 403, "private-response-password"},
		{"bad-token", "/nacos/v1/auth/login", 200, `{"message":"private-response-password"}`},
		{"denied-read", "/nacos/v1/cs/configs", 403, "private-response-password"},
		{"missing-document", "/nacos/v1/cs/configs", 404, "private-response-password"},
		{"empty-document", "/nacos/v1/cs/configs", 200, ""},
		{"invalid-yaml", "/nacos/v1/cs/configs", 200, "app: [private-response-password"},
		{"invalid-schema", "/nacos/v1/cs/configs", 200, strings.Replace(remoteFixture, "port: 1639", "port: private-response-password", 1)},
		{"oversized-document", "/nacos/v1/cs/configs", 200, strings.Repeat("s", maxNacosConfigBytes+1)},
		{"invalid-port", "/nacos/v1/cs/configs", 200, strings.Replace(remoteFixture, "port: 1639", "port: 0", 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			isolatedNacosEnvironment(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == test.path {
					w.WriteHeader(test.status)
					_, _ = w.Write([]byte(test.body))
				} else {
					_, _ = w.Write([]byte(`{"accessToken":"private-token"}`))
				}
			}))
			defer server.Close()
			t.Setenv("NACOS_SERVER_ADDR", server.URL)
			t.Setenv("NACOS_DATA_ID", "game")
			t.Setenv("NACOS_USERNAME", "reader")
			t.Setenv("NACOS_PASSWORD", "private-password")
			filename := filepath.Join(t.TempDir(), "casino.yml")
			require.NoError(t, os.WriteFile(filename, []byte(remoteFixture), 0600))
			previous, previousSource := config, loadedSource
			err := Init(filename)
			require.Error(t, err)
			for _, secret := range []string{"private-password", "private-token", "private-response-password"} {
				require.NotContains(t, err.Error(), secret)
			}
			require.Same(t, previous, config)
			require.Equal(t, previousSource, loadedSource)
		})
	}
}

func TestNacosNamespaceAndRedirectProtection(t *testing.T) {
	redirectCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirectCalls++ }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "deployment-namespace-id", r.URL.Query().Get("tenant"))
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer server.Close()
	_, err := readNacosConfig(context.Background(), nacosSettings{baseURL: server.URL, namespace: "deployment-namespace-id", group: "G", dataID: "game", timeout: time.Second})
	require.Error(t, err)
	require.Zero(t, redirectCalls)
}

func TestNacosTimeoutDoesNotExposeRequestToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	_, err := readNacosConfig(context.Background(), nacosSettings{baseURL: server.URL, dataID: "game", timeout: 30 * time.Millisecond})
	require.ErrorContains(t, err, "connectivity")
}

func TestConfigurationSourceSelectionAndSettingsValidation(t *testing.T) {
	isolatedNacosEnvironment(t)
	source, err := configurationSource()
	require.NoError(t, err)
	require.Equal(t, "local", source)
	t.Setenv("NACOS_SERVER_ADDR", "127.0.0.1:8848")
	t.Setenv("NACOS_DATA_ID", "game")
	source, err = configurationSource()
	require.NoError(t, err)
	require.Equal(t, "nacos", source)
	s, err := nacosFromEnvironment()
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:8848/nacos", s.baseURL)
	t.Setenv("NACOS_SERVER_ADDR", "https://example.com/nacos/")
	s, err = nacosFromEnvironment()
	require.NoError(t, err)
	require.Equal(t, "https://example.com/nacos", s.baseURL)
	t.Setenv("CASINO_CONFIG_SOURCE", "local")
	source, err = configurationSource()
	require.NoError(t, err)
	require.Equal(t, "local", source)
	t.Setenv("CASINO_CONFIG_SOURCE", "invalid")
	_, err = configurationSource()
	require.Error(t, err)
	t.Setenv("NACOS_USERNAME", "reader")
	_, err = nacosFromEnvironment()
	require.Error(t, err)
	t.Setenv("NACOS_PASSWORD", "private-password")
	t.Setenv("NACOS_TIMEOUT_SECONDS", "0")
	_, err = nacosFromEnvironment()
	require.Error(t, err)
	t.Setenv("NACOS_TIMEOUT_SECONDS", "15")
	t.Setenv("NACOS_SERVER_ADDR", "http://reader:private-password@example.com")
	_, err = nacosFromEnvironment()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "private-password")
}
