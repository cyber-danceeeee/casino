package systemadmin

import (
	"fmt"
	"strings"
)

type brandSetupItem struct {
	Key    string `json:"key"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

func (s *Service) BrandOverview(actor Actor) (map[string]any, error) {
	if actor.Platform {
		return nil, NewError("TENANT_SCOPE_REQUIRED", "brand overview requires tenant scope")
	}
	row, err := s.brandSiteRow(actor)
	if err != nil {
		return nil, err
	}
	var tenant TenantRow
	if err := s.db.Table("game_saas_tenants t").
		Select(tenantSelect()).
		Where("t.id=? AND t.deleted_at IS NULL", row.TenantID).
		Take(&tenant).Error; err != nil {
		return nil, rowNotFound("tenant")
	}
	var gamesEnabled int64
	_ = s.db.Table("game_games").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL AND platform_enabled=1 AND provider_enabled=1", row.TenantID, row.ID).
		Count(&gamesEnabled).Error
	var providersEnabled int64
	_ = s.db.Table("game_game_providers").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL AND status=1", row.TenantID, row.ID).
		Count(&providersEnabled).Error
	var gamesConfigured int64
	_ = s.db.Table("game_games").
		Where("tenant_id=? AND site_id=? AND deleted_at IS NULL AND COALESCE(platform_enabled_draft,platform_enabled)=1", row.TenantID, row.ID).
		Count(&gamesConfigured).Error
	gameLimit, providerLimit := -1, -1
	if s.db.Migrator().HasTable("game_saas_tenant_packages") {
		var limits struct {
			Games     int `gorm:"column:game_limit"`
			Providers int `gorm:"column:provider_limit"`
		}
		if err := s.db.Table("game_saas_tenant_packages").Select("game_limit,provider_limit").
			Where("package_code=? AND deleted_at IS NULL", tenant.PackageCode).Take(&limits).Error; err == nil {
			gameLimit, providerLimit = limits.Games, limits.Providers
		}
	}
	meta := siteConfigMeta(row)
	draft := siteDraftConfig(row)
	maintenance := parseJSONMap(firstDraft(row.MaintenanceDraft, row.MaintenanceConfig))
	modules := []string{}
	if pending, _, pendingErr := pendingReleaseModulesIfAvailable(s, row); pendingErr == nil {
		modules = pending
	}
	setup := s.brandSetupChecklist(row, draft, gamesEnabled, providersEnabled)
	completed := 0
	for _, item := range setup {
		if item.Status == "COMPLETED" {
			completed++
		}
	}
	updatedBy := ""
	_ = s.db.Table("game_admin_operation_logs").Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", row.TenantID, row.ID).Order("created_at DESC").Limit(1).Pluck("operator_username", &updatedBy).Error
	changes := make([]map[string]any, 0, len(modules))
	for _, module := range modules {
		change := map[string]any{"module": module, "summary": module + " draft updated", "updatedBy": updatedBy, "updatedAt": row.ConfigDraftUpdatedAt}
		if module == "BRAND" {
			change["before"] = businessConfig(sitePublishedConfig(row))
			change["after"] = businessConfig(draft)
		}
		changes = append(changes, change)
	}
	return map[string]any{
		"tenantID":           row.TenantID,
		"siteID":             row.ID,
		"tenantCode":         tenant.TenantCode,
		"tenantName":         tenant.TenantName,
		"brandName":          firstNonEmpty(stringFromMap(draft, "siteName"), row.SiteName, tenant.TenantName),
		"countryCode":        stringFromMap(draft, "countryCode"),
		"currencyCode":       stringFromMap(draft, "currencyCode"),
		"defaultLanguage":    stringFromMap(draft, "defaultLanguage"),
		"timezone":           stringFromMap(draft, "timezone"),
		"currentSite":        row.SiteName,
		"siteCode":           row.SiteCode,
		"primaryDomain":      row.PrimaryDomain,
		"adminDomain":        row.AdminDomain,
		"h5Domain":           row.H5Domain,
		"maintenance":        maintenance,
		"maintenanceEnabled": truthyValue(maintenance["enabled"]) || strings.EqualFold(row.Status, "MAINTENANCE"),
		"gamesEnabled":       gamesEnabled,
		"gamesConfigured":    gamesConfigured,
		"providersEnabled":   providersEnabled,
		"packageCode":        tenant.PackageCode,
		"gameLimit":          gameLimit,
		"providerLimit":      providerLimit,
		"publishStatus":      meta["status"],
		"publishedVersion":   meta["configVersion"],
		"draftVersion":       draftVersion(meta),
		"publishedAt":        row.ConfigPublishedAt,
		"draftUpdatedAt":     row.ConfigDraftUpdatedAt,
		"hasUnpublished":     strings.EqualFold(stringFromAny(meta["status"]), "DRAFT"),
		"unpublishedChanges": len(modules),
		"unpublishedModules": modules,
		"pendingChanges":     changes,
		"lastUpdatedBy":      updatedBy,
		"isInitialized":      row.IsInitialized,
		"setupChecklist":     setup,
		"setupCompleted":     completed,
		"setupTotal":         len(setup),
		"setupPercent":       setupPercent(completed, len(setup)),
		"publishedConfig":    sitePublishedConfig(row),
		"draftConfig":        draft,
		"configMeta":         meta,
	}, nil
}

func pendingReleaseModulesIfAvailable(s *Service, row SiteRow) ([]string, bool, error) {
	if !s.db.Migrator().HasTable("game_config_releases") {
		if strings.EqualFold(row.ConfigStatus, "DRAFT") {
			return []string{"BRAND"}, true, nil
		}
		return []string{}, false, nil
	}
	checks := []string{"game_lobby_sections", "game_games", "game_site_banners", "game_site_announcements", "game_tenant_popup_campaigns", "game_h5_navigation_items"}
	for _, table := range checks {
		if !s.db.Migrator().HasTable(table) {
			if strings.EqualFold(row.ConfigStatus, "DRAFT") {
				return []string{"BRAND"}, true, nil
			}
			return []string{}, false, nil
		}
	}
	return pendingReleaseModules(s.db, row)
}

func (s *Service) brandSetupChecklist(row SiteRow, draft map[string]any, gamesEnabled, providersEnabled int64) []brandSetupItem {
	brand := configMap(draft["brandConfig"])
	theme := configMap(draft["themeConfig"])
	country := firstNonEmpty(stringFromMap(draft, "countryCode"), row.CountryCode)
	currency := firstNonEmpty(stringFromMap(draft, "currencyCode"), row.CurrencyCode)
	language := firstNonEmpty(stringFromMap(draft, "defaultLanguage"), row.DefaultLanguage)

	var domainCount, paymentCount, lobbyCount, bannerCount, supportCount int64
	if s.db.Migrator().HasTable("game_saas_domains") {
		_ = s.db.Table("game_saas_domains").Where("tenant_id=? AND site_id=? AND enabled=1 AND status='ACTIVE' AND deleted_at IS NULL", row.TenantID, row.ID).Count(&domainCount).Error
	}
	if s.db.Migrator().HasTable("game_payment_channels") {
		_ = s.db.Table("game_payment_channels").Where("tenant_id IN (0, ?) AND site_id IN (0, ?) AND status=1 AND country_code IN ('', ?) AND currency_code IN ('', ?) AND deleted_at IS NULL", row.TenantID, row.ID, country, currency).Count(&paymentCount).Error
	}
	if s.db.Migrator().HasTable("game_lobby_sections") {
		_ = s.db.Table("game_lobby_sections").Where("tenant_id=? AND site_id=? AND enabled=1 AND deleted_at IS NULL", row.TenantID, row.ID).Count(&lobbyCount).Error
	}
	if s.db.Migrator().HasTable("game_site_banners") {
		_ = s.db.Table("game_site_banners").Where("tenant_id=? AND site_id=? AND enabled=1 AND deleted_at IS NULL", row.TenantID, row.ID).Count(&bannerCount).Error
	}
	if s.db.Migrator().HasTable("game_social_links") {
		_ = s.db.Table("game_social_links").Where("tenant_id=? AND site_id=? AND enabled=1 AND deleted_at IS NULL", row.TenantID, row.ID).Count(&supportCount).Error
	}
	if firstNonEmpty(stringFromMap(brand, "customerService"), stringFromMap(brand, "supportUrl")) != "" {
		supportCount++
	}

	status := func(ok bool) string {
		if ok {
			return "COMPLETED"
		}
		return "INCOMPLETE"
	}
	return []brandSetupItem{
		{Key: "brandInformation", Status: status(firstNonEmpty(stringFromMap(brand, "name"), stringFromMap(draft, "siteName"), row.SiteName) != ""), Path: "/system/brand"},
		{Key: "logo", Status: status(stringFromMap(brand, "logo") != ""), Path: "/system/brand"},
		{Key: "theme", Status: status(stringFromMap(theme, "primary") != ""), Path: "/system/brand"},
		{Key: "country", Status: status(country != ""), Path: "/system/brand"},
		{Key: "currency", Status: status(currency != ""), Path: "/system/brand"},
		{Key: "language", Status: status(language != ""), Path: "/system/brand"},
		{Key: "domain", Status: status(domainCount > 0 || firstNonEmpty(row.PrimaryDomain, row.H5Domain) != ""), Path: "/system/domains"},
		{Key: "paymentChannel", Status: status(paymentCount > 0), Path: "/funds/payment-channels"},
		{Key: "provider", Status: status(providersEnabled > 0), Path: "/games/game-list"},
		{Key: "games", Status: status(gamesEnabled > 0), Path: "/games/game-list"},
		{Key: "lobby", Status: status(lobbyCount > 0), Path: "/system/lobby"},
		{Key: "banner", Status: status(bannerCount > 0), Path: "/operations/banners"},
		{Key: "customerSupport", Status: status(supportCount > 0), Path: "/operations/social-links"},
		{Key: "preview", Status: status(row.ConfigDraftUpdatedAt != nil || row.ConfigPublishedAt != nil), Path: "/system/brand"},
		{Key: "firstPublish", Status: status(row.ConfigPublishedAt != nil), Path: "/system/brand"},
	}
}

func setupPercent(completed, total int) int {
	if total == 0 {
		return 0
	}
	return completed * 100 / total
}

func configMap(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case string:
		return parseJSONMap(typed)
	default:
		return map[string]any{}
	}
}

func businessConfig(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, item := range typed {
			lower := strings.ToLower(strings.ReplaceAll(key, "_", ""))
			if lower == "id" || strings.HasSuffix(lower, "id") || strings.Contains(lower, "uuid") {
				continue
			}
			result[key] = businessConfig(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, businessConfig(item))
		}
		return result
	default:
		return value
	}
}

func (s *Service) brandSiteRow(actor Actor) (SiteRow, error) {
	if actor.SiteID > 0 {
		return s.siteRowForActor(s.db, actor.SiteID, actor, false)
	}
	var row SiteRow
	query := s.db.Table("game_saas_sites s").
		Select(siteSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id AND t.deleted_at IS NULL").
		Where("s.tenant_id=? AND s.deleted_at IS NULL", actor.TenantID).
		Order("s.enabled DESC,s.id ASC")
	if err := query.Take(&row).Error; err != nil {
		return row, rowNotFound("site")
	}
	return row, nil
}

func draftVersion(meta map[string]any) uint64 {
	version := uint64(0)
	switch value := meta["configVersion"].(type) {
	case uint64:
		version = value
	case int:
		version = uint64(value)
	case int64:
		version = uint64(value)
	}
	if version == 0 {
		version = 1
	}
	if strings.EqualFold(stringFromAny(meta["status"]), "DRAFT") {
		return version + 1
	}
	return version
}

func stringFromMap(values map[string]any, key string) string {
	return stringFromAny(values[key])
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(fmt.Sprint(value), "\x00", ""))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truthyValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case int:
		return typed != 0
	case int64:
		return typed != 0
	case float64:
		return typed != 0
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "1" || normalized == "true" || normalized == "enabled" || normalized == "yes"
	default:
		return false
	}
}
