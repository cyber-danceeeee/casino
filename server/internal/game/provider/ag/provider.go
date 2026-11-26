package ag

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"casino/internal/game/domain"
)

const (
	pathCategories = "/game/v5/categories"
	pathGames      = "/game/v5/games"
	pathLaunchURL  = "/game/v5/game/url"
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
	categories, err := p.categories(ctx, req)
	if err != nil {
		return nil, err
	}
	categoryGames := make([][]domain.Game, len(categories))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(6)
	for index, category := range categories {
		index, category := index, category
		categoryCode := strings.TrimSpace(category.CategoryCode)
		if categoryCode == "" {
			continue
		}
		group.Go(func() error {
			games, err := p.gamesByCategory(groupCtx, req, category)
			if err != nil {
				return err
			}
			categoryGames[index] = games
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	games := make([]domain.Game, 0)
	for _, items := range categoryGames {
		games = append(games, items...)
	}
	return deduplicateGames(games), nil
}

func (p *Provider) categories(ctx context.Context, req domain.GetGamesRequest) ([]categoryItem, error) {
	payload := apiRequest{
		"currencyCode": p.cfg.SyncCurrency(req.Currency),
		"lang":         languageOrDefault(req.Language),
	}
	envelope, err := p.client.post(ctx, pathCategories, payload)
	if err != nil {
		return nil, err
	}
	var data categoryListResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return nil, err
	}
	return data.CategoryDataList, nil
}

func (p *Provider) gamesByCategory(ctx context.Context, req domain.GetGamesRequest, category categoryItem) ([]domain.Game, error) {
	result := make([]domain.Game, 0)
	total := 0
	for pageIndex := 1; pageIndex <= 200; pageIndex++ {
		payload := apiRequest{
			"categoryCode": category.CategoryCode,
			"currencyCode": p.cfg.SyncCurrency(req.Currency),
			"lang":         languageOrDefault(req.Language),
			"pageIndex":    pageIndex,
		}
		envelope, err := p.client.post(ctx, pathGames, payload)
		if err != nil {
			return nil, err
		}
		var data gameListResponse
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return nil, err
		}
		if total == 0 {
			total = data.TotalSize
		}
		if len(data.GameItemList) == 0 {
			break
		}
		for _, item := range data.GameItemList {
			item.ProviderCategoryCode = category.CategoryCode
			item.ProviderCategoryName = category.CategoryName
			result = append(result, mapGame(item))
		}
		if total > 0 && len(result) >= total {
			break
		}
	}
	return result, nil
}

func (p *Provider) CreatePlayer(ctx context.Context, req domain.CreatePlayerRequest) (*domain.ProviderPlayer, error) {
	return &domain.ProviderPlayer{
		ProviderCode:     Code,
		UserID:           req.UserID,
		ProviderPlayerID: req.ExternalPlayerID,
		Currency:         req.Currency,
	}, nil
}

func (p *Provider) LaunchGame(ctx context.Context, req domain.LaunchGameRequest) (*domain.LaunchGameResult, error) {
	payload := apiRequest{
		"gameCode":     req.ProviderGameID,
		"playerId":     req.ProviderPlayerID,
		"currencyCode": p.CurrencyMapping().ToProviderCurrency(req.Country, req.Currency),
		"language":     languageOrDefault(req.Language),
		"terminalType": terminalType(req.Device),
		"ipAddress":    req.IPAddress,
		"nickName":     req.ProviderPlayerID,
	}
	setNonEmpty(payload, "returnUrl", req.ReturnURL)
	setNonEmpty(payload, "subMerchantCode", p.cfg.SubMerchantCode)
	setNonEmpty(payload, "clientIp", req.IPAddress)
	setNonEmpty(payload, "deviceCode", req.SessionID)
	setNonEmpty(payload, "domain", launchDomain(req.ReturnURL))
	envelope, trace, err := p.client.postWithTrace(ctx, pathLaunchURL, payload)
	if err != nil {
		return nil, err
	}
	var data launchResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return nil, &providerError{Message: err.Error(), RequestTrace: traceRequestJSON(trace), ResponseTrace: traceResponseJSON(trace)}
	}
	launchURL := strings.TrimSpace(data.GameURL)
	if launchURL == "" {
		launchURL = strings.TrimSpace(data.GameURLAlt)
	}
	if launchURL == "" {
		return nil, &providerError{Message: "ag launch response missing data.gameUrl", RequestTrace: traceRequestJSON(trace), ResponseTrace: traceResponseJSON(trace)}
	}
	result := &domain.LaunchGameResult{
		LaunchURL:             launchURL,
		ProviderCode:          Code,
		ProviderPlayerID:      req.ProviderPlayerID,
		ProviderRequestTrace:  traceRequestJSON(trace),
		ProviderResponseTrace: traceResponseJSON(trace),
	}
	if data.ExpireTime != "" {
		if expiresAt, err := time.Parse(time.RFC3339, data.ExpireTime); err == nil {
			result.ExpiresAt = &expiresAt
		}
	}
	return result, nil
}

func (p *Provider) LaunchDemo(ctx context.Context, req domain.LaunchDemoRequest) (*domain.LaunchGameResult, error) {
	return nil, domain.ErrNotSupported
}

func (p *Provider) GetProviderBalance(ctx context.Context) (*domain.ProviderBalance, error) {
	if _, err := p.categories(ctx, domain.GetGamesRequest{Language: "en"}); err != nil {
		return nil, err
	}
	return &domain.ProviderBalance{}, nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	_, err := p.categories(ctx, domain.GetGamesRequest{Language: "en"})
	return err
}

func setNonEmpty(payload apiRequest, key string, value string) {
	if value = strings.TrimSpace(value); value != "" {
		payload[key] = value
	}
}

func languageOrDefault(language string) string {
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return "en"
	}
	return language
}

func terminalType(device string) string {
	switch strings.ToLower(strings.TrimSpace(device)) {
	case "desktop", "pc", "web":
		return "PC"
	default:
		return "PHONE"
	}
}

func launchDomain(returnURL string) string {
	returnURL = strings.TrimSpace(returnURL)
	if returnURL == "" {
		return ""
	}
	parsed, err := url.Parse(returnURL)
	if err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return returnURL
}
