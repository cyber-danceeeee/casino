package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/game/connections"
	"casino/internal/model"
	"casino/internal/operationadmin"
)

func (h *gameBannersHandler) Home(c *gin.Context) {
	db := database.GetDB()
	now := time.Now()
	scope := h5ContentScope(c)

	banners := h5HomeSiteBanners(c, db, scope, now)

	var categories []model.GameGameCategories
	_ = h5FindWithPlatformFallback(db.Model(&model.GameGameCategories{}), "game_game_categories", scope, &categories, func(query *gorm.DB) *gorm.DB {
		return query.
			Where("game_game_categories.category_key = ? OR game_game_categories.status = ?", "hot", 1).
			Order("CASE WHEN game_game_categories.category_key = 'hot' THEN 0 ELSE 1 END ASC, sort ASC, id ASC")
	})
	hasHot := false
	for i := range categories {
		hasHot = hasHot || categories[i].CategoryKey == "hot"
		categories[i].IconURL = h5AssetURL(categories[i].IconURL)
	}
	if !hasHot {
		categories = append([]model.GameGameCategories{{
			CategoryKey: "hot",
			Label:       "HOT",
			ShortLabel:  "HOT",
			Sort:        0,
			Status:      0,
		}}, categories...)
	}

	var featureCards []model.GameHomeFeatureCards
	_ = h5FindWithPlatformFallback(db.Model(&model.GameHomeFeatureCards{}), "", scope, &featureCards, func(query *gorm.DB) *gorm.DB {
		return query.Where("status = ?", 1).Order("sort ASC, id ASC")
	})
	for i := range featureCards {
		featureCards[i].ImageURL = h5AssetURL(featureCards[i].ImageURL)
	}

	var jackpot model.GameRebateJackpots
	_ = db.Where("status = ?", 1).
		Order("id DESC").
		First(&jackpot).Error

	response.Success(c, gin.H{
		"banners":       banners,
		"categories":    categories,
		"featureCards":  featureCards,
		"rebateJackpot": h5JackpotPayload(jackpot),
		"socialLinks":   h5SocialLinks(c, db),
		"providers":     h5BootstrapProviders(c, db, scope),
	})
}

func h5HomeSiteBanners(c *gin.Context, db *gorm.DB, scope content.Scope, now time.Time) []gin.H {
	if scope.SiteID == 0 {
		return []gin.H{}
	}
	_, _, loggedIn := h5AuthenticatedUser(db, c)
	language := strings.TrimSpace(c.Query("language"))
	if language == "" {
		language = strings.TrimSpace(c.GetHeader("Accept-Language"))
		if index := strings.Index(language, ","); index >= 0 {
			language = language[:index]
		}
	}
	country := strings.TrimSpace(c.Query("country"))
	if language == "" || country == "" {
		site := h5BootstrapSite(db, scope.SiteID)
		if language == "" {
			language = site.DefaultLanguage
		}
		if country == "" {
			country = site.CountryCode
		}
	}
	rows, err := operationadmin.New(db).PublicSiteBanners(scope, language, country, loggedIn, now)
	if err != nil || len(rows) == 0 {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		imageURL := row.MobileImageURL
		if imageURL == "" {
			imageURL = row.DesktopImageURL
		}
		out = append(out, gin.H{
			"id":              row.ID,
			"bannerCode":      row.Name,
			"title":           row.Title,
			"subtitle":        row.Subtitle,
			"imageURL":        imageURL,
			"desktopImageURL": row.DesktopImageURL,
			"mobileImageURL":  row.MobileImageURL,
			"actionType":      row.ActionType,
			"actionTarget":    row.ActionTarget,
			"sort":            row.SortOrder,
			"configVersion":   row.ConfigVersion,
		})
	}
	return out
}

func h5DefaultGameCategories() []model.GameGameCategories {
	return []model.GameGameCategories{
		{CategoryKey: "hot", Label: "HOT", ShortLabel: "HOT", Sort: 1, Status: 1},
		{CategoryKey: "slots", Label: "SLOTS", ShortLabel: "SLOTS", Sort: 10, Status: 1},
		{CategoryKey: "live-casino", Label: "LIVE", ShortLabel: "LIVE", Sort: 20, Status: 1},
		{CategoryKey: "table-games", Label: "TABLE", ShortLabel: "TABLE", Sort: 30, Status: 1},
		{CategoryKey: "crash", Label: "CRASH", ShortLabel: "CRASH", Sort: 40, Status: 1},
		{CategoryKey: "fishing", Label: "FISHING", ShortLabel: "FISH", Sort: 50, Status: 1},
		{CategoryKey: "sports", Label: "SPORTS", ShortLabel: "SPORTS", Sort: 60, Status: 1},
	}
}

func h5SupportedGameProviderCodes(db *gorm.DB) []string {
	return connections.New(db).PublicCodes(context.Background())
}

func h5GameSections(ctx context.Context, categories []model.GameGameCategories, games []model.GameGames, relations []model.GameGameCategoryRelations, providers map[uint64]model.GameGameProviders) []gin.H {
	categoryGameSet := map[uint64]map[uint64]bool{}
	for _, relation := range relations {
		if categoryGameSet[relation.CategoryID] == nil {
			categoryGameSet[relation.CategoryID] = map[uint64]bool{}
		}
		categoryGameSet[relation.CategoryID][relation.GameID] = true
	}

	sections := make([]gin.H, 0, len(categories))
	for _, category := range categories {
		key := category.CategoryKey
		gameType := h5GameType(key)
		items := make([]h5GameDTO, 0)
		relationGames := categoryGameSet[category.ID]
		for _, game := range games {
			if relationGames != nil {
				if relationGames[game.ID] {
					items = append(items, h5GameDTOFromModel(ctx, game, providers[game.ProviderID]))
				}
				continue
			}
			if key == "hot" {
				if game.HotScore > 0 {
					items = append(items, h5GameDTOFromModel(ctx, game, providers[game.ProviderID]))
				}
				continue
			}
			if game.PlatformCategoryCode == key || h5GameMatchesCategory(game.GameType, gameType) {
				items = append(items, h5GameDTOFromModel(ctx, game, providers[game.ProviderID]))
			}
		}
		if key == "hot" && len(items) == 0 {
			items = append(items, h5GameDTOs(ctx, games, providers)...)
		}
		sections = append(sections, gin.H{
			"key":        key,
			"title":      category.Label,
			"shortLabel": category.ShortLabel,
			"iconUrl":    h5AssetURL(category.IconURL),
			"games":      items,
		})
	}
	return sections
}

func h5GameMatchesCategory(gameType string, categoryType string) bool {
	switch categoryType {
	case "chance":
		return gameType == "chance" || gameType == "live" || gameType == "table"
	case "live":
		return gameType == "live" || gameType == "chance"
	case "table":
		return gameType == "table"
	default:
		return gameType == categoryType
	}
}

func h5JackpotPayload(jackpot model.GameRebateJackpots) gin.H {
	return gin.H{
		"periodType": h5FirstNonEmpty(jackpot.PeriodType, "weekly"),
		"amount":     h5DecimalString(jackpot.JackpotAmount),
		"digitText":  h5FirstNonEmpty(jackpot.DigitText, "0000000000"),
		"periodEnd":  jackpot.PeriodEnd,
	}
}
