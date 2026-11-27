package systemadmin

import (
	"casino/internal/singlesite"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SiteRow struct {
	ID uint64 `json:"id"`

	SiteCode             string     `json:"siteCode" gorm:"column:site_code"`
	SiteName             string     `json:"siteName" gorm:"column:site_name"`
	PrimaryDomain        string     `json:"primaryDomain" gorm:"column:primary_domain"`
	AdminDomain          string     `json:"adminDomain" gorm:"column:admin_domain"`
	H5Domain             string     `json:"h5Domain" gorm:"column:h5_domain"`
	AndroidPackageName   string     `json:"androidPackageName" gorm:"column:android_package_name"`
	AndroidAPKObjectKey  string     `json:"androidAPKObjectKey" gorm:"column:android_apk_object_key"`
	AndroidAPKName       string     `json:"androidAPKName" gorm:"column:android_apk_name"`
	AndroidAPKSize       int64      `json:"androidAPKSize" gorm:"column:android_apk_size"`
	AndroidAPKUploadedAt *time.Time `json:"androidAPKUploadedAt" gorm:"column:android_apk_uploaded_at"`
	IOSBundleID          string     `json:"iosBundleID" gorm:"column:ios_bundle_id"`
	CountryCode          string     `json:"countryCode" gorm:"column:country_code"`
	CurrencyCode         string     `json:"currencyCode" gorm:"column:currency_code"`
	DefaultLanguage      string     `json:"defaultLanguage" gorm:"column:default_language"`
	Timezone             string     `json:"timezone"`
	BrandConfig          string     `json:"brandConfig" gorm:"column:brand_config"`
	ThemeConfig          string     `json:"themeConfig" gorm:"column:theme_config"`
	FeatureConfig        string     `json:"featureConfig" gorm:"column:feature_config"`
	MaintenanceConfig    string     `json:"maintenanceConfig" gorm:"column:maintenance_config"`
	SEOConfig            string     `json:"seoConfig" gorm:"column:seo_config"`
	ConfigVersion        uint64     `json:"configVersion" gorm:"column:config_version"`
	IsInitialized        bool       `json:"isInitialized" gorm:"column:is_initialized"`
	Enabled              bool       `json:"enabled"`
	Status               string     `json:"status"`
	Remark               string     `json:"remark"`
	CreatedAt            time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type SiteInput struct {
	SiteCode            string         `json:"siteCode"`
	SiteName            string         `json:"siteName"`
	PrimaryDomain       string         `json:"primaryDomain"`
	AdminDomain         string         `json:"adminDomain"`
	H5Domain            string         `json:"h5Domain"`
	AndroidPackageName  string         `json:"androidPackageName"`
	AndroidAPKObjectKey string         `json:"androidAPKObjectKey"`
	AndroidAPKName      string         `json:"androidAPKName"`
	AndroidAPKSize      int64          `json:"androidAPKSize"`
	IOSBundleID         string         `json:"iosBundleID"`
	CountryCode         string         `json:"countryCode"`
	CurrencyCode        string         `json:"currencyCode"`
	DefaultLanguage     string         `json:"defaultLanguage"`
	Timezone            string         `json:"timezone"`
	BrandConfig         map[string]any `json:"brandConfig"`
	ThemeConfig         map[string]any `json:"themeConfig"`
	FeatureConfig       map[string]any `json:"featureConfig"`
	MaintenanceConfig   map[string]any `json:"maintenanceConfig"`
	SEOConfig           map[string]any `json:"seoConfig"`
	Enabled             *bool          `json:"enabled"`
	Status              string         `json:"status"`
	Remark              string         `json:"remark"`
	Reason              string         `json:"reason"`
}

func normalizeStatus(enabled bool, value string, active string, disabled string) string {
	status := strings.ToUpper(strings.TrimSpace(value))
	if status == "" {
		if enabled {
			return active
		}
		return disabled
	}
	return status
}

func validateSiteInput(input SiteInput, creating bool) error {
	if strings.TrimSpace(input.SiteName) == "" {
		return NewError("SITE_NAME_REQUIRED", "site name is required")
	}
	if strings.TrimSpace(input.Timezone) == "" {
		return NewError("TIMEZONE_REQUIRED", "timezone is required")
	}
	if key := strings.TrimSpace(input.AndroidAPKObjectKey); key != "" && !strings.HasPrefix(key, "admin/uploads/apk/") {
		return NewError("INVALID_APK_REFERENCE", "invalid Android APK reference")
	}
	return nil
}

func siteSelect() string {
	return `s.id,s.site_code,s.site_name,s.primary_domain,s.admin_domain,s.h5_domain,s.android_package_name,s.ios_bundle_id,s.country_code,s.currency_code,s.default_language,s.timezone,COALESCE(CAST(s.brand_config AS CHAR),'{}') brand_config,COALESCE(CAST(s.theme_config AS CHAR),'{}') theme_config,COALESCE(CAST(s.feature_config AS CHAR),'{}') feature_config,COALESCE(CAST(s.maintenance_config AS CHAR),'{}') maintenance_config,COALESCE(CAST(s.seo_config AS CHAR),'{}') seo_config,s.config_version,s.is_initialized,s.enabled,s.status,s.remark,s.created_at,s.updated_at`
}

func (s *Service) SiteDetail(id uint64, actor Actor) (map[string]any, error) {
	var row SiteRow
	db := s.db.Table("game_site_config s").Select(siteSelect()).Where("s.id=? AND s.deleted_at IS NULL", id)
	if !actor.Platform {
		db = db
		if actor.SiteID > 0 {
			db = db.Where("s.id=?", actor.SiteID)
		}
	}
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("site")
	}
	if s.db.Migrator().HasColumn("game_site_config", "android_apk_object_key") {
		_ = s.db.Table("game_site_config").Select("android_apk_object_key,android_apk_name,android_apk_size,android_apk_uploaded_at").Where("id=?", id).Take(&row).Error
	}
	admins := []map[string]any{}
	_ = s.db.Table("game_admin_accounts").Select("id,username,display_name displayName,mobile,email,status,created_at createdAt").Where("site_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Limit(100).Find(&admins).Error
	return map[string]any{
		"overview":          row,
		"administrators":    admins,
		"brandConfig":       parseJSONMap(row.BrandConfig),
		"themeConfig":       parseJSONMap(row.ThemeConfig),
		"featureConfig":     parseJSONMap(row.FeatureConfig),
		"maintenanceConfig": parseJSONMap(row.MaintenanceConfig),
		"seoConfig":         parseJSONMap(row.SEOConfig),
		"configuration":     siteConfiguration(row),
		"configMeta":        siteConfigMeta(row),
		"operationLogs":     s.operationLogs("site", id),
	}, nil
}

func (s *Service) UpdateSite(id uint64, input SiteInput, actor Actor) error {
	if err := validateSiteInput(input, false); err != nil {
		return err
	}
	var before map[string]any
	db := s.db.Table("game_site_config").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db
		if actor.SiteID > 0 {
			db = db.Where("id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("site")
	}
	if scope, ok := singlesite.Scope(); ok {
		if id != scope.SiteID {
			return NewError("SITE_NOT_FOUND", "site not found")
		}
		if err := s.validateSingleSiteMarket(input, scope); err != nil {
			return err
		}
	}
	input = preserveSiteDefaults(input, before)
	enabled := boolDefault(input.Enabled, dbTruthy(before["enabled"]))
	updates := siteImmediateInputRow(input, enabled, actor.ID)
	stripUnavailableSiteColumns(s.db, updates)
	action := "system.site.update"

	return s.db.Transaction(func(tx *gorm.DB) error {
		updateDB := tx.Table("game_site_config").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			updateDB = updateDB
			if actor.SiteID > 0 {
				updateDB = updateDB.Where("id=?", actor.SiteID)
			}
		}
		if err := updateDB.Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:site:update", action, "site", id, before, updates, input.Reason)
	})
}

func (s *Service) SiteStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var before struct {
		Enabled bool   `gorm:"column:enabled"`
		Status  string `gorm:"column:status"`
	}
	if err := s.db.Table("game_site_config").Select("enabled, status").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("site")
	}
	if !actor.Platform && actor.SiteID > 0 && id != actor.SiteID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	updates := map[string]any{"enabled": boolValue(enabled), "status": normalizeStatus(enabled, "", "ONLINE", "DISABLED"), "updated_by": actor.ID, "version": gorm.Expr("version+1"), "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_site_config").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:site:status", "system.site.status", "site", id, before, updates, reason)
	})
}

func siteInputRow(input SiteInput, enabled bool, actorID uint64) map[string]any {
	brand := marshal(input.BrandConfig)
	theme := marshal(input.ThemeConfig)
	features := marshal(input.FeatureConfig)
	maintenance := marshal(input.MaintenanceConfig)
	seo := marshal(input.SEOConfig)
	return map[string]any{
		"site_name":      strings.TrimSpace(input.SiteName),
		"primary_domain": strings.TrimSpace(input.PrimaryDomain), "admin_domain": strings.TrimSpace(input.AdminDomain), "h5_domain": strings.TrimSpace(input.H5Domain),
		"android_package_name": strings.TrimSpace(input.AndroidPackageName), "ios_bundle_id": strings.TrimSpace(input.IOSBundleID),
		"android_apk_object_key": strings.TrimSpace(input.AndroidAPKObjectKey), "android_apk_name": strings.TrimSpace(input.AndroidAPKName), "android_apk_size": input.AndroidAPKSize,
		"android_apk_uploaded_at": apkUploadedAt(input.AndroidAPKObjectKey),
		"country_code":            strings.ToUpper(strings.TrimSpace(input.CountryCode)), "currency_code": strings.ToUpper(strings.TrimSpace(input.CurrencyCode)),
		"default_language": strings.TrimSpace(input.DefaultLanguage), "timezone": strings.TrimSpace(input.Timezone), "brand_config": brand,
		"theme_config": theme, "feature_config": features, "maintenance_config": maintenance, "seo_config": seo,

		"config_version": 1,
		"is_initialized": 0,
		"enabled":        boolValue(enabled), "status": normalizeStatus(enabled, input.Status, "ONLINE", "DISABLED"), "remark": strings.TrimSpace(input.Remark),
		"created_by": actorID, "updated_by": actorID, "created_at": now(), "updated_at": now(),
	}
}

func siteImmediateInputRow(input SiteInput, enabled bool, actorID uint64) map[string]any {
	values := siteInputRow(input, enabled, actorID)
	for _, key := range []string{"created_by", "created_at", "config_version", "is_initialized"} {
		delete(values, key)
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["version"] = gorm.Expr("version+1")
	return values
}

func apkUploadedAt(objectKey string) any {
	if strings.TrimSpace(objectKey) == "" {
		return nil
	}
	return now()
}

func stripUnavailableSiteColumns(db *gorm.DB, values map[string]any) {
	if db.Migrator().HasColumn("game_site_config", "android_apk_object_key") {
		return
	}
	for _, key := range []string{"android_apk_object_key", "android_apk_name", "android_apk_size", "android_apk_uploaded_at"} {
		delete(values, key)
	}
}

func preserveSiteDefaults(input SiteInput, before map[string]any) SiteInput {
	if input.BrandConfig == nil {
		input.BrandConfig = parseJSONMap(mapString(before, "brand_config"))
	}
	if input.ThemeConfig == nil {
		input.ThemeConfig = parseJSONMap(mapString(before, "theme_config"))
	}
	if input.FeatureConfig == nil {
		input.FeatureConfig = parseJSONMap(mapString(before, "feature_config"))
	}
	if input.MaintenanceConfig == nil {
		input.MaintenanceConfig = parseJSONMap(mapString(before, "maintenance_config"))
	}
	if input.SEOConfig == nil {
		input.SEOConfig = parseJSONMap(mapString(before, "seo_config"))
	}
	return input
}

func mapString(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		data, _ := json.Marshal(typed)
		return string(data)
	}
}

func parseJSONMap(value string) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func dbTruthy(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case int:
		return typed != 0
	case int8:
		return typed != 0
	case int16:
		return typed != 0
	case int32:
		return typed != 0
	case int64:
		return typed != 0
	case uint:
		return typed != 0
	case uint8:
		return typed != 0
	case uint16:
		return typed != 0
	case uint32:
		return typed != 0
	case uint64:
		return typed != 0
	case []byte:
		return string(typed) != "" && string(typed) != "0"
	case string:
		return strings.TrimSpace(typed) != "" && strings.TrimSpace(typed) != "0"
	default:
		return false
	}
}
