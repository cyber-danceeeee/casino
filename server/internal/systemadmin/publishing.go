package systemadmin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/eventbus"
	"casino/internal/operationadmin"
	"casino/internal/storage"
	"casino/internal/tenancy"
)

func sitePublishedConfig(row SiteRow) map[string]any {
	return map[string]any{
		"siteName":          row.SiteName,
		"countryCode":       row.CountryCode,
		"currencyCode":      row.CurrencyCode,
		"defaultLanguage":   row.DefaultLanguage,
		"timezone":          row.Timezone,
		"brandConfig":       parseJSONMap(row.BrandConfig),
		"themeConfig":       parseJSONMap(row.ThemeConfig),
		"featureConfig":     parseJSONMap(row.FeatureConfig),
		"maintenanceConfig": parseJSONMap(row.MaintenanceConfig),
		"seoConfig":         parseJSONMap(row.SEOConfig),
	}
}

func siteDraftConfig(row SiteRow) map[string]any {
	return map[string]any{
		"siteName":          firstDraft(row.SiteNameDraft, row.SiteName),
		"countryCode":       firstDraft(row.CountryCodeDraft, row.CountryCode),
		"currencyCode":      firstDraft(row.CurrencyCodeDraft, row.CurrencyCode),
		"defaultLanguage":   firstDraft(row.LanguageDraft, row.DefaultLanguage),
		"timezone":          firstDraft(row.TimezoneDraft, row.Timezone),
		"brandConfig":       parseJSONMap(firstDraft(row.BrandConfigDraft, row.BrandConfig)),
		"themeConfig":       parseJSONMap(firstDraft(row.ThemeConfigDraft, row.ThemeConfig)),
		"featureConfig":     parseJSONMap(firstDraft(row.FeatureConfigDraft, row.FeatureConfig)),
		"maintenanceConfig": parseJSONMap(firstDraft(row.MaintenanceDraft, row.MaintenanceConfig)),
		"seoConfig":         parseJSONMap(firstDraft(row.SEOConfigDraft, row.SEOConfig)),
	}
}

func siteConfigMeta(row SiteRow) map[string]any {
	version := row.ConfigVersion
	if version == 0 {
		version = 1
	}
	status := strings.ToUpper(strings.TrimSpace(row.ConfigStatus))
	if status == "" {
		status = "PUBLISHED"
	}
	return map[string]any{
		"status":         status,
		"configVersion":  version,
		"publishedAt":    row.ConfigPublishedAt,
		"draftUpdatedAt": row.ConfigDraftUpdatedAt,
		"tenantId":       row.TenantID,
		"siteId":         row.ID,
	}
}

func firstDraft(draft, published string) string {
	if strings.TrimSpace(draft) != "" {
		return draft
	}
	return published
}

func (s *Service) PreviewSiteConfig(id uint64, actor Actor) (map[string]any, error) {
	row, err := s.siteRowForActor(s.db, id, actor, false)
	if err != nil {
		return nil, err
	}
	bannerDrafts, err := previewBannerDrafts(s.db, row.TenantID, row.ID)
	if err != nil {
		return nil, err
	}
	announcementDrafts, err := previewAnnouncementDrafts(s.db, row.TenantID, row.ID)
	if err != nil {
		return nil, err
	}
	popupDrafts, err := previewPopupDrafts(s.db, row.TenantID, row.ID)
	if err != nil {
		return nil, err
	}
	navigationDrafts, err := operationadmin.PreviewH5NavigationDrafts(s.db, row.TenantID, row.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"overview":          row,
		"publishedConfig":   sitePublishedConfig(row),
		"draftConfig":       siteDraftConfig(row),
		"configMeta":        siteConfigMeta(row),
		"bannerDrafts":      bannerDrafts,
		"announcements":     announcementDrafts,
		"popups":            popupDrafts,
		"navigationEntries": navigationDrafts,
		"previewMode":       true,
	}, nil
}

func (s *Service) PublishSiteConfig(id uint64, reason string, actor Actor) (map[string]any, error) {
	return s.PublishSiteConfigWithKey(id, reason, "", actor)
}

func (s *Service) PublishSiteConfigWithKey(id uint64, reason, idempotencyKey string, actor Actor) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, err
	}
	var result map[string]any
	var failedRelease failedReleaseAttempt
	err := s.db.Transaction(func(tx *gorm.DB) error {
		row, err := s.siteRowForActor(tx, id, actor, true)
		if err != nil {
			return err
		}
		key := releaseIdempotencyKey(row.TenantID, row.ID, idempotencyKey)
		if existing, ok, err := existingReleaseByKey(tx, row.TenantID, row.ID, key); err != nil {
			return err
		} else if ok {
			result = releaseResult(existing, true)
			return nil
		}
		modules, pending, err := pendingReleaseModules(tx, row)
		if err != nil {
			return err
		}
		if !pending {
			result = map[string]any{"tenantId": row.TenantID, "siteId": row.ID, "configVersion": row.ConfigVersion, "publishedAt": row.ConfigPublishedAt, "idempotent": true, "modules": []string{}}
			return nil
		}
		nextVersion, err := nextReleaseVersion(tx, row.TenantID, row.ID, row.ConfigVersion)
		if err != nil {
			return err
		}
		nowValue := now()
		releaseID, err := createPublishingRelease(tx, row.TenantID, row.ID, nextVersion, modules, key, 0, actor.ID)
		if err != nil {
			return err
		}
		failedRelease = failedReleaseAttempt{TenantID: row.TenantID, SiteID: row.ID, Version: nextVersion, Key: key, Modules: strings.Join(modules, ","), CreatedBy: actor.ID}
		published := siteDraftConfig(row)
		updates := map[string]any{
			"site_name":                published["siteName"],
			"country_code":             published["countryCode"],
			"currency_code":            published["currencyCode"],
			"default_language":         published["defaultLanguage"],
			"timezone":                 published["timezone"],
			"brand_config":             marshal(published["brandConfig"]),
			"theme_config":             marshal(published["themeConfig"]),
			"feature_config":           marshal(published["featureConfig"]),
			"maintenance_config":       marshal(published["maintenanceConfig"]),
			"seo_config":               marshal(published["seoConfig"]),
			"site_name_draft":          published["siteName"],
			"country_code_draft":       published["countryCode"],
			"currency_code_draft":      published["currencyCode"],
			"default_language_draft":   published["defaultLanguage"],
			"timezone_draft":           published["timezone"],
			"brand_config_draft":       marshal(published["brandConfig"]),
			"theme_config_draft":       marshal(published["themeConfig"]),
			"feature_config_draft":     marshal(published["featureConfig"]),
			"maintenance_config_draft": marshal(published["maintenanceConfig"]),
			"seo_config_draft":         marshal(published["seoConfig"]),
			"config_status":            "PUBLISHED",
			"config_version":           nextVersion,
			"config_published_at":      nowValue,
			"config_published_by":      actor.ID,
			"is_initialized":           1,
			"updated_by":               actor.ID,
			"version":                  gorm.Expr("version+1"),
			"updated_at":               nowValue,
		}
		updateDB := tx.Table("game_saas_sites").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			updateDB = updateDB.Where("tenant_id=?", actor.TenantID)
			if actor.SiteID > 0 {
				updateDB = updateDB.Where("id=?", actor.SiteID)
			}
		}
		if err := updateDB.Updates(updates).Error; err != nil {
			return err
		}
		if err := publishLobbySectionDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion); err != nil {
			return err
		}
		if err := validatePublishedGameLimit(tx, row.TenantID, row.ID); err != nil {
			return err
		}
		gameDraftsPublished, err := publishGameDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion)
		if err != nil {
			return err
		}
		bannerDraftsPublished, err := publishBannerDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion)
		if err != nil {
			return err
		}
		announcementDraftsPublished, err := publishAnnouncementDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion)
		if err != nil {
			return err
		}
		popupDraftsPublished, err := publishPopupDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion)
		if err != nil {
			return err
		}
		navigationDraftsPublished, err := operationadmin.PublishH5NavigationDrafts(tx, row.TenantID, row.ID, actor.ID, nextVersion)
		if err != nil {
			return NewError("H5_NAVIGATION_PUBLISH_INVALID", err.Error())
		}
		if err := clearWhiteLabelConfigCache(tx, row.TenantID, row.ID); err != nil {
			return err
		}
		targetActor := targetAuditActor(actor, row.TenantID, row.ID)
		if err := audit(tx, targetActor, "release:publish", "system.site.config.publish", "site", id, sitePublishedConfig(row), published, reason); err != nil {
			return err
		}
		if _, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope:          tenancy.Scope{TenantID: row.TenantID, SiteID: row.ID, Platform: actor.Platform},
			TenantID:       row.TenantID,
			SiteID:         row.ID,
			EventType:      eventbus.EventWhiteLabelConfigPublished,
			AggregateType:  "white_label_config",
			AggregateID:    fmt.Sprintf("%d", row.ID),
			Source:         "systemadmin",
			IdempotencyKey: fmt.Sprintf("white_label.config_published:%d:%d", row.ID, nextVersion),
			TraceID:        actor.CorrelationID,
			CorrelationID:  actor.CorrelationID,
			Payload: map[string]any{
				"tenantId":      row.TenantID,
				"siteId":        row.ID,
				"configVersion": nextVersion,
				"publishedAt":   nowValue,
			},
		}); err != nil {
			return err
		}
		snapshot, err := captureReleaseSnapshot(tx, row.TenantID, row.ID)
		if err != nil {
			return err
		}
		if err := tx.Table("game_config_releases").Where("id=?", releaseID).Updates(map[string]any{"status": "PUBLISHED", "snapshot": marshal(snapshot), "validation_result": marshal(map[string]any{"valid": true}), "published_by": actor.ID, "published_at": nowValue, "updated_at": nowValue}).Error; err != nil {
			return err
		}
		result = map[string]any{
			"releaseID":                   releaseID,
			"tenantId":                    row.TenantID,
			"siteId":                      row.ID,
			"configVersion":               nextVersion,
			"publishedAt":                 nowValue,
			"gameDraftsPublished":         gameDraftsPublished,
			"bannerDraftsPublished":       bannerDraftsPublished,
			"announcementDraftsPublished": announcementDraftsPublished,
			"popupDraftsPublished":        popupDraftsPublished,
			"navigationDraftsPublished":   navigationDraftsPublished,
			"cacheCleared":                true,
			"eventType":                   eventbus.EventWhiteLabelConfigPublished,
			"modules":                     modules,
			"idempotent":                  false,
		}
		return nil
	})
	if err != nil && failedRelease.Key != "" {
		s.recordFailedRelease(failedRelease, err)
	}
	return result, err
}

func validatePublishedGameLimit(tx *gorm.DB, tenantID, siteID uint64) error {
	if !tx.Migrator().HasTable("game_saas_tenant_packages") {
		return nil
	}
	var limits []struct {
		GameLimit int `gorm:"column:game_limit"`
	}
	err := tx.Table("game_saas_tenants t").
		Select("p.game_limit").
		Joins("JOIN game_saas_tenant_packages p ON p.package_code=t.package_code AND p.deleted_at IS NULL").
		Where("t.id=? AND t.deleted_at IS NULL", tenantID).
		Limit(1).Find(&limits).Error
	if err != nil || len(limits) == 0 || limits[0].GameLimit < 0 {
		return err
	}
	gameLimit := limits[0].GameLimit
	var enabledCount int64
	if err := tx.Table("game_games").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).
		Where("COALESCE(platform_enabled_draft,platform_enabled)=1").
		Count(&enabledCount).Error; err != nil {
		return err
	}
	var publishedEnabled int64
	if err := tx.Table("game_games").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL AND platform_enabled=1", tenantID, siteID).
		Count(&publishedEnabled).Error; err != nil {
		return err
	}
	allowed := int64(gameLimit)
	if publishedEnabled > allowed {
		allowed = publishedEnabled
	}
	if enabledCount > allowed {
		return NewError("TENANT_GAME_LIMIT_EXCEEDED", "tenant package game limit exceeded")
	}
	return nil
}

type siteBannerPublishRow struct {
	ID             uint64 `gorm:"column:id"`
	DraftPayload   string `gorm:"column:draft_payload"`
	DesktopAssetID uint64 `gorm:"column:desktop_asset_id_draft"`
	MobileAssetID  uint64 `gorm:"column:mobile_asset_id_draft"`
}

type siteBannerPublishPayload struct {
	Name            string     `json:"name"`
	Title           string     `json:"title"`
	Subtitle        string     `json:"subtitle"`
	DesktopAssetID  uint64     `json:"desktopAssetID"`
	MobileAssetID   uint64     `json:"mobileAssetID"`
	DesktopImageURL string     `json:"desktopImageUrl"`
	MobileImageURL  string     `json:"mobileImageUrl"`
	ActionType      string     `json:"actionType"`
	ActionTarget    string     `json:"actionTarget"`
	SortOrder       int        `json:"sortOrder"`
	LanguageCode    string     `json:"languageCode"`
	CountryCode     string     `json:"countryCode"`
	CurrencyCode    string     `json:"currencyCode"`
	VisibilityScope string     `json:"visibilityScope"`
	StartAt         *time.Time `json:"startAt"`
	EndAt           *time.Time `json:"endAt"`
	Enabled         bool       `json:"enabled"`
	Deleted         bool       `json:"deleted"`
}

func previewBannerDrafts(db *gorm.DB, tenantID, siteID uint64) ([]map[string]any, error) {
	var rows []siteBannerPublishRow
	if err := db.Table("game_site_banners").Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload,desktop_asset_id_draft,mobile_asset_id_draft").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(row.DraftPayload), &payload); err != nil {
			return nil, NewError("INVALID_BANNER_DRAFT", "invalid banner draft")
		}
		payload["id"] = row.ID
		out = append(out, payload)
	}
	return out, nil
}

func publishBannerDrafts(tx *gorm.DB, tenantID, siteID, actorID, configVersion uint64) (int, error) {
	var rows []siteBannerPublishRow
	if err := tx.Table("game_site_banners").Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload,desktop_asset_id_draft,mobile_asset_id_draft").
		Where("tenant_id=? AND site_id=? AND publish_status=? AND deleted_at IS NULL", tenantID, siteID, "DRAFT").Order("id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	nowValue := now()
	for _, row := range rows {
		var payload siteBannerPublishPayload
		if err := json.Unmarshal([]byte(row.DraftPayload), &payload); err != nil {
			return 0, NewError("INVALID_BANNER_DRAFT", "invalid banner draft")
		}
		if payload.Deleted {
			if err := tx.Table("game_site_banners").Where("id=? AND tenant_id=? AND site_id=?", row.ID, tenantID, siteID).
				Updates(map[string]any{"deleted_at": nowValue, "updated_at": nowValue, "updated_by": actorID}).Error; err != nil {
				return 0, err
			}
			continue
		}
		if err := validatePublishBannerAsset(tx, tenantID, siteID, payload.DesktopAssetID, payload.DesktopImageURL, true); err != nil {
			return 0, err
		}
		if payload.MobileAssetID > 0 {
			if err := validatePublishBannerAsset(tx, tenantID, siteID, payload.MobileAssetID, payload.MobileImageURL, false); err != nil {
				return 0, err
			}
		}
		updates := map[string]any{
			"name": strings.TrimSpace(payload.Name), "title": strings.TrimSpace(payload.Title), "subtitle": strings.TrimSpace(payload.Subtitle),
			"desktop_asset_id": payload.DesktopAssetID, "mobile_asset_id": payload.MobileAssetID,
			"desktop_asset_id_draft": payload.DesktopAssetID, "mobile_asset_id_draft": payload.MobileAssetID,
			"action_type": payload.ActionType, "action_target": strings.TrimSpace(payload.ActionTarget), "sort_order": payload.SortOrder,
			"language_code": payload.LanguageCode, "country_code": payload.CountryCode, "currency_code": payload.CurrencyCode,
			"visibility_scope": payload.VisibilityScope, "start_at": payload.StartAt, "end_at": payload.EndAt, "enabled": boolValue(payload.Enabled),
			"published_payload": row.DraftPayload, "publish_status": "PUBLISHED", "config_version": configVersion,
			"published_at": nowValue, "published_by": actorID, "updated_by": actorID, "updated_at": nowValue,
		}
		if err := tx.Table("game_site_banners").Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", row.ID, tenantID, siteID).Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

func validatePublishBannerAsset(tx *gorm.DB, tenantID, siteID, assetID uint64, imageURL string, required bool) error {
	if assetID == 0 {
		if storage.IsAssetURL(imageURL) {
			return nil
		}
		if required {
			return NewError("BANNER_ASSET_INVALID", "banner desktop asset is required")
		}
		return nil
	}
	var count int64
	err := tx.Table("game_asset_library").Where("id=? AND asset_type='IMAGE' AND status='ACTIVE' AND deleted_at IS NULL", assetID).
		Where("(asset_scope='PLATFORM' AND is_public=1) OR (tenant_id=? AND site_id IN (0,?))", tenantID, siteID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return NewError("BANNER_ASSET_INVALID", "banner asset is missing, inactive, or outside the site scope")
	}
	return nil
}

type gameDraftPublishRow struct {
	ID                      uint64         `gorm:"column:id"`
	ProviderEnabled         bool           `gorm:"column:provider_enabled"`
	PlatformEnabled         bool           `gorm:"column:platform_enabled"`
	PlatformEnabledDraft    sql.NullBool   `gorm:"column:platform_enabled_draft"`
	IsHot                   bool           `gorm:"column:is_hot"`
	IsHotDraft              sql.NullBool   `gorm:"column:is_hot_draft"`
	IsFeatured              bool           `gorm:"column:is_featured"`
	IsFeaturedDraft         sql.NullBool   `gorm:"column:is_featured_draft"`
	IsNew                   bool           `gorm:"column:is_new"`
	IsNewDraft              sql.NullBool   `gorm:"column:is_new_draft"`
	Sort                    int            `gorm:"column:sort"`
	SortDraft               sql.NullInt64  `gorm:"column:sort_draft"`
	Status                  int            `gorm:"column:status"`
	StatusDraft             sql.NullInt64  `gorm:"column:status_draft"`
	MaintenanceMessage      string         `gorm:"column:maintenance_message"`
	MaintenanceMessageDraft sql.NullString `gorm:"column:maintenance_message_draft"`
	CategoryIDsDraft        sql.NullString `gorm:"column:category_ids_draft"`
}

type gameCategoryDraftItem struct {
	CategoryID uint64 `json:"categoryId"`
	Sort       int    `json:"sort"`
}

func publishGameDrafts(tx *gorm.DB, tenantID, siteID, actorID, configVersion uint64) (int, error) {
	var rows []gameDraftPublishRow
	if err := tx.Table("game_games").
		Select(`id,provider_enabled,platform_enabled,platform_enabled_draft,is_hot,is_hot_draft,is_featured,is_featured_draft,
			is_new,is_new_draft,sort,sort_draft,status,status_draft,maintenance_message,maintenance_message_draft,
			CAST(category_ids_draft AS CHAR) category_ids_draft`).
		Where("tenant_id=? AND site_id=? AND config_status=? AND deleted_at IS NULL", tenantID, siteID, "DRAFT").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return 0, err
	}
	nowValue := now()
	for _, row := range rows {
		platformEnabled := row.PlatformEnabled
		if row.PlatformEnabledDraft.Valid {
			platformEnabled = row.PlatformEnabledDraft.Bool
		}
		isHot := row.IsHot
		if row.IsHotDraft.Valid {
			isHot = row.IsHotDraft.Bool
		}
		isFeatured := row.IsFeatured
		if row.IsFeaturedDraft.Valid {
			isFeatured = row.IsFeaturedDraft.Bool
		}
		isNew := row.IsNew
		if row.IsNewDraft.Valid {
			isNew = row.IsNewDraft.Bool
		}
		sortOrder := row.Sort
		if row.SortDraft.Valid {
			sortOrder = int(row.SortDraft.Int64)
		}
		status := row.Status
		if row.StatusDraft.Valid {
			status = int(row.StatusDraft.Int64)
		} else if row.ProviderEnabled && platformEnabled {
			status = 1
		} else {
			status = 3
		}
		maintenanceMessage := row.MaintenanceMessage
		if row.MaintenanceMessageDraft.Valid {
			maintenanceMessage = row.MaintenanceMessageDraft.String
		}
		if row.CategoryIDsDraft.Valid {
			items, err := parseGameCategoryDraft(row.CategoryIDsDraft.String)
			if err != nil {
				return 0, err
			}
			if err := publishGameCategoryDraft(tx, tenantID, siteID, row.ID, items); err != nil {
				return 0, err
			}
		}
		updates := map[string]any{
			"platform_enabled":          boolValue(platformEnabled),
			"platform_enabled_draft":    boolValue(platformEnabled),
			"is_hot":                    boolValue(isHot),
			"is_hot_draft":              boolValue(isHot),
			"is_featured":               boolValue(isFeatured),
			"is_featured_draft":         boolValue(isFeatured),
			"is_new":                    boolValue(isNew),
			"is_new_draft":              boolValue(isNew),
			"sort":                      sortOrder,
			"sort_draft":                sortOrder,
			"status":                    status,
			"status_draft":              status,
			"maintenance_message":       maintenanceMessage,
			"maintenance_message_draft": maintenanceMessage,
			"config_status":             "PUBLISHED",
			"config_version":            configVersion,
			"config_published_at":       nowValue,
			"config_published_by":       actorID,
			"updated_at":                nowValue,
		}
		if err := tx.Table("game_games").
			Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", row.ID, tenantID, siteID).
			Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

func parseGameCategoryDraft(raw string) ([]gameCategoryDraftItem, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var items []gameCategoryDraftItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, NewError("INVALID_GAME_CATEGORY_DRAFT", "invalid game category draft")
	}
	return items, nil
}

func publishGameCategoryDraft(tx *gorm.DB, tenantID, siteID, gameID uint64, items []gameCategoryDraftItem) error {
	nowValue := now()
	if err := tx.Table("game_game_category_relations").
		Where("tenant_id=? AND site_id=? AND game_id=? AND deleted_at IS NULL", tenantID, siteID, gameID).
		Updates(map[string]any{"deleted_at": nowValue, "updated_at": nowValue}).Error; err != nil {
		return err
	}
	seen := map[uint64]bool{}
	for index, item := range items {
		if item.CategoryID == 0 || seen[item.CategoryID] {
			continue
		}
		seen[item.CategoryID] = true
		var count int64
		if err := tx.Table("game_game_categories").
			Where("tenant_id=? AND site_id=? AND id=? AND deleted_at IS NULL", tenantID, siteID, item.CategoryID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return NewError("INVALID_GAME_CATEGORY_DRAFT", "invalid game category draft")
		}
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "game_id": gameID, "category_id": item.CategoryID,
			"sort": index, "created_at": nowValue, "updated_at": nowValue, "deleted_at": nil,
		}
		if err := tx.Table("game_game_category_relations").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "game_id"}, {Name: "category_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"sort", "updated_at", "deleted_at"}),
		}).Create(row).Error; err != nil {
			return err
		}
	}
	return syncPublishedGamePrimaryCategory(tx, tenantID, siteID, gameID)
}

func syncPublishedGamePrimaryCategory(tx *gorm.DB, tenantID, siteID, gameID uint64) error {
	var primary struct {
		ID   uint64 `gorm:"column:id"`
		Code string `gorm:"column:code"`
	}
	result := tx.Table("game_game_category_relations r").
		Select("c.id,c.category_key code").
		Joins("JOIN game_game_categories c ON c.id=r.category_id AND c.deleted_at IS NULL").
		Where("r.tenant_id=? AND r.site_id=? AND r.game_id=? AND r.deleted_at IS NULL", tenantID, siteID, gameID).
		Order("r.sort ASC,r.id ASC").Limit(1).Find(&primary)
	if result.Error != nil {
		return result.Error
	}
	return tx.Table("game_games").
		Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", gameID, tenantID, siteID).
		Updates(map[string]any{"platform_category_id": primary.ID, "platform_category_code": primary.Code, "updated_at": now()}).Error
}

func (s *Service) siteRowForActor(db *gorm.DB, id uint64, actor Actor, forUpdate bool) (SiteRow, error) {
	var row SiteRow
	query := db.Table("game_saas_sites s").
		Select(siteSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id AND t.deleted_at IS NULL").
		Where("s.id=? AND s.deleted_at IS NULL", id)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if !actor.Platform {
		query = query.Where("s.tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			query = query.Where("s.id=?", actor.SiteID)
		}
	}
	if err := query.Take(&row).Error; err != nil {
		return row, rowNotFound("site")
	}
	return row, nil
}

func clearWhiteLabelConfigCache(tx *gorm.DB, tenantID, siteID uint64) error {
	return nil
}

type lobbyDraftRow struct {
	ID           uint64 `gorm:"column:id"`
	DraftPayload string `gorm:"column:draft_payload"`
}

type lobbyDraftPayload struct {
	SectionCode         string         `json:"sectionCode"`
	SectionType         string         `json:"sectionType"`
	TitleZh             string         `json:"titleZh"`
	TitleEn             string         `json:"titleEn"`
	DisplayStyle        string         `json:"displayStyle"`
	SourceType          string         `json:"sourceType"`
	SourceRef           string         `json:"sourceRef"`
	MaxItems            int            `json:"maxItems"`
	SortOrder           int            `json:"sortOrder"`
	Enabled             bool           `json:"enabled"`
	SupportedCountries  string         `json:"supportedCountries"`
	SupportedCurrencies string         `json:"supportedCurrencies"`
	Config              map[string]any `json:"config"`
}

func publishLobbySectionDrafts(tx *gorm.DB, tenantID, siteID, actorID, configVersion uint64) error {
	var rows []lobbyDraftRow
	if err := tx.Table("game_lobby_sections").
		Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload").
		Where("tenant_id=? AND site_id=? AND publish_status=? AND deleted_at IS NULL", tenantID, siteID, "DRAFT").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return err
	}
	nowValue := now()
	for _, row := range rows {
		payload := lobbyDraftPayload{Config: map[string]any{}}
		if strings.TrimSpace(row.DraftPayload) != "" {
			if err := json.Unmarshal([]byte(row.DraftPayload), &payload); err != nil {
				return NewError("INVALID_LOBBY_DRAFT", "invalid lobby section draft")
			}
		}
		if payload.MaxItems <= 0 {
			payload.MaxItems = 12
		}
		configJSON := "{}"
		if payload.Config != nil {
			configJSON = marshal(payload.Config)
		}
		updates := map[string]any{
			"section_code":         normalizeSaaSCode(payload.SectionCode),
			"section_type":         strings.TrimSpace(payload.SectionType),
			"title_zh":             strings.TrimSpace(payload.TitleZh),
			"title_en":             strings.TrimSpace(payload.TitleEn),
			"display_style":        strings.TrimSpace(payload.DisplayStyle),
			"source_type":          strings.TrimSpace(payload.SourceType),
			"source_ref":           strings.TrimSpace(payload.SourceRef),
			"max_items":            payload.MaxItems,
			"sort_order":           payload.SortOrder,
			"enabled":              boolValue(payload.Enabled),
			"supported_countries":  strings.TrimSpace(payload.SupportedCountries),
			"supported_currencies": strings.TrimSpace(payload.SupportedCurrencies),
			"config_json":          configJSON,
			"draft_payload":        row.DraftPayload,
			"publish_status":       "PUBLISHED",
			"config_version":       configVersion,
			"published_at":         nowValue,
			"published_by":         actorID,
			"updated_at":           nowValue,
		}
		if err := tx.Table("game_lobby_sections").
			Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", row.ID, tenantID, siteID).
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}
