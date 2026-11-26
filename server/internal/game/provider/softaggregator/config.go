package softaggregator

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const Code = "softaggregator"

type Config struct {
	Enabled                   bool
	APIBaseURL                string
	APILogin                  string
	APIPassword               string
	SaltKey                   string
	CallbackPath              string
	CallbackIPAllowlist       []string
	MaxTimestampDrift         time.Duration
	RequestTimeout            time.Duration
	ConnectTimeout            time.Duration
	MaxTransactionAmountMinor int64
}

func LoadConfigFromEnv() Config {
	return Config{
		Enabled:                   envBool("GAME_SOFTAGGREGATOR_ENABLED", false),
		APIBaseURL:                envString("GAME_SOFTAGGREGATOR_API_BASE_URL", "https://api.softaggregator.com"),
		APILogin:                  os.Getenv("GAME_SOFTAGGREGATOR_API_LOGIN"),
		APIPassword:               os.Getenv("GAME_SOFTAGGREGATOR_API_PASSWORD"),
		SaltKey:                   os.Getenv("GAME_SOFTAGGREGATOR_SALT_KEY"),
		CallbackPath:              envString("GAME_SOFTAGGREGATOR_CALLBACK_PATH", "/api/v1/game/providers/softaggregator/callback"),
		CallbackIPAllowlist:       envCSV("GAME_SOFTAGGREGATOR_CALLBACK_IP_ALLOWLIST"),
		MaxTimestampDrift:         time.Duration(envInt("GAME_SOFTAGGREGATOR_MAX_TIMESTAMP_DRIFT_SECONDS", 30)) * time.Second,
		RequestTimeout:            time.Duration(envInt("GAME_SOFTAGGREGATOR_REQUEST_TIMEOUT_SECONDS", 10)) * time.Second,
		ConnectTimeout:            time.Duration(envInt("GAME_SOFTAGGREGATOR_CONNECT_TIMEOUT_SECONDS", 3)) * time.Second,
		MaxTransactionAmountMinor: int64(envInt("GAME_SOFTAGGREGATOR_MAX_TRANSACTION_AMOUNT_MINOR", 100000000)),
	}
}

func (c Config) EndpointURL() string {
	base := strings.TrimRight(strings.TrimSpace(c.APIBaseURL), "/")
	if base == "" {
		base = "https://api.softaggregator.com"
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return base
	}
	if strings.HasSuffix(u.Path, "/api/v1") || u.Path == "/api/v1" {
		return u.String()
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/api/v1"
	return u.String()
}

func (c Config) OutboundReady() bool {
	return c.Enabled && strings.TrimSpace(c.APILogin) != "" && strings.TrimSpace(c.APIPassword) != ""
}

func (c Config) CallbackReady() bool {
	return c.Enabled && strings.TrimSpace(c.SaltKey) != ""
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSV(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
