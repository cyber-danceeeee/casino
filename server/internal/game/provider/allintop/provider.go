// Package allintop implements the AllInTop v2 single-wallet API.
package allintop

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"casino/internal/game/domain"
	"github.com/shopspring/decimal"
)

const Code = "allintop"

type Config struct {
	Enabled                                          bool
	APIBaseURL, ClientID, ClientSecret, CallbackPath string
	CallbackIPAllowlist                              []string
	RequestTimeout, ConnectTimeout                   time.Duration
	MaxTransactionAmountMinor                        int64
}
type Provider struct {
	cfg    Config
	client *http.Client
	token  *tokenEntry
}
type tokenEntry struct {
	sync.Mutex
	value               string
	expires, retryAfter time.Time
}

// The connection store constructs adapters per request. Cache across instances,
// with credentials in the cache key, to respect the 5 tokens / 30 seconds limit.
var tokenCache sync.Map

func NewProvider(c Config, client *http.Client) *Provider {
	c.APIBaseURL = strings.TrimRight(c.APIBaseURL, "/")
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 15 * time.Second
	}
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = 3 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: c.RequestTimeout, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: c.ConnectTimeout}).DialContext, TLSHandshakeTimeout: c.ConnectTimeout}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	key := sha256.Sum256([]byte(c.APIBaseURL + "\x00" + c.ClientID + "\x00" + c.ClientSecret))
	entry, _ := tokenCache.LoadOrStore(key, &tokenEntry{})
	return &Provider{cfg: c, client: client, token: entry.(*tokenEntry)}
}
func (p *Provider) Code() string                           { return Code }
func (p *Provider) CurrencyMapping() domain.CurrencyMapper { return currencyMapper{} }

type currencyMapper struct{}

func (currencyMapper) ToProviderCurrency(country, currency string) string {
	if strings.EqualFold(country, "ID") && strings.EqualFold(currency, "IDR") {
		return "IDR"
	}
	return ""
}
func (currencyMapper) ToPlatformCurrency(c string) string {
	if strings.EqualFold(c, "IDR") {
		return "IDR"
	}
	return ""
}
func (p *Provider) request(ctx context.Context, path string, input any, token string, out any) error {
	if !p.cfg.Enabled || p.cfg.ClientID == "" || p.cfg.ClientSecret == "" {
		return domain.ErrProviderDisabled
	}
	u, err := url.Parse(p.cfg.APIBaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("invalid AllInTop API URL")
	}
	var body io.Reader
	method := http.MethodGet
	if input != nil {
		b, e := json.Marshal(input)
		if e != nil {
			return e
		}
		body = bytes.NewReader(b)
		method = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, method, p.cfg.APIBaseURL+path, body)
	if err != nil {
		return errors.New("invalid AllInTop request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return errors.New("AllInTop API connection failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AllInTop API returned HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, (16<<20)+1))
	if err != nil || len(b) > 16<<20 {
		return errors.New("invalid AllInTop response size")
	}
	if err = json.Unmarshal(b, out); err != nil {
		return errors.New("invalid AllInTop JSON response")
	}
	return nil
}
func (p *Provider) bearer(ctx context.Context) (string, error) {
	e := p.token
	e.Lock()
	defer e.Unlock()
	now := time.Now()
	if e.value != "" && now.Add(10*time.Second).Before(e.expires) {
		return e.value, nil
	}
	if now.Before(e.retryAfter) {
		return "", errors.New("AllInTop token request cooling down; check credentials and provider IP allowlist")
	}
	var out struct {
		Token      string `json:"token"`
		Expiration int64  `json:"expiration"`
	}
	// Back off after any failed authentication; never automatically loop on 401/403/429.
	e.retryAfter = now.Add(30 * time.Second)
	err := p.request(ctx, "/auth/createtoken", map[string]string{"clientId": p.cfg.ClientID, "clientSecret": p.cfg.ClientSecret}, "", &out)
	if err != nil {
		return "", err
	}
	if out.Token == "" || out.Expiration <= now.Add(10*time.Second).Unix() {
		return "", errors.New("AllInTop returned an invalid token")
	}
	e.value = out.Token
	e.expires = time.Unix(out.Expiration, 0)
	e.retryAfter = time.Time{}
	return e.value, nil
}
func (p *Provider) call(ctx context.Context, path string, input, out any) error {
	token, err := p.bearer(ctx)
	if err != nil {
		return err
	}
	var env struct {
		Success   bool            `json:"success"`
		Message   json.RawMessage `json:"message"`
		ErrorCode int             `json:"errorCode"`
	}
	if err = p.request(ctx, path, input, token, &env); err != nil {
		return err
	}
	if !env.Success || env.ErrorCode != 0 {
		return fmt.Errorf("AllInTop API error %d", env.ErrorCode)
	}
	if out != nil {
		if err = json.Unmarshal(env.Message, out); err != nil {
			return errors.New("invalid AllInTop result")
		}
	}
	return nil
}

type Vendor struct {
	VendorCode string `json:"vendorCode"`
	Type       int    `json:"type"`
	Name       string `json:"name"`
}
type upstreamGame struct {
	VendorCode       string `json:"vendorCode"`
	GameCode         string `json:"gameCode"`
	GameName         string `json:"gameName"`
	Slug             string `json:"slug"`
	Thumbnail        string `json:"thumbnail"`
	UpdatedAt        string `json:"updatedAt"`
	UnderMaintenance bool   `json:"underMaintenance"`
}

func (p *Provider) GetGames(ctx context.Context, r domain.GetGamesRequest) ([]domain.Game, error) {
	var vendors []Vendor
	if err := p.call(ctx, "/vendors/list", nil, &vendors); err != nil {
		return nil, err
	}
	language := r.Language
	if language == "" {
		language = "id"
	}
	games := []domain.Game{}
	seen := map[string]bool{}
	for _, v := range vendors {
		if v.VendorCode == "" || strings.Contains(v.VendorCode, "/") {
			return nil, errors.New("invalid AllInTop vendor code")
		}
		var rows []json.RawMessage
		if err := p.call(ctx, "/games/list", map[string]string{"vendorCode": v.VendorCode, "language": language}, &rows); err != nil {
			return nil, fmt.Errorf("AllInTop game catalog incomplete: %w", err)
		}
		for _, raw := range rows {
			var g upstreamGame
			if json.Unmarshal(raw, &g) != nil || g.GameCode == "" || g.GameName == "" {
				return nil, errors.New("invalid AllInTop game entry")
			}
			id := v.VendorCode + "/" + g.GameCode
			if len(id) > 128 {
				return nil, errors.New("AllInTop game identifier exceeds 128 bytes")
			}
			if seen[id] {
				return nil, errors.New("duplicate AllInTop game identifier")
			}
			seen[id] = true
			category := map[int]string{1: "live", 2: "slots", 3: "arcade", 4: "fishing", 5: "sports", 6: "table-games"}[v.Type]
			if category == "" {
				category = "other"
			}
			metadata, _ := json.Marshal(map[string]any{"vendor": v, "game": json.RawMessage(raw)})
			slug := g.Slug
			if slug == "" {
				slug = g.GameCode
			}
			row := domain.Game{ProviderCode: Code, ProviderGameID: id, Name: g.GameName, Slug: catalogSlug(id, v.VendorCode+"-"+slug), Category: category, GameType: map[int]string{1: "live", 2: "slot", 3: "arcade", 4: "fish", 5: "sports", 6: "table"}[v.Type], ThumbnailURL: g.Thumbnail, IsEnabled: !g.UnderMaintenance, RawMetadata: metadata}
			if t, err := time.Parse(time.RFC3339, g.UpdatedAt); err == nil {
				row.LastProviderUpdate = &t
			}
			games = append(games, row)
		}
	}
	// Empty responses must not disable a previously imported catalog.
	if len(games) == 0 {
		return nil, errors.New("AllInTop returned an empty catalog")
	}
	return games, nil
}
func (p *Provider) CreatePlayer(_ context.Context, r domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	if r.Currency != "IDR" || r.ExternalPlayerID == "" {
		return nil, domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "AllInTop account requires IDR", nil)
	}
	// Seamless users are created by launch, not the transfer-wallet create-user API.
	return &domain.ProviderPlayer{ProviderCode: Code, UserID: r.UserID, ProviderPlayerID: r.ExternalPlayerID, Currency: "IDR"}, nil
}
func (p *Provider) LaunchGame(ctx context.Context, r domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	if p.CurrencyMapping().ToProviderCurrency(r.Country, r.Currency) == "" {
		return nil, domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "AllInTop account supports ID / IDR only", nil)
	}
	parts := strings.SplitN(r.ProviderGameID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || r.ProviderPlayerID == "" {
		return nil, errors.New("invalid AllInTop launch parameters")
	}
	language := r.Language
	if language == "" {
		language = "id"
	}
	req := map[string]string{"vendorCode": parts[0], "gameCode": parts[1], "userCode": r.ProviderPlayerID, "language": language}
	if r.ReturnURL != "" {
		req["lobbyUrl"] = r.ReturnURL
	}
	var link string
	if err := p.call(ctx, "/game/launch-url", req, &link); err != nil {
		return nil, err
	}
	u, err := url.Parse(link)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return nil, errors.New("invalid AllInTop launch URL")
	}
	return &domain.LaunchGameResult{ProviderCode: Code, LaunchURL: link, SessionID: r.SessionID, ProviderPlayerID: r.ProviderPlayerID}, nil
}
func (p *Provider) LaunchDemo(context.Context, domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	return nil, domain.ErrNotSupported
}
func (p *Provider) TestConnection(ctx context.Context) error {
	var vendors []Vendor
	return p.call(ctx, "/vendors/list", nil, &vendors)
}
func (p *Provider) GetProviderBalance(ctx context.Context) (*domain.ProviderBalance, error) {
	var value decimal.Decimal
	if err := p.call(ctx, "/agent/balance", nil, &value); err != nil {
		return nil, err
	}
	minor, err := moneyMinor(value)
	if err != nil {
		return nil, err
	}
	return &domain.ProviderBalance{BalanceMinor: minor, Currency: "IDR"}, nil
}
func moneyMinor(d decimal.Decimal) (int64, error) {
	n := d.Mul(decimal.NewFromInt(100))
	if !n.Equal(n.Truncate(0)) || n.GreaterThan(decimal.NewFromInt(9223372036854775807)) || n.LessThan(decimal.NewFromInt(-9223372036854775807)) {
		return 0, errors.New("invalid AllInTop amount precision or range")
	}
	return n.IntPart(), nil
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func catalogSlug(id, value string) string {
	slug := strings.Trim(slugPattern.ReplaceAllString(strings.ToLower(value), "-"), "-")
	if len(slug) > 90 {
		slug = strings.Trim(slug[:90], "-")
	}
	if slug == "" {
		slug = "game"
	}
	sum := sha256.Sum256([]byte(id))
	return slug + "-" + hex.EncodeToString(sum[:6])
}
