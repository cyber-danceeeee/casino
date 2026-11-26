package softaggregator

import (
	"context"
	"encoding/json"
	"strings"

	"casino/internal/game/domain"
)

type Provider struct {
	cfg    Config
	client *Client
}

func NewProvider(cfg Config, client *Client) *Provider {
	if client == nil {
		client = NewClient(cfg)
	}
	return &Provider{cfg: cfg, client: client}
}

func (p *Provider) Code() string {
	return Code
}

func (p *Provider) CurrencyMapping() domain.CurrencyMapper {
	return currencyMapper{}
}

func (p *Provider) GetGames(ctx context.Context, req domain.GetGamesRequest) ([]domain.Game, error) {
	envelope, err := p.client.call(ctx, "getGameList", apiRequest{}, true)
	if err != nil {
		return nil, err
	}
	var items []gameListItem
	if err := json.Unmarshal(envelope.Response, &items); err != nil {
		return nil, err
	}
	games := make([]domain.Game, 0, len(items))
	for _, item := range items {
		games = append(games, mapGame(item))
	}
	return games, nil
}

func (p *Provider) CreatePlayer(ctx context.Context, req domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	payload := apiRequest{
		"user_username": req.ExternalPlayerID,
		"user_password": req.ProviderPlayerPassword,
		"currency":      req.Currency,
	}
	if _, err := p.client.call(ctx, "createPlayer", payload, false); err != nil {
		return nil, err
	}
	return &domain.ProviderPlayer{
		ProviderCode:     Code,
		UserID:           req.UserID,
		ProviderPlayerID: req.ExternalPlayerID,
		Currency:         req.Currency,
	}, nil
}

func (p *Provider) LaunchGame(ctx context.Context, req domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	payload := apiRequest{
		"user_username": req.ProviderPlayerID,
		"user_password": req.ProviderPlayerPassword,
		"gameid":        req.ProviderGameID,
		"lang":          req.Language,
		"currency":      req.Currency,
		"homeurl":       req.ReturnURL,
	}
	envelope, trace, err := p.client.callWithTrace(ctx, "getGame", payload, false)
	if err != nil {
		return nil, err
	}
	launchURL, err := parseLaunchResponse(envelope.Response)
	if err != nil {
		return nil, &providerError{Message: err.Error(), RequestTrace: traceRequestJSON(trace), ResponseTrace: traceResponseJSON(trace)}
	}
	return &domain.LaunchGameResult{
		LaunchURL:             launchURL,
		ProviderCode:          Code,
		ProviderPlayerID:      req.ProviderPlayerID,
		ProviderRequestTrace:  traceRequestJSON(trace),
		ProviderResponseTrace: traceResponseJSON(trace),
	}, nil
}

func (p *Provider) LaunchDemo(ctx context.Context, req domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	payload := apiRequest{
		"gameid":       req.ProviderGameID,
		"lang":         req.Language,
		"currency":     req.Currency,
		"homeurl":      req.ReturnURL,
		"play_for_fun": true,
	}
	envelope, trace, err := p.client.callWithTrace(ctx, "getGameDemo", payload, false)
	if err != nil {
		return nil, err
	}
	launchURL, err := parseLaunchResponse(envelope.Response)
	if err != nil {
		return nil, &providerError{Message: err.Error(), RequestTrace: traceRequestJSON(trace), ResponseTrace: traceResponseJSON(trace)}
	}
	return &domain.LaunchGameResult{
		LaunchURL:             launchURL,
		ProviderCode:          Code,
		ProviderRequestTrace:  traceRequestJSON(trace),
		ProviderResponseTrace: traceResponseJSON(trace),
	}, nil
}

func (p *Provider) GetProviderBalance(ctx context.Context) (*domain.ProviderBalance, error) {
	return nil, domain.ErrNotSupported
}

func (p *Provider) TestConnection(ctx context.Context) error {
	_, err := p.client.call(ctx, "getGameList", apiRequest{}, true)
	return err
}

type launchResponse struct {
	GameURL string `json:"gameurl"`
}

func parseLaunchResponse(raw json.RawMessage) (string, error) {
	var response launchResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", err
	}
	launchURL := strings.TrimSpace(response.GameURL)
	if launchURL == "" {
		return "", &providerError{Message: "softaggregator launch response missing response.gameurl"}
	}
	return launchURL, nil
}
