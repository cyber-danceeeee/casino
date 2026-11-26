package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type h5LobbyBannerDTO struct {
	ID           uint64 `json:"id"`
	BannerCode   string `json:"bannerCode"`
	Position     string `json:"position"`
	Title        string `json:"title"`
	ImageURL     string `json:"imageUrl"`
	LinkType     string `json:"linkType"`
	LinkURL      string `json:"linkUrl"`
	ActivitySlug string `json:"activitySlug"`
	SortOrder    int    `json:"sortOrder"`
}

type h5LobbyCategoryDTO struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	NameZh    string `json:"nameZh"`
	NameEn    string `json:"nameEn"`
	IconURL   string `json:"iconUrl"`
	BannerURL string `json:"bannerUrl"`
	SortOrder int    `json:"sortOrder"`
}

type h5LobbyProviderDTO struct {
	ID                  uint64 `json:"id"`
	ProviderCode        string `json:"providerCode"`
	Name                string `json:"name"`
	LogoURL             string `json:"logoUrl"`
	WalletMode          string `json:"walletMode"`
	SupportedCountries  string `json:"supportedCountries"`
	SupportedCurrencies string `json:"supportedCurrencies"`
	Maintenance         bool   `json:"maintenance"`
	MaintenanceMessage  string `json:"maintenanceMessage,omitempty"`
	SortOrder           int    `json:"sortOrder"`
}

type h5LobbySectionDTO struct {
	ID                  uint64 `json:"id"`
	SectionCode         string `json:"sectionCode"`
	SectionType         string `json:"sectionType"`
	TitleZh             string `json:"titleZh"`
	TitleEn             string `json:"titleEn"`
	DisplayStyle        string `json:"displayStyle"`
	SourceType          string `json:"sourceType"`
	SourceRef           string `json:"sourceRef"`
	MaxItems            int    `json:"maxItems"`
	SortOrder           int    `json:"sortOrder"`
	SupportedCountries  string `json:"supportedCountries"`
	SupportedCurrencies string `json:"supportedCurrencies"`
	Config              any    `json:"config,omitempty"`
	Items               any    `json:"items,omitempty"`
}

func (h *gameGamesHandler) Lobby(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	country := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(c.Query("country"), c.GetHeader("X-Country-Code"))))
	currency := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(c.Query("currency"), c.GetHeader("X-Currency-Code"))))

	sections, err := h5LobbySections(db, scope, country, currency)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "lobby config failed")
		return
	}

	banners := h5LobbyBanners(c.Request.Context(), db, scope)
	categories := h5LobbyCategories(c.Request.Context(), db, scope)
	providers := h5LobbyProviders(c.Request.Context(), db, scope, country, currency)
	gameBuckets := h5LobbyGameBuckets(c, db, scope, country, currency, sections)
	jackpot := h5LobbyJackpot(db)

	sectionPayloads := make([]h5LobbySectionDTO, 0, len(sections))
	for _, section := range sections {
		item := h5LobbySectionDTO{
			ID:                  section.ID,
			SectionCode:         section.SectionCode,
			SectionType:         section.SectionType,
			TitleZh:             section.TitleZh,
			TitleEn:             section.TitleEn,
			DisplayStyle:        section.DisplayStyle,
			SourceType:          section.SourceType,
			SourceRef:           section.SourceRef,
			MaxItems:            h5LobbyLimit(section.MaxItems),
			SortOrder:           section.SortOrder,
			SupportedCountries:  section.SupportedCountries,
			SupportedCurrencies: section.SupportedCurrencies,
			Config:              h5JSONValue(section.ConfigJSON),
		}
		switch section.SectionCode {
		case "top-banner":
			item.Items = banners
		case "jackpot":
			item.Items = jackpot
		case "hot-games":
			item.Items = gameBuckets["hotGames"]
		case "new-games":
			item.Items = gameBuckets["newGames"]
		case "recommended-games":
			item.Items = gameBuckets["recommendedGames"]
		case "continue-playing":
			item.Items = gameBuckets["continuePlaying"]
		case "recently-played":
			item.Items = gameBuckets["recentlyPlayed"]
		case "categories":
			item.Items = categories
		case "providers":
			item.Items = providers
		}
		sectionPayloads = append(sectionPayloads, item)
	}

	response.Success(c, gin.H{
		"banners":          banners,
		"categories":       categories,
		"hotGames":         gameBuckets["hotGames"],
		"newGames":         gameBuckets["newGames"],
		"recommendedGames": gameBuckets["recommendedGames"],
		"continuePlaying":  gameBuckets["continuePlaying"],
		"recentlyPlayed":   gameBuckets["recentlyPlayed"],
		"providers":        providers,
		"jackpot":          jackpot,
		"sections":         sectionPayloads,
	})
}

func h5LobbySections(db *gorm.DB, scope content.Scope, country string, currency string) ([]model.GameLobbySections, error) {
	var rows []model.GameLobbySections
	q := db.Model(&model.GameLobbySections{}).Where("deleted_at IS NULL AND enabled = ? AND publish_status = ?", true, "PUBLISHED")
	q = h5LobbyScope(q, scope)
	q = h5Availability(q, "game_lobby_sections", country, currency)
	if err := q.Order("site_id DESC, sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]model.GameLobbySections, 0, len(rows))
	for _, row := range rows {
		if seen[row.SectionCode] {
			continue
		}
		seen[row.SectionCode] = true
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder == out[j].SortOrder {
			return out[i].ID < out[j].ID
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out, nil
}

func h5LobbyScope(db *gorm.DB, scope content.Scope) *gorm.DB {
	if scope.SiteID > 0 {
		db = db
	}
	if scope.SiteID > 0 {
		db = db.Where("site_id IN (0, ?)", scope.SiteID)
	}
	return db
}

func h5LobbyBanners(ctx context.Context, db *gorm.DB, scope content.Scope) []h5LobbyBannerDTO {
	now := time.Now()
	var rows []model.GameBanners
	_ = content.ApplySite(db.Model(&model.GameBanners{}), "", scope).
		Where("status = ? AND position IN ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", 1, []string{"home", "top-banner"}, now, now).
		Order("sort ASC, id ASC").
		Limit(20).
		Find(&rows).Error
	out := make([]h5LobbyBannerDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, h5LobbyBannerDTO{ID: row.ID, BannerCode: row.BannerCode, Position: row.Position, Title: row.Title, ImageURL: h5SignedAsset(ctx, row.ImageURL), LinkType: row.LinkType, LinkURL: row.LinkURL, ActivitySlug: row.ActivitySlug, SortOrder: row.Sort})
	}
	return out
}

func h5LobbyCategories(ctx context.Context, db *gorm.DB, scope content.Scope) []h5LobbyCategoryDTO {
	var rows []model.GameGameCategories
	_ = content.ApplySite(db.Model(&model.GameGameCategories{}), "", scope).
		Where("status = ?", 1).
		Order("sort ASC, id ASC").
		Find(&rows).Error
	out := make([]h5LobbyCategoryDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, h5LobbyCategoryDTO{ID: row.ID, Code: row.CategoryKey, NameZh: row.NameZh, NameEn: row.NameEn, IconURL: h5SignedAsset(ctx, row.IconURL), BannerURL: h5SignedAsset(ctx, row.BannerURL), SortOrder: row.Sort})
	}
	return out
}

func h5LobbyProviders(ctx context.Context, db *gorm.DB, scope content.Scope, country string, currency string) []h5LobbyProviderDTO {
	var rows []model.GameGameProviders
	q := content.ApplySite(db.Model(&model.GameGameProviders{}), "", scope).Where("status = ?", 1)
	q = h5Availability(q, "game_game_providers", country, currency)
	if supported := h5SupportedGameProviderCodes(); len(supported) > 0 {
		q = q.Where("provider_code IN ?", supported)
	} else {
		q = q.Where("1 = 0")
	}
	_ = q.Order("sort ASC, id ASC").Find(&rows).Error
	out := make([]h5LobbyProviderDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, h5LobbyProviderDTO{ID: row.ID, ProviderCode: row.ProviderCode, Name: row.Name, LogoURL: h5SignedAsset(ctx, row.LogoURL), WalletMode: row.WalletMode, SupportedCountries: row.SupportedCountries, SupportedCurrencies: row.SupportedCurrencies, Maintenance: row.Status == 2, MaintenanceMessage: row.MaintenanceMessage, SortOrder: row.Sort})
	}
	return out
}

func h5LobbyGameBuckets(c *gin.Context, db *gorm.DB, scope content.Scope, country string, currency string, sections []model.GameLobbySections) map[string][]h5GameDTO {
	limits := h5LobbySectionLimits(sections)
	out := map[string][]h5GameDTO{
		"hotGames":         h5LobbyGames(c.Request.Context(), db, scope, country, currency, h5LobbyLimitValue(limits, "hot-games"), "hot"),
		"newGames":         h5LobbyGames(c.Request.Context(), db, scope, country, currency, h5LobbyLimitValue(limits, "new-games"), "new"),
		"recommendedGames": h5LobbyGames(c.Request.Context(), db, scope, country, currency, h5LobbyLimitValue(limits, "recommended-games"), "featured"),
		"continuePlaying":  {},
		"recentlyPlayed":   {},
	}
	if user, _, ok := h5AuthenticatedUser(db, c); ok {
		recentLimit := h5LobbyLimitValue(limits, "recently-played")
		out["recentlyPlayed"] = h5LobbyRecentGames(c.Request.Context(), db, scope, country, currency, user.ID, recentLimit)
		out["continuePlaying"] = out["recentlyPlayed"]
	}
	return out
}

func h5LobbyGames(ctx context.Context, db *gorm.DB, scope content.Scope, country string, currency string, limit int, bucket string) []h5GameDTO {
	var games []model.GameGames
	catalogScope := scope
	q := h5LobbyGameQuery(db, catalogScope, country, currency)
	switch bucket {
	case "hot":
		games = h5ResolveHotGames(ctx, q, scope)
	case "new":
		q = q.Where("(game_games.is_new = 1 OR game_games.badge = ?)", "new").Order("game_games.created_at DESC, game_games.sort ASC, game_games.id ASC")
		_ = q.Limit(limit).Find(&games).Error
	default:
		q = q.Where("(game_games.is_featured = 1 OR game_games.badge IN ?)", []string{"featured", "recommend", "recommended"}).Order("game_games.sort ASC, game_games.hot_score DESC, game_games.id ASC")
		_ = q.Limit(limit).Find(&games).Error
	}
	providers := h5ProvidersByID(db, catalogScope, games)
	return h5GameDTOs(ctx, games, providers)
}

func h5LobbyRecentGames(ctx context.Context, db *gorm.DB, scope content.Scope, country string, currency string, userID uint64, limit int) []h5GameDTO {
	var ids []uint64
	_ = content.ApplySite(db.Table("game_game_launch_logs"), "", scope).
		Where("user_id = ? AND game_id > 0 AND deleted_at IS NULL", userID).
		Group("game_id").
		Order("MAX(created_at) DESC").
		Limit(limit).
		Pluck("game_id", &ids).Error
	if len(ids) == 0 {
		return nil
	}
	var games []model.GameGames
	_ = h5LobbyGameQuery(db, scope, country, currency).
		Where("game_games.id IN ?", ids).
		Order(h5IDOrderExpr("game_games.id", ids)).
		Find(&games).Error
	providers := h5ProvidersByID(db, scope, games)
	return h5GameDTOs(ctx, games, providers)
}

func h5LobbyGameQuery(db *gorm.DB, scope content.Scope, country string, currency string) *gorm.DB {
	q := db.Model(&model.GameGames{}).
		Joins("JOIN game_game_providers gp ON gp.id = game_games.provider_id AND gp.deleted_at IS NULL").
		Where("game_games.status = ? AND game_games.provider_enabled = ? AND game_games.platform_enabled = ? AND gp.status = ?", 1, true, true, 1)
	q = content.ApplySite(q, "game_games", scope)
	q = content.ApplySite(q, "gp", scope)
	q = h5Availability(q, "game_games", country, currency)
	q = h5Availability(q, "gp", country, currency)
	if supported := h5SupportedGameProviderCodes(); len(supported) > 0 {
		q = q.Where("gp.provider_code IN ?", supported)
	} else {
		q = q.Where("1 = 0")
	}
	return q
}

func h5Availability(db *gorm.DB, alias string, country string, currency string) *gorm.DB {
	if country != "" {
		db = db.Where("("+alias+".supported_countries = '' OR FIND_IN_SET(?, "+alias+".supported_countries))", country)
	}
	if currency != "" {
		db = db.Where("("+alias+".supported_currencies = '' OR FIND_IN_SET(?, "+alias+".supported_currencies))", currency)
	}
	return db
}

func h5LobbyJackpot(db *gorm.DB) gin.H {
	var jackpot model.GameRebateJackpots
	_ = db.Where("status = ?", 1).Order("id DESC").First(&jackpot).Error
	return h5JackpotPayload(jackpot)
}

func h5LobbySectionLimits(sections []model.GameLobbySections) map[string]int {
	out := make(map[string]int, len(sections))
	for _, section := range sections {
		out[section.SectionCode] = h5LobbyLimit(section.MaxItems)
	}
	return out
}

func h5LobbyLimitValue(limits map[string]int, sectionCode string) int {
	if value := limits[sectionCode]; value > 0 {
		return value
	}
	return 12
}

func h5LobbyLimit(value int) int {
	if value <= 0 {
		return 12
	}
	if value > 50 {
		return 50
	}
	return value
}

func h5IDOrderExpr(column string, ids []uint64) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(id, 10))
	}
	return fmt.Sprintf("FIELD(%s,%s)", column, strings.Join(parts, ","))
}
