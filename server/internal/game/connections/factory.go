package connections

import (
	"casino/internal/config"
	"context"
	"errors"
	"strings"
	"time"

	"casino/internal/game/domain"
	"casino/internal/game/provider/ag"
	"casino/internal/game/provider/allintop"
	"casino/internal/game/provider/softaggregator"
)

type AdapterFactory struct {
	Fields   []string
	Required []string
	Build    func(Config, map[string]string) (domain.GameProvider, domain.ProviderCallbackAdapter)
}

// New protocols implement GameProvider/ProviderCallbackAdapter and add one factory.
var factories = map[string]AdapterFactory{
	"allintop": {Fields: []string{"clientId", "clientSecret"}, Required: []string{"clientId", "clientSecret"}, Build: func(c Config, secrets map[string]string) (domain.GameProvider, domain.ProviderCallbackAdapter) {
		source := config.Get().Game.AllInTop
		cfg := allintop.Config{Enabled: true, APIBaseURL: c.APIBaseURL, ClientID: secrets["clientId"], ClientSecret: secrets["clientSecret"], CallbackPath: "/api/v1/game/providers/" + c.ProviderCode + "/callback", CallbackIPAllowlist: c.CallbackIPAllowlist, RequestTimeout: seconds(source.RequestTimeoutSeconds, 15), ConnectTimeout: seconds(source.ConnectTimeoutSeconds, 3), MaxTransactionAmountMinor: amountLimit(source.MaxTransactionAmountMinor)}
		return allintop.NewProvider(cfg, nil), allintop.NewCallbackAdapter(cfg)
	}},
	"ag": {Fields: []string{"merchantCode", "secretKey", "subMerchantCode"}, Required: []string{"merchantCode", "secretKey"}, Build: func(c Config, secrets map[string]string) (domain.GameProvider, domain.ProviderCallbackAdapter) {
		source := config.Get().Game.AG
		cfg := ag.Config{Enabled: true, APIBaseURL: c.APIBaseURL, MerchantCode: secrets["merchantCode"], SecretKey: secrets["secretKey"], SubMerchantCode: secrets["subMerchantCode"], DefaultCurrency: c.MarketCurrency, CallbackIPAllowlist: c.CallbackIPAllowlist, RequestTimeout: seconds(source.RequestTimeoutSeconds, 10), ConnectTimeout: seconds(source.ConnectTimeoutSeconds, 3), MaxTimestampDrift: seconds(source.MaxTimestampDriftSeconds, 30), MaxTransactionAmountMinor: amountLimit(source.MaxTransactionAmountMinor)}
		return ag.NewProvider(cfg, nil), ag.NewCallbackAdapter(cfg)
	}},
	"softaggregator": {Fields: []string{"apiLogin", "apiPassword", "saltKey"}, Required: []string{"apiLogin", "apiPassword", "saltKey"}, Build: func(c Config, secrets map[string]string) (domain.GameProvider, domain.ProviderCallbackAdapter) {
		source := config.Get().Game.SoftAggregator
		cfg := softaggregator.Config{Enabled: true, APIBaseURL: c.APIBaseURL, APILogin: secrets["apiLogin"], APIPassword: secrets["apiPassword"], SaltKey: secrets["saltKey"], CallbackIPAllowlist: c.CallbackIPAllowlist, CallbackPath: "/api/v1/game/providers/" + c.ProviderCode + "/callback", RequestTimeout: seconds(source.RequestTimeoutSeconds, 10), ConnectTimeout: seconds(source.ConnectTimeoutSeconds, 3), MaxTimestampDrift: seconds(source.MaxTimestampDriftSeconds, 30), MaxTransactionAmountMinor: amountLimit(source.MaxTransactionAmountMinor)}
		return softaggregator.NewProvider(cfg, nil), softaggregator.NewCallbackAdapter(cfg)
	}},
}

func ready(c Config, secrets map[string]string) bool {
	f, ok := factories[c.AdapterType]
	if !ok {
		return false
	}
	for _, key := range f.Required {
		if strings.TrimSpace(secrets[key]) == "" {
			return false
		}
	}
	return true
}
func build(c Config, secrets map[string]string) (domain.GameProvider, domain.ProviderCallbackAdapter, error) {
	f, ok := factories[c.AdapterType]
	if !ok {
		return nil, nil, errors.New("unsupported provider adapter")
	}
	if !ready(c, secrets) {
		return nil, nil, errors.New("provider credentials are incomplete")
	}
	p, cb := f.Build(c, secrets)
	return &namedProvider{GameProvider: p, code: c.ProviderCode}, &namedCallback{ProviderCallbackAdapter: cb, code: c.ProviderCode}, nil
}

type namedProvider struct {
	domain.GameProvider
	code string
}

func (p *namedProvider) Code() string { return p.code }
func (p *namedProvider) GetGames(ctx context.Context, r domain.GetGamesRequest) ([]domain.Game, error) {
	games, err := p.GameProvider.GetGames(ctx, r)
	for i := range games {
		games[i].ProviderCode = p.code
	}
	return games, err
}
func (p *namedProvider) CreatePlayer(ctx context.Context, r domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	row, err := p.GameProvider.CreatePlayer(ctx, r)
	if row != nil {
		row.ProviderCode = p.code
	}
	return row, err
}
func (p *namedProvider) TestConnection(ctx context.Context) error {
	tester, ok := p.GameProvider.(domain.ProviderConnectionTester)
	if !ok {
		return domain.ErrNotSupported
	}
	return tester.TestConnection(ctx)
}

type namedCallback struct {
	domain.ProviderCallbackAdapter
	code string
}

func (c *namedCallback) Code() string { return c.code }
func (c *namedCallback) ParseCommand(ctx context.Context, r domain.ProviderCallbackRequest) (*domain.WalletCommand, error) {
	row, err := c.ProviderCallbackAdapter.ParseCommand(ctx, r)
	if row != nil {
		row.ProviderCode = c.code
	}
	return row, err
}

func (p *namedProvider) LaunchGame(ctx context.Context, r domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	result, err := p.GameProvider.LaunchGame(ctx, r)
	if result != nil {
		result.ProviderCode = p.code
	}
	return result, err
}
func (p *namedProvider) LaunchDemo(ctx context.Context, r domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	result, err := p.GameProvider.LaunchDemo(ctx, r)
	if result != nil {
		result.ProviderCode = p.code
	}
	return result, err
}

func seconds(value, fallback int) time.Duration {
	if value <= 0 {
		value = fallback
	}
	return time.Duration(value) * time.Second
}
func amountLimit(value int64) int64 {
	if value <= 0 {
		return 100000000
	}
	return value
}

func (c *namedCallback) ParseBatchCommands(ctx context.Context, r domain.ProviderCallbackRequest) ([]domain.WalletCommand, error) {
	adapter, ok := c.ProviderCallbackAdapter.(domain.ProviderBatchCallbackAdapter)
	if !ok {
		return nil, domain.ErrNotSupported
	}
	commands, err := adapter.ParseBatchCommands(ctx, r)
	for i := range commands {
		commands[i].ProviderCode = c.code
	}
	return commands, err
}
