package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type h5BootstrapSiteRow struct {
	ID uint64 `gorm:"column:id"`

	SiteCode          string     `gorm:"column:site_code"`
	SiteName          string     `gorm:"column:site_name"`
	PrimaryDomain     string     `gorm:"column:primary_domain"`
	H5Domain          string     `gorm:"column:h5_domain"`
	CountryCode       string     `gorm:"column:country_code"`
	CurrencyCode      string     `gorm:"column:currency_code"`
	DefaultLanguage   string     `gorm:"column:default_language"`
	Timezone          string     `gorm:"column:timezone"`
	BrandConfig       string     `gorm:"column:brand_config"`
	ThemeConfig       string     `gorm:"column:theme_config"`
	FeatureConfig     string     `gorm:"column:feature_config"`
	MaintenanceConfig string     `gorm:"column:maintenance_config"`
	SEOConfig         string     `gorm:"column:seo_config"`
	ConfigVersion     uint64     `gorm:"column:config_version"`
	ConfigPublishedAt *time.Time `gorm:"column:config_published_at"`
	Enabled           bool       `gorm:"column:enabled"`
	Status            string     `gorm:"column:status"`
}

type h5BootstrapCountryRow struct {
	CountryCode           string `gorm:"column:country_code"`
	ISOCountryCode        string `gorm:"column:iso_country_code"`
	CountryNameZh         string `gorm:"column:country_name_zh"`
	CountryNameEn         string `gorm:"column:country_name_en"`
	CurrencyCode          string `gorm:"column:currency_code"`
	CurrencyNameZh        string `gorm:"column:currency_name_zh"`
	CurrencyNameEn        string `gorm:"column:currency_name_en"`
	CurrencySymbol        string `gorm:"column:currency_symbol"`
	CurrencyDecimalPlaces int    `gorm:"column:currency_decimal_places"`
	Timezone              string `gorm:"column:timezone"`
	DefaultLanguage       string `gorm:"column:default_language"`
	SupportedLanguages    string `gorm:"column:supported_languages"`
	PhoneCountryCode      string `gorm:"column:phone_country_code"`
	PhoneLength           int    `gorm:"column:phone_length"`
	BrandingConfig        string `gorm:"column:branding_config"`
	ThemeConfig           string `gorm:"column:theme_config"`
	Status                bool   `gorm:"column:status"`
}

func (h *gameGamesHandler) Bootstrap(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	site := h5BootstrapSite(db, scope.SiteID)
	countryCode := strings.ToUpper(strings.TrimSpace(site.CountryCode))
	country := h5BootstrapCountry(db, countryCode)

	brandConfig := h5BootstrapJSON(site.BrandConfig)
	countryBranding := h5BootstrapJSON(country.BrandingConfig)
	branding := h5MergeMaps(countryBranding, brandConfig)
	branding = h5NormalizeBranding(branding, site, country)

	theme := h5MergeMaps(h5BootstrapJSON(country.ThemeConfig), h5BootstrapJSON(site.ThemeConfig))
	theme = h5NormalizeTheme(theme)
	features := h5NormalizeFeatures(h5BootstrapJSON(site.FeatureConfig))
	maintenance := h5EffectiveMaintenance(c, db, scope)

	providers := h5BootstrapProviders(c, db, scope)
	languageDefault := h5FirstNonEmpty(site.DefaultLanguage, country.DefaultLanguage, "en")
	currencyCode := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CurrencyCode, country.CurrencyCode)))

	response.Success(c, gin.H{

		"siteId":        site.ID,
		"configVersion": h5ConfigVersion(site.ConfigVersion),
		"publishedAt":   site.ConfigPublishedAt,
		"site": gin.H{
			"id": site.ID, "siteCode": site.SiteCode, "siteName": site.SiteName,
			"primaryDomain": site.PrimaryDomain, "h5Domain": site.H5Domain, "status": site.Status,
		},
		"branding": branding,
		"theme":    theme,
		"country": gin.H{
			"code": country.CountryCode, "isoCode": country.ISOCountryCode, "nameZh": country.CountryNameZh,
			"nameEn": country.CountryNameEn, "phoneCountryCode": country.PhoneCountryCode, "phoneLength": country.PhoneLength,
		},
		"currency": gin.H{
			"code": currencyCode, "nameZh": country.CurrencyNameZh, "nameEn": country.CurrencyNameEn,
			"symbol": country.CurrencySymbol, "decimalPlaces": country.CurrencyDecimalPlaces,
		},
		"language": gin.H{
			"default": languageDefault, "supported": h5CSV(country.SupportedLanguages),
		},
		"providers":   providers,
		"features":    features,
		"maintenance": maintenance,
		"timezone":    h5FirstNonEmpty(site.Timezone, country.Timezone),
		"domain":      h5RequestHost(c),
	})
}

func h5BootstrapSite(db interface {
	Table(string, ...interface{}) *gorm.DB
}, siteID uint64) h5BootstrapSiteRow {
	var row h5BootstrapSiteRow
	if siteID == 0 {
		return row
	}
	_ = db.Table("game_site_config").
		Select("id, site_code, site_name, primary_domain, h5_domain, country_code, currency_code, default_language, timezone, COALESCE(CAST(brand_config AS CHAR), '') AS brand_config, COALESCE(CAST(theme_config AS CHAR), '') AS theme_config, COALESCE(CAST(feature_config AS CHAR), '') AS feature_config, COALESCE(CAST(maintenance_config AS CHAR), '') AS maintenance_config, COALESCE(CAST(seo_config AS CHAR), '') AS seo_config, config_version, config_published_at, enabled, status").
		Where("id=? AND deleted_at IS NULL", siteID).
		Take(&row).Error
	return row
}

func h5ConfigVersion(value uint64) uint64 {
	if value == 0 {
		return 1
	}
	return value
}

func h5BootstrapCountry(db interface {
	Table(string, ...interface{}) *gorm.DB
}, countryCode string) h5BootstrapCountryRow {
	var row h5BootstrapCountryRow
	query := db.Table("game_country_configs").
		Select("country_code,iso_country_code,country_name_zh,country_name_en,currency_code,currency_name_zh,currency_name_en,currency_symbol,currency_decimal_places,timezone,default_language,supported_languages,phone_country_code,phone_length,COALESCE(CAST(branding_config AS CHAR),'') branding_config,COALESCE(CAST(theme_config AS CHAR),'') theme_config,status").
		Where("deleted_at IS NULL AND status=1")
	if countryCode != "" {
		if err := query.Where("country_code=?", countryCode).Take(&row).Error; err == nil {
			return row
		}
	}
	_ = db.Table("game_country_configs").
		Select("country_code,iso_country_code,country_name_zh,country_name_en,currency_code,currency_name_zh,currency_name_en,currency_symbol,currency_decimal_places,timezone,default_language,supported_languages,phone_country_code,phone_length,COALESCE(CAST(branding_config AS CHAR),'') branding_config,COALESCE(CAST(theme_config AS CHAR),'') theme_config,status").
		Where("deleted_at IS NULL AND status=1").
		Order("is_default DESC,id ASC").
		Take(&row).Error
	return row
}

func h5BootstrapProviders(c *gin.Context, db *gorm.DB, scope content.Scope) []gin.H {
	var rows []model.GameGameProviders
	q := content.ApplySite(db.Model(&model.GameGameProviders{}), "", scope).
		Where("status IN ?", []int{1, 2})
	if supported := h5SupportedGameProviderCodes(db); len(supported) > 0 {
		q = q.Where("provider_code IN ?", supported)
	} else {
		q = q.Where("1=0")
	}
	_ = q.Order("sort ASC,id ASC").Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	country, currency, _ := h5PromotionAudience(c, db, scope.SiteID)
	for _, row := range rows {
		if h5ProviderMaintenanceHidden(db, scope, row.ProviderCode, country, currency) {
			continue
		}
		out = append(out, gin.H{
			"id": row.ID, "providerCode": row.ProviderCode, "name": row.Name, "logoUrl": h5AssetURL(row.LogoURL),
			"status": row.Status, "maintenance": row.Status == 2, "maintenanceMessage": row.MaintenanceMessage,
			"walletMode": row.WalletMode, "sortOrder": row.Sort,
		})
	}
	return out
}

func h5BootstrapJSON(raw string) gin.H {
	value := h5JSONValue(raw)
	if value == nil {
		return gin.H{}
	}
	if out, ok := value.(map[string]interface{}); ok {
		return gin.H(out)
	}
	if out, ok := value.(gin.H); ok {
		return out
	}
	return gin.H{}
}

func h5MergeMaps(base gin.H, override gin.H) gin.H {
	out := gin.H{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range override {
		out[key] = value
	}
	return out
}

func h5NormalizeBranding(config gin.H, site h5BootstrapSiteRow, country h5BootstrapCountryRow) gin.H {
	appName := h5AnyString(config, "appName", h5FirstNonEmpty(site.SiteName, "Casino"))
	seo := h5BootstrapJSON(site.SEOConfig)
	out := gin.H{
		"appName":         appName,
		"logo":            h5AnyAsset(config, "logo", "logoUrl"),
		"darkLogo":        h5AnyAsset(config, "darkLogo", "darkLogoUrl"),
		"favicon":         h5AnyAsset(config, "favicon", "faviconUrl"),
		"loadingLogo":     h5AnyAsset(config, "loadingLogo", "loadingLogoUrl"),
		"splash":          h5AnyAsset(config, "splash", "splashUrl"),
		"customerService": h5AnyString(config, "customerService", ""),
		"telegram":        h5AnyString(config, "telegram", ""),
		"email":           h5AnyString(config, "email", ""),
		"facebook":        h5AnyString(config, "facebook", ""),
		"x":               h5AnyString(config, "x", ""),
		"copyright":       h5AnyString(config, "copyright", ""),
		"seoTitle":        h5AnyString(seo, "title", h5AnyString(config, "seoTitle", appName)),
		"seoDescription":  h5AnyString(seo, "description", h5AnyString(config, "seoDescription", "")),
		"countryCode":     country.CountryCode,
	}
	return h5SignJSONAssetReferences(out).(gin.H)
}

func h5NormalizeTheme(config gin.H) gin.H {
	return gin.H{
		"primary":    h5AnyString(config, "primary", h5AnyString(config, "primaryColor", "")),
		"secondary":  h5AnyString(config, "secondary", h5AnyString(config, "secondaryColor", "")),
		"accent":     h5AnyString(config, "accent", ""),
		"background": h5AnyString(config, "background", h5AnyString(config, "backgroundColor", "")),
		"card":       h5AnyString(config, "card", ""),
		"button":     h5AnyString(config, "button", h5AnyString(config, "buttonColor", "")),
		"text":       h5AnyString(config, "text", h5AnyString(config, "textColor", "")),
	}
}

func h5NormalizeFeatures(config gin.H) gin.H {
	defaults := gin.H{"vip": true, "promotion": true, "referral": true, "deposit": true, "withdraw": true, "chat": true, "mission": true, "jackpot": true}
	for key, value := range config {
		defaults[strings.ToLower(key)] = value
	}
	return defaults
}

func h5NormalizeMaintenance(config gin.H, site h5BootstrapSiteRow) gin.H {
	enabled, _ := config["enabled"].(bool)
	enabled = enabled || !site.Enabled || strings.EqualFold(site.Status, "DISABLED")
	return gin.H{"enabled": enabled, "message": h5AnyString(config, "message", "")}
}

func h5AnyString(values gin.H, key string, fallback string) string {
	if value, ok := values[key]; ok {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}
	return fallback
}

func h5AnyAsset(values gin.H, keys ...string) string {
	for _, key := range keys {
		if value := h5AnyString(values, key, ""); value != "" {
			return h5AssetURL(value)
		}
	}
	return ""
}

func h5CSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
