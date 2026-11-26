package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"casino/internal/game/domain"
	"casino/internal/game/provider/ag"
	"casino/internal/game/provider/softaggregator"
	"casino/internal/game/repository"
	"casino/internal/model"
	"casino/internal/tenancy"
)

type GameService struct {
	registry        domain.ProviderRegistry
	repo            *repository.Repository
	providerBaseURL map[string]string
}

func NewGameService(registry domain.ProviderRegistry, repo *repository.Repository, providerBaseURL map[string]string) *GameService {
	if providerBaseURL == nil {
		providerBaseURL = map[string]string{}
	}
	return &GameService{registry: registry, repo: repo, providerBaseURL: providerBaseURL}
}

func (s *GameService) SyncProviderGames(ctx context.Context, providerCode string) (repository.SyncSummary, error) {
	return s.SyncProviderGamesScoped(ctx, tenancy.Scope{}, providerCode)
}

func (s *GameService) SyncProviderGamesScoped(ctx context.Context, scope tenancy.Scope, providerCode string) (repository.SyncSummary, error) {
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	summary := repository.SyncSummary{TenantID: scope.TenantID, SiteID: scope.SiteID, ProviderCode: providerCode}
	startedAt := time.Now()
	defer func() { summary.DurationMS = time.Since(startedAt).Milliseconds() }()

	if err := s.repo.EnsureCatalogSchema(ctx); err != nil {
		_ = s.repo.RecordSyncFailureScoped(ctx, scope, providerCode, err)
		return summary, err
	}

	provider, err := s.registry.Get(providerCode)
	if err != nil {
		_ = s.repo.RecordSyncFailureScoped(ctx, scope, providerCode, err)
		return summary, err
	}

	providerRow, err := s.repo.EnsureProviderScoped(ctx, scope, providerCode, displayProviderName(providerCode), s.providerAPIBaseURL(providerCode), true)
	if err != nil {
		_ = s.repo.RecordSyncFailureScoped(ctx, scope, providerCode, err)
		return summary, err
	}

	games, err := provider.GetGames(ctx, domain.GetGamesRequest{})
	if err != nil {
		_ = s.repo.RecordSyncFailureScoped(ctx, scope, providerCode, err)
		return summary, err
	}
	if encoded, marshalErr := json.Marshal(syncRawResponse(games)); marshalErr == nil {
		summary.RawResponse = string(encoded)
	}

	for index := range games {
		if games[index].ProviderCode == "" {
			games[index].ProviderCode = providerCode
		}
	}
	games = uniqueProviderGames(games)
	providerGameIDs := make([]string, 0, len(games))
	for _, game := range games {
		providerGameIDs = append(providerGameIDs, game.ProviderGameID)
	}

	var summaryMu sync.Mutex
	jobs := make(chan domain.Game)
	workerCount := 16
	if len(games) < workerCount {
		workerCount = len(games)
	}
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for game := range jobs {
				s.processSyncGame(ctx, providerCode, providerRow, game, &summary, &summaryMu)
			}
		}()
	}
	for _, game := range games {
		jobs <- game
	}
	close(jobs)
	workers.Wait()
	s.syncGameThumbnails(ctx, providerCode, providerRow.ID, games, &summary, &summaryMu)

	disabled, err := s.repo.MarkMissingProviderGames(ctx, providerRow.ID, providerGameIDs)
	if err != nil {
		_ = s.repo.RecordSyncFailureScoped(ctx, scope, providerCode, err)
		return summary, err
	}
	summary.Disabled = disabled
	summary.DurationMS = time.Since(startedAt).Milliseconds()

	if err := s.repo.RecordSyncSuccess(ctx, summary); err != nil {
		return summary, err
	}
	return summary, nil
}

func uniqueProviderGames(games []domain.Game) []domain.Game {
	unique := make([]domain.Game, 0, len(games))
	seen := make(map[string]struct{}, len(games))
	for _, game := range games {
		game.ProviderGameID = strings.TrimSpace(game.ProviderGameID)
		if game.ProviderGameID == "" {
			continue
		}
		key := strings.ToLower(game.ProviderGameID)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, game)
	}
	return unique
}

func (s *GameService) processSyncGame(ctx context.Context, providerCode string, providerRow *model.GameGameProviders, game domain.Game, summary *repository.SyncSummary, summaryMu *sync.Mutex) {
	created, updated, err := s.repo.UpsertGame(ctx, providerRow, game)
	if err != nil {
		summaryMu.Lock()
		summary.Failed++
		summary.Items = append(summary.Items, repository.SyncItem{ProviderGameID: game.ProviderGameID, GameName: game.Name, ChangeType: "FAILED", ErrorMessage: err.Error()})
		summaryMu.Unlock()
		return
	}

	summaryMu.Lock()
	summary.Synced++
	switch {
	case created:
		summary.Inserted++
		summary.Items = append(summary.Items, repository.SyncItem{ProviderGameID: game.ProviderGameID, GameName: game.Name, ChangeType: "NEW"})
	case updated:
		summary.Updated++
		summary.Items = append(summary.Items, repository.SyncItem{ProviderGameID: game.ProviderGameID, GameName: game.Name, ChangeType: "UPDATED"})
	default:
		summary.Unchanged++
	}
	summaryMu.Unlock()

}

func syncRawResponse(games []domain.Game) map[string]any {
	const sampleLimit = 50
	sampleSize := len(games)
	if sampleSize > sampleLimit {
		sampleSize = sampleLimit
	}
	return map[string]any{
		"total":     len(games),
		"truncated": len(games) > sampleLimit,
		"sample":    games[:sampleSize],
	}
}

func (s *GameService) syncGameThumbnails(ctx context.Context, providerCode string, providerID uint64, games []domain.Game, summary *repository.SyncSummary, summaryMu *sync.Mutex) {
	jobs := make(chan domain.Game)
	workerCount := 16
	if len(games) < workerCount {
		workerCount = len(games)
	}
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for game := range jobs {
				if strings.TrimSpace(game.ThumbnailURL) == "" {
					continue
				}
				localURL, thumbnailErr := downloadGameThumbnail(ctx, providerCode, game.ProviderGameID, game.ThumbnailURL)
				if thumbnailErr == nil {
					thumbnailErr = s.repo.UpdateGameLocalThumbnail(ctx, providerID, game.ProviderGameID, localURL)
				}
				if thumbnailErr != nil {
					summaryMu.Lock()
					summary.Items = append(summary.Items, repository.SyncItem{
						ProviderGameID: game.ProviderGameID,
						GameName:       game.Name,
						ChangeType:     "THUMBNAIL_FAILED",
						ErrorMessage:   thumbnailErr.Error(),
					})
					summaryMu.Unlock()
				}
			}
		}()
	}
	for _, game := range games {
		jobs <- game
	}
	close(jobs)
	workers.Wait()
}

func displayProviderName(providerCode string) string {
	switch providerCode {
	case ag.Code:
		return "AG"
	case softaggregator.Code:
		return "SoftAggregator"
	default:
		return providerCode
	}
}

func (s *GameService) providerAPIBaseURL(providerCode string) string {
	return s.providerBaseURL[providerCode]
}
