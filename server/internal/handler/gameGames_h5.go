package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/game/connections"
	"casino/internal/game/domain"
	"casino/internal/game/market"
	gameservice "casino/internal/game/service"
	"casino/internal/model"
	"casino/internal/storage"
)

type h5GameDTO struct {
	ID                   uint64   `json:"id"`
	GameCode             string   `json:"gameCode"`
	ProviderID           uint64   `json:"providerID"`
	ProviderCode         string   `json:"providerCode"`
	ProviderGameID       string   `json:"providerGameId"`
	ProviderGameCode     string   `json:"providerGameCode"`
	Name                 string   `json:"name"`
	NameI18n             any      `json:"nameI18n,omitempty"`
	Slug                 string   `json:"slug"`
	CategoryCode         string   `json:"categoryCode"`
	PlatformCategoryID   uint64   `json:"platformCategoryID"`
	PlatformCategoryCode string   `json:"platformCategoryCode"`
	ProviderCategoryCode string   `json:"providerCategoryCode"`
	GameType             string   `json:"gameType"`
	GameTypeCode         string   `json:"gameTypeCode"`
	ThumbnailURL         string   `json:"thumbnailUrl"`
	CoverURL             string   `json:"coverURL"`
	BackgroundURL        string   `json:"backgroundUrl"`
	LaunchImageURL       string   `json:"launchImageUrl"`
	RTP                  *float64 `json:"rtp,omitempty"`
	Volatility           string   `json:"volatility,omitempty"`
	SupportsDemo         bool     `json:"supportsDemo"`
	SupportsFreeSpins    bool     `json:"supportsFreeSpins"`
	ProviderEnabled      bool     `json:"providerEnabled"`
	PlatformEnabled      bool     `json:"platformEnabled"`
	EffectiveEnabled     bool     `json:"effectiveEnabled"`
	IsEnabled            bool     `json:"isEnabled"`
	IsHot                bool     `json:"isHot"`
	IsFeatured           bool     `json:"isFeatured"`
	IsNew                bool     `json:"isNew"`
	MaintenanceMessage   string   `json:"maintenanceMessage,omitempty"`
	Badge                string   `json:"badge,omitempty"`
	PlayerCount          uint     `json:"playerCount"`
	HotScore             int      `json:"hotScore"`
	SortOrder            int      `json:"sortOrder"`
	Status               int      `json:"status"`
}

type h5GameListFilters struct {
	Category     string
	GameType     string
	Keyword      string
	ProviderCode string
	SortMode     string
	FavoriteOnly bool
}

func (h *gameGamesHandler) PublicList(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	filters := h5GameListFilters{
		Category:     strings.TrimSpace(c.Query("category")),
		GameType:     strings.TrimSpace(c.Query("type")),
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		ProviderCode: strings.ToLower(strings.TrimSpace(c.Query("provider"))),
		SortMode:     strings.ToLower(strings.TrimSpace(c.Query("sort"))),
		FavoriteOnly: strings.EqualFold(strings.TrimSpace(c.Query("favorite")), "1") ||
			strings.EqualFold(strings.TrimSpace(c.Query("favorite")), "true"),
	}
	page := h5Page(c)
	pageSize := h5PageSize(c, 30)

	queryScope := scope
	q, categoryKey := h5BuildPublicGameQuery(c, db, queryScope, filters)
	if categoryKey == "hot" && !h5CategoryConfigured(db, queryScope, "hot") {
		// HOT must only use games already visible to this site. Redis is only
		// the ranking source; it must never expand the site's game catalogue.
		h5RespondHotGames(c, db, q, queryScope, scope, page, pageSize)
		return
	}

	var total int64
	_ = q.Count(&total).Error
	if total == 0 && scope.SiteScoped() && scope.SiteID != 1 && !filters.FavoriteOnly {
		queryScope = h5PlatformCatalogScope()
		q, categoryKey = h5BuildPublicGameQuery(c, db, queryScope, filters)
		if categoryKey == "hot" && !h5CategoryConfigured(db, queryScope, "hot") {
			h5RespondHotGames(c, db, q, queryScope, scope, page, pageSize)
			return
		}
		_ = q.Count(&total).Error
	}

	var games []model.GameGames
	orderExpr := "game_games.hot_score DESC, game_games.sort ASC, game_games.id ASC"
	if categoryKey != "" && h5CategoryConfigured(db, queryScope, categoryKey) {
		orderExpr = "display_relation.sort ASC,game_games.sort ASC,game_games.id ASC"
	}
	switch filters.SortMode {
	case "name-asc":
		orderExpr = "game_games.name ASC, game_games.id ASC"
	case "name-desc":
		orderExpr = "game_games.name DESC, game_games.id ASC"
	case "provider":
		orderExpr = "gp.provider_code ASC, game_games.sort ASC, game_games.id ASC"
	case "newest":
		orderExpr = "game_games.created_at DESC, game_games.id DESC"
	}
	_ = q.Order(orderExpr).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&games).Error

	providers := h5ProvidersByID(db, queryScope, games)
	items := h5GameDTOs(c.Request.Context(), games, providers)
	response.Success(c, gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func h5BuildPublicGameQuery(c *gin.Context, db *gorm.DB, scope content.Scope, filters h5GameListFilters) (*gorm.DB, string) {
	q := db.Model(&model.GameGames{}).
		Joins("JOIN game_game_providers gp ON gp.id = game_games.provider_id AND gp.deleted_at IS NULL").
		Where("game_games.status IN (1,2) AND game_games.provider_enabled = ? AND game_games.platform_enabled = ? AND gp.status = ?", true, true, 1)
	q = h5ApplyReadableScope(q, "game_games", scope)
	q = h5ApplyReadableScope(q, "gp", scope)
	deployment, marketErr := market.Load(db.WithContext(c.Request.Context()))
	if marketErr != nil {
		q = q.Where("1=0")
	}
	q = market.ApplyGames(q, "game_games", deployment)
	country, currency := deployment.Country, deployment.Currency
	q = h5ApplyGameMaintenance(q, scope, country, currency)
	supportedProviderCodes := h5SupportedGameProviderCodes(db)
	if len(supportedProviderCodes) > 0 {
		q = q.Where("gp.provider_code IN ?", supportedProviderCodes)
	} else {
		q = q.Where("1 = 0")
	}
	if filters.ProviderCode != "" {
		q = q.Where("gp.provider_code = ?", filters.ProviderCode)
	}
	if filters.FavoriteOnly {
		user, _, ok := h5AuthenticatedUser(db, c)
		if !ok {
			q = q.Where("1 = 0")
		} else {
			q = q.Joins(
				"JOIN game_user_favorite_games gufg ON gufg.game_id = game_games.id AND gufg.user_id = ? AND gufg.deleted_at IS NULL",
				user.ID,
			)
		}
	}
	categoryKey := ""
	if filters.Category != "" {
		categoryKey = h5NormalizeCategoryKey(filters.Category)
		var category model.GameGameCategories
		err := h5ApplyReadableScope(db.Model(&category), "", scope).Where("category_key=? AND status=1", categoryKey).Take(&category).Error
		if err == nil {
			q = q.Joins("JOIN game_game_category_relations display_relation ON display_relation.game_id=game_games.id AND display_relation.category_id=? AND display_relation.deleted_at IS NULL", category.ID)
			q = h5ApplyReadableScope(q, "display_relation", scope)
		} else if categoryKey == "hot" {
			var configured int64
			h5ApplyReadableScope(db.Model(&model.GameGameCategories{}), "", scope).Where("category_key=?", "hot").Count(&configured)
			if configured > 0 {
				q = q.Where("1=0")
			}
		} else {
			q = q.Where("1=0")
		}
	}

	if filters.GameType != "" {
		q = q.Where("game_games.game_type = ?", h5GameType(filters.GameType))
	}
	if filters.Keyword != "" {
		q = q.Where("game_games.name LIKE ?", "%"+filters.Keyword+"%")
	}
	return q, categoryKey
}

func h5FindPublicGameByID(db *gorm.DB, scope content.Scope, id uint64, onlineOnly bool) (model.GameGames, content.Scope, error) {
	find := func(current content.Scope) (model.GameGames, error) {
		var game model.GameGames
		m, err := market.Load(db)
		if err != nil {
			return game, err
		}
		q := h5ApplyReadableScope(db.Model(&model.GameGames{}), "game_games", current).Joins("JOIN game_game_providers gp ON gp.id=game_games.provider_id AND gp.deleted_at IS NULL").Where("game_games.id=? AND game_games.status IN (1,2) AND game_games.provider_enabled=1 AND game_games.platform_enabled=1 AND gp.status=1", id)
		if onlineOnly {
			q = q.Where("game_games.status=1")
		}
		q = market.ApplyGames(q, "game_games", m).Where("gp.provider_code IN ?", connections.New(db).PublicCodes(context.Background()))
		return game, q.First(&game).Error
	}
	game, err := find(scope)
	if err == nil {
		return game, scope, nil
	}
	if scope.SiteScoped() && errors.Is(err, gorm.ErrRecordNotFound) {
		catalog := h5PlatformCatalogScope()
		if game, e := find(catalog); e == nil {
			return game, catalog, nil
		}
	}
	return model.GameGames{}, scope, err
}

func (h *gameGamesHandler) PublicDetail(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	gameID := h5PathUint64(c, "id")
	if gameID == 0 {
		h5Error(c, http.StatusBadRequest, "invalid game id")
		return
	}

	game, gameScope, err := h5FindPublicGameByID(db, scope, gameID, false)
	if err != nil {
		h5Error(c, http.StatusNotFound, "game not found")
		return
	}

	providers := h5ProvidersByID(db, gameScope, []model.GameGames{game})
	response.Success(c, h5GameDTOFromModel(c.Request.Context(), game, providers[game.ProviderID]))
}

func (h *gameGamesHandler) Launch(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	gameID := h5PathUint64(c, "id")
	if gameID == 0 {
		h5Error(c, http.StatusBadRequest, "invalid game id")
		return
	}

	game, gameScope, err := h5FindPublicGameByID(db, scope, gameID, true)
	if err != nil {
		h5Error(c, http.StatusNotFound, "game not found")
		return
	}

	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope = h5UserContentScope(db, user.ID, scope)
	providers := h5ProvidersByID(db, gameScope, []model.GameGames{game})
	provider := providers[game.ProviderID]
	result, err := gameservice.NewRuntime(db).LaunchService.Launch(c.Request.Context(), gameservice.LaunchRequest{SiteID: 1, ProviderCode: provider.ProviderCode, GameID: game.ProviderGameCode, UserID: user.ID, Device: "web", IPAddress: c.ClientIP()})
	if err != nil {
		var walletError *domain.WalletError
		if errors.As(err, &walletError) {
			h5MachineError(c, http.StatusBadRequest, string(walletError.Code), walletError.Error())
		} else {
			h5Error(c, http.StatusBadGateway, "game provider is unavailable")
		}
		return
	}
	response.Success(c, gin.H{"game": h5GameDTOFromModel(c.Request.Context(), game, provider), "launchUrl": result.LaunchURL, "token": result.SessionID})
}

func (h *gameGamesHandler) ToggleFavorite(c *gin.Context) {
	db := database.GetDB()
	gameID := h5PathUint64(c, "id")
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	if gameID == 0 {
		h5Error(c, http.StatusBadRequest, "invalid game")
		return
	}

	var favorite model.GameUserFavoriteGames
	err := db.Where("user_id = ? AND game_id = ?", user.ID, gameID).First(&favorite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		favorite = model.GameUserFavoriteGames{UserID: user.ID, GameID: gameID}
		_ = db.Create(&favorite).Error
		response.Success(c, gin.H{"favorite": true})
		return
	}
	if err == nil {
		_ = db.Delete(&favorite).Error
	}
	response.Success(c, gin.H{"favorite": false})
}

func (h *gameGamesHandler) FavoriteStatus(c *gin.Context) {
	db := database.GetDB()
	gameID := h5PathUint64(c, "id")
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	if gameID == 0 {
		h5Error(c, http.StatusBadRequest, "invalid game")
		return
	}

	var count int64
	if err := db.Model(&model.GameUserFavoriteGames{}).
		Where("user_id = ? AND game_id = ?", user.ID, gameID).
		Count(&count).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "favorite status failed")
		return
	}
	response.Success(c, gin.H{"favorite": count > 0})
}

func h5Page(c *gin.Context) int {
	page, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	if err != nil || page <= 0 {
		return 1
	}
	return page
}

func h5PageSize(c *gin.Context, defaultSize int) int {
	pageSize, err := strconv.Atoi(strings.TrimSpace(h5FirstNonEmpty(c.Query("pageSize"), c.Query("limit"), strconv.Itoa(defaultSize))))
	if err != nil || pageSize <= 0 {
		return defaultSize
	}
	if pageSize > 200 {
		return 200
	}
	return pageSize
}

func h5ProvidersByID(db *gorm.DB, scope content.Scope, games []model.GameGames) map[uint64]model.GameGameProviders {
	providerIDs := make([]uint64, 0)
	seen := map[uint64]bool{}
	for _, game := range games {
		if game.ProviderID == 0 || seen[game.ProviderID] {
			continue
		}
		seen[game.ProviderID] = true
		providerIDs = append(providerIDs, game.ProviderID)
	}
	if len(providerIDs) == 0 {
		return map[uint64]model.GameGameProviders{}
	}
	var providers []model.GameGameProviders
	_ = h5ApplyReadableScope(db.Model(&model.GameGameProviders{}), "", scope).Where("id IN ?", providerIDs).Find(&providers).Error
	out := make(map[uint64]model.GameGameProviders, len(providers))
	for _, provider := range providers {
		out[provider.ID] = provider
	}
	if scope.SiteScoped() && len(out) < len(providerIDs) {
		var platformProviders []model.GameGameProviders
		_ = h5ApplyReadableScope(db.Model(&model.GameGameProviders{}), "", h5PlatformCatalogScope()).
			Where("id IN ?", providerIDs).
			Find(&platformProviders).Error
		for _, provider := range platformProviders {
			if _, exists := out[provider.ID]; !exists {
				out[provider.ID] = provider
			}
		}
	}
	return out
}

func h5RespondHotGames(c *gin.Context, db *gorm.DB, base *gorm.DB, catalogScope content.Scope, activityScope content.Scope, page int, pageSize int) {
	const hotLimit = 12
	page = 1
	pageSize = hotLimit
	games := h5ResolveHotGames(c.Request.Context(), base, activityScope)

	providerScope := catalogScope
	if activityScope.SiteScoped() {
		providerScope = activityScope
	}
	providers := h5ProvidersByID(db, providerScope, games)
	items := h5GameDTOs(c.Request.Context(), games, providers)
	response.Success(c, gin.H{
		"list":     items,
		"total":    len(items),
		"page":     page,
		"pageSize": pageSize,
	})
}

func h5ResolveHotGames(ctx context.Context, base *gorm.DB, activityScope content.Scope) []model.GameGames {
	const hotLimit = 12
	games := make([]model.GameGames, 0, hotLimit)

	if rankedIDs, err := h5HotRankingStore().TopGameIDs(ctx, activityScope, time.Now(), hotLimit); err == nil && len(rankedIDs) > 0 {
		_ = base.Session(&gorm.Session{}).
			Where("game_games.id IN ?", rankedIDs).
			Order(h5IDOrderExpr("game_games.id", rankedIDs)).
			Limit(hotLimit).
			Find(&games).Error
	}

	// Fill missing slots with eligible random games, without duplicates. This
	// also makes the endpoint return twelve items whenever twelve eligible games
	// exist, even when Redis has no ranking entries yet.
	ids := make([]uint64, 0, len(games))
	for _, game := range games {
		ids = append(ids, game.ID)
	}
	randomQuery := base.Session(&gorm.Session{})
	if len(ids) > 0 {
		randomQuery = randomQuery.Where("game_games.id NOT IN ?", ids)
	}
	var randomGames []model.GameGames
	_ = randomQuery.Order("RAND()").Limit(hotLimit - len(games)).Find(&randomGames).Error
	games = append(games, randomGames...)
	if len(games) >= hotLimit {
		return games
	}

	// A site may have fewer than twelve eligible games. Repeat only those
	// site games so the HOT response still has twelve items without leaking
	// games that the site has not imported.
	if len(games) == 0 {
		return games
	}
	repeatSource := append([]model.GameGames(nil), games...)
	for len(games) < hotLimit {
		games = append(games, repeatSource[(len(games)-len(repeatSource))%len(repeatSource)])
	}
	return games[:hotLimit]
}

func h5GameDTOs(ctx context.Context, games []model.GameGames, providers map[uint64]model.GameGameProviders) []h5GameDTO {
	items := make([]h5GameDTO, 0, len(games))
	for _, game := range games {
		items = append(items, h5GameDTOFromModel(ctx, game, providers[game.ProviderID]))
	}
	return items
}

func h5GameDTOFromModel(ctx context.Context, game model.GameGames, provider model.GameGameProviders) h5GameDTO {
	providerEnabled := game.ProviderEnabled && (provider.ID == 0 || provider.Status == 1)
	platformEnabled := game.PlatformEnabled
	effectiveEnabled := providerEnabled && platformEnabled && game.Status == 1
	badge := strings.ToLower(strings.TrimSpace(game.Badge))
	gameType := h5GameTypeCode(game.GameType)
	platformCategoryCode := h5FirstNonEmpty(game.PlatformCategoryCode, h5GameCategoryCode(game.GameType))
	thumbnailURL := strings.TrimSpace(game.LocalThumbnailURL)
	if thumbnailURL == "" {
		thumbnailURL = strings.TrimSpace(game.CoverURL)
	}
	if thumbnailURL == "" {
		thumbnailURL = "/static/h5/images/game-placeholder.svg"
	}
	if signedURL, err := storage.PublicURL(ctx, thumbnailURL); err == nil && signedURL != "" {
		thumbnailURL = signedURL
	}
	backgroundURL := h5SignedAsset(ctx, game.BackgroundURL)
	launchImageURL := h5SignedAsset(ctx, game.LaunchImageURL)
	return h5GameDTO{
		ID:                   game.ID,
		GameCode:             game.GameCode,
		ProviderID:           game.ProviderID,
		ProviderCode:         provider.ProviderCode,
		ProviderGameID:       game.ProviderGameCode,
		ProviderGameCode:     game.ProviderGameCode,
		Name:                 game.Name,
		NameI18n:             h5JSONValue(game.NameI18n),
		Slug:                 game.Slug,
		CategoryCode:         platformCategoryCode,
		PlatformCategoryID:   game.PlatformCategoryID,
		PlatformCategoryCode: platformCategoryCode,
		ProviderCategoryCode: game.Category,
		GameType:             gameType,
		GameTypeCode:         game.GameType,
		ThumbnailURL:         thumbnailURL,
		CoverURL:             thumbnailURL,
		BackgroundURL:        backgroundURL,
		LaunchImageURL:       launchImageURL,
		RTP:                  game.RTP,
		Volatility:           game.Volatility,
		SupportsDemo:         game.SupportsDemo,
		SupportsFreeSpins:    game.SupportsFreeSpins,
		ProviderEnabled:      providerEnabled,
		PlatformEnabled:      platformEnabled,
		EffectiveEnabled:     effectiveEnabled,
		IsEnabled:            effectiveEnabled,
		IsHot:                game.IsHot || game.HotScore > 0 || badge == "hot",
		IsFeatured:           game.IsFeatured || badge == "featured" || badge == "recommend" || badge == "recommended",
		IsNew:                game.IsNew || badge == "new",
		MaintenanceMessage:   game.MaintenanceMessage,
		Badge:                game.Badge,
		PlayerCount:          game.PlayerCount,
		HotScore:             game.HotScore,
		SortOrder:            game.Sort,
		Status:               game.Status,
	}
}

func h5SignedAsset(ctx context.Context, value string) string {
	if signedURL, err := storage.PublicURL(ctx, value); err == nil && signedURL != "" {
		return signedURL
	}
	return value
}

func h5JSONValue(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var value any
	if json.Unmarshal([]byte(raw), &value) != nil {
		return raw
	}
	return value
}

func h5GameCategoryCode(gameType string) string {
	switch strings.ToLower(strings.TrimSpace(gameType)) {
	case "slot", "slots":
		return "slots"
	case "live", "live-casino":
		return "live-casino"
	case "chance", "casino":
		return "live-casino"
	case "fish", "fishing":
		return "fishing"
	case "sports", "sport":
		return "sports"
	case "table", "table-games":
		return "table-games"
	case "crash", "instant":
		return "crash"
	case "lottery", "lotto":
		return "lottery"
	case "arcade":
		return "arcade"
	default:
		return strings.ToLower(strings.TrimSpace(gameType))
	}
}

func h5NormalizeCategoryKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func h5GameTypeCode(gameType string) string {
	switch strings.ToLower(strings.TrimSpace(gameType)) {
	case "slot", "slots":
		return "SLOT"
	case "live", "live-casino":
		return "LIVE_CASINO"
	case "chance", "casino":
		return "CHANCE"
	case "fish", "fishing":
		return "FISHING"
	case "sports", "sport":
		return "SPORTS"
	case "table", "table-games":
		return "TABLE"
	default:
		return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(gameType), "-", "_"))
	}
}

func h5CategoryConfigured(db *gorm.DB, scope content.Scope, code string) bool {
	var count int64
	h5ApplyReadableScope(db.Model(&model.GameGameCategories{}), "", scope).Where("category_key=? AND status=1", code).Count(&count)
	return count > 0
}
