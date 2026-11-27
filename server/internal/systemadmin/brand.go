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
		return nil, NewError("SITE_REQUIRED", "brand overview requires H5 content scope")
	}
	row, err := s.brandSiteRow(actor)
	if err != nil {
		return nil, err
	}
	var gamesEnabled int64
	_ = s.db.Table("game_games").
		Where("site_id = ? AND deleted_at IS NULL AND platform_enabled = 1 AND provider_enabled = 1", row.ID).
		Count(&gamesEnabled).Error
	var providersEnabled int64
	_ = s.db.Table("game_game_providers").
		Where("site_id = ? AND deleted_at IS NULL AND status = 1", row.ID).
		Count(&providersEnabled).Error
	meta := siteConfigMeta(row)
	configuration := siteConfiguration(row)
	maintenance := parseJSONMap(row.MaintenanceConfig)
	setup := s.brandSetupChecklist(row, configuration, gamesEnabled, providersEnabled)
	completed := 0
	for _, item := range setup {
		if item.Status == "COMPLETED" {
			completed++
		}
	}
	updatedBy := ""
	_ = s.db.Table("game_admin_operation_logs").Where("site_id = ? AND deleted_at IS NULL", row.ID).Order("created_at DESC").Limit(1).Pluck("operator_username", &updatedBy).Error
	return map[string]any{
		"siteID":             row.ID,
		"brandName":          firstNonEmpty(stringFromMap(configuration, "siteName"), row.SiteName),
		"countryCode":        stringFromMap(configuration, "countryCode"),
		"currencyCode":       stringFromMap(configuration, "currencyCode"),
		"defaultLanguage":    stringFromMap(configuration, "defaultLanguage"),
		"timezone":           stringFromMap(configuration, "timezone"),
		"currentSite":        row.SiteName,
		"siteCode":           row.SiteCode,
		"primaryDomain":      row.PrimaryDomain,
		"adminDomain":        row.AdminDomain,
		"h5Domain":           row.H5Domain,
		"maintenance":        maintenance,
		"maintenanceEnabled": truthyValue(maintenance["enabled"]) || strings.EqualFold(row.Status, "MAINTENANCE"),
		"gamesEnabled":       gamesEnabled,
		"gamesConfigured":    gamesEnabled,
		"providersEnabled":   providersEnabled,
		"lastUpdatedBy":      updatedBy,
		"isInitialized":      row.IsInitialized,
		"setupChecklist":     setup,
		"setupCompleted":     completed,
		"setupTotal":         len(setup),
		"setupPercent":       setupPercent(completed, len(setup)),
		"configuration":      configuration, "configMeta": meta,
	}, nil
}

func (s *Service) brandSetupChecklist(row SiteRow, configuration map[string]any, gamesEnabled int64, providersEnabled int64) []brandSetupItem {
	brand := configMap(configuration["brandConfig"])
	country := firstNonEmpty(stringFromMap(configuration, "countryCode"), row.CountryCode)
	currency := firstNonEmpty(stringFromMap(configuration, "currencyCode"), row.CurrencyCode)
	language := firstNonEmpty(stringFromMap(configuration, "defaultLanguage"), row.DefaultLanguage)

	var domainCount, paymentCount, lobbyCount, bannerCount, supportCount int64
	if s.db.Migrator().HasTable("game_payment_channels") {
		_ = s.db.Table("game_payment_channels").Where("site_id IN (0, ?) AND status = 1 AND country_code IN ('', ?) AND currency_code IN ('', ?) AND deleted_at IS NULL", row.ID, country, currency).Count(&paymentCount).Error
	}
	if s.db.Migrator().HasTable("game_lobby_sections") {
		_ = s.db.Table("game_lobby_sections").Where("site_id = ? AND enabled = 1 AND deleted_at IS NULL", row.ID).Count(&lobbyCount).Error
	}
	if s.db.Migrator().HasTable("game_site_banners") {
		_ = s.db.Table("game_site_banners").Where("site_id = ? AND enabled = 1 AND deleted_at IS NULL", row.ID).Count(&bannerCount).Error
	}
	if s.db.Migrator().HasTable("game_social_links") {
		_ = s.db.Table("game_social_links").Where("site_id = ? AND enabled = 1 AND deleted_at IS NULL", row.ID).Count(&supportCount).Error
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
		{Key: "brandInformation", Status: status(firstNonEmpty(stringFromMap(brand, "name"), stringFromMap(configuration, "siteName"), row.SiteName) != ""), Path: "/system/brand"},
		{Key: "country", Status: status(country != ""), Path: "/system/brand"},
		{Key: "currency", Status: status(currency != ""), Path: "/system/brand"},
		{Key: "language", Status: status(language != ""), Path: "/system/brand"},
		{Key: "domain", Status: status(domainCount > 0 || firstNonEmpty(row.PrimaryDomain, row.H5Domain) != ""), Path: "/system/brand"},
		{Key: "paymentChannel", Status: status(paymentCount > 0), Path: "/funds/deposit-channels"},
		{Key: "provider", Status: status(providersEnabled > 0), Path: "/games/list"},
		{Key: "games", Status: status(gamesEnabled > 0), Path: "/games/list"},
		{Key: "lobby", Status: status(lobbyCount > 0), Path: "/games/lobby-sections"},
		{Key: "banner", Status: status(bannerCount > 0), Path: "/operations/banners"},
		{Key: "customerSupport", Status: status(supportCount > 0), Path: "/operations/social-links"},
	}
}

func setupPercent(completed int, total int) int {
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
	query := s.db.Table("game_site_config s").
		Select(siteSelect()).
		Where("s.deleted_at IS NULL").
		Order("s.enabled DESC,s.id ASC")
	if err := query.Take(&row).Error; err != nil {
		return row, rowNotFound("site")
	}
	return row, nil
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
