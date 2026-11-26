package gameadmin

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/game/domain"
	"casino/internal/game/market"
	"casino/internal/model"
)

func (s *Service) ListGames(q ListQuery) (PageResult, error) {
	var rows []gameRecord
	db := s.db.Table("game_games g").Select(gameListSelect(q.Scope)).
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").Where("g.deleted_at IS NULL")
	db = content.ApplySite(db, "g", q.Scope)
	db = content.ApplySite(db, "p", q.Scope)
	db = applyGameListFilters(db, q)
	result, err := paginate(db.Order("g.sort ASC,g.id DESC"), q, &rows)
	for i := range rows {
		signGameAsset(&rows[i])
	}
	result.List = rows
	return result, err
}

func gameListSelect(scope content.Scope) string {
	return `g.id,g.game_code,g.provider_id,p.provider_code,g.provider_game_code,g.name,g.name_i18n,g.slug,g.game_type,g.rtp,g.volatility,g.category,
		g.cover_url,g.local_thumbnail_url,g.background_url,g.launch_image_url,g.maintenance_message,g.maintenance_message_draft,g.supports_demo,g.supports_free_spins,
		g.provider_enabled,g.platform_enabled,g.platform_enabled_draft,g.is_hot,g.is_hot_draft,g.is_featured,g.is_featured_draft,g.is_new,g.is_new_draft,
		g.status,g.status_draft,g.sort,g.sort_draft,g.category_ids_draft,g.config_status,g.config_version,g.config_published_at,g.config_published_by,g.draft_updated_at,g.draft_updated_by,
		g.supported_countries,g.supported_currencies,g.supported_devices,g.last_synced_at,g.created_at,g.updated_at,
		(SELECT GROUP_CONCAT(c.category_key ORDER BY r.sort,r.id SEPARATOR ',') FROM game_game_category_relations r JOIN game_game_categories c ON c.id=r.category_id AND c.deleted_at IS NULL WHERE r.game_id=g.id AND r.deleted_at IS NULL AND ` + categoryRelationScopePredicate("r", scope) + ` AND ` + categoryRelationScopePredicate("c", scope) + `) platform_category_codes,
		(SELECT GROUP_CONCAT(t.code ORDER BY t.sort_order,t.id SEPARATOR ',') FROM game_game_tag_relations tr JOIN game_game_tags t ON t.id=tr.tag_id AND t.deleted_at IS NULL WHERE tr.game_id=g.id AND tr.deleted_at IS NULL) tag_codes`
}

func applyGameListFilters(db *gorm.DB, q ListQuery) *gorm.DB {
	if id := strings.TrimSpace(q.Values["gameID"]); id != "" {
		db = db.Where("g.game_code=? OR g.id=?", id, parseID(id))
	}
	db = contains(db, "g.name", q.Values["gameName"])
	db = equal(db, "p.provider_code", q.Values["providerCode"])
	db = contains(db, "g.provider_game_code", q.Values["providerGameID"])
	db = equal(db, "g.category", q.Values["providerCategory"])
	db = equal(db, "g.game_type", q.Values["gameType"])
	db = boolFilter(db, "g.platform_enabled", q.Values["enabled"])
	db = boolFilter(db, "g.supports_demo", q.Values["supportsDemo"])
	db = boolFilter(db, "g.supports_free_spins", q.Values["supportsFreeSpins"])
	db = boolFilter(db, "g.is_hot", q.Values["isHot"])
	db = boolFilter(db, "g.is_featured", q.Values["isFeatured"])
	db = boolFilter(db, "g.is_new", q.Values["isNew"])
	db = timeRange(db, "g.last_synced_at", q.Values["lastSyncedFrom"], q.Values["lastSyncedTo"])
	if code := strings.TrimSpace(q.Values["platformCategory"]); code != "" {
		db = db.Where("EXISTS (SELECT 1 FROM game_game_category_relations r JOIN game_game_categories c ON c.id=r.category_id WHERE r.game_id=g.id AND r.deleted_at IS NULL AND c.deleted_at IS NULL AND "+categoryRelationScopePredicate("r", q.Scope)+" AND "+categoryRelationScopePredicate("c", q.Scope)+" AND c.category_key=?)", code)
	}
	if country := strings.TrimSpace(q.Values["country"]); country != "" {
		db = db.Where("g.supported_countries='' OR FIND_IN_SET(?,g.supported_countries)", country)
	}
	if currency := strings.TrimSpace(q.Values["currency"]); currency != "" {
		db = db.Where("g.supported_currencies='' OR FIND_IN_SET(?,g.supported_currencies)", currency)
	}
	return db
}

func (s *Service) ListImportableGames(q ListQuery) (PageResult, error) {
	if !q.Scope.SiteScoped() || q.Scope.SiteID == 0 {
		return PageResult{}, newError("SITE_REQUIRED", "H5 scope is required")
	}
	var rows []gameRecord
	db := s.db.Table("game_games g").Select(gameListSelect(q.Scope)).
		Joins("JOIN game_game_providers AS p ON p.id = g.provider_id AND p.site_id = 0 AND p.deleted_at IS NULL").
		Joins("JOIN game_game_providers AS tp ON tp.provider_code = p.provider_code AND tp.site_id = ? AND tp.status = 1 AND tp.deleted_at IS NULL", q.Scope.SiteID).
		Where("g.site_id = 0 AND g.deleted_at IS NULL").
		Where("NOT EXISTS(SELECT 1 FROM game_games AS tg JOIN game_game_providers AS tgp ON tgp.id = tg.provider_id AND tgp.deleted_at IS NULL WHERE tg.site_id = ? AND tg.deleted_at IS NULL AND tgp.provider_code = p.provider_code AND tg.provider_game_code = g.provider_game_code)", q.Scope.SiteID)
	db = applyGameListFilters(db, q)
	result, err := paginate(db.Order("p.sort ASC,g.sort ASC,g.id DESC"), q, &rows)
	for i := range rows {
		signGameAsset(&rows[i])
	}
	result.List = rows
	return result, err
}

func (s *Service) ListImportableProviderGameTypes(providerCode string, scope content.Scope) ([]providerGameTypeOption, error) {
	providerCode = strings.TrimSpace(providerCode)
	if providerCode == "" {
		return nil, newError("INVALID_REQUEST", "provider code is required")
	}
	if !scope.SiteScoped() || scope.SiteID == 0 {
		return nil, newError("SITE_REQUIRED", "H5 scope is required")
	}
	var siteProvider struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := s.db.Table("game_game_providers").
		Select("id").
		Where("site_id = ? AND provider_code = ? AND status = 1 AND deleted_at IS NULL", scope.SiteID, providerCode).
		Take(&siteProvider).Error; err != nil {
		return nil, notFound(err, "PROVIDER_NOT_AUTHORIZED")
	}
	var platformProvider struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := s.db.Table("game_game_providers").
		Select("id").
		Where("site_id = 0 AND provider_code = ? AND deleted_at IS NULL", providerCode).
		Take(&platformProvider).Error; err != nil {
		return nil, notFound(err, "PROVIDER_NOT_FOUND")
	}
	var rows []providerGameTypeOption
	err := s.db.Table("game_games g").
		Select(`? provider_code,
			CASE WHEN TRIM(g.category)<>'' THEN 'providerCategory' ELSE 'gameType' END source,
			CASE WHEN TRIM(g.category)<>'' THEN g.category ELSE g.game_type END code,
			COUNT(g.id) game_count`, providerCode).
		Where("g.site_id = 0 AND g.provider_id = ? AND g.deleted_at IS NULL", platformProvider.ID).
		Where("NOT EXISTS(SELECT 1 FROM game_games AS tg WHERE tg.site_id = ? AND tg.provider_id = ? AND tg.provider_game_code = g.provider_game_code AND tg.deleted_at IS NULL)", scope.SiteID, siteProvider.ID).
		Group("source,code").
		Having("code<>''").
		Order("source ASC,code ASC").
		Find(&rows).Error
	return rows, err
}

func (s *Service) ImportGames(in GameImportInput, actor Actor) (map[string]any, error) {
	if !actor.Scope.SiteScoped() || actor.Scope.SiteID == 0 {
		return nil, newError("SITE_REQUIRED", "H5 scope is required")
	}
	gameIDs := uniqueIDs(in.IDs)
	if len(gameIDs) == 0 {
		return nil, newError("INVALID_REQUEST", "game ids are required")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "import_platform_games"
	}
	result := map[string]any{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var sources []model.GameGames
		if err := tx.Model(&model.GameGames{}).
			Where("site_id = 0 AND id IN (?) AND deleted_at IS NULL", gameIDs).
			Order("id ASC").
			Find(&sources).Error; err != nil {
			return err
		}
		if len(sources) != len(gameIDs) {
			return newError("GAME_NOT_FOUND", "one or more games do not exist")
		}
		providerIDs := make([]uint64, 0, len(sources))
		for _, src := range sources {
			providerIDs = append(providerIDs, src.ProviderID)
		}
		var platformProviders []model.GameGameProviders
		if err := tx.Model(&model.GameGameProviders{}).
			Where("site_id = 0 AND id IN (?) AND deleted_at IS NULL", uniqueIDs(providerIDs)).
			Find(&platformProviders).Error; err != nil {
			return err
		}
		providerCodes := map[uint64]string{}
		authorizedCodes := make([]string, 0, len(platformProviders))
		for _, provider := range platformProviders {
			code := strings.TrimSpace(provider.ProviderCode)
			if code == "" {
				continue
			}
			providerCodes[provider.ID] = code
			authorizedCodes = append(authorizedCodes, code)
		}
		if len(providerCodes) != len(uniqueIDs(providerIDs)) {
			return newError("PROVIDER_NOT_FOUND", "record not found")
		}
		var siteProviders []model.GameGameProviders
		if err := tx.Model(&model.GameGameProviders{}).
			Where("site_id = ? AND provider_code IN (?) AND status = 1 AND deleted_at IS NULL", actor.Scope.SiteID, uniqueStrings(authorizedCodes)).
			Find(&siteProviders).Error; err != nil {
			return err
		}
		siteProviderByCode := map[string]model.GameGameProviders{}
		for _, provider := range siteProviders {
			siteProviderByCode[strings.TrimSpace(provider.ProviderCode)] = provider
		}
		existingKeys := map[string]bool{}
		var existing []struct {
			ProviderCode     string `gorm:"column:provider_code"`
			ProviderGameCode string `gorm:"column:provider_game_code"`
		}
		if err := tx.Table("game_games tg").
			Select("p.provider_code,tg.provider_game_code").
			Joins("JOIN game_game_providers p ON p.id=tg.provider_id AND p.deleted_at IS NULL").
			Where("tg.site_id = ? AND tg.deleted_at IS NULL", actor.Scope.SiteID).
			Find(&existing).Error; err != nil {
			return err
		}
		for _, row := range existing {
			existingKeys[gameImportKey(row.ProviderCode, row.ProviderGameCode)] = true
		}
		nowValue := time.Now()
		rows := make([]map[string]any, 0, len(sources))
		importedIDs := make([]uint64, 0, len(sources))
		skippedIDs := make([]uint64, 0)
		for _, src := range sources {
			providerCode := providerCodes[src.ProviderID]
			siteProvider, ok := siteProviderByCode[providerCode]
			if !ok {
				return newError("PROVIDER_NOT_AUTHORIZED", "provider is not authorized for site")
			}
			if existingKeys[gameImportKey(providerCode, src.ProviderGameCode)] {
				skippedIDs = append(skippedIDs, src.ID)
				continue
			}
			rows = append(rows, map[string]any{

				"site_id":                   actor.Scope.SiteID,
				"game_code":                 src.GameCode,
				"provider_id":               siteProvider.ID,
				"provider_game_code":        src.ProviderGameCode,
				"name":                      src.Name,
				"name_i18n":                 jsonObjectOrDefault(src.NameI18n),
				"slug":                      src.Slug,
				"category":                  src.Category,
				"game_type":                 src.GameType,
				"rtp":                       src.RTP,
				"volatility":                src.Volatility,
				"cover_url":                 src.CoverURL,
				"local_thumbnail_url":       src.LocalThumbnailURL,
				"background_url":            src.BackgroundURL,
				"launch_image_url":          src.LaunchImageURL,
				"maintenance_message":       src.MaintenanceMessage,
				"maintenance_message_draft": src.MaintenanceMessage,
				"supports_demo":             src.SupportsDemo,
				"supports_free_spins":       src.SupportsFreeSpins,
				"provider_enabled":          src.ProviderEnabled && siteProvider.Status == 1,
				"platform_enabled":          false,
				"platform_enabled_draft":    false,
				"is_hot":                    false,
				"is_hot_draft":              false,
				"is_featured":               false,
				"is_featured_draft":         false,
				"is_new":                    false,
				"is_new_draft":              false,
				"supported_countries":       src.SupportedCountries,
				"supported_currencies":      src.SupportedCurrencies,
				"supported_devices":         src.SupportedDevices,
				"raw_metadata":              jsonObjectOrDefault(src.RawMetadata),
				"metadata_hash":             src.MetadataHash,
				"last_synced_at":            src.LastSyncedAt,
				"badge":                     src.Badge,
				"player_count":              0,
				"hot_score":                 src.HotScore,
				"launch_url":                src.LaunchURL,
				"status":                    3,
				"status_draft":              3,
				"sort":                      src.Sort,
				"sort_draft":                src.Sort,
				"platform_category_id":      0,
				"platform_category_code":    "",
				"category_ids_draft":        "[]",
				"config_status":             "DRAFT",
				"config_version":            0,
				"config_published_by":       0,
				"draft_updated_at":          nowValue,
				"draft_updated_by":          actor.ID,
				"created_at":                nowValue,
				"updated_at":                nowValue,
			})
			importedIDs = append(importedIDs, src.ID)
			existingKeys[gameImportKey(providerCode, src.ProviderGameCode)] = true
		}
		importedCount := int64(0)
		if len(rows) > 0 {
			db := insertSiteGameImportRows(tx, rows)
			if db.Error != nil {
				return duplicateCode(db.Error, "GAME_ALREADY_IMPORTED", "one or more games are already imported")
			}
			importedCount = db.RowsAffected
		}
		after := map[string]any{"sourceGameIDs": importedIDs, "skippedSourceGameIDs": skippedIDs, "importedCount": importedCount}
		if err := audit(tx, actor, "game:update", "import_games", "game", 0, nil, after, reason); err != nil {
			return err
		}
		result = map[string]any{"importedCount": importedCount, "skippedCount": len(skippedIDs), "sourceGameIDs": importedIDs}
		return nil
	})
	return result, err
}

func gameImportKey(providerCode string, providerGameCode string) string {
	return strings.ToLower(strings.TrimSpace(providerCode)) + ":" + strings.TrimSpace(providerGameCode)
}

func insertSiteGameImportRows(tx *gorm.DB, rows []map[string]any) *gorm.DB {
	db := tx.Table("game_games")
	if tx.Dialector != nil && tx.Dialector.Name() == "mysql" {
		return db.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(rows)
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(rows)
}

func jsonObjectOrDefault(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return value
}

func (s *Service) ListProviderGameTypes(providerCode string, scope content.Scope) ([]providerGameTypeOption, error) {
	providerCode = strings.TrimSpace(providerCode)
	if providerCode == "" {
		return nil, newError("INVALID_REQUEST", "provider code is required")
	}
	var exists int64
	if err := content.ApplySite(s.db.Table("game_game_providers"), "", scope).Where("provider_code=? AND deleted_at IS NULL", providerCode).Count(&exists).Error; err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, newError("PROVIDER_NOT_FOUND", "record not found")
	}
	var rows []providerGameTypeOption
	db := s.db.Table("game_games g").
		Select(`p.provider_code,
			CASE WHEN TRIM(g.category)<>'' THEN 'providerCategory' ELSE 'gameType' END source,
			CASE WHEN TRIM(g.category)<>'' THEN g.category ELSE g.game_type END code,
			COUNT(g.id) game_count`).
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").
		Where("p.provider_code=? AND g.deleted_at IS NULL", providerCode).
		Group("source,code,p.provider_code").
		Having("code<>''").
		Order("source ASC,code ASC")
	db = content.ApplySite(db, "g", scope)
	db = content.ApplySite(db, "p", scope)
	err := db.Find(&rows).Error
	return rows, err
}

func (s *Service) ApplyProviderCategoryMapping(in ProviderCategoryMappingInput, actor Actor) (map[string]any, error) {
	providerCode := strings.TrimSpace(in.ProviderCode)
	source := strings.TrimSpace(in.Source)
	code := strings.TrimSpace(in.Code)
	categoryIDs := uniqueIDs(in.CategoryIDs)
	if providerCode == "" || code == "" || (source != "providerCategory" && source != "gameType") {
		return nil, newError("INVALID_REQUEST", "invalid provider category mapping")
	}
	if len(categoryIDs) == 0 {
		return nil, newError("INVALID_RELATION", "one or more relation targets do not exist")
	}
	result := map[string]any{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var provider struct {
			ID uint64 `gorm:"column:id"`
		}
		if err := content.ApplySite(tx.Table("game_game_providers"), "", actor.Scope).Select("id").Where("provider_code=? AND deleted_at IS NULL", providerCode).Take(&provider).Error; err != nil {
			return notFound(err, "PROVIDER_NOT_FOUND")
		}
		var categories []categoryRecord
		categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := categoryDB.Where("id IN ? AND deleted_at IS NULL", categoryIDs).Find(&categories).Error; err != nil {
			return err
		}
		if len(categories) != len(categoryIDs) {
			return newError("INVALID_RELATION", "one or more relation targets do not exist")
		}
		gameQuery := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("provider_id=? AND deleted_at IS NULL", provider.ID)
		if source == "gameType" {
			gameQuery = gameQuery.Where("game_type=?", code)
		} else {
			gameQuery = gameQuery.Where("category=?", code)
		}
		var gameIDs []uint64
		if err := gameQuery.Pluck("id", &gameIDs).Error; err != nil {
			return err
		}
		if len(gameIDs) > 0 {
			items := categoryDraftFromIDs(categoryIDs)
			for _, gameID := range gameIDs {
				var err error
				if actor.Scope.Platform {
					err = s.savePublishedGameCategories(tx, actor, gameID, items)
				} else {
					err = s.saveGameCategoryDraft(tx, actor, gameID, items)
				}
				if err != nil {
					return err
				}
			}
		}
		before := map[string]any{"providerCode": providerCode, "source": source, "code": code}
		after := map[string]any{"categoryIds": categoryIDs, "matchedCount": len(gameIDs), "updatedCount": len(gameIDs)}
		action := "provider_category_mapping_draft"
		if actor.Scope.Platform {
			action = "provider_category_mapping"
		}
		if err := audit(tx, actor, "game:category:assign", action, "game_category_mapping", 0, before, after, in.Reason); err != nil {
			return err
		}
		result = map[string]any{
			"providerCode": providerCode,
			"source":       source,
			"code":         code,
			"categoryIds":  categoryIDs,
			"matchedCount": len(gameIDs),
			"updatedCount": len(gameIDs),
		}
		return nil
	})
	return result, err
}

func (s *Service) GameDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row gameRecord
	db := s.db.Table("game_games g").Select("g.*,p.provider_code").Joins("JOIN game_game_providers p ON p.id=g.provider_id").Where("g.id=? AND g.deleted_at IS NULL", id)
	db = content.ApplySite(db, "g", scope)
	db = content.ApplySite(db, "p", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "GAME_NOT_FOUND")
	}
	signGameAsset(&row)
	var categories []categoryRecord
	categoryDB := s.db.Table("game_game_categories c").Select("c.*").Joins("JOIN game_game_category_relations r ON r.category_id=c.id AND r.deleted_at IS NULL").Where("r.game_id=? AND c.deleted_at IS NULL", id)
	categoryDB = applyCategoryResourceScope(categoryDB, "c", scope)
	categoryDB = applyCategoryResourceScope(categoryDB, "r", scope)
	_ = categoryDB.Order("r.sort,r.id").Find(&categories).Error
	for i := range categories {
		categories[i].Enabled = categories[i].Status == 1
		signCategoryAsset(&categories[i])
	}
	var tags []tagRecord
	tagDB := s.db.Table("game_game_tags t").Select("t.*").Joins("JOIN game_game_tag_relations r ON r.tag_id=t.id AND r.deleted_at IS NULL").Where("r.game_id=? AND t.deleted_at IS NULL", id)
	tagDB = content.ApplySite(tagDB, "t", scope)
	tagDB = content.ApplySite(tagDB, "r", scope)
	_ = tagDB.Order("t.sort_order,t.id").Find(&tags).Error
	for i := range tags {
		signTagAsset(&tags[i])
	}
	var collections []collectionRecord
	collectionDB := s.db.Table("game_game_collections c").Select("c.*").Joins("JOIN game_game_collection_relations r ON r.collection_id=c.id AND r.deleted_at IS NULL").Where("r.game_id=? AND c.deleted_at IS NULL", id)
	collectionDB = content.ApplySite(collectionDB, "c", scope)
	collectionDB = content.ApplySite(collectionDB, "r", scope)
	_ = collectionDB.Order("r.sort_order,r.id").Find(&collections).Error
	for i := range collections {
		signCollectionAsset(&collections[i])
	}
	providerData := map[string]any{"providerCode": row.ProviderCode, "providerGameId": row.ProviderGameID, "providerCategoryCode": row.ProviderCategoryCode, "providerEnabled": row.ProviderEnabled, "rawMetadata": redactJSON(row.RawMetadata)}
	draftConfiguration := map[string]any{
		"status":             strings.ToUpper(firstNonEmpty(row.ConfigStatus, "PUBLISHED")),
		"configVersion":      row.ConfigVersion,
		"publishedAt":        row.ConfigPublishedAt,
		"publishedBy":        row.ConfigPublishedBy,
		"draftUpdatedAt":     row.DraftUpdatedAt,
		"draftUpdatedBy":     row.DraftUpdatedBy,
		"platformEnabled":    draftBool(row.PlatformEnabledDraft, row.PlatformEnabled),
		"isHot":              draftBool(row.IsHotDraft, row.IsHot),
		"isFeatured":         draftBool(row.IsFeaturedDraft, row.IsFeatured),
		"isNew":              draftBool(row.IsNewDraft, row.IsNew),
		"sortOrder":          draftInt(row.SortDraft, row.SortOrder),
		"statusValue":        draftInt(row.StatusDraft, row.Status),
		"maintenanceMessage": draftString(row.MaintenanceMessageDraft, row.MaintenanceMessage),
		"categoryDraft":      parseGameCategoryDraftItems(row.CategoryIDsDraft),
	}
	return map[string]any{
		"overview":                row,
		"providerData":            providerData,
		"platformConfiguration":   map[string]any{"platformEnabled": row.PlatformEnabled, "isHot": row.IsHot, "isFeatured": row.IsFeatured, "isNew": row.IsNew, "sortOrder": row.SortOrder, "localThumbnailUrl": row.LocalThumbnailURL},
		"draftConfiguration":      draftConfiguration,
		"categoriesAndTags":       map[string]any{"categories": categories, "tags": tags, "collections": collections},
		"countryAndCurrencyRules": map[string]any{"countries": row.SupportedCountries, "currencies": row.SupportedCurrencies, "devices": row.SupportedDevices},
		"launchConfiguration":     map[string]any{"supportsDemo": row.SupportsDemo, "supportsFreeSpins": row.SupportsFreeSpins, "effectiveEnabled": row.ProviderEnabled && row.PlatformEnabled},
		"operationLogs":           s.operationLogs("game", id),
	}, nil
}

func (s *Service) UpdateGame(id uint64, in GameInput, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row gameRecord
		if err := content.ApplySite(tx.Table("game_games").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return notFound(err, "GAME_NOT_FOUND")
		}
		nameI18n := row.NameI18n
		if in.NameI18n != nil {
			encoded, err := json.Marshal(in.NameI18n)
			if err != nil {
				return newError("INVALID_REQUEST", "invalid name i18n")
			}
			nameI18n = string(encoded)
		}
		updates := map[string]any{
			"name":                 strings.TrimSpace(in.Name),
			"name_i18n":            nameI18n,
			"local_thumbnail_url":  adminAssetRef(in.LocalThumbnailURL),
			"background_url":       adminAssetRef(in.BackgroundURL),
			"launch_image_url":     adminAssetRef(in.LaunchImageURL),
			"rtp":                  in.RTP,
			"volatility":           strings.TrimSpace(in.Volatility),
			"supported_countries":  strings.TrimSpace(in.SupportedCountries),
			"supported_currencies": strings.TrimSpace(in.SupportedCurrencies),
			"supported_devices":    strings.TrimSpace(in.SupportedDevices),
		}
		if in.SupportsDemo != nil {
			updates["supports_demo"] = *in.SupportsDemo
		}
		if in.SupportsFreeSpins != nil {
			updates["supports_free_spins"] = *in.SupportsFreeSpins
		}
		if actor.Scope.Platform {
			updates["platform_enabled"] = in.PlatformEnabled
			updates["platform_enabled_draft"] = in.PlatformEnabled
			updates["is_hot"] = in.IsHot
			updates["is_hot_draft"] = in.IsHot
			updates["is_featured"] = in.IsFeatured
			updates["is_featured_draft"] = in.IsFeatured
			updates["is_new"] = in.IsNew
			updates["is_new_draft"] = in.IsNew
			updates["sort"] = in.SortOrder
			updates["sort_draft"] = in.SortOrder
			updates["status"] = effectiveStatus(row.ProviderEnabled, in.PlatformEnabled)
			updates["status_draft"] = effectiveStatus(row.ProviderEnabled, in.PlatformEnabled)
			updates["maintenance_message"] = strings.TrimSpace(in.MaintenanceMessage)
			updates["maintenance_message_draft"] = strings.TrimSpace(in.MaintenanceMessage)
			updates = gamePublishedUpdate(actor, updates)
			if err := tx.Table("game_games").Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
				return err
			}
			return audit(tx, actor, "game:update", "update", "game", id, row, updates, in.Reason)
		}
		updates = gameDraftUpdate(actor, updates)
		updates["platform_enabled_draft"] = in.PlatformEnabled
		updates["is_hot_draft"] = in.IsHot
		updates["is_featured_draft"] = in.IsFeatured
		updates["is_new_draft"] = in.IsNew
		updates["sort_draft"] = in.SortOrder
		updates["status_draft"] = effectiveStatus(row.ProviderEnabled, in.PlatformEnabled)
		updates["maintenance_message_draft"] = strings.TrimSpace(in.MaintenanceMessage)
		if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "game:update", "update_draft", "game", id, row, updates, in.Reason)
	})
}

func (s *Service) SetGameStatus(id uint64, command ReasonCommand, actor Actor) error {
	var row gameRecord
	if err := content.ApplySite(s.db.Table("game_games"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return notFound(err, "GAME_NOT_FOUND")
	}
	return s.UpdateGame(id, GameInput{
		Name:                row.Name,
		LocalThumbnailURL:   row.LocalThumbnailURL,
		BackgroundURL:       row.BackgroundURL,
		LaunchImageURL:      row.LaunchImageURL,
		PlatformEnabled:     command.Enabled,
		IsHot:               row.IsHot,
		IsFeatured:          row.IsFeatured,
		IsNew:               row.IsNew,
		RTP:                 row.RTP,
		Volatility:          row.Volatility,
		MaintenanceMessage:  row.MaintenanceMessage,
		SortOrder:           row.SortOrder,
		SupportedCountries:  row.SupportedCountries,
		SupportedCurrencies: row.SupportedCurrencies,
		SupportedDevices:    row.SupportedDevices,
		Reason:              command.Reason,
	}, actor)
}

func (s *Service) BatchUpdateGames(in GameBatchInput, actor Actor) (map[string]any, error) {
	gameIDs := uniqueIDs(in.IDs)
	for _, item := range in.SortOrders {
		if item.ID > 0 {
			gameIDs = append(gameIDs, item.ID)
		}
	}
	gameIDs = uniqueIDs(gameIDs)
	if len(gameIDs) == 0 {
		return nil, newError("INVALID_REQUEST", "game ids are required")
	}
	updates := map[string]any{}
	auditUpdates := map[string]any{}
	if in.PlatformEnabled != nil {
		if actor.Scope.Platform {
			updates["platform_enabled"] = *in.PlatformEnabled
			updates["platform_enabled_draft"] = *in.PlatformEnabled
			updates["status"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND ? THEN 1 ELSE 3 END", *in.PlatformEnabled)
			updates["status_draft"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND ? THEN 1 ELSE 3 END", *in.PlatformEnabled)
		} else {
			updates["platform_enabled_draft"] = *in.PlatformEnabled
			updates["status_draft"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND ? THEN 1 ELSE 3 END", *in.PlatformEnabled)
		}
		auditUpdates["platformEnabled"] = *in.PlatformEnabled
	}
	if in.IsHot != nil {
		if actor.Scope.Platform {
			updates["is_hot"] = *in.IsHot
		}
		updates["is_hot_draft"] = *in.IsHot
		auditUpdates["isHot"] = *in.IsHot
	}
	if in.IsFeatured != nil {
		if actor.Scope.Platform {
			updates["is_featured"] = *in.IsFeatured
		}
		updates["is_featured_draft"] = *in.IsFeatured
		auditUpdates["isFeatured"] = *in.IsFeatured
	}
	if in.IsNew != nil {
		if actor.Scope.Platform {
			updates["is_new"] = *in.IsNew
		}
		updates["is_new_draft"] = *in.IsNew
		auditUpdates["isNew"] = *in.IsNew
	}
	if in.Maintenance != nil {
		if *in.Maintenance {
			updates["status_draft"] = 2
			if actor.Scope.Platform {
				updates["status"] = 2
			}
		} else if in.PlatformEnabled != nil {
			updates["status_draft"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND ? THEN 1 ELSE 3 END", *in.PlatformEnabled)
			if actor.Scope.Platform {
				updates["status"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND ? THEN 1 ELSE 3 END", *in.PlatformEnabled)
			}
		} else {
			updates["status_draft"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND COALESCE(platform_enabled_draft,platform_enabled) = 1 THEN 1 ELSE 3 END")
			if actor.Scope.Platform {
				updates["status"] = gorm.Expr("CASE WHEN provider_enabled = 1 AND platform_enabled = 1 THEN 1 ELSE 3 END")
			}
		}
		auditUpdates["maintenance"] = *in.Maintenance
	}
	if in.MaintenanceMessage != nil {
		updates["maintenance_message_draft"] = strings.TrimSpace(*in.MaintenanceMessage)
		if actor.Scope.Platform {
			updates["maintenance_message"] = strings.TrimSpace(*in.MaintenanceMessage)
		}
		auditUpdates["maintenanceMessage"] = strings.TrimSpace(*in.MaintenanceMessage)
	}
	sortUpdates := map[uint64]int{}
	for _, item := range in.SortOrders {
		if item.ID > 0 {
			sortUpdates[item.ID] = item.SortOrder
		}
	}
	if len(sortUpdates) > 0 {
		auditUpdates["sortOrders"] = sortUpdates
	}
	if in.CategoryIDs != nil {
		auditUpdates["categoryIds"] = uniqueIDs(*in.CategoryIDs)
	}
	if in.AddCategoryIDs != nil {
		auditUpdates["addCategoryIds"] = uniqueIDs(*in.AddCategoryIDs)
	}
	if in.CategoryIDs != nil && in.AddCategoryIDs != nil {
		return nil, newError("INVALID_REQUEST", "categoryIds and addCategoryIds cannot be used together")
	}
	if in.TagIDs != nil {
		auditUpdates["tagIds"] = uniqueIDs(*in.TagIDs)
	}
	if len(auditUpdates) == 0 {
		return nil, newError("INVALID_REQUEST", "no game fields selected")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "batch_update_games"
	}
	result := map[string]any{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before []gameRecord
		if err := content.ApplySite(tx.Table("game_games").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor.Scope).Where("id IN ? AND deleted_at IS NULL", gameIDs).Find(&before).Error; err != nil {
			return err
		}
		if len(before) != len(gameIDs) {
			return newError("GAME_NOT_FOUND", "one or more games do not exist")
		}
		var affected int64
		if len(updates) > 0 {
			values := updates
			if actor.Scope.Platform {
				values["updated_at"] = time.Now()
			} else {
				values = gameDraftUpdate(actor, updates)
			}
			db := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id IN ? AND deleted_at IS NULL", gameIDs).Updates(values)
			if db.Error != nil {
				return db.Error
			}
			affected = db.RowsAffected
		}
		for gameID, sortOrder := range sortUpdates {
			values := map[string]any{"sort_draft": sortOrder}
			if actor.Scope.Platform {
				values = map[string]any{"sort": sortOrder, "sort_draft": sortOrder, "updated_at": time.Now()}
			} else {
				values = gameDraftUpdate(actor, values)
			}
			db := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", gameID).Updates(values)
			if db.Error != nil {
				return db.Error
			}
			affected += db.RowsAffected
		}
		if in.CategoryIDs != nil {
			items := categoryDraftFromIDs(*in.CategoryIDs)
			for _, gameID := range gameIDs {
				if actor.Scope.Platform {
					normalized, err := s.normalizeGameCategoryDraft(tx, actor, items)
					if err != nil {
						return err
					}
					ids := make([]uint64, 0, len(normalized))
					for _, item := range normalized {
						ids = append(ids, item.CategoryID)
					}
					if err := s.batchReplaceRelations(tx, actor, []uint64{gameID}, ids, "category"); err != nil {
						return err
					}
					encoded, err := json.Marshal(normalized)
					if err != nil {
						return newError("INVALID_REQUEST", "invalid category configuration")
					}
					if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).
						Where("id=? AND deleted_at IS NULL", gameID).
						Updates(map[string]any{"category_ids_draft": string(encoded), "updated_at": time.Now()}).Error; err != nil {
						return err
					}
				} else {
					if err := s.saveGameCategoryDraft(tx, actor, gameID, items); err != nil {
						return err
					}
				}
			}
			affected += int64(len(gameIDs))
		}
		if in.AddCategoryIDs != nil {
			additions := categoryDraftFromIDs(*in.AddCategoryIDs)
			if len(additions) == 0 {
				return newError("INVALID_REQUEST", "category ids are required")
			}
			for _, gameID := range gameIDs {
				var items []gameCategoryDraftItem
				var err error
				if actor.Scope.Platform {
					items, err = s.loadPublishedGameCategories(tx, actor, gameID)
				} else {
					items, err = s.loadGameCategoryDraft(tx, actor, gameID)
				}
				if err != nil {
					return err
				}
				seen := make(map[uint64]bool, len(items))
				for _, item := range items {
					seen[item.CategoryID] = true
				}
				for _, addition := range additions {
					if seen[addition.CategoryID] {
						continue
					}
					addition.Sort = len(items)
					items = append(items, addition)
					seen[addition.CategoryID] = true
				}
				if actor.Scope.Platform {
					err = s.savePublishedGameCategories(tx, actor, gameID, items)
				} else {
					err = s.saveGameCategoryDraft(tx, actor, gameID, items)
				}
				if err != nil {
					return err
				}
			}
			affected += int64(len(gameIDs))
		}
		if in.TagIDs != nil {
			if err := s.batchReplaceRelations(tx, actor, gameIDs, uniqueIDs(*in.TagIDs), "tag"); err != nil {
				return err
			}
		}
		if actor.Scope.Platform {
			if err := tx.Table("game_games").Where("id IN ? AND deleted_at IS NULL", gameIDs).
				Updates(gamePublishedUpdate(actor, map[string]any{})).Error; err != nil {
				return err
			}
		}
		after := map[string]any{"ids": gameIDs, "updates": auditUpdates, "updatedCount": affected}
		action := "batch_update_draft"
		if actor.Scope.Platform {
			action = "batch_update"
		}
		if err := audit(tx, actor, "game:update", action, "game", 0, before, after, reason); err != nil {
			return err
		}
		result = map[string]any{"updatedCount": affected}
		return nil
	})
	return result, err
}

func (s *Service) batchReplaceRelations(tx *gorm.DB, actor Actor, gameIDs []uint64, targetIDs []uint64, kind string) error {
	config := map[string]struct {
		relationTable string
		foreign       string
		targetTable   string
		orderColumn   string
	}{
		"category": {"game_game_category_relations", "category_id", "game_game_categories", "sort"},
		"tag":      {"game_game_tag_relations", "tag_id", "game_game_tags", ""},
	}[kind]
	if config.relationTable == "" {
		return newError("INVALID_REQUEST", "invalid relation type")
	}
	if len(targetIDs) > 0 {
		var count int64
		targetDB := content.ApplySite(tx.Table(config.targetTable), "", actor.Scope)
		if kind == "category" {
			var err error
			targetDB, err = applyCategoryWriteScope(tx.Table(config.targetTable), "", actor.Scope)
			if err != nil {
				return err
			}
		}
		if err := targetDB.Where("id IN ? AND deleted_at IS NULL", targetIDs).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(targetIDs)) {
			return newError("INVALID_RELATION", "one or more relation targets do not exist")
		}
	}
	now := time.Now()
	relationDB := content.ApplySite(tx.Table(config.relationTable), "", actor.Scope)
	if kind == "category" {
		relationDB = applyCategoryResourceScope(tx.Table(config.relationTable), "", actor.Scope)
	}
	if err := relationDB.Where("game_id IN ? AND deleted_at IS NULL", gameIDs).Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
		return err
	}
	if len(targetIDs) > 0 {
		rows := make([]map[string]any, 0, len(gameIDs)*len(targetIDs))
		for _, gameID := range gameIDs {
			for index, targetID := range targetIDs {
				row := map[string]any{

					"site_id":      actor.Scope.SiteID,
					"game_id":      gameID,
					config.foreign: targetID,
					"created_at":   now,
					"updated_at":   now,
					"deleted_at":   nil,
				}
				if config.orderColumn != "" {
					row[config.orderColumn] = index
				}
				rows = append(rows, row)
			}
		}
		assignments := []string{"updated_at", "deleted_at"}
		if config.orderColumn != "" {
			assignments = append(assignments, config.orderColumn)
		}
		if err := tx.Table(config.relationTable).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "game_id"}, {Name: config.foreign}},
			DoUpdates: clause.AssignmentColumns(assignments),
		}).CreateInBatches(rows, 500).Error; err != nil {
			return err
		}
	}
	if kind != "category" {
		return nil
	}
	primaryID := uint64(0)
	primaryCode := ""
	if len(targetIDs) > 0 {
		var category categoryRecord
		categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := categoryDB.Where("id=? AND deleted_at IS NULL", targetIDs[0]).Take(&category).Error; err != nil {
			return err
		}
		primaryID = category.ID
		primaryCode = category.Code
	}
	return content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id IN ? AND deleted_at IS NULL", gameIDs).Updates(map[string]any{
		"platform_category_id":   primaryID,
		"platform_category_code": primaryCode,
		"updated_at":             now,
	}).Error
}

func effectiveStatus(providerEnabled bool, platformEnabled bool) int {
	if providerEnabled && platformEnabled {
		return 1
	}
	return 3
}

func (s *Service) SetGameCategories(id uint64, in RelationInput, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError("GAME_NOT_FOUND", "record not found")
		}
		before, err := s.loadGameCategoryDraft(tx, actor, id)
		if actor.Scope.Platform {
			before, err = s.loadPublishedGameCategories(tx, actor, id)
		}
		if err != nil {
			return err
		}
		after := categoryDraftFromIDs(in.IDs)
		if actor.Scope.Platform {
			err = s.savePublishedGameCategories(tx, actor, id, after)
		} else {
			err = s.saveGameCategoryDraft(tx, actor, id, after)
		}
		if err != nil {
			return err
		}
		action := "assign_category_draft"
		if actor.Scope.Platform {
			action = "assign_category"
		}
		return audit(tx, actor, "game:category:assign", action, "game", id, before, after, in.Reason)
	})
}
func (s *Service) SetGameTags(id uint64, in RelationInput, actor Actor) error {
	return s.replaceRelations(id, in, actor, "tag")
}
func (s *Service) SetGameCollections(id uint64, in RelationInput, actor Actor) error {
	return s.replaceRelations(id, in, actor, "collection")
}

func (s *Service) replaceRelations(gameID uint64, in RelationInput, actor Actor, kind string) error {
	config := map[string]struct{ table, foreign, target, permission string }{
		"category":   {"game_game_category_relations", "category_id", "game_category", "game:category:assign"},
		"tag":        {"game_game_tag_relations", "tag_id", "game_tag", "game:tag:assign"},
		"collection": {"game_game_collection_relations", "collection_id", "game_collection", "game:collection:assign"},
	}[kind]
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", gameID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError("GAME_NOT_FOUND", "record not found")
		}
		var before []uint64
		_ = content.ApplySite(tx.Table(config.table), "", actor.Scope).Where("game_id=? AND deleted_at IS NULL", gameID).Pluck(config.foreign, &before).Error
		if err := content.ApplySite(tx.Table(config.table), "", actor.Scope).Where("game_id=? AND deleted_at IS NULL", gameID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		for index, targetID := range uniqueIDs(in.IDs) {
			row := map[string]any{"site_id": actor.Scope.SiteID, "game_id": gameID, config.foreign: targetID, "created_at": time.Now(), "updated_at": time.Now(), "deleted_at": nil}
			if kind == "category" {
				row["sort"] = index
			}
			if kind == "collection" {
				row["sort_order"] = index
			}
			if err := tx.Table(config.table).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "game_id"}, {Name: config.foreign}}, DoUpdates: clause.Assignments(row)}).Create(row).Error; err != nil {
				return newError("INVALID_RELATION", "one or more relation targets do not exist")
			}
		}
		return audit(tx, actor, config.permission, "assign_"+kind, "game", gameID, before, in.IDs, in.Reason)
	})
}

func uniqueIDs(values []uint64) []uint64 {
	seen := map[uint64]bool{}
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value > 0 && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func (s *Service) LaunchDemo(ctx context.Context, id uint64, scope content.Scope) (map[string]any, error) {
	var game gameRecord
	db := s.db.Table("game_games g").Select("g.*,p.provider_code").Joins("JOIN game_game_providers p ON p.id=g.provider_id").Where("g.id=? AND g.deleted_at IS NULL", id)
	db = content.ApplySite(db, "g", scope)
	db = content.ApplySite(db, "p", scope)
	if err := db.Take(&game).Error; err != nil {
		return nil, notFound(err, "GAME_NOT_FOUND")
	}
	if !game.SupportsDemo {
		return nil, newError("DEMO_NOT_SUPPORTED", "demo mode is not supported")
	}
	deployment, err := market.Load(s.db.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var rules struct {
		SupportedCountries  string
		SupportedCurrencies string
	}
	if err := s.db.Table("game_game_providers").Select("supported_countries,supported_currencies").Where("id=?", game.ProviderID).Take(&rules).Error; err != nil {
		return nil, err
	}
	if !deployment.ProviderAllowed(rules.SupportedCountries, rules.SupportedCurrencies) || !deployment.GameAllowed(game.SupportedCountries, game.SupportedCurrencies) {
		return nil, newError("MARKET_UNAVAILABLE", "game is not available in the deployment market")
	}
	provider, err := s.runtime.ProviderRegistry.Get(game.ProviderCode)
	if err != nil {
		return nil, newError("PROVIDER_NOT_REGISTERED", "provider adapter is not registered")
	}
	result, err := provider.LaunchDemo(ctx, domain.LaunchDemoRequest{ProviderCode: game.ProviderCode, ProviderGameID: game.ProviderGameID, Currency: deployment.Currency, Language: "en", Country: deployment.Country, Device: "desktop", SessionID: generatedCode("ADMIN-DEMO")})
	if err != nil {
		return nil, newError("DEMO_LAUNCH_FAILED", "demo launch failed")
	}
	return map[string]any{"launchUrl": result.LaunchURL, "expiresAt": result.ExpiresAt}, nil
}
