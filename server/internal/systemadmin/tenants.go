package systemadmin

import (
	"casino/internal/singlesite"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TenantRow struct {
	ID                 uint64     `json:"id"`
	TenantCode         string     `json:"tenantCode" gorm:"column:tenant_code"`
	TenantName         string     `json:"tenantName" gorm:"column:tenant_name"`
	CompanyName        string     `json:"companyName" gorm:"column:company_name"`
	ContactName        string     `json:"contactName" gorm:"column:contact_name"`
	ContactEmail       string     `json:"contactEmail" gorm:"column:contact_email"`
	ContactPhone       string     `json:"contactPhone" gorm:"column:contact_phone"`
	PlanCode           string     `json:"planCode" gorm:"column:plan_code"`
	CountryCode        string     `json:"countryCode" gorm:"column:country_code"`
	PackageCode        string     `json:"packageCode" gorm:"column:package_code"`
	DefaultCurrency    string     `json:"defaultCurrency" gorm:"column:default_currency"`
	DefaultLanguage    string     `json:"defaultLanguage" gorm:"column:default_language"`
	Timezone           string     `json:"timezone"`
	ExpiresAt          *time.Time `json:"expiresAt" gorm:"column:expires_at"`
	Enabled            bool       `json:"enabled"`
	Status             string     `json:"status"`
	MaintenanceEnabled bool       `json:"maintenanceEnabled" gorm:"column:maintenance_enabled"`
	MaintenanceMessage string     `json:"maintenanceMessage" gorm:"column:maintenance_message"`
	SiteCount          int64      `json:"siteCount" gorm:"column:site_count"`
	AdminCount         int64      `json:"adminCount" gorm:"column:admin_count"`
	PrimaryDomain      string     `json:"primaryDomain" gorm:"column:primary_domain"`
	AdminDomain        string     `json:"adminDomain" gorm:"column:admin_domain"`
	H5Domain           string     `json:"h5Domain" gorm:"column:h5_domain"`
	DomainStatus       string     `json:"domainStatus" gorm:"column:domain_status"`
	OwnerAdminID       uint64     `json:"ownerAdminID" gorm:"column:owner_admin_id"`
	OwnerUsername      string     `json:"ownerUsername" gorm:"column:owner_username"`
	GameCount          int64      `json:"gameCount" gorm:"column:game_count"`
	PlayerCount        int64      `json:"playerCount" gorm:"column:player_count"`
	LastLoginTime      *time.Time `json:"lastLoginTime" gorm:"column:last_login_at"`
	Remark             string     `json:"remark"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type SiteRow struct {
	ID                   uint64     `json:"id"`
	TenantID             uint64     `json:"tenantID" gorm:"column:tenant_id"`
	Tenant               string     `json:"tenant"`
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
	SiteNameDraft        string     `json:"siteNameDraft" gorm:"column:site_name_draft"`
	CountryCodeDraft     string     `json:"countryCodeDraft" gorm:"column:country_code_draft"`
	CurrencyCodeDraft    string     `json:"currencyCodeDraft" gorm:"column:currency_code_draft"`
	LanguageDraft        string     `json:"languageDraft" gorm:"column:default_language_draft"`
	TimezoneDraft        string     `json:"timezoneDraft" gorm:"column:timezone_draft"`
	BrandConfigDraft     string     `json:"brandConfigDraft" gorm:"column:brand_config_draft"`
	ThemeConfigDraft     string     `json:"themeConfigDraft" gorm:"column:theme_config_draft"`
	FeatureConfigDraft   string     `json:"featureConfigDraft" gorm:"column:feature_config_draft"`
	MaintenanceDraft     string     `json:"maintenanceConfigDraft" gorm:"column:maintenance_config_draft"`
	SEOConfigDraft       string     `json:"seoConfigDraft" gorm:"column:seo_config_draft"`
	ConfigStatus         string     `json:"configStatus" gorm:"column:config_status"`
	ConfigVersion        uint64     `json:"configVersion" gorm:"column:config_version"`
	ConfigPublishedAt    *time.Time `json:"configPublishedAt" gorm:"column:config_published_at"`
	ConfigDraftUpdatedAt *time.Time `json:"configDraftUpdatedAt" gorm:"column:config_draft_updated_at"`
	IsInitialized        bool       `json:"isInitialized" gorm:"column:is_initialized"`
	Enabled              bool       `json:"enabled"`
	Status               string     `json:"status"`
	Remark               string     `json:"remark"`
	CreatedAt            time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type TenantInput struct {
	TenantCode         string     `json:"tenantCode"`
	TenantName         string     `json:"tenantName"`
	CompanyName        string     `json:"companyName"`
	ContactName        string     `json:"contactName"`
	ContactEmail       string     `json:"contactEmail"`
	ContactPhone       string     `json:"contactPhone"`
	PlanCode           string     `json:"planCode"`
	CountryCode        string     `json:"countryCode"`
	PackageCode        string     `json:"packageCode"`
	DefaultCurrency    string     `json:"defaultCurrency"`
	DefaultLanguage    string     `json:"defaultLanguage"`
	Timezone           string     `json:"timezone"`
	PrimaryDomain      *string    `json:"primaryDomain"`
	AdminDomain        *string    `json:"adminDomain"`
	H5Domain           *string    `json:"h5Domain"`
	ExpiresAt          *time.Time `json:"expiresAt"`
	Enabled            *bool      `json:"enabled"`
	Status             string     `json:"status"`
	MaintenanceEnabled *bool      `json:"maintenanceEnabled"`
	MaintenanceMessage string     `json:"maintenanceMessage"`
	Remark             string     `json:"remark"`
	OwnerUsername      string     `json:"ownerUsername"`
	OwnerPassword      string     `json:"ownerPassword"`
	OwnerDisplayName   string     `json:"ownerDisplayName"`
	OwnerEmail         string     `json:"ownerEmail"`
	OwnerMobile        string     `json:"ownerMobile"`
	Reason             string     `json:"reason"`
}

type SiteInput struct {
	TenantID            uint64         `json:"tenantID"`
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

func normalizeSaaSCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateSaaSCode(value string) error {
	matched, _ := regexp.MatchString(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])$`, normalizeSaaSCode(value))
	if !matched {
		return NewError("INVALID_CODE", "invalid SaaS code")
	}
	return nil
}

func normalizeStatus(enabled bool, value, active, disabled string) string {
	status := strings.ToUpper(strings.TrimSpace(value))
	if status == "" {
		if enabled {
			return active
		}
		return disabled
	}
	return status
}

func validateTenantInput(input TenantInput, creating bool) error {
	if creating && validateSaaSCode(input.TenantCode) != nil {
		return NewError("INVALID_TENANT_CODE", "invalid tenant code")
	}
	if strings.TrimSpace(input.TenantName) == "" {
		return NewError("TENANT_NAME_REQUIRED", "tenant name is required")
	}
	if strings.TrimSpace(input.DefaultCurrency) == "" {
		return NewError("CURRENCY_REQUIRED", "default currency is required")
	}
	if strings.TrimSpace(input.DefaultLanguage) == "" {
		return NewError("LANGUAGE_REQUIRED", "default language is required")
	}
	if strings.TrimSpace(input.Timezone) == "" {
		return NewError("TIMEZONE_REQUIRED", "timezone is required")
	}
	if creating && strings.TrimSpace(input.OwnerUsername) != "" {
		if validateCode(input.OwnerUsername) != nil {
			return NewError("INVALID_OWNER_USERNAME", "invalid owner username")
		}
		if len(input.OwnerPassword) < 10 {
			return NewError("WEAK_PASSWORD", "password must contain at least 10 characters")
		}
	}
	return nil
}

func validateSiteInput(input SiteInput, creating bool) error {
	if input.TenantID == 0 {
		return NewError("TENANT_REQUIRED", "tenant is required")
	}
	if creating && validateSaaSCode(input.SiteCode) != nil {
		return NewError("INVALID_SITE_CODE", "invalid site code")
	}
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

func tenantSelect() string {
	return `t.id,t.tenant_code,t.tenant_name,t.company_name,t.contact_name,t.contact_email,t.contact_phone,t.plan_code,t.country_code,t.package_code,t.default_currency,t.default_language,t.timezone,t.expires_at,t.enabled,t.status,t.maintenance_enabled,t.maintenance_message,t.remark,t.created_at,t.updated_at,
		(SELECT COUNT(*) FROM game_saas_sites s WHERE s.tenant_id=t.id AND s.deleted_at IS NULL) site_count,
		(SELECT COUNT(*) FROM game_admin_accounts a WHERE a.tenant_id=t.id AND a.deleted_at IS NULL) admin_count,
		COALESCE((SELECT d.domain FROM game_saas_domains d WHERE d.tenant_id=t.id AND d.domain_type='PRIMARY' AND d.deleted_at IS NULL ORDER BY d.is_primary DESC,d.id ASC LIMIT 1),(SELECT s.primary_domain FROM game_saas_sites s WHERE s.tenant_id=t.id AND s.deleted_at IS NULL ORDER BY s.id ASC LIMIT 1),'') primary_domain,
		COALESCE((SELECT s.admin_domain FROM game_saas_sites s WHERE s.tenant_id=t.id AND s.deleted_at IS NULL ORDER BY s.id ASC LIMIT 1),'') admin_domain,
		COALESCE((SELECT s.h5_domain FROM game_saas_sites s WHERE s.tenant_id=t.id AND s.deleted_at IS NULL ORDER BY s.id ASC LIMIT 1),'') h5_domain,
		COALESCE((SELECT d.status FROM game_saas_domains d WHERE d.tenant_id=t.id AND d.domain_type='PRIMARY' AND d.deleted_at IS NULL ORDER BY d.is_primary DESC,d.id ASC LIMIT 1),'PENDING') domain_status,
		COALESCE((SELECT a.id FROM game_admin_accounts a WHERE a.tenant_id=t.id AND a.deleted_at IS NULL ORDER BY a.id ASC LIMIT 1),0) owner_admin_id,
		COALESCE((SELECT a.username FROM game_admin_accounts a WHERE a.tenant_id=t.id AND a.deleted_at IS NULL ORDER BY a.id ASC LIMIT 1),'') owner_username,
		(SELECT COUNT(*) FROM game_games g WHERE g.tenant_id=t.id AND g.platform_enabled=1 AND g.deleted_at IS NULL) game_count,
		(SELECT COUNT(*) FROM game_users u WHERE u.tenant_id=t.id AND u.deleted_at IS NULL) player_count,
		(SELECT MAX(a.last_login_at) FROM game_admin_accounts a WHERE a.tenant_id=t.id AND a.deleted_at IS NULL) last_login_at`
}

func siteSelect() string {
	return `s.id,s.tenant_id,COALESCE(t.tenant_name,'') tenant,s.site_code,s.site_name,s.primary_domain,s.admin_domain,s.h5_domain,s.android_package_name,s.ios_bundle_id,s.country_code,s.currency_code,s.default_language,s.timezone,COALESCE(CAST(s.brand_config AS CHAR),'') brand_config,COALESCE(CAST(s.theme_config AS CHAR),'') theme_config,COALESCE(CAST(s.feature_config AS CHAR),'') feature_config,COALESCE(CAST(s.maintenance_config AS CHAR),'') maintenance_config,COALESCE(CAST(s.seo_config AS CHAR),'') seo_config,COALESCE(s.site_name_draft,'') site_name_draft,COALESCE(s.country_code_draft,'') country_code_draft,COALESCE(s.currency_code_draft,'') currency_code_draft,COALESCE(s.default_language_draft,'') default_language_draft,COALESCE(s.timezone_draft,'') timezone_draft,COALESCE(CAST(s.brand_config_draft AS CHAR),'') brand_config_draft,COALESCE(CAST(s.theme_config_draft AS CHAR),'') theme_config_draft,COALESCE(CAST(s.feature_config_draft AS CHAR),'') feature_config_draft,COALESCE(CAST(s.maintenance_config_draft AS CHAR),'') maintenance_config_draft,COALESCE(CAST(s.seo_config_draft AS CHAR),'') seo_config_draft,s.config_status,s.config_version,s.config_published_at,s.config_draft_updated_at,s.is_initialized,s.enabled,s.status,s.remark,s.created_at,s.updated_at`
}

func (s *Service) ListTenants(q Query) (Result, error) {
	rows := []TenantRow{}
	db := s.db.Table("game_saas_tenants t").Select(tenantSelect()).Where("t.deleted_at IS NULL")
	db = exact(db, "t.id", value(q, "tenantID"))
	db = like(db, "t.tenant_code", value(q, "tenantCode"))
	db = like(db, "t.tenant_name", value(q, "tenantName"))
	db = like(db, "t.company_name", value(q, "companyName"))
	db = exact(db, "t.plan_code", value(q, "planCode"))
	db = exact(db, "t.country_code", strings.ToUpper(value(q, "countryCode")))
	db = exact(db, "t.package_code", strings.ToUpper(value(q, "packageCode")))
	db = exact(db, "t.default_currency", value(q, "defaultCurrency"))
	db = exact(db, "t.enabled", value(q, "status"))
	db = exact(db, "t.status", value(q, "lifecycleStatus"))
	db = timeRange(db, "t.created_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("t.created_at DESC,t.id DESC"), q, &rows)
}

func (s *Service) TenantDetail(id uint64) (map[string]any, error) {
	var row TenantRow
	if err := s.db.Table("game_saas_tenants t").Select(tenantSelect()).Where("t.id=? AND t.deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("tenant")
	}
	sites := []SiteRow{}
	_ = s.db.Table("game_saas_sites s").Select(siteSelect()).Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id").Where("s.tenant_id=? AND s.deleted_at IS NULL", id).Order("s.created_at DESC").Limit(100).Find(&sites).Error
	admins := []map[string]any{}
	_ = s.db.Table("game_admin_accounts").Select("id,site_id siteID,username,display_name displayName,mobile,email,status,mfa_enabled mfaEnabled,last_login_ip lastLoginIP,last_login_at lastLoginTime,created_at createdAt").Where("tenant_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Limit(100).Find(&admins).Error
	domains := []DomainRow{}
	_ = s.db.Table("game_saas_domains d").Select(domainSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=d.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites s ON s.id=d.site_id AND s.deleted_at IS NULL").
		Where("d.tenant_id=? AND d.deleted_at IS NULL", id).Order("d.is_primary DESC,d.domain_type ASC,d.id ASC").Find(&domains).Error
	providers := []map[string]any{}
	_ = s.db.Table("game_game_providers").
		Select("id,site_id siteID,provider_code providerCode,name,api_status apiStatus,status,maintenance_message maintenanceMessage,sort,created_at createdAt,updated_at updatedAt").
		Where("tenant_id=? AND deleted_at IS NULL", id).Order("site_id,sort,id").Limit(200).Find(&providers).Error
	games := []map[string]any{}
	_ = s.db.Table("game_games g").
		Select("g.id,g.site_id siteID,g.provider_id providerID,g.provider_game_code providerGameCode,g.name,g.game_type gameType,g.cover_url coverURL,g.platform_enabled platformEnabled,g.status,g.config_status configStatus,g.config_version configVersion,g.updated_at updatedAt").
		Where("g.tenant_id=? AND g.deleted_at IS NULL", id).Order("g.site_id,g.sort,g.id").Limit(500).Find(&games).Error
	assets := []map[string]any{}
	_ = s.db.Table("game_asset_library").
		Select("id,site_id siteID,asset_type assetType,usage_type usageType,display_name displayName,public_url publicURL,file_size fileSize,width,height,duration_ms durationMS,reference_count referenceCount,status,created_at createdAt").
		Where("tenant_id=? AND deleted_at IS NULL", id).Order("created_at DESC,id DESC").Limit(200).Find(&assets).Error
	ledgerAccounts, _ := s.TenantLedgerAccounts(id, Query{}, Actor{Platform: true})
	auditResult, _ := s.TenantAudit(id, Query{PageSize: 100}, Actor{Platform: true})
	return map[string]any{
		"overview": row, "sites": sites, "administrators": admins, "domains": domains,
		"providerAuthorizations": providers, "gameAuthorizations": games, "assets": assets,
		"ledgerAccounts": ledgerAccounts, "audit": auditResult, "operationLogs": s.operationLogs("tenant", id),
	}, nil
}

func (s *Service) CreateTenant(input TenantInput, actor Actor) (uint64, error) {
	if !actor.Platform {
		return 0, NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := validateTenantInput(input, true); err != nil {
		return 0, err
	}
	enabled := boolDefault(input.Enabled, true)
	row := map[string]any{
		"tenant_code": normalizeSaaSCode(input.TenantCode), "tenant_name": strings.TrimSpace(input.TenantName),
		"company_name": strings.TrimSpace(input.CompanyName), "contact_name": strings.TrimSpace(input.ContactName),
		"contact_email": strings.TrimSpace(input.ContactEmail), "contact_phone": strings.TrimSpace(input.ContactPhone),
		"plan_code":        strings.ToUpper(strings.TrimSpace(defaultString(input.PlanCode, "STANDARD"))),
		"country_code":     strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		"package_code":     strings.ToUpper(strings.TrimSpace(defaultString(input.PackageCode, defaultString(input.PlanCode, "STANDARD")))),
		"default_currency": strings.ToUpper(strings.TrimSpace(input.DefaultCurrency)),
		"default_language": strings.TrimSpace(input.DefaultLanguage), "timezone": strings.TrimSpace(input.Timezone),
		"expires_at": input.ExpiresAt,
		"enabled":    boolValue(enabled), "status": normalizeStatus(enabled, input.Status, "ACTIVE", "DISABLED"),
		"maintenance_enabled": boolValue(boolDefault(input.MaintenanceEnabled, false)),
		"maintenance_message": strings.TrimSpace(input.MaintenanceMessage),
		"remark":              strings.TrimSpace(input.Remark), "created_by": actor.ID, "updated_by": actor.ID, "created_at": now(), "updated_at": now(),
	}
	var tenantID uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureTenantCodeAvailable(tx, row["tenant_code"]); err != nil {
			return err
		}
		result := tx.Table("game_saas_tenants").Create(row)
		if result.Error != nil {
			return duplicateError(result.Error)
		}
		if err := tx.Table("game_saas_tenants").Select("id").Where("tenant_code=? AND deleted_at IS NULL", row["tenant_code"]).Take(&tenantID).Error; err != nil {
			return err
		}
		if strings.TrimSpace(input.OwnerUsername) != "" {
			if err := createTenantOwner(tx, tenantID, input, actor); err != nil {
				return err
			}
		}
		return audit(tx, actor, "system:tenant:create", "system.tenant.create", "tenant", tenantID, nil, row, input.Reason)
	})
	return tenantID, err
}

func ensureTenantCodeAvailable(tx *gorm.DB, code any) error {
	var count int64
	if err := tx.Table("game_saas_tenants").
		Where("tenant_code=? AND deleted_at IS NULL", code).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return NewError("TENANT_CODE_EXISTS", "tenant code already exists")
	}
	return nil
}

func createTenantOwner(tx *gorm.DB, tenantID uint64, input TenantInput, actor Actor) error {
	return createTenantOwnerWithSite(tx, tenantID, 0, input, actor)
}

func createTenantOwnerWithSite(tx *gorm.DB, tenantID, siteID uint64, input TenantInput, actor Actor) error {
	roleID, err := ensureTenantOwnerRole(tx, tenantID)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	displayName := strings.TrimSpace(input.OwnerDisplayName)
	if displayName == "" {
		displayName = "Tenant Owner"
	}
	preferredLanguage := "en-US"
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(input.DefaultLanguage)), "zh") {
		preferredLanguage = "zh-CN"
	}
	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	account := map[string]any{
		"tenant_id": tenantID, "site_id": siteID, "username": strings.TrimSpace(input.OwnerUsername), "display_name": displayName,
		"password_hash": string(hash), "mobile": strings.TrimSpace(input.OwnerMobile), "email": strings.TrimSpace(input.OwnerEmail),
		"preferred_language": preferredLanguage, "timezone": timezone,
		"status": 1, "last_password_changed_at": now(), "must_change_password": 1, "created_at": now(), "updated_at": now(),
	}
	result := tx.Table("game_admin_accounts").Create(account)
	if result.Error != nil {
		return duplicateError(result.Error)
	}
	var adminID uint64
	if err := tx.Table("game_admin_accounts").Select("id").Where("tenant_id=? AND username=? AND deleted_at IS NULL", tenantID, account["username"]).Take(&adminID).Error; err != nil {
		return err
	}
	if err := assignAdminRoles(tx, adminID, tenantID, []uint64{roleID}); err != nil {
		return err
	}
	return audit(tx, actor, "system:admin:create", "system.tenant.owner.create", "administrator", adminID, nil, map[string]any{"tenantID": tenantID, "username": input.OwnerUsername}, input.Reason)
}

func ensureTenantOwnerRole(tx *gorm.DB, tenantID uint64) (uint64, error) {
	row := map[string]any{
		"tenant_id": tenantID, "code": "tenant-owner", "name": "Tenant Owner", "name_zh": "租户负责人", "name_en": "Tenant Owner",
		"description_zh": "客户站点负责人", "description_en": "Tenant site owner", "data_scope": "ALL", "status": 1,
		"system_protected": 1, "created_at": now(), "updated_at": now(), "deleted_at": nil,
	}
	if err := tx.Table("game_admin_roles").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "code"}},
		DoUpdates: clause.Assignments(map[string]any{"deleted_at": nil, "status": 1, "updated_at": now()}),
	}).Create(row).Error; err != nil {
		return 0, err
	}
	var roleID uint64
	if err := tx.Table("game_admin_roles").Select("id").Where("tenant_id=? AND code=? AND deleted_at IS NULL", tenantID, "tenant-owner").Take(&roleID).Error; err != nil {
		return 0, err
	}
	if err := tx.Exec(`
		INSERT INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
		SELECT ?, p.id, NOW(3), NOW(3), NULL
		FROM game_admin_permissions p
		WHERE p.enabled=1 AND p.deleted_at IS NULL
		  AND ((p.code NOT LIKE 'system:%' AND p.code NOT IN ?) OR p.code IN ?)
		ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at)`, roleID, tenantOwnerDeniedPermissionCodes(), tenantOwnerPermissionCodes()).Error; err != nil {
		return 0, err
	}
	if err := tx.Exec(`
		UPDATE game_admin_role_permissions rp
		JOIN game_admin_permissions p ON p.id=rp.permission_id
		SET rp.deleted_at=NOW(3), rp.updated_at=NOW(3)
		WHERE rp.role_id=? AND rp.deleted_at IS NULL
		  AND NOT ((p.code NOT LIKE 'system:%' AND p.code NOT IN ?) OR p.code IN ?)`, roleID, tenantOwnerDeniedPermissionCodes(), tenantOwnerPermissionCodes()).Error; err != nil {
		return 0, err
	}
	if err := tx.Exec(`
		INSERT INTO game_admin_role_menus (role_id, menu_id, created_at, updated_at, deleted_at)
		SELECT ?, m.id, NOW(3), NOW(3), NULL
		FROM game_admin_menus m
		WHERE m.deleted_at IS NULL
		  AND m.code IN ?
		ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at)`, roleID, tenantOwnerMenuCodes()).Error; err != nil {
		return 0, err
	}
	if err := tx.Exec(`
		UPDATE game_admin_role_menus rm
		JOIN game_admin_menus m ON m.id=rm.menu_id
		SET rm.deleted_at=NOW(3), rm.updated_at=NOW(3)
		WHERE rm.role_id=? AND rm.deleted_at IS NULL AND m.code NOT IN ?`, roleID, tenantOwnerMenuCodes()).Error; err != nil {
		return 0, err
	}
	return roleID, nil
}

func tenantOwnerPermissionCodes() []string {
	return []string{
		"system:site:list", "system:site:view", "system:site:update", "system:site:config:preview", "system:site:config:publish",
		"system:domain:list", "system:domain:view",
		"jobs:read", "jobs:retry", "alerts:read", "alerts:acknowledge", "alerts:resolve",
	}
}

func tenantOwnerDeniedPermissionCodes() []string {
	return []string{
		"game:provider:list", "game:provider:view", "game:provider:update", "game:provider:test", "game:provider:sync", "game:sync-log:view",
		"bet:callback-log:list", "bet:callback-log:view", "bet:reconciliation:list", "bet:reconciliation:view", "bet:reconciliation:generate", "bet:reconciliation:resolve",
		"report:provider:view", "system:event:list", "system:event:view", "system:event:dispatch", "system:event:replay",
		"health:system:read", "jobs:execute",
		"backup:read", "backup:create", "backup:restore", "backup:delete",
	}
}

func tenantOwnerMenuCodes() []string {
	return []string{
		"dashboard", "users", "user-list", "login-records", "device-management", "risk-users",
		"funds", "wallet-accounts", "wallet-ledgers", "manual-adjustments", "deposit-orders", "withdrawal-orders", "financial-reconciliation", "deposit-methods", "deposit-channels", "deposit-amount-options",
		"games", "game-list", "game-categories", "lobby-sections", "game-launch-records",
		"bets", "bet-records", "rollback-records", "abnormal-bets",
		"operations", "asset-library", "banners", "announcements", "activities", "vip", "push-messages", "popups", "music", "social-links", "h5-navigation", "vouchers",
		"risk", "risk-dashboard", "risk-rules", "risk-events", "review-queue", "user-associations", "blacklist", "withdrawal-risk", "risk-operation-logs",
		"agents", "agent-list", "agent-hierarchy", "channel-management", "commission-plans", "commission-records", "agent-settlements", "agent-wallets", "agent-operation-logs",
		"reports", "financial-report", "game-report", "user-report", "channel-report", "agent-report", "promotion-report", "export-tasks",
		"support-center", "support-player-search", "support-player-timeline", "support-order-search", "support-payment-diagnostics", "support-game-diagnostics", "support-cases", "support-exports",
		"ops-stability", "payment-monitor", "provider-monitor", "maintenance-center", "release-center", "jobs-retries", "alert-center",
		"system", "sites", "brand-config", "system-audit-logs", "system-login-logs",
	}
}

func (s *Service) UpdateTenant(id uint64, input TenantInput, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := validateTenantInput(input, false); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("tenant")
	}
	enabled := boolDefault(input.Enabled, dbTruthy(before["enabled"]))
	packageCode := strings.ToUpper(strings.TrimSpace(defaultString(input.PackageCode, defaultString(input.PlanCode, "STANDARD"))))
	planCode := strings.ToUpper(strings.TrimSpace(defaultString(input.PlanCode, packageCode)))
	updates := map[string]any{
		"tenant_name": strings.TrimSpace(input.TenantName), "company_name": strings.TrimSpace(input.CompanyName),
		"contact_name": strings.TrimSpace(input.ContactName), "contact_email": strings.TrimSpace(input.ContactEmail), "contact_phone": strings.TrimSpace(input.ContactPhone),
		"plan_code": planCode, "country_code": strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		"package_code": packageCode, "default_currency": strings.ToUpper(strings.TrimSpace(input.DefaultCurrency)),
		"default_language": strings.TrimSpace(input.DefaultLanguage), "timezone": strings.TrimSpace(input.Timezone), "expires_at": input.ExpiresAt, "enabled": boolValue(enabled),
		"status": normalizeStatus(enabled, input.Status, "ACTIVE", "DISABLED"), "remark": strings.TrimSpace(input.Remark), "updated_by": actor.ID,
		"maintenance_enabled": boolValue(boolDefault(input.MaintenanceEnabled, dbTruthy(before["maintenance_enabled"]))),
		"maintenance_message": strings.TrimSpace(input.MaintenanceMessage),
		"version":             gorm.Expr("version+1"), "updated_at": now(),
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := s.updateTenantPrimarySiteDomains(tx, id, input, actor); err != nil {
			return err
		}
		return audit(tx, actor, "system:tenant:update", "system.tenant.update", "tenant", id, before, updates, input.Reason)
	})
}

func (s *Service) TenantStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(reason); err != nil {
		return err
	}
	var before struct {
		Enabled bool   `gorm:"column:enabled"`
		Status  string `gorm:"column:status"`
	}
	if err := s.db.Table("game_saas_tenants").Select("enabled,status").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("tenant")
	}
	updates := map[string]any{"enabled": boolValue(enabled), "status": normalizeStatus(enabled, "", "ACTIVE", "DISABLED"), "updated_by": actor.ID, "version": gorm.Expr("version+1"), "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_tenants").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:tenant:status", "system.tenant.status", "tenant", id, before, updates, reason)
	})
}

func (s *Service) DeleteTenant(id uint64, reason string, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(reason); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("tenant")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		nowValue := now()
		if err := softDeleteDomainsByQuery(tx, tx.Table("game_saas_domains").Where("tenant_id=? AND deleted_at IS NULL", id), actor.ID, nowValue); err != nil {
			return err
		}
		if err := tx.Table("game_saas_sites").Where("tenant_id=? AND deleted_at IS NULL", id).Updates(map[string]any{
			"enabled":    0,
			"status":     "DELETED",
			"deleted_at": nowValue,
			"updated_by": actor.ID,
			"updated_at": nowValue,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_admin_accounts").Where("tenant_id=? AND deleted_at IS NULL", id).Updates(map[string]any{
			"status":     0,
			"deleted_at": nowValue,
			"updated_at": nowValue,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_admin_sessions").
			Where("admin_id IN (SELECT id FROM game_admin_accounts WHERE tenant_id=?) AND revoked_at IS NULL", id).
			Updates(map[string]any{"revoked_at": nowValue, "updated_at": nowValue}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_saas_brand_deliveries").Where("tenant_id=? AND deleted_at IS NULL", id).Updates(map[string]any{
			"status":        "DELETED",
			"domain_status": "DISABLED",
			"deleted_at":    nowValue,
			"updated_at":    nowValue,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", id).Updates(map[string]any{
			"enabled":    0,
			"status":     "DELETED",
			"deleted_at": nowValue,
			"updated_by": actor.ID,
			"version":    gorm.Expr("version+1"),
			"updated_at": nowValue,
		}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:tenant:delete", "system.tenant.delete", "tenant", id, before, nil, reason)
	})
}

type tenantSiteDomains struct {
	ID            uint64 `gorm:"column:id"`
	PrimaryDomain string `gorm:"column:primary_domain"`
	AdminDomain   string `gorm:"column:admin_domain"`
	H5Domain      string `gorm:"column:h5_domain"`
}

func (s *Service) updateTenantPrimarySiteDomains(tx *gorm.DB, tenantID uint64, input TenantInput, actor Actor) error {
	if input.PrimaryDomain == nil && input.AdminDomain == nil && input.H5Domain == nil {
		return nil
	}
	var site tenantSiteDomains
	if err := tx.Table("game_saas_sites").
		Select("id,primary_domain,admin_domain,h5_domain").
		Where("tenant_id=? AND deleted_at IS NULL", tenantID).
		Order("id ASC").
		Take(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	siteUpdates := map[string]any{"updated_by": actor.ID, "updated_at": now()}
	if input.PrimaryDomain != nil {
		next := normalizeDomain(*input.PrimaryDomain)
		siteUpdates["primary_domain"] = next
		if err := s.syncTenantDomain(tx, tenantID, site.ID, "PRIMARY", site.PrimaryDomain, next, true, actor); err != nil {
			return err
		}
	}
	if input.H5Domain != nil {
		next := normalizeDomain(*input.H5Domain)
		siteUpdates["h5_domain"] = next
		primaryDomain := normalizeDomain(site.PrimaryDomain)
		if input.PrimaryDomain != nil {
			primaryDomain = normalizeDomain(*input.PrimaryDomain)
		}
		if next == primaryDomain {
			if site.H5Domain != "" && normalizeDomain(site.H5Domain) != primaryDomain {
				if err := softDeleteDomainsByQuery(tx, tx.Table("game_saas_domains").
					Where("tenant_id=? AND site_id=? AND domain_type=? AND domain=? AND deleted_at IS NULL", tenantID, site.ID, "H5", normalizeDomain(site.H5Domain)), actor.ID, now()); err != nil {
					return err
				}
			}
		} else {
			if err := s.syncTenantDomain(tx, tenantID, site.ID, "H5", site.H5Domain, next, false, actor); err != nil {
				return err
			}
		}
	}
	if input.AdminDomain != nil {
		next := normalizeDomain(*input.AdminDomain)
		siteUpdates["admin_domain"] = next
		if err := s.syncTenantDomain(tx, tenantID, site.ID, "ADMIN", site.AdminDomain, next, true, actor); err != nil {
			return err
		}
	}
	return tx.Table("game_saas_sites").Where("id=? AND tenant_id=? AND deleted_at IS NULL", site.ID, tenantID).Updates(siteUpdates).Error
}

func (s *Service) syncTenantDomain(tx *gorm.DB, tenantID, siteID uint64, domainType, previous, next string, primary bool, actor Actor) error {
	previous = normalizeDomain(previous)
	next = normalizeDomain(next)
	nowValue := now()
	if previous != "" && previous != next {
		if err := softDeleteDomainsByQuery(tx, tx.Table("game_saas_domains").
			Where("tenant_id=? AND site_id=? AND domain_type=? AND domain=? AND deleted_at IS NULL", tenantID, siteID, domainType, previous), actor.ID, nowValue); err != nil {
			return err
		}
	}
	if next == "" {
		return nil
	}
	if primary {
		if err := softDeleteDomainsByQuery(tx, tx.Table("game_saas_domains").
			Where("tenant_id=? AND site_id=? AND domain_type=? AND domain<>? AND deleted_at IS NULL", tenantID, siteID, domainType, next), actor.ID, nowValue); err != nil {
			return err
		}
		if err := tx.Table("game_saas_domains").
			Where("tenant_id=? AND site_id=? AND domain_type=? AND deleted_at IS NULL", tenantID, siteID, domainType).
			Update("is_primary", 0).Error; err != nil {
			return err
		}
	}
	updates := map[string]any{
		"domain":      next,
		"domain_key":  next,
		"domain_type": domainType,
		"is_primary":  boolValue(primary),
		"enabled":     1,
		"status":      "PENDING",
		"updated_by":  actor.ID,
		"updated_at":  nowValue,
	}
	var existing struct{ ID uint64 }
	err := tx.Table("game_saas_domains").
		Select("id").
		Where("tenant_id=? AND site_id=? AND domain_type=? AND domain_key=? AND deleted_at IS NULL", tenantID, siteID, domainType, next).
		Take(&existing).Error
	if err == nil {
		return tx.Table("game_saas_domains").Where("id=?", existing.ID).Updates(updates).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	row := domainInputRow(DomainInput{TenantID: tenantID, SiteID: siteID, Domain: next, DomainType: domainType, IsPrimary: primary, Status: "PENDING", Reason: "tenant domain updated"}, true, actor.ID)
	return tx.Table("game_saas_domains").Create(row).Error
}

func softDeleteDomainsByQuery(tx *gorm.DB, query *gorm.DB, actorID uint64, deletedAt time.Time) error {
	rows := []struct {
		ID        uint64 `gorm:"column:id"`
		DomainKey string `gorm:"column:domain_key"`
	}{}
	if err := query.Select("id,domain_key").Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		suffix := "#deleted#" + strconv.FormatUint(row.ID, 10) + "-" + strconv.FormatInt(deletedAt.UnixNano(), 10)
		if err := tx.Table("game_saas_domains").Where("id=?", row.ID).Updates(map[string]any{
			"domain_key": row.DomainKey + suffix,
			"enabled":    0,
			"status":     "DELETED",
			"deleted_at": deletedAt,
			"updated_by": actorID,
			"updated_at": deletedAt,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListSites(q Query, actor Actor) (Result, error) {
	rows := []SiteRow{}
	db := s.db.Table("game_saas_sites s").Select(siteSelect()).Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id AND t.deleted_at IS NULL").Where("s.deleted_at IS NULL")
	if !actor.Platform {
		db = db.Where("s.tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("s.id=?", actor.SiteID)
		}
	}
	db = exact(db, "s.tenant_id", value(q, "tenantID"))
	db = like(db, "t.tenant_name", value(q, "tenant"))
	db = like(db, "s.site_code", value(q, "siteCode"))
	db = like(db, "s.site_name", value(q, "siteName"))
	db = like(db, "s.primary_domain", value(q, "primaryDomain"))
	db = like(db, "s.admin_domain", value(q, "adminDomain"))
	db = like(db, "s.h5_domain", value(q, "h5Domain"))
	db = exact(db, "s.country_code", value(q, "countryCode"))
	db = exact(db, "s.currency_code", value(q, "currencyCode"))
	db = exact(db, "s.enabled", value(q, "status"))
	db = exact(db, "s.status", value(q, "lifecycleStatus"))
	db = timeRange(db, "s.created_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("s.created_at DESC,s.id DESC"), q, &rows)
}

func (s *Service) SiteDetail(id uint64, actor Actor) (map[string]any, error) {
	var row SiteRow
	db := s.db.Table("game_saas_sites s").Select(siteSelect()).Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id AND t.deleted_at IS NULL").Where("s.id=? AND s.deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("s.tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("s.id=?", actor.SiteID)
		}
	}
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("site")
	}
	if s.db.Migrator().HasColumn("game_saas_sites", "android_apk_object_key") {
		_ = s.db.Table("game_saas_sites").Select("android_apk_object_key,android_apk_name,android_apk_size,android_apk_uploaded_at").Where("id=?", id).Take(&row).Error
	}
	admins := []map[string]any{}
	_ = s.db.Table("game_admin_accounts").Select("id,username,display_name displayName,mobile,email,status,created_at createdAt").Where("site_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Limit(100).Find(&admins).Error
	domains := []DomainRow{}
	_ = s.db.Table("game_saas_domains").Where("site_id=? AND deleted_at IS NULL", id).Order("is_primary DESC,domain_type ASC,id ASC").Find(&domains).Error
	return map[string]any{
		"overview":          row,
		"administrators":    admins,
		"brandConfig":       parseJSONMap(row.BrandConfig),
		"themeConfig":       parseJSONMap(row.ThemeConfig),
		"featureConfig":     parseJSONMap(row.FeatureConfig),
		"maintenanceConfig": parseJSONMap(row.MaintenanceConfig),
		"seoConfig":         parseJSONMap(row.SEOConfig),
		"publishedConfig":   sitePublishedConfig(row),
		"draftConfig":       siteDraftConfig(row),
		"configMeta":        siteConfigMeta(row),
		"domains":           domains,
		"operationLogs":     s.operationLogs("site", id),
	}, nil
}

func (s *Service) CreateSite(input SiteInput, actor Actor) (uint64, error) {
	if !actor.Platform {
		input.TenantID = actor.TenantID
	}
	if err := validateSiteInput(input, true); err != nil {
		return 0, err
	}
	if err := s.ensureTenantAvailable(input.TenantID); err != nil {
		return 0, err
	}
	enabled := boolDefault(input.Enabled, true)
	row := siteInputRow(input, enabled, actor.ID)
	stripUnavailableSiteColumns(s.db, row)
	row["site_code"] = normalizeSaaSCode(input.SiteCode)
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureSiteCodeAvailable(tx, row["site_code"]); err != nil {
			return err
		}
		result := tx.Table("game_saas_sites").Create(row)
		if result.Error != nil {
			return duplicateError(result.Error)
		}
		if err := tx.Table("game_saas_sites").Select("id").Where("tenant_id=? AND site_code=? AND deleted_at IS NULL", input.TenantID, row["site_code"]).Take(&id).Error; err != nil {
			return err
		}
		if err := provisionTenantOperationDefaults(tx, input.TenantID, id, input.CurrencyCode, input.CountryCode, input.DefaultLanguage, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "system:site:create", "system.site.create", "site", id, nil, row, input.Reason)
	})
	return id, err
}

func ensureSiteCodeAvailable(tx *gorm.DB, code any) error {
	var count int64
	if err := tx.Table("game_saas_sites").
		Where("site_code=? AND deleted_at IS NULL", code).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return NewError("SITE_CODE_EXISTS", "site code already exists")
	}
	return nil
}

func (s *Service) UpdateSite(id uint64, input SiteInput, actor Actor) error {
	if !actor.Platform {
		input.TenantID = actor.TenantID
	}
	if err := validateSiteInput(input, false); err != nil {
		return err
	}
	if err := s.ensureTenantAvailable(input.TenantID); err != nil {
		return err
	}
	var before map[string]any
	db := s.db.Table("game_saas_sites").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("site")
	}
	if !actor.Platform && !singlesite.Active() {
		// Country and currency are commercial dimensions fixed during platform
		// provisioning. Tenant operators cannot alter them through a crafted request.
		input.CountryCode = mapString(before, "country_code")
		input.CurrencyCode = mapString(before, "currency_code")
	}
	if scope, ok := singlesite.Scope(); ok {
		if id != scope.SiteID {
			return NewError("SITE_NOT_FOUND", "site not found")
		}
		if err := s.validateSingleSiteMarket(input, scope); err != nil {
			return err
		}
	}
	input = preserveSiteDraftDefaults(input, before)
	updates := siteDraftInputRow(input, actor.ID)
	action := "system.site.config.draft"
	if actor.Platform {
		enabled := boolDefault(input.Enabled, dbTruthy(before["enabled"]))
		updates = siteImmediateInputRow(input, enabled, actor.ID)
		stripUnavailableSiteColumns(s.db, updates)
		action = "system.site.update"
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
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
		return audit(tx, actor, "system:site:update", action, "site", id, before, updates, input.Reason)
	})
}

func (s *Service) SiteStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var before struct {
		TenantID uint64 `gorm:"column:tenant_id"`
		Enabled  bool   `gorm:"column:enabled"`
		Status   string `gorm:"column:status"`
	}
	if err := s.db.Table("game_saas_sites").Select("tenant_id,enabled,status").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("site")
	}
	if !actor.Platform && before.TenantID != actor.TenantID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if !actor.Platform && actor.SiteID > 0 && id != actor.SiteID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	updates := map[string]any{"enabled": boolValue(enabled), "status": normalizeStatus(enabled, "", "ONLINE", "DISABLED"), "updated_by": actor.ID, "version": gorm.Expr("version+1"), "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_sites").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:site:status", "system.site.status", "site", id, before, updates, reason)
	})
}

func (s *Service) DeleteSite(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var before map[string]any
	db := s.db.Table("game_saas_sites").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("site")
	}
	var adminCount int64
	_ = s.db.Table("game_admin_accounts").Where("site_id=? AND deleted_at IS NULL", id).Count(&adminCount).Error
	if adminCount > 0 {
		return NewError("SITE_IN_USE", "site has administrators")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		deleteDB := tx.Table("game_saas_sites").Where("id=?", id)
		if !actor.Platform {
			deleteDB = deleteDB.Where("tenant_id=?", actor.TenantID)
		}
		if err := deleteDB.Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:site:delete", "system.site.delete", "site", id, before, nil, reason)
	})
}

func (s *Service) ensureTenantAvailable(id uint64) error {
	if scope, ok := singlesite.Scope(); ok {
		if id != scope.TenantID {
			return NewError("SITE_NOT_FOUND", "site not found")
		}
		return nil
	}

	var count int64
	if err := s.db.Table("game_saas_tenants").Where("id=? AND enabled=1 AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return rowNotFound("tenant")
	}
	return nil
}

func siteInputRow(input SiteInput, enabled bool, actorID uint64) map[string]any {
	brand := marshal(input.BrandConfig)
	theme := marshal(input.ThemeConfig)
	features := marshal(input.FeatureConfig)
	maintenance := marshal(input.MaintenanceConfig)
	seo := marshal(input.SEOConfig)
	return map[string]any{
		"tenant_id": input.TenantID, "site_name": strings.TrimSpace(input.SiteName),
		"primary_domain": strings.TrimSpace(input.PrimaryDomain), "admin_domain": strings.TrimSpace(input.AdminDomain), "h5_domain": strings.TrimSpace(input.H5Domain),
		"android_package_name": strings.TrimSpace(input.AndroidPackageName), "ios_bundle_id": strings.TrimSpace(input.IOSBundleID),
		"android_apk_object_key": strings.TrimSpace(input.AndroidAPKObjectKey), "android_apk_name": strings.TrimSpace(input.AndroidAPKName), "android_apk_size": input.AndroidAPKSize,
		"android_apk_uploaded_at": apkUploadedAt(input.AndroidAPKObjectKey),
		"country_code":            strings.ToUpper(strings.TrimSpace(input.CountryCode)), "currency_code": strings.ToUpper(strings.TrimSpace(input.CurrencyCode)),
		"default_language": strings.TrimSpace(input.DefaultLanguage), "timezone": strings.TrimSpace(input.Timezone), "brand_config": brand,
		"theme_config": theme, "feature_config": features, "maintenance_config": maintenance, "seo_config": seo,
		"site_name_draft": strings.TrimSpace(input.SiteName), "country_code_draft": strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		"currency_code_draft": strings.ToUpper(strings.TrimSpace(input.CurrencyCode)), "default_language_draft": strings.TrimSpace(input.DefaultLanguage),
		"timezone_draft": strings.TrimSpace(input.Timezone), "brand_config_draft": brand, "theme_config_draft": theme,
		"feature_config_draft": features, "maintenance_config_draft": maintenance, "seo_config_draft": seo,
		"config_status": "PUBLISHED", "config_version": 1, "config_published_at": now(), "config_published_by": actorID,
		"is_initialized": 0,
		"enabled":        boolValue(enabled), "status": normalizeStatus(enabled, input.Status, "ONLINE", "DISABLED"), "remark": strings.TrimSpace(input.Remark),
		"created_by": actorID, "updated_by": actorID, "created_at": now(), "updated_at": now(),
	}
}

func siteImmediateInputRow(input SiteInput, enabled bool, actorID uint64) map[string]any {
	values := siteInputRow(input, enabled, actorID)
	for _, key := range []string{"tenant_id", "created_by", "created_at", "config_version", "is_initialized"} {
		delete(values, key)
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["config_status"] = "PUBLISHED"
	values["config_published_at"] = now()
	values["config_published_by"] = actorID
	values["config_draft_updated_at"] = now()
	values["config_draft_updated_by"] = actorID
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
	if db.Migrator().HasColumn("game_saas_sites", "android_apk_object_key") {
		return
	}
	for _, key := range []string{"android_apk_object_key", "android_apk_name", "android_apk_size", "android_apk_uploaded_at"} {
		delete(values, key)
	}
}

func siteDraftInputRow(input SiteInput, actorID uint64) map[string]any {
	return map[string]any{
		"site_name_draft":          strings.TrimSpace(input.SiteName),
		"country_code_draft":       strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		"currency_code_draft":      strings.ToUpper(strings.TrimSpace(input.CurrencyCode)),
		"default_language_draft":   strings.TrimSpace(input.DefaultLanguage),
		"timezone_draft":           strings.TrimSpace(input.Timezone),
		"brand_config_draft":       marshal(input.BrandConfig),
		"theme_config_draft":       marshal(input.ThemeConfig),
		"feature_config_draft":     marshal(input.FeatureConfig),
		"maintenance_config_draft": marshal(input.MaintenanceConfig),
		"seo_config_draft":         marshal(input.SEOConfig),
		"config_status":            "DRAFT",
		"config_draft_updated_at":  now(),
		"config_draft_updated_by":  actorID,
		"updated_by":               actorID,
		"version":                  gorm.Expr("version+1"),
		"updated_at":               now(),
	}
}

func preserveSiteDraftDefaults(input SiteInput, before map[string]any) SiteInput {
	if input.BrandConfig == nil {
		input.BrandConfig = parseJSONMap(firstDraft(mapString(before, "brand_config_draft"), mapString(before, "brand_config")))
	}
	if input.ThemeConfig == nil {
		input.ThemeConfig = parseJSONMap(firstDraft(mapString(before, "theme_config_draft"), mapString(before, "theme_config")))
	}
	if input.FeatureConfig == nil {
		input.FeatureConfig = parseJSONMap(firstDraft(mapString(before, "feature_config_draft"), mapString(before, "feature_config")))
	}
	if input.MaintenanceConfig == nil {
		input.MaintenanceConfig = parseJSONMap(firstDraft(mapString(before, "maintenance_config_draft"), mapString(before, "maintenance_config")))
	}
	if input.SEOConfig == nil {
		input.SEOConfig = parseJSONMap(firstDraft(mapString(before, "seo_config_draft"), mapString(before, "seo_config")))
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

func defaultString(value, fallback string) string {
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
