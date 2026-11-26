package repository

import (
	"context"
	"crypto/sha1" //nolint:gosec // Used only for deterministic non-secret catalog identifiers.
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/game/domain"
	"casino/internal/model"
)

type Repository struct {
	db *gorm.DB
}

const (
	syncedGameNameI18nDefault         = "{}"
	syncedGameCategoryIDsDraftDefault = "[]"
)

type SyncSummary struct {
	SiteID       uint64     `json:"siteID,omitempty"`
	ProviderCode string     `json:"providerCode"`
	Synced       int        `json:"synced"`
	Inserted     int        `json:"inserted"`
	Updated      int        `json:"updated"`
	Unchanged    int        `json:"unchanged"`
	Disabled     int        `json:"disabled"`
	Failed       int        `json:"failed"`
	DurationMS   int64      `json:"durationMs"`
	RawResponse  string     `json:"-"`
	Items        []SyncItem `json:"-"`
}

type SyncItem struct {
	ProviderGameID string
	GameName       string
	ChangeType     string
	ErrorMessage   string
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) EnsureCatalogSchema(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&model.GameGameCategories{},
		&model.GameGameCategoryRelations{},
		&model.GameProviderCategoryMappings{},
		&model.GameGames{},
	)
}

func (r *Repository) EnsureProvider(ctx context.Context, code string, name string, apiBaseURL string, enabled bool) (*model.GameGameProviders, error) {
	return r.EnsureProviderScoped(ctx, content.Scope{}, code, name, apiBaseURL, enabled)
}

func (r *Repository) EnsureProviderScoped(ctx context.Context, scope content.Scope, code string, name string, apiBaseURL string, enabled bool) (*model.GameGameProviders, error) {
	status := 3
	if enabled {
		status = 1
	}
	var provider model.GameGameProviders
	err := content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGameProviders{}), "", scope).Where("provider_code = ?", code).First(&provider).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		provider = model.GameGameProviders{

			SiteID:       scope.SiteID,
			ProviderCode: code,
			Name:         name,
			ApiBaseURL:   apiBaseURL,
			Status:       status,
		}
		if err := r.db.WithContext(ctx).Create(&provider).Error; err != nil {
			return nil, err
		}
		return &provider, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{
		"name":         name,
		"api_base_url": apiBaseURL,
	}
	if err := content.ApplySite(r.db.WithContext(ctx).Model(&provider), "", scope).Updates(updates).Error; err != nil {
		return nil, err
	}
	provider.Name = name
	provider.ApiBaseURL = apiBaseURL
	return &provider, nil
}

func (r *Repository) ProviderByCode(ctx context.Context, code string) (*model.GameGameProviders, error) {
	var provider model.GameGameProviders
	if err := r.db.WithContext(ctx).Where("provider_code = ?", code).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *Repository) ProviderByCodeScoped(ctx context.Context, scope content.Scope, code string) (*model.GameGameProviders, error) {
	var provider model.GameGameProviders
	query := content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGameProviders{}), "", scope).
		Where("provider_code = ?", code)
	if err := query.First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *Repository) UpsertGame(ctx context.Context, provider *model.GameGameProviders, game domain.Game) (created bool, updated bool, err error) {
	now := time.Now()
	scope := content.Scope{SiteID: provider.SiteID}
	rawMetadata := string(game.RawMetadata)
	rawMetadata = syncedJSONOrDefault(rawMetadata, "{}")
	rawMetadataHash := metadataHash(rawMetadata)
	platformCategory, err := r.ResolvePlatformCategory(ctx, scope, provider.ProviderCode, game.Category, game.GameType)
	if err != nil {
		return false, false, err
	}
	var platformCategoryID uint64
	var platformCategoryCode string
	if platformCategory != nil {
		platformCategoryID = platformCategory.ID
		platformCategoryCode = platformCategory.CategoryKey
	}

	var existing model.GameGames
	err = content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGames{}), "", scope).
		Where("provider_id = ? AND provider_game_code = ?", provider.ID, game.ProviderGameID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := &model.GameGames{

			SiteID:               scope.SiteID,
			GameCode:             gameCode(game.ProviderCode, game.Slug),
			ProviderID:           provider.ID,
			ProviderGameCode:     game.ProviderGameID,
			Name:                 game.Name,
			NameI18n:             syncedGameNameI18nDefault,
			Slug:                 game.Slug,
			Category:             game.Category,
			PlatformCategoryID:   platformCategoryID,
			PlatformCategoryCode: platformCategoryCode,
			GameType:             game.GameType,
			CoverURL:             game.ThumbnailURL,
			LocalThumbnailURL:    game.LocalThumbnailURL,
			SupportsDemo:         game.SupportsDemo,
			SupportsFreeSpins:    game.SupportsFreeSpins,
			ProviderEnabled:      game.IsEnabled,
			PlatformEnabled:      true,
			RawMetadata:          rawMetadata,
			CategoryIDsDraft:     syncedGameCategoryIDsDraftDefault,
			MetadataHash:         rawMetadataHash,
			LastSyncedAt:         &now,
			Status:               effectiveGameStatus(game.IsEnabled, true),
		}
		if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
			return false, false, err
		}
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}

	updates := map[string]any{
		"name":                   game.Name,
		"slug":                   game.Slug,
		"category":               game.Category,
		"platform_category_id":   platformCategoryID,
		"platform_category_code": platformCategoryCode,
		"game_type":              game.GameType,
		"cover_url":              game.ThumbnailURL,
		"supports_demo":          game.SupportsDemo,
		"supports_free_spins":    game.SupportsFreeSpins,
		"provider_enabled":       game.IsEnabled,
		"raw_metadata":           rawMetadata,
		"metadata_hash":          rawMetadataHash,
		"last_synced_at":         now,
		"status":                 effectiveGameStatus(game.IsEnabled, existing.PlatformEnabled),
	}
	if game.LocalThumbnailURL != "" {
		updates["local_thumbnail_url"] = game.LocalThumbnailURL
	}
	changed := existing.Name != game.Name ||
		existing.Slug != game.Slug ||
		existing.Category != game.Category ||
		existing.PlatformCategoryID != platformCategoryID ||
		existing.PlatformCategoryCode != platformCategoryCode ||
		existing.GameType != game.GameType ||
		existing.CoverURL != game.ThumbnailURL ||
		(game.LocalThumbnailURL != "" && existing.LocalThumbnailURL != game.LocalThumbnailURL) ||
		existing.SupportsDemo != game.SupportsDemo ||
		existing.SupportsFreeSpins != game.SupportsFreeSpins ||
		existing.ProviderEnabled != game.IsEnabled ||
		existing.MetadataHash != rawMetadataHash
	if err := content.ApplySite(r.db.WithContext(ctx).Model(&existing), "", scope).Updates(updates).Error; err != nil {
		return false, false, err
	}
	return false, changed, nil
}

func (r *Repository) UpdateGameLocalThumbnail(ctx context.Context, providerID uint64, providerGameID string, localURL string) error {
	localURL = strings.TrimSpace(localURL)
	if providerID == 0 || strings.TrimSpace(providerGameID) == "" || localURL == "" {
		return nil
	}
	var provider model.GameGameProviders
	_ = r.db.WithContext(ctx).Model(&model.GameGameProviders{}).Select("site_id").Where("id=?", providerID).Take(&provider).Error
	scope := content.Scope{SiteID: provider.SiteID}
	return content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGames{}), "", scope).
		Where("provider_id = ? AND provider_game_code = ? AND deleted_at IS NULL", providerID, providerGameID).
		Updates(map[string]any{"local_thumbnail_url": localURL, "updated_at": time.Now()}).Error
}

func (r *Repository) MarkMissingProviderGames(ctx context.Context, providerID uint64, providerGameIDs []string) (int, error) {
	var provider model.GameGameProviders
	_ = r.db.WithContext(ctx).Model(&model.GameGameProviders{}).Select("site_id").Where("id=?", providerID).Take(&provider).Error
	scope := content.Scope{SiteID: provider.SiteID}
	db := content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGames{}), "", scope).
		Where("provider_id = ? AND deleted_at IS NULL AND provider_enabled = 1", providerID)
	if len(providerGameIDs) > 0 {
		db = db.Where("provider_game_code NOT IN ?", providerGameIDs)
	}
	result := db.Updates(map[string]any{"provider_enabled": false, "status": 3})
	return int(result.RowsAffected), result.Error
}

func metadataHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func syncedJSONOrDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" || !json.Valid([]byte(value)) {
		return fallback
	}
	return value
}

func effectiveGameStatus(providerEnabled bool, platformEnabled bool) int {
	if providerEnabled && platformEnabled {
		return 1
	}
	return 3
}

func (r *Repository) ResolvePlatformCategory(ctx context.Context, scope content.Scope, providerCode string, providerCategoryCode string, gameType string) (*model.GameGameCategories, error) {
	providerCode = normalizeCatalogCode(providerCode)
	providerCategoryCode = normalizeProviderCategoryCode(providerCategoryCode)

	if providerCode != "" && providerCategoryCode != "" {
		var mapped model.GameGameCategories
		query := content.ApplySite(r.db.WithContext(ctx).Table("game_provider_category_mapping AS m"), "m", scope)
		query = content.ApplySite(query, "c", scope)
		err := query.
			Select("c.*").
			Joins("JOIN game_game_categories c ON c.id = m.platform_category_id AND c.deleted_at IS NULL AND c.status = ?", 1).
			Where("m.provider_code = ? AND m.provider_category_code = ? AND m.status = ? AND m.deleted_at IS NULL", providerCode, providerCategoryCode, 1).
			First(&mapped).Error
		if err == nil {
			return &mapped, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	categoryCode := defaultPlatformCategoryCode(providerCategoryCode, gameType)
	if categoryCode == "" {
		return nil, nil
	}
	return r.ensurePlatformCategory(ctx, scope, categoryCode)
}

func (r *Repository) ensurePlatformCategory(ctx context.Context, scope content.Scope, code string) (*model.GameGameCategories, error) {
	def := platformCategoryDefault(code)
	var category model.GameGameCategories
	err := content.ApplySite(r.db.WithContext(ctx).Model(&model.GameGameCategories{}), "", scope).Where("category_key = ?", def.CategoryKey).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category = def
		category.SiteID = scope.SiteID
		category.SiteID = scope.SiteID
		if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
			return nil, err
		}
		return &category, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) syncPrimaryCategoryRelation(ctx context.Context, gameID uint64, oldCategoryID uint64, newCategoryID uint64) error {
	if gameID == 0 {
		return nil
	}
	if oldCategoryID > 0 && oldCategoryID != newCategoryID {
		if err := r.db.WithContext(ctx).
			Unscoped().
			Where("game_id = ? AND category_id = ?", gameID, oldCategoryID).
			Delete(&model.GameGameCategoryRelations{}).Error; err != nil {
			return err
		}
	}
	if newCategoryID == 0 {
		return nil
	}

	relation := model.GameGameCategoryRelations{GameID: gameID, CategoryID: newCategoryID}
	return r.db.WithContext(ctx).
		Where("game_id = ? AND category_id = ?", gameID, newCategoryID).
		FirstOrCreate(&relation).Error
}

func normalizeCatalogCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeProviderCategoryCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func defaultPlatformCategoryCode(providerCategoryCode string, gameType string) string {
	switch normalizeProviderCategoryCode(providerCategoryCode) {
	case "slot", "slots", "video_slots", "videoslots":
		return "slots"
	case "pragmaticplaylive", "live", "live_casino", "livecasino", "casino":
		return "live-casino"
	case "table", "table_games", "tablegames", "card", "cards":
		return "table-games"
	case "crash", "instant", "instant_games", "instantgames":
		return "crash"
	case "fish", "fishing":
		return "fishing"
	case "sports", "sport", "virtual", "virtual_sports":
		return "sports"
	case "lottery", "lotto":
		return "lottery"
	case "arcade":
		return "arcade"
	}

	switch strings.ToLower(strings.TrimSpace(gameType)) {
	case "slot", "slots":
		return "slots"
	case "live", "live-casino", "casino":
		return "live-casino"
	case "table", "table-games", "card":
		return "table-games"
	case "crash", "instant":
		return "crash"
	case "fish", "fishing":
		return "fishing"
	case "sports", "sport":
		return "sports"
	default:
		return "slots"
	}
}

func platformCategoryDefault(code string) model.GameGameCategories {
	definitions := map[string]model.GameGameCategories{
		"slots":       {CategoryKey: "slots", Label: "Slots", ShortLabel: "S", Sort: 10, Status: 1},
		"live-casino": {CategoryKey: "live-casino", Label: "Live Casino", ShortLabel: "LC", Sort: 20, Status: 1},
		"table-games": {CategoryKey: "table-games", Label: "Table Games", ShortLabel: "TG", Sort: 30, Status: 1},
		"crash":       {CategoryKey: "crash", Label: "Crash", ShortLabel: "CR", Sort: 40, Status: 1},
		"fishing":     {CategoryKey: "fishing", Label: "Fishing", ShortLabel: "F", Sort: 50, Status: 1},
		"sports":      {CategoryKey: "sports", Label: "Sports", ShortLabel: "SP", Sort: 60, Status: 1},
		"lottery":     {CategoryKey: "lottery", Label: "Lottery", ShortLabel: "L", Sort: 70, Status: 1},
		"arcade":      {CategoryKey: "arcade", Label: "Arcade", ShortLabel: "A", Sort: 80, Status: 1},
	}
	if def, ok := definitions[code]; ok {
		return def
	}

	label := strings.ReplaceAll(code, "-", " ")
	if label != "" {
		label = strings.ToUpper(label[:1]) + label[1:]
	}
	shortLabel := ""
	if code != "" {
		shortLabel = strings.ToUpper(code[:1])
	}
	return model.GameGameCategories{CategoryKey: code, Label: label, ShortLabel: shortLabel, Sort: 100, Status: 1}
}

func (r *Repository) RecordSyncSuccess(ctx context.Context, summary SyncSummary) error {
	now := time.Now()
	row := model.GameProviderSyncs{

		SiteID:                 summary.SiteID,
		ProviderCode:           summary.ProviderCode,
		LastSuccessfulSyncAt:   &now,
		LastAttemptAt:          &now,
		LastStatus:             "SUCCESS",
		LastSyncedGameCount:    summary.Synced,
		LastInsertedGameCount:  summary.Inserted,
		LastUpdatedGameCount:   summary.Updated,
		LastUnchangedGameCount: summary.Unchanged,
	}
	updates := map[string]any{
		"last_successful_sync_at":     now,
		"last_attempt_at":             now,
		"last_status":                 "SUCCESS",
		"last_error":                  "",
		"last_synced_game_count":      summary.Synced,
		"last_inserted_game_count":    summary.Inserted,
		"last_updated_game_count":     summary.Updated,
		"last_unchanged_game_count":   summary.Unchanged,
		"last_provider_error_message": "",
	}
	scope := content.Scope{SiteID: summary.SiteID}
	result := content.ApplySiteStrict(r.db.WithContext(ctx).Model(&model.GameProviderSyncs{}), "", scope).
		Where("provider_code = ?", summary.ProviderCode).
		Updates(updates)
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (r *Repository) RecordSyncFailure(ctx context.Context, providerCode string, err error) error {
	return r.RecordSyncFailureScoped(ctx, content.Scope{}, providerCode, err)
}

func (r *Repository) RecordSyncFailureScoped(ctx context.Context, scope content.Scope, providerCode string, err error) error {
	now := time.Now()
	message := ""
	if err != nil {
		message = err.Error()
		if len(message) > 500 {
			message = message[:500]
		}
	}
	row := model.GameProviderSyncs{

		SiteID:        scope.SiteID,
		ProviderCode:  providerCode,
		LastAttemptAt: &now,
		LastStatus:    "FAILED",
		LastError:     message,
	}
	result := content.ApplySiteStrict(r.db.WithContext(ctx).Model(&model.GameProviderSyncs{}), "", scope).
		Where("provider_code = ?", providerCode).
		Updates(map[string]any{
			"last_attempt_at": now,
			"last_status":     "FAILED",
			"last_error":      message,
		})
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (r *Repository) FindLaunchGame(ctx context.Context, scope content.Scope, providerCode string, gameID string) (*model.GameGameProviders, *model.GameGames, error) {
	provider, game, err := r.findLaunchGame(ctx, scope, providerCode, gameID)
	if err == nil {
		return provider, game, nil
	}
	if scope.SiteScoped() && errors.Is(err, gorm.ErrRecordNotFound) {
		return r.findLaunchGame(ctx, content.Scope{Platform: true}, providerCode, gameID)
	}
	return nil, nil, err
}

func (r *Repository) findLaunchGame(ctx context.Context, scope content.Scope, providerCode string, gameID string) (*model.GameGameProviders, *model.GameGames, error) {
	provider, err := r.providerByCodeReadable(ctx, scope, providerCode)
	if err != nil {
		return nil, nil, err
	}
	var game model.GameGames
	err = applyReadableScope(r.db.WithContext(ctx).Model(&model.GameGames{}), "", scope).
		Where("provider_id = ? AND (provider_game_code = ? OR slug = ? OR game_code = ?)", provider.ID, gameID, gameID, gameID).
		First(&game).Error
	if err != nil {
		return nil, nil, err
	}
	return provider, &game, nil
}

func (r *Repository) providerByCodeReadable(ctx context.Context, scope content.Scope, code string) (*model.GameGameProviders, error) {
	var provider model.GameGameProviders
	query := applyReadableScope(r.db.WithContext(ctx).Model(&model.GameGameProviders{}), "", scope).
		Where("provider_code = ?", code)
	if err := query.First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func applyReadableScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	if scope.Platform {
		return db.Where(repositoryColumn(alias, "site_id")+" = ?", 0)
	}
	return content.ApplySite(db, alias, scope)
}

func repositoryColumn(alias string, name string) string {
	if strings.TrimSpace(alias) == "" {
		return name
	}
	return alias + "." + name
}

func (r *Repository) PlayerMapping(ctx context.Context, scope content.Scope, providerCode string, userID uint64, currency string) (*model.GameProviderPlayers, error) {
	var player model.GameProviderPlayers
	err := content.ApplySite(r.db.WithContext(ctx).Model(&model.GameProviderPlayers{}), "", scope).
		Where("provider_code = ? AND user_id = ? AND currency = ?", providerCode, userID, currency).
		First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *Repository) PlayerByProviderID(ctx context.Context, providerCode string, providerPlayerID string) (*model.GameProviderPlayers, error) {
	var player model.GameProviderPlayers
	err := r.db.WithContext(ctx).
		Where("provider_code = ? AND provider_player_id = ? AND status = ?", providerCode, providerPlayerID, 1).
		First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *Repository) CreatePlayerMapping(ctx context.Context, player *model.GameProviderPlayers) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(player).Error
	if err != nil {
		return err
	}
	if player.ID > 0 {
		return nil
	}
	existing, err := r.PlayerMapping(ctx, content.Scope{SiteID: player.SiteID}, player.ProviderCode, player.UserID, player.Currency)
	if err != nil {
		return err
	}
	*player = *existing
	return nil
}

func gameStatus(enabled bool) int {
	if enabled {
		return 1
	}
	return 3
}

func gameCode(providerCode string, slug string) string {
	code := strings.ToLower(strings.Trim(providerCode+"-"+slug, "-"))
	if len(code) <= 64 {
		return code
	}
	sum := sha1.Sum([]byte(code)) //nolint:gosec
	suffix := hex.EncodeToString(sum[:])[:8]
	return strings.TrimRight(code[:55], "-") + "-" + suffix
}
