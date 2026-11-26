package ag

import (
	"net/url"
	"strings"
	"time"
)

const (
	Code     = "ag"
	signType = "HmacSHA256"
)

type Config struct {
	Enabled                   bool
	APIBaseURL                string
	MerchantCode              string
	SecretKey                 string
	SubMerchantCode           string
	DefaultCurrency           string
	CallbackIPAllowlist       []string
	MaxTimestampDrift         time.Duration
	RequestTimeout            time.Duration
	ConnectTimeout            time.Duration
	MaxTransactionAmountMinor int64
}

func (c Config) EndpointURL() string {
	base := strings.TrimRight(strings.TrimSpace(c.APIBaseURL), "/")
	if base == "" {
		base = "https://pre-api-area.agsoftgames.com"
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return base
	}
	return u.String()
}

func (c Config) OutboundReady() bool {
	return c.Enabled && strings.TrimSpace(c.MerchantCode) != "" && strings.TrimSpace(c.SecretKey) != ""
}

func (c Config) CallbackReady() bool {
	return c.Enabled && strings.TrimSpace(c.MerchantCode) != "" && strings.TrimSpace(c.SecretKey) != ""
}

func (c Config) SyncCurrency(currency string) string {
	if strings.TrimSpace(currency) == "" {
		currency = c.DefaultCurrency
	}
	return strings.ToUpper(strings.TrimSpace(currency))
}
