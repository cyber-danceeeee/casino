package systemadmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"casino/internal/eventbus"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/onboardingemail"
	"casino/internal/tenancy"
)

type TenantProvisionInput struct {
	TenantCode          string         `json:"tenantCode"`
	TenantName          string         `json:"tenantName"`
	CompanyName         string         `json:"companyName"`
	ContactName         string         `json:"contactName"`
	ContactEmail        string         `json:"contactEmail"`
	ContactPhone        string         `json:"contactPhone"`
	PlanCode            string         `json:"planCode"`
	PackageCode         string         `json:"packageCode"`
	ExpiresAt           *time.Time     `json:"expiresAt"`
	CountryCode         string         `json:"countryCode"`
	CurrencyCode        string         `json:"currencyCode"`
	Language            string         `json:"language"`
	Timezone            string         `json:"timezone"`
	SiteCode            string         `json:"siteCode"`
	SiteName            string         `json:"siteName"`
	PrimaryDomain       string         `json:"primaryDomain"`
	H5Domain            string         `json:"h5Domain"`
	AdminDomain         string         `json:"adminDomain"`
	AndroidAPKObjectKey string         `json:"androidAPKObjectKey"`
	AndroidAPKName      string         `json:"androidAPKName"`
	AndroidAPKSize      int64          `json:"androidAPKSize"`
	AppName             string         `json:"appName"`
	Logo                string         `json:"logo"`
	DarkLogo            string         `json:"darkLogo"`
	Favicon             string         `json:"favicon"`
	LoadingLogo         string         `json:"loadingLogo"`
	Splash              string         `json:"splash"`
	CustomerService     string         `json:"customerService"`
	Telegram            string         `json:"telegram"`
	Email               string         `json:"email"`
	Facebook            string         `json:"facebook"`
	X                   string         `json:"x"`
	Copyright           string         `json:"copyright"`
	SEOTitle            string         `json:"seoTitle"`
	SEODescription      string         `json:"seoDescription"`
	PrimaryColor        string         `json:"primaryColor"`
	SecondaryColor      string         `json:"secondaryColor"`
	AccentColor         string         `json:"accentColor"`
	BackgroundColor     string         `json:"backgroundColor"`
	CardColor           string         `json:"cardColor"`
	ButtonColor         string         `json:"buttonColor"`
	TextColor           string         `json:"textColor"`
	BrandConfig         map[string]any `json:"brandConfig"`
	ThemeConfig         map[string]any `json:"themeConfig"`
	FeatureConfig       map[string]any `json:"featureConfig"`
	MaintenanceConfig   map[string]any `json:"maintenanceConfig"`
	SEOConfig           map[string]any `json:"seoConfig"`
	OwnerUsername       string         `json:"ownerUsername"`
	OwnerPassword       string         `json:"ownerPassword"`
	OwnerEmail          string         `json:"ownerEmail"`
	ProviderCodes       []string       `json:"providerCodes"`
	Reason              string         `json:"reason"`
}

type TenantProvisionResult struct {
	BrandName               string `json:"brandName"`
	TenantID                uint64 `json:"tenantID"`
	SiteID                  uint64 `json:"siteID"`
	AdminID                 uint64 `json:"adminID"`
	RoleID                  uint64 `json:"roleID"`
	AdminUsername           string `json:"adminUsername"`
	TemporaryPassword       string `json:"temporaryPassword"`
	PrimaryDomain           string `json:"primaryDomain"`
	H5Domain                string `json:"h5Domain"`
	AdminDomain             string `json:"adminDomain"`
	H5URL                   string `json:"h5Url"`
	AdminURL                string `json:"adminUrl"`
	BootstrapURL            string `json:"bootstrapUrl"`
	LobbyURL                string `json:"lobbyUrl"`
	DomainStatus            string `json:"domainStatus"`
	InitialConfigVersion    uint64 `json:"initialConfigVersion"`
	LobbySections           int    `json:"lobbySections"`
	Providers               int    `json:"providers"`
	Games                   int    `json:"games"`
	DeliveryEmail           string `json:"deliveryEmail"`
	EmailNotificationStatus string `json:"emailNotificationStatus"`
}

type provisionCountryDefaults struct {
	CountryCode     string
	CurrencyCode    string
	DefaultLanguage string
	Timezone        string
}

type provisionPackageLimits struct {
	GameLimit     int `gorm:"column:game_limit"`
	ProviderLimit int `gorm:"column:provider_limit"`
}

func (s *Service) ProvisionTenant(input TenantProvisionInput, actor Actor) (TenantProvisionResult, error) {
	if !actor.Platform {
		return TenantProvisionResult{}, NewError("PERMISSION_DENIED", "permission denied")
	}
	normalized, err := s.normalizeProvisionInput(input)
	if err != nil {
		return TenantProvisionResult{}, err
	}
	var result TenantProvisionResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		packageLimits, err := provisionLimits(tx, normalized.PackageCode)
		if err != nil {
			return err
		}
		defaults, err := s.provisionCountryDefaults(tx, normalized.CountryCode)
		if err != nil {
			return err
		}
		normalized.CountryCode = defaults.CountryCode
		normalized.CurrencyCode = defaults.CurrencyCode
		normalized.Language = defaults.DefaultLanguage
		normalized.Timezone = defaults.Timezone
		tenantInput := provisionTenantInput(normalized, defaults)
		if err := validateTenantInput(tenantInput, true); err != nil {
			return err
		}
		tenantRow := provisionTenantRow(tenantInput, actor.ID)
		if err := ensureTenantCodeAvailable(tx, tenantRow["tenant_code"]); err != nil {
			return err
		}
		if err := tx.Table("game_saas_tenants").Create(tenantRow).Error; err != nil {
			return duplicateError(err)
		}
		var tenantID uint64
		if err := tx.Table("game_saas_tenants").Select("id").Where("tenant_code=? AND deleted_at IS NULL", tenantInput.TenantCode).Take(&tenantID).Error; err != nil {
			return err
		}

		siteInput := provisionSiteInput(normalized, tenantID, tenantInput, defaults)
		if err := validateSiteInput(siteInput, true); err != nil {
			return err
		}
		siteRow := siteInputRow(siteInput, true, actor.ID)
		stripUnavailableSiteColumns(tx, siteRow)
		siteRow["site_code"] = normalizeSaaSCode(siteInput.SiteCode)
		if err := ensureSiteCodeAvailable(tx, siteRow["site_code"]); err != nil {
			return err
		}
		if err := tx.Table("game_saas_sites").Create(siteRow).Error; err != nil {
			return duplicateError(err)
		}
		var siteID uint64
		if err := tx.Table("game_saas_sites").Select("id").Where("tenant_id=? AND site_code=? AND deleted_at IS NULL", tenantID, siteRow["site_code"]).Take(&siteID).Error; err != nil {
			return err
		}

		if err := provisionDomains(tx, tenantID, siteID, normalized, actor.ID); err != nil {
			return err
		}
		lobbyCount, err := provisionLobbySections(tx, tenantID, siteID)
		if err != nil {
			return err
		}
		if err := provisionTenantOperationDefaults(tx, tenantID, siteID, defaults.CurrencyCode, defaults.CountryCode, defaults.DefaultLanguage, actor.ID); err != nil {
			return err
		}
		providerCount, err := provisionProviders(tx, tenantID, siteID, normalized.ProviderCodes, packageLimits.ProviderLimit)
		if err != nil {
			return err
		}
		gameCount, err := provisionGames(tx, tenantID, siteID, normalized.ProviderCodes, packageLimits.GameLimit, actor.ID)
		if err != nil {
			return err
		}
		adminID, roleID, err := createProvisionTenantAdmin(tx, tenantID, siteID, tenantInput, actor)
		if err != nil {
			return err
		}
		if err := audit(tx, actor, "system:tenant:create", "system.tenant.provision", "tenant", tenantID, nil, map[string]any{
			"tenantID": tenantID, "siteID": siteID, "adminID": adminID, "primaryDomain": siteInput.PrimaryDomain,
		}, normalized.Reason); err != nil {
			return err
		}
		result = TenantProvisionResult{
			BrandName:               siteInput.SiteName,
			TenantID:                tenantID,
			SiteID:                  siteID,
			AdminID:                 adminID,
			RoleID:                  roleID,
			AdminUsername:           normalized.OwnerUsername,
			TemporaryPassword:       normalized.OwnerPassword,
			PrimaryDomain:           siteInput.PrimaryDomain,
			H5Domain:                siteInput.H5Domain,
			AdminDomain:             siteInput.AdminDomain,
			H5URL:                   provisionURL(siteInput.H5Domain, ""),
			AdminURL:                provisionURL(siteInput.AdminDomain, "/tenant/login"),
			BootstrapURL:            provisionURL(siteInput.H5Domain, "/api/v1/bootstrap"),
			LobbyURL:                provisionURL(siteInput.H5Domain, "/api/v1/lobby"),
			DomainStatus:            "PENDING",
			InitialConfigVersion:    1,
			LobbySections:           lobbyCount,
			Providers:               providerCount,
			Games:                   gameCount,
			DeliveryEmail:           normalized.OwnerEmail,
			EmailNotificationStatus: "PENDING",
		}
		if err := tx.Table("game_saas_brand_deliveries").Create(map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "admin_id": adminID,
			"admin_username": normalized.OwnerUsername, "primary_domain": siteInput.PrimaryDomain,
			"h5_url": result.H5URL, "admin_url": result.AdminURL, "domain_status": result.DomainStatus,
			"initial_config_version": result.InitialConfigVersion, "status": "DELIVERED",
			"delivered_by": actor.ID, "delivered_at": now(), "created_at": now(), "updated_at": now(),
		}).Error; err != nil {
			return err
		}
		if err := queueTenantOnboardingEmail(tx, normalized, result, actor); err != nil {
			return err
		}
		return nil
	})
	return result, err
}

func provisionLimits(tx *gorm.DB, packageCode string) (provisionPackageLimits, error) {
	limits := provisionPackageLimits{GameLimit: -1, ProviderLimit: -1}
	if !tx.Migrator().HasTable("game_saas_tenant_packages") {
		return limits, nil
	}
	err := tx.Table("game_saas_tenant_packages").
		Select("game_limit,provider_limit").
		Where("package_code=? AND deleted_at IS NULL", strings.ToUpper(strings.TrimSpace(packageCode))).
		Take(&limits).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return provisionPackageLimits{}, NewError("PACKAGE_NOT_FOUND", "tenant package is not configured")
	}
	return limits, err
}

func (s *Service) normalizeProvisionInput(input TenantProvisionInput) (TenantProvisionInput, error) {
	input.TenantCode = normalizeSaaSCode(input.TenantCode)
	input.SiteCode = normalizeSaaSCode(defaultString(input.SiteCode, input.TenantCode))
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	input.CurrencyCode = strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	input.PlanCode = strings.ToUpper(strings.TrimSpace(defaultString(input.PlanCode, "STANDARD")))
	input.PackageCode = strings.ToUpper(strings.TrimSpace(defaultString(input.PackageCode, input.PlanCode)))
	input.PrimaryDomain = normalizeDomain(input.PrimaryDomain)
	input.H5Domain = normalizeDomain(defaultString(input.H5Domain, input.PrimaryDomain))
	input.AdminDomain = normalizeDomain(input.AdminDomain)
	input.AndroidAPKObjectKey = strings.TrimSpace(input.AndroidAPKObjectKey)
	input.AndroidAPKName = strings.TrimSpace(input.AndroidAPKName)
	input.Language = strings.TrimSpace(input.Language)
	input.Timezone = strings.TrimSpace(input.Timezone)
	input.TenantName = strings.TrimSpace(input.TenantName)
	input.SiteName = strings.TrimSpace(defaultString(input.SiteName, input.TenantName))
	input.OwnerUsername = strings.TrimSpace(input.OwnerUsername)
	input.OwnerEmail = strings.ToLower(strings.TrimSpace(input.OwnerEmail))
	input.Reason = strings.TrimSpace(defaultString(input.Reason, "tenant provisioning"))
	input.ProviderCodes = normalizeProvisionProviderCodes(input.ProviderCodes)
	if input.AndroidAPKObjectKey != "" && !strings.HasPrefix(input.AndroidAPKObjectKey, "admin/uploads/apk/") {
		return input, NewError("INVALID_APK_REFERENCE", "invalid Android APK reference")
	}
	if input.PrimaryDomain == "" {
		return input, NewError("DOMAIN_REQUIRED", "primary domain is required")
	}
	if input.OwnerUsername == "" {
		return input, NewError("OWNER_REQUIRED", "tenant admin username is required")
	}
	if input.OwnerEmail == "" {
		return input, NewError("OWNER_EMAIL_REQUIRED", "tenant admin email is required")
	}
	address, err := mail.ParseAddress(input.OwnerEmail)
	if err != nil || !strings.EqualFold(address.Address, input.OwnerEmail) {
		return input, NewError("INVALID_OWNER_EMAIL", "tenant admin email is invalid")
	}
	if input.OwnerPassword == "" {
		input.OwnerPassword, err = randomPassword()
		if err != nil {
			return input, err
		}
	}
	if err := validatePasswordStrength(input.OwnerPassword); err != nil {
		return input, err
	}
	input.ContactEmail = strings.TrimSpace(defaultString(input.ContactEmail, input.OwnerEmail))
	return input, nil
}

func (s *Service) provisionCountryDefaults(tx *gorm.DB, countryCode string) (provisionCountryDefaults, error) {
	var row struct {
		CountryCode     string `gorm:"column:country_code"`
		CurrencyCode    string `gorm:"column:currency_code"`
		DefaultLanguage string `gorm:"column:default_language"`
		Timezone        string `gorm:"column:timezone"`
	}
	query := tx.Table("game_country_configs").
		Select("country_code,currency_code,default_language,timezone").
		Where("deleted_at IS NULL AND status=1")
	if countryCode != "" {
		if err := query.Where("country_code=?", countryCode).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return provisionCountryDefaults{}, NewError("COUNTRY_NOT_FOUND", "country is not configured or enabled")
			}
			return provisionCountryDefaults{}, err
		}
		return validateProvisionCountryDefaults(row)
	}
	err := tx.Table("game_country_configs").
		Select("country_code,currency_code,default_language,timezone").
		Where("deleted_at IS NULL AND status=1").
		Order("is_default DESC,id ASC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return provisionCountryDefaults{}, NewError("COUNTRY_NOT_FOUND", "no enabled country configuration is available")
		}
		return provisionCountryDefaults{}, err
	}
	return validateProvisionCountryDefaults(row)
}

func validateProvisionCountryDefaults(row struct {
	CountryCode     string `gorm:"column:country_code"`
	CurrencyCode    string `gorm:"column:currency_code"`
	DefaultLanguage string `gorm:"column:default_language"`
	Timezone        string `gorm:"column:timezone"`
}) (provisionCountryDefaults, error) {
	defaults := provisionCountryDefaults{
		CountryCode: strings.ToUpper(strings.TrimSpace(row.CountryCode)), CurrencyCode: strings.ToUpper(strings.TrimSpace(row.CurrencyCode)),
		DefaultLanguage: strings.TrimSpace(row.DefaultLanguage), Timezone: strings.TrimSpace(row.Timezone),
	}
	if defaults.CountryCode == "" || defaults.CurrencyCode == "" || defaults.DefaultLanguage == "" || defaults.Timezone == "" {
		return provisionCountryDefaults{}, NewError("COUNTRY_CONFIG_INCOMPLETE", "country currency, language, and timezone must be configured")
	}
	return defaults, nil
}

func provisionTenantInput(input TenantProvisionInput, defaults provisionCountryDefaults) TenantInput {
	enabled := true
	return TenantInput{
		TenantCode:       input.TenantCode,
		TenantName:       input.TenantName,
		CompanyName:      input.CompanyName,
		ContactName:      input.ContactName,
		ContactEmail:     input.ContactEmail,
		ContactPhone:     input.ContactPhone,
		PlanCode:         input.PlanCode,
		PackageCode:      input.PackageCode,
		CountryCode:      defaults.CountryCode,
		DefaultCurrency:  defaults.CurrencyCode,
		DefaultLanguage:  defaults.DefaultLanguage,
		Timezone:         defaults.Timezone,
		ExpiresAt:        input.ExpiresAt,
		Enabled:          &enabled,
		Status:           "ACTIVE",
		OwnerUsername:    input.OwnerUsername,
		OwnerPassword:    input.OwnerPassword,
		OwnerDisplayName: input.OwnerUsername,
		OwnerEmail:       input.OwnerEmail,
		OwnerMobile:      "",
		Reason:           input.Reason,
	}
}

func queueTenantOnboardingEmail(tx *gorm.DB, input TenantProvisionInput, result TenantProvisionResult, actor Actor) error {
	encryptedPayload, err := onboardingemail.EncryptPayload(onboardingemail.ProvisionedPayload{
		BrandName: result.BrandName, TenantCode: input.TenantCode, CountryCode: input.CountryCode,
		CurrencyCode: input.CurrencyCode, AdminUsername: result.AdminUsername, TemporaryPassword: result.TemporaryPassword,
		AdminURL: result.AdminURL, H5URL: result.H5URL, Language: input.Language,
	})
	if err != nil {
		return err
	}
	createdAt := now()
	job := onboardingemail.Job{
		TenantID: result.TenantID, SiteID: result.SiteID, RecipientEmail: input.OwnerEmail,
		TemplateCode: onboardingemail.TemplateTenantProvisioned, EncryptedPayload: encryptedPayload, Status: "PENDING",
		IdempotencyKey: fmt.Sprintf("tenant-provision-email:%d:%d", result.TenantID, result.AdminID), CreatedAt: createdAt, UpdatedAt: createdAt,
	}
	if err := tx.Create(&job).Error; err != nil {
		return duplicateError(err)
	}
	record, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
		Scope:    tenancy.Scope{TenantID: result.TenantID, SiteID: result.SiteID, Platform: true},
		TenantID: result.TenantID, SiteID: result.SiteID, EventType: eventbus.EventTenantProvisioned,
		AggregateType: "tenant", AggregateID: fmt.Sprint(result.TenantID), Source: "systemadmin",
		IdempotencyKey: fmt.Sprintf("tenant.provisioned:%d", result.TenantID), TraceID: actor.CorrelationID,
		CorrelationID: actor.CorrelationID, OccurredAt: createdAt,
		TargetConsumers: []string{eventbus.ConsumerTenantOnboardingEmail},
		Payload: map[string]any{
			"emailJobID": job.ID, "adminID": result.AdminID,
			onboardingemail.EventEncryptedPayloadKey: encryptedPayload,
		},
		Metadata: map[string]any{"operatorID": actor.ID},
	})
	if err != nil {
		return err
	}
	return tx.Model(&onboardingemail.Job{}).Where("id=?", job.ID).Update("event_id", record.EventID).Error
}

func provisionSiteInput(input TenantProvisionInput, tenantID uint64, tenant TenantInput, defaults provisionCountryDefaults) SiteInput {
	brandConfig := mergeProvisionMaps(defaultBrandConfig(input, tenant), input.BrandConfig)
	themeConfig := mergeProvisionMaps(defaultThemeConfig(input), input.ThemeConfig)
	featureConfig := mergeProvisionMaps(defaultFeatureConfig(), input.FeatureConfig)
	seoConfig := mergeProvisionMaps(defaultSEOConfig(input, tenant), input.SEOConfig)
	maintenanceConfig := mergeProvisionMaps(map[string]any{"enabled": false, "message": ""}, input.MaintenanceConfig)
	return SiteInput{
		TenantID:            tenantID,
		SiteCode:            input.SiteCode,
		SiteName:            input.SiteName,
		PrimaryDomain:       input.PrimaryDomain,
		H5Domain:            input.H5Domain,
		AdminDomain:         input.AdminDomain,
		AndroidAPKObjectKey: input.AndroidAPKObjectKey,
		AndroidAPKName:      input.AndroidAPKName,
		AndroidAPKSize:      input.AndroidAPKSize,
		CountryCode:         defaults.CountryCode,
		CurrencyCode:        defaults.CurrencyCode,
		DefaultLanguage:     defaults.DefaultLanguage,
		Timezone:            defaults.Timezone,
		BrandConfig:         brandConfig,
		ThemeConfig:         themeConfig,
		FeatureConfig:       featureConfig,
		MaintenanceConfig:   maintenanceConfig,
		SEOConfig:           seoConfig,
		Status:              "ONLINE",
		Remark:              "Provisioned tenant site",
		Reason:              input.Reason,
	}
}

func provisionTenantRow(input TenantInput, actorID uint64) map[string]any {
	return map[string]any{
		"tenant_code":         normalizeSaaSCode(input.TenantCode),
		"tenant_name":         strings.TrimSpace(input.TenantName),
		"company_name":        strings.TrimSpace(input.CompanyName),
		"contact_name":        strings.TrimSpace(input.ContactName),
		"contact_email":       strings.TrimSpace(input.ContactEmail),
		"contact_phone":       strings.TrimSpace(input.ContactPhone),
		"plan_code":           strings.ToUpper(strings.TrimSpace(defaultString(input.PlanCode, "STANDARD"))),
		"country_code":        strings.ToUpper(strings.TrimSpace(input.CountryCode)),
		"package_code":        strings.ToUpper(strings.TrimSpace(defaultString(input.PackageCode, defaultString(input.PlanCode, "STANDARD")))),
		"default_currency":    strings.ToUpper(strings.TrimSpace(input.DefaultCurrency)),
		"default_language":    strings.TrimSpace(input.DefaultLanguage),
		"timezone":            strings.TrimSpace(input.Timezone),
		"expires_at":          input.ExpiresAt,
		"enabled":             1,
		"status":              "ACTIVE",
		"maintenance_enabled": 0,
		"maintenance_message": "",
		"remark":              strings.TrimSpace(input.Remark),
		"created_by":          actorID,
		"updated_by":          actorID,
		"created_at":          now(),
		"updated_at":          now(),
	}
}

func provisionDomains(tx *gorm.DB, tenantID, siteID uint64, input TenantProvisionInput, actorID uint64) error {
	seen := map[string]bool{}
	domains := []struct {
		domain     string
		domainType string
		primary    bool
	}{
		{domain: input.PrimaryDomain, domainType: "PRIMARY", primary: true},
		{domain: input.H5Domain, domainType: "H5", primary: true},
		{domain: input.AdminDomain, domainType: "ADMIN", primary: false},
	}
	for _, item := range domains {
		if item.domain == "" || seen[item.domain] {
			continue
		}
		seen[item.domain] = true
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "domain": item.domain, "domain_key": item.domain,
			"domain_type": item.domainType, "is_primary": boolValue(item.primary), "enabled": 1, "status": "PENDING",
			"created_by": actorID, "updated_by": actorID, "created_at": now(), "updated_at": now(),
		}
		if err := tx.Table("game_saas_domains").Create(row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

func provisionLobbySections(tx *gorm.DB, tenantID, siteID uint64) (int, error) {
	var rows []model.GameLobbySections
	if err := tx.Model(&model.GameLobbySections{}).
		Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").
		Order("sort_order ASC,id ASC").
		Find(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		rows = defaultProvisionLobbySections()
	}
	for _, src := range rows {
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "section_code": src.SectionCode, "section_type": src.SectionType,
			"title_zh": src.TitleZh, "title_en": src.TitleEn, "display_style": src.DisplayStyle, "source_type": src.SourceType,
			"source_ref": src.SourceRef, "max_items": src.MaxItems, "sort_order": src.SortOrder, "enabled": boolValue(src.Enabled),
			"supported_countries": src.SupportedCountries, "supported_currencies": src.SupportedCurrencies, "config_json": defaultString(src.ConfigJSON, "{}"),
			"created_at": now(), "updated_at": now(),
		}
		if err := tx.Table("game_lobby_sections").Create(row).Error; err != nil {
			return 0, duplicateError(err)
		}
	}
	return len(rows), nil
}

func provisionH5NavigationEntries(tx *gorm.DB, tenantID, siteID uint64) error {
	return tx.Exec(`
		INSERT IGNORE INTO game_h5_navigation_items
		(tenant_id,site_id,platform_source_id,menu_code,label_zh,label_en,labels_json,display_position,action_type,action_target,
		 page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,
		 country_codes,language_codes,enabled,publish_status,config_version,created_at,updated_at,published_at)
		SELECT ?,?,0,menu_code,label_zh,label_en,labels_json,'BOTTOM_NAV',action_type,action_target,
		       page_route,provider_code,game_id,provider_game_code,game_name,sort_order,'NONE','ALL',
		       '[]','[]',enabled,'PUBLISHED',1,NOW(3),NOW(3),NOW(3)
		FROM game_h5_navigation_items
		WHERE tenant_id=0 AND site_id=0 AND deleted_at IS NULL`, tenantID, siteID).Error
}

func provisionTenantOperationDefaults(tx *gorm.DB, tenantID, siteID uint64, currencyCode, countryCode, languageCode string, operatorID uint64) error {
	if err := provisionGameCategories(tx, tenantID, siteID); err != nil {
		return err
	}
	if err := provisionH5NavigationEntries(tx, tenantID, siteID); err != nil {
		return err
	}
	if err := provisionDepositAmountOptions(tx, tenantID, siteID, currencyCode); err != nil {
		return err
	}
	if err := provisionSiteBanners(tx, tenantID, siteID, currencyCode, countryCode, languageCode, operatorID); err != nil {
		return err
	}
	if err := provisionSocialLinks(tx, tenantID, siteID); err != nil {
		return err
	}
	if err := provisionMusic(tx, tenantID, siteID); err != nil {
		return err
	}
	if err := provisionVIPLevels(tx, tenantID, siteID); err != nil {
		return err
	}
	if err := provisionDepositVouchers(tx, tenantID, siteID, currencyCode); err != nil {
		return err
	}
	return provisionPromotions(tx, tenantID, siteID, currencyCode, countryCode, languageCode)
}

// provisionGameCategories gives every site its own editable copy of the
// platform catalog. Game relations are created by provisionGames after the
// tenant providers and game rows have been provisioned.
func provisionGameCategories(tx *gorm.DB, tenantID, siteID uint64) error {
	if !tx.Migrator().HasTable("game_game_categories") {
		return nil
	}
	var sources []model.GameGameCategories
	if err := tx.Model(&model.GameGameCategories{}).
		Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").
		Order("sort ASC,id ASC").Find(&sources).Error; err != nil {
		return err
	}
	for _, source := range sources {
		var count int64
		if err := tx.Table("game_game_categories").
			Where("tenant_id=? AND site_id=? AND category_key=?", tenantID, siteID, source.CategoryKey).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "category_key": source.CategoryKey,
			"label": source.Label, "name_zh": source.NameZh, "name_en": source.NameEn,
			"short_label": source.ShortLabel, "icon_url": source.IconURL, "banner_url": source.BannerURL,
			"sort": source.Sort, "status": source.Status, "created_at": now(), "updated_at": now(),
		}
		if err := tx.Table("game_game_categories").Create(row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

func provisionDepositAmountOptions(tx *gorm.DB, tenantID, siteID uint64, currencyCode string) error {
	if !tx.Migrator().HasTable("game_deposit_amount_options") {
		return nil
	}
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	if currencyCode == "" {
		currencyCode = "PKR"
	}
	rows := []model.GameDepositAmountOptions{}
	hasCurrencyColumn := tx.Migrator().HasColumn("game_deposit_amount_options", "currency_code")
	selectColumns := "id,tenant_id,site_id,option_code,amount,bonus_reward,cash_reward,sort,status,created_at,updated_at,deleted_at"
	if hasCurrencyColumn {
		selectColumns = "id,tenant_id,site_id,option_code,currency_code,amount,bonus_reward,cash_reward,sort,status,created_at,updated_at,deleted_at"
	}
	query := tx.Table("game_deposit_amount_options").Select(selectColumns).Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL")
	if hasCurrencyColumn {
		query = query.Where("currency_code='' OR currency_code=? OR currency_code='PKR'", currencyCode)
	}
	if err := query.Order("sort ASC,id ASC").Find(&rows).Error; err != nil {
		return err
	}
	for _, src := range rows {
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "option_code": src.OptionCode, "amount": src.Amount,
			"bonus_reward": src.BonusReward, "cash_reward": src.CashReward, "sort": src.Sort, "status": src.Status,
			"created_at": now(), "updated_at": now(),
		}
		if hasCurrencyColumn {
			row["currency_code"] = currencyCode
		}
		if err := tx.Table("game_deposit_amount_options").Create(row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

type provisionBannerSource struct {
	ID             uint64
	BannerCode     string
	Title          string
	TitleZh        string
	TitleEn        string
	ImageURL       string
	MobileImageURL string
	TargetType     string
	TargetValue    string
	SortOrder      int
	Status         int
	Enabled        bool
	StartAt        *time.Time
	EndAt          *time.Time
}

func provisionSiteBanners(tx *gorm.DB, tenantID, siteID uint64, currencyCode, countryCode, languageCode string, operatorID uint64) error {
	if !tx.Migrator().HasTable("game_site_banners") || !tx.Migrator().HasTable("game_banners") {
		return nil
	}
	var rows []provisionBannerSource
	if err := tx.Table("game_banners").
		Select("id,banner_code,title,title_zh,title_en,image_url,mobile_image_url,target_type,target_value,sort_order,status,enabled,start_at,end_at").
		Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").
		Order("sort_order ASC,id ASC").Find(&rows).Error; err != nil {
		return err
	}
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	languageCode = strings.TrimSpace(languageCode)
	for _, src := range rows {
		name := defaultString(src.BannerCode, firstNonEmpty(src.TitleEn, src.TitleZh, src.Title, fmt.Sprintf("banner-%d", src.ID)))
		var count int64
		if err := tx.Table("game_site_banners").Where("tenant_id=? AND site_id=? AND name=? AND deleted_at IS NULL", tenantID, siteID, name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		actionType, actionTarget := provisionBannerAction(src.TargetType, src.TargetValue)
		enabled := src.Enabled
		if !enabled && src.Status == 1 {
			enabled = true
		}
		payload := map[string]any{
			"name": name, "title": firstNonEmpty(src.TitleEn, src.TitleZh, src.Title, name), "subtitle": "",
			"desktopAssetID": 0, "mobileAssetID": 0, "desktopImageUrl": src.ImageURL, "mobileImageUrl": firstNonEmpty(src.MobileImageURL, src.ImageURL),
			"actionType": actionType, "actionTarget": actionTarget, "sortOrder": src.SortOrder,
			"languageCode": firstNonEmpty(languageCode, "ALL"), "countryCode": firstNonEmpty(countryCode, "ALL"), "currencyCode": firstNonEmpty(currencyCode, "ALL"),
			"visibilityScope": "ALL", "startAt": src.StartAt, "endAt": src.EndAt, "enabled": enabled,
		}
		raw, _ := json.Marshal(payload)
		nowValue := now()
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "name": name, "title": payload["title"], "subtitle": "",
			"desktop_asset_id": 0, "mobile_asset_id": 0, "desktop_asset_id_draft": 0, "mobile_asset_id_draft": 0,
			"action_type": actionType, "action_target": actionTarget, "sort_order": src.SortOrder,
			"language_code": firstNonEmpty(languageCode, "ALL"), "country_code": firstNonEmpty(countryCode, "ALL"), "currency_code": firstNonEmpty(currencyCode, "ALL"),
			"visibility_scope": "ALL", "start_at": src.StartAt, "end_at": src.EndAt, "enabled": enabled,
			"draft_payload": string(raw), "published_payload": string(raw), "publish_status": "PUBLISHED", "config_version": 1,
			"created_by": operatorID, "updated_by": operatorID, "draft_updated_at": nowValue, "published_at": nowValue, "published_by": operatorID,
			"created_at": nowValue, "updated_at": nowValue,
		}
		if err := tx.Table("game_site_banners").Create(row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

func provisionBannerAction(kind, target string) (string, string) {
	target = strings.TrimSpace(target)
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case "EXTERNAL_URL":
		if target != "" {
			return "OPEN_EXTERNAL_URL", target
		}
	case "INTERNAL_ROUTE":
		if target != "" {
			return "OPEN_INTERNAL_PAGE", target
		}
	case "GAME":
		if target != "" {
			return "OPEN_GAME", target
		}
	case "PROMOTION":
		if target != "" {
			return "OPEN_PROMOTION", target
		}
	}
	return "NONE", ""
}

func provisionSocialLinks(tx *gorm.DB, tenantID, siteID uint64) error {
	if !tx.Migrator().HasTable("game_social_links") {
		return nil
	}
	var rows []model.GameSocialLinks
	if err := tx.Model(&model.GameSocialLinks{}).Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").Order("sort ASC,id ASC").Find(&rows).Error; err != nil {
		return err
	}
	for _, src := range rows {
		var count int64
		if err := tx.Table("game_social_links").Where("tenant_id=? AND site_id=? AND platform_code=? AND deleted_at IS NULL", tenantID, siteID, src.PlatformCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		row := src
		row.ID = 0
		row.TenantID = tenantID
		row.SiteID = siteID
		row.CreatedAt = now()
		row.UpdatedAt = now()
		row.DeletedAt = gorm.DeletedAt{}
		if err := tx.Create(&row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

func provisionMusic(tx *gorm.DB, tenantID, siteID uint64) error {
	if !tx.Migrator().HasTable("game_operation_music") {
		return nil
	}
	var rows []model.GameOperationMusic
	if err := tx.Model(&model.GameOperationMusic{}).Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").Order("sort_order ASC,id ASC").Find(&rows).Error; err != nil {
		return err
	}
	for _, src := range rows {
		var count int64
		if err := tx.Table("game_operation_music").Where("tenant_id=? AND site_id=? AND code=? AND deleted_at IS NULL", tenantID, siteID, src.Code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		row := src
		row.ID = 0
		row.TenantID = tenantID
		row.SiteID = siteID
		row.CreatedAt = now()
		row.UpdatedAt = now()
		row.DeletedAt = gorm.DeletedAt{}
		if err := tx.Create(&row).Error; err != nil {
			return duplicateError(err)
		}
	}
	return nil
}

func provisionVIPLevels(tx *gorm.DB, tenantID, siteID uint64) error {
	if !tx.Migrator().HasTable("game_vip_levels") {
		return nil
	}
	nowValue := now()
	if err := tx.Exec(`
		INSERT INTO game_vip_levels
		(tenant_id,site_id,level_no,code,name,name_zh,name_en,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,
		 required_experience,required_deposit,required_valid_bet,withdrawal_limit,cashback_rate,birthday_reward,sort_order,status,enabled,created_at,updated_at)
		SELECT ?,?,p.level_no,p.code,p.name,p.name_zh,p.name_en,p.icon_url,p.level_up_reward,p.weekly_reward,p.monthly_reward,p.daily_withdraw_count,
		       p.required_experience,p.required_deposit,p.required_valid_bet,p.withdrawal_limit,p.cashback_rate,p.birthday_reward,p.sort_order,p.status,p.enabled,?,?
		FROM game_vip_levels p
		WHERE p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_vip_levels t
		    WHERE t.tenant_id=? AND t.site_id=? AND t.level_no=p.level_no AND t.deleted_at IS NULL
		  )`, tenantID, siteID, nowValue, nowValue, tenantID, siteID).Error; err != nil {
		return duplicateError(err)
	}
	if !tx.Migrator().HasTable("game_vip_benefits") {
		return nil
	}
	return tx.Exec(`
		INSERT INTO game_vip_benefits
		(tenant_id,site_id,vip_level_id,benefit_code,name_zh,name_en,benefit_value,value_unit,enabled,sort_order,created_at,updated_at)
		SELECT ?,?,tv.id,b.benefit_code,b.name_zh,b.name_en,b.benefit_value,b.value_unit,b.enabled,b.sort_order,?,?
		FROM game_vip_benefits b
		JOIN game_vip_levels pv ON pv.id=b.vip_level_id AND pv.tenant_id=0 AND pv.site_id=0 AND pv.deleted_at IS NULL
		JOIN game_vip_levels tv ON tv.tenant_id=? AND tv.site_id=? AND tv.level_no=pv.level_no AND tv.deleted_at IS NULL
		WHERE b.tenant_id=0 AND b.site_id=0 AND b.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_vip_benefits tb
		    WHERE tb.tenant_id=? AND tb.site_id=? AND tb.vip_level_id=tv.id AND tb.benefit_code=b.benefit_code AND tb.deleted_at IS NULL
		  )`, tenantID, siteID, nowValue, nowValue, tenantID, siteID, tenantID, siteID).Error
}

func provisionDepositVouchers(tx *gorm.DB, tenantID, siteID uint64, currencyCode string) error {
	if !tx.Migrator().HasTable("game_deposit_vouchers") {
		return nil
	}
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	if currencyCode == "" {
		currencyCode = "PKR"
	}
	nowValue := now()
	if err := tx.Exec(`
		INSERT INTO game_deposit_vouchers
		(tenant_id,site_id,voucher_code,title,title_zh,title_en,description_zh,description_en,voucher_type,icon_url,
		 min_amount,max_amount,max_reward,bonus_reward,cash_reward,sort,status,start_at,end_at,new_users_only,first_deposit_only,
		 daily_limit,weekly_limit,per_user_limit,countries,currencies,method_codes,created_at,updated_at)
		SELECT ?,?,p.voucher_code,p.title,p.title_zh,p.title_en,p.description_zh,p.description_en,p.voucher_type,p.icon_url,
		       p.min_amount,p.max_amount,p.max_reward,p.bonus_reward,p.cash_reward,p.sort,p.status,p.start_at,p.end_at,p.new_users_only,p.first_deposit_only,
		       p.daily_limit,p.weekly_limit,p.per_user_limit,p.countries,?,p.method_codes,?,?
		FROM game_deposit_vouchers p
		WHERE p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_deposit_vouchers t
		    WHERE t.tenant_id=? AND t.site_id=? AND t.voucher_code=p.voucher_code AND t.deleted_at IS NULL
		  )`, tenantID, siteID, currencyCode, nowValue, nowValue, tenantID, siteID).Error; err != nil {
		return duplicateError(err)
	}
	if !tx.Migrator().HasTable("game_deposit_voucher_rewards") {
		return nil
	}
	return tx.Exec(`
		INSERT INTO game_deposit_voucher_rewards
		(tenant_id,site_id,voucher_id,amount_option_id,amount,bonus_reward,cash_reward,reward_amount,sort,status,created_at,updated_at)
		SELECT ?,?,tv.id,COALESCE(ta.id,0),r.amount,r.bonus_reward,r.cash_reward,r.reward_amount,r.sort,r.status,?,?
		FROM game_deposit_voucher_rewards r
		JOIN game_deposit_vouchers pv ON pv.id=r.voucher_id AND pv.tenant_id=0 AND pv.site_id=0 AND pv.deleted_at IS NULL
		JOIN game_deposit_vouchers tv ON tv.tenant_id=? AND tv.site_id=? AND tv.voucher_code=pv.voucher_code AND tv.deleted_at IS NULL
		LEFT JOIN game_deposit_amount_options ta ON ta.tenant_id=? AND ta.site_id=? AND ta.amount=r.amount AND ta.deleted_at IS NULL
		WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_deposit_voucher_rewards tr
		    WHERE tr.tenant_id=? AND tr.site_id=? AND tr.voucher_id=tv.id AND tr.amount=r.amount AND tr.deleted_at IS NULL
		  )`, tenantID, siteID, nowValue, nowValue, tenantID, siteID, tenantID, siteID, tenantID, siteID).Error
}

func provisionPromotions(tx *gorm.DB, tenantID, siteID uint64, currencyCode, countryCode, languageCode string) error {
	if !tx.Migrator().HasTable("game_activity_events") {
		return nil
	}
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	languageCode = strings.TrimSpace(languageCode)
	nowValue := now()
	if err := tx.Exec(`
		INSERT INTO game_activity_events
		(tenant_id,site_id,activity_code,slug,kind,title,name_zh,name_en,subtitle,description_zh,description_en,promotion_type,
		 trigger_type,claim_mode,stacking_group,priority,countries,currencies,languages,reward_text,max_reward_text,note,cta_text,
		 action_url,hero_url,card_image_url,status,status_code,total_budget,per_user_limit,total_claim_limit,claim_count,reward_amount,
		 sort,start_at,end_at,created_at,updated_at)
		SELECT ?,?,p.activity_code,p.slug,p.kind,p.title,p.name_zh,p.name_en,p.subtitle,p.description_zh,p.description_en,p.promotion_type,
		       p.trigger_type,p.claim_mode,p.stacking_group,p.priority,COALESCE(NULLIF(?,''),p.countries),COALESCE(NULLIF(?,''),p.currencies),COALESCE(NULLIF(?,''),p.languages),
		       p.reward_text,p.max_reward_text,p.note,p.cta_text,p.action_url,p.hero_url,p.card_image_url,p.status,p.status_code,p.total_budget,
		       p.per_user_limit,p.total_claim_limit,0,0,p.sort,p.start_at,p.end_at,?,?
		FROM game_activity_events p
		WHERE p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_activity_events t
		    WHERE t.tenant_id=? AND t.site_id=? AND t.activity_code=p.activity_code AND t.deleted_at IS NULL
		  )`, tenantID, siteID, countryCode, currencyCode, languageCode, nowValue, nowValue, tenantID, siteID).Error; err != nil {
		return duplicateError(err)
	}
	if err := provisionPromotionParticipationRules(tx, tenantID, siteID, nowValue); err != nil {
		return err
	}
	if err := provisionPromotionRewardRules(tx, tenantID, siteID, currencyCode, nowValue); err != nil {
		return err
	}
	return provisionPromotionEligibleGames(tx, tenantID, siteID, nowValue)
}

func provisionPromotionParticipationRules(tx *gorm.DB, tenantID, siteID uint64, nowValue time.Time) error {
	if !tx.Migrator().HasTable("game_promotion_participation_rules") {
		return nil
	}
	return tx.Exec(`
		INSERT INTO game_promotion_participation_rules
		(tenant_id,site_id,promotion_id,rule_type,comparison_operator,numeric_value,text_value,enabled,sort_order,created_at,updated_at)
		SELECT ?,?,ta.id,r.rule_type,r.comparison_operator,r.numeric_value,r.text_value,r.enabled,r.sort_order,?,?
		FROM game_promotion_participation_rules r
		JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
		JOIN game_activity_events ta ON ta.tenant_id=? AND ta.site_id=? AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
		WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_promotion_participation_rules tr
		    WHERE tr.tenant_id=? AND tr.site_id=? AND tr.promotion_id=ta.id AND tr.rule_type=r.rule_type AND tr.sort_order=r.sort_order AND tr.deleted_at IS NULL
		  )`, tenantID, siteID, nowValue, nowValue, tenantID, siteID, tenantID, siteID).Error
}

func provisionPromotionRewardRules(tx *gorm.DB, tenantID, siteID uint64, currencyCode string, nowValue time.Time) error {
	if !tx.Migrator().HasTable("game_promotion_reward_rules") {
		return nil
	}
	return tx.Exec(`
		INSERT INTO game_promotion_reward_rules
		(tenant_id,site_id,promotion_id,reward_type,calculation_type,reward_amount,percentage_rate,currency,wagering_multiplier,
		 minimum_amount,maximum_reward,expiration_days,enabled,sort_order,created_at,updated_at)
		SELECT ?,?,ta.id,r.reward_type,r.calculation_type,r.reward_amount,r.percentage_rate,?,r.wagering_multiplier,
		       r.minimum_amount,r.maximum_reward,r.expiration_days,r.enabled,r.sort_order,?,?
		FROM game_promotion_reward_rules r
		JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
		JOIN game_activity_events ta ON ta.tenant_id=? AND ta.site_id=? AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
		WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_promotion_reward_rules tr
		    WHERE tr.tenant_id=? AND tr.site_id=? AND tr.promotion_id=ta.id AND tr.reward_type=r.reward_type AND tr.sort_order=r.sort_order AND tr.deleted_at IS NULL
		  )`, tenantID, siteID, currencyCode, nowValue, nowValue, tenantID, siteID, tenantID, siteID).Error
}

func provisionPromotionEligibleGames(tx *gorm.DB, tenantID, siteID uint64, nowValue time.Time) error {
	if !tx.Migrator().HasTable("game_promotion_eligible_games") {
		return nil
	}
	return tx.Exec(`
		INSERT INTO game_promotion_eligible_games
		(tenant_id,site_id,promotion_id,game_id,created_at,updated_at)
		SELECT ?,?,ta.id,r.game_id,?,?
		FROM game_promotion_eligible_games r
		JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
		JOIN game_activity_events ta ON ta.tenant_id=? AND ta.site_id=? AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
		WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM game_promotion_eligible_games tr
		    WHERE tr.tenant_id=? AND tr.site_id=? AND tr.promotion_id=ta.id AND tr.game_id=r.game_id AND tr.deleted_at IS NULL
		  )`, tenantID, siteID, nowValue, nowValue, tenantID, siteID, tenantID, siteID).Error
}

func provisionProviders(tx *gorm.DB, tenantID, siteID uint64, providerCodes []string, providerLimit int) (int, error) {
	var rows []model.GameGameProviders
	q := tx.Model(&model.GameGameProviders{}).Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL")
	if len(providerCodes) > 0 {
		q = q.Where("provider_code IN ?", providerCodes)
	}
	if err := q.Order("sort ASC,id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	uniqueCount := 0
	uniqueCodes := map[string]bool{}
	for _, row := range rows {
		code := strings.ToLower(strings.TrimSpace(row.ProviderCode))
		if code != "" && !uniqueCodes[code] {
			uniqueCodes[code] = true
			uniqueCount++
		}
	}
	if providerLimit >= 0 && uniqueCount > providerLimit {
		return 0, NewError("TENANT_PROVIDER_LIMIT_EXCEEDED", "tenant package provider limit exceeded")
	}
	count := 0
	seen := map[string]bool{}
	for _, src := range rows {
		code := strings.ToLower(strings.TrimSpace(src.ProviderCode))
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		row := map[string]any{
			"tenant_id": tenantID, "site_id": siteID, "provider_code": code, "name": strings.TrimSpace(src.Name),
			"logo_url": src.LogoURL, "api_base_url": src.ApiBaseURL, "environment": defaultString(src.Environment, "production"),
			"api_status": defaultString(src.APIStatus, "UNKNOWN"), "status": 1, "wallet_mode": defaultString(src.WalletMode, "SEAMLESS"),
			"callback_url": src.CallbackURL, "supported_countries": src.SupportedCountries, "supported_currencies": src.SupportedCurrencies,
			"maintenance_message": src.MaintenanceMessage, "sort": src.Sort, "created_at": now(), "updated_at": now(),
		}
		if err := tx.Table("game_game_providers").Create(row).Error; err != nil {
			return count, duplicateError(err)
		}
		count++
	}
	return count, nil
}

func provisionGames(tx *gorm.DB, tenantID, siteID uint64, providerCodes []string, gameLimit int, operatorID uint64) (int, error) {
	if !tx.Migrator().HasTable("game_games") ||
		!tx.Migrator().HasTable("game_game_categories") ||
		!tx.Migrator().HasTable("game_game_category_relations") {
		return 0, nil
	}
	providerCodes = normalizeProvisionProviderCodes(providerCodes)
	if len(providerCodes) == 0 {
		return 0, nil
	}

	var categories []model.GameGameCategories
	if err := tx.Model(&model.GameGameCategories{}).
		Where("tenant_id=? AND site_id=? AND status=1 AND deleted_at IS NULL AND category_key <> ?", tenantID, siteID, "hot").
		Order("sort ASC,id ASC").Find(&categories).Error; err != nil {
		return 0, err
	}
	if len(categories) == 0 {
		return 0, nil
	}

	var tenantProviders []model.GameGameProviders
	if err := tx.Model(&model.GameGameProviders{}).
		Where("tenant_id=? AND site_id=? AND provider_code IN ? AND status=1 AND deleted_at IS NULL", tenantID, siteID, providerCodes).
		Find(&tenantProviders).Error; err != nil {
		return 0, err
	}
	tenantProviderByCode := make(map[string]model.GameGameProviders, len(tenantProviders))
	for _, provider := range tenantProviders {
		tenantProviderByCode[strings.ToLower(strings.TrimSpace(provider.ProviderCode))] = provider
	}
	if len(tenantProviderByCode) == 0 {
		return 0, nil
	}

	var platformProviders []model.GameGameProviders
	if err := tx.Model(&model.GameGameProviders{}).
		Where("tenant_id=0 AND site_id=0 AND provider_code IN ? AND deleted_at IS NULL", providerCodes).
		Find(&platformProviders).Error; err != nil {
		return 0, err
	}
	platformProviderIDs := make([]uint64, 0, len(platformProviders))
	platformProviderCodes := make(map[uint64]string, len(platformProviders))
	for _, provider := range platformProviders {
		code := strings.ToLower(strings.TrimSpace(provider.ProviderCode))
		if _, ok := tenantProviderByCode[code]; !ok {
			continue
		}
		platformProviderIDs = append(platformProviderIDs, provider.ID)
		platformProviderCodes[provider.ID] = code
	}
	if len(platformProviderIDs) == 0 {
		return 0, nil
	}

	poolSize := 50
	if gameLimit > 0 && gameLimit < poolSize {
		poolSize = gameLimit
	}
	randomOrder := "RAND()"
	if tx.Dialector != nil && tx.Dialector.Name() == "sqlite" {
		randomOrder = "RANDOM()"
	}
	tenantGameBySource := make(map[uint64]uint64, poolSize*len(categories))
	tenantGameCategories := make(map[uint64][]uint64, poolSize*len(categories))
	tenantGameIDs := make([]uint64, 0, poolSize*len(categories))
	created := 0
	nowValue := now()
	for _, category := range categories {
		// Each category gets an independent random sample. Do not reuse one
		// preselected pool across all categories.
		var sources []model.GameGames
		if err := tx.Model(&model.GameGames{}).
			Where("tenant_id=0 AND site_id=0 AND provider_id IN ? AND status=1 AND provider_enabled=1 AND platform_enabled=1 AND deleted_at IS NULL", platformProviderIDs).
			Order(randomOrder).Limit(poolSize).Find(&sources).Error; err != nil {
			return created, err
		}
		for _, source := range sources {
			providerCode := platformProviderCodes[source.ProviderID]
			tenantProvider, ok := tenantProviderByCode[providerCode]
			if !ok {
				continue
			}

			gameID, exists := tenantGameBySource[source.ID]
			if !exists {
				if gameLimit > 0 && created >= gameLimit {
					// Keep package limits intact. Reuse an existing tenant game
					// for this category instead of creating a new row.
					for _, candidateID := range tenantGameIDs {
						if !containsUint64(tenantGameCategories[candidateID], category.ID) {
							gameID = candidateID
							break
						}
					}
					if gameID == 0 {
						continue
					}
				} else {
					row := map[string]any{
						"tenant_id": tenantID, "site_id": siteID, "game_code": source.GameCode,
						"provider_id": tenantProvider.ID, "provider_game_code": source.ProviderGameCode,
						"name": source.Name, "name_i18n": provisionJSONOrDefault(source.NameI18n), "slug": source.Slug,
						"category": source.Category, "platform_category_id": category.ID, "platform_category_code": category.CategoryKey,
						"game_type": source.GameType, "rtp": source.RTP, "volatility": source.Volatility,
						"cover_url": source.CoverURL, "local_thumbnail_url": source.LocalThumbnailURL,
						"background_url": source.BackgroundURL, "launch_image_url": source.LaunchImageURL,
						"maintenance_message": source.MaintenanceMessage, "maintenance_message_draft": source.MaintenanceMessage,
						"supports_demo": source.SupportsDemo, "supports_free_spins": source.SupportsFreeSpins,
						"provider_enabled": true, "platform_enabled": true, "platform_enabled_draft": true,
						"is_hot": false, "is_hot_draft": false, "is_featured": false, "is_featured_draft": false,
						"is_new": false, "is_new_draft": false, "supported_countries": source.SupportedCountries,
						"supported_currencies": source.SupportedCurrencies, "supported_devices": source.SupportedDevices,
						"raw_metadata": provisionJSONOrDefault(source.RawMetadata), "metadata_hash": source.MetadataHash,
						"last_synced_at": source.LastSyncedAt, "badge": source.Badge, "player_count": 0,
						"hot_score": source.HotScore, "launch_url": source.LaunchURL, "status": 1, "status_draft": 1,
						"sort": source.Sort, "sort_draft": source.Sort, "category_ids_draft": "[]",
						"config_status": "PUBLISHED", "config_version": 1, "config_published_at": nowValue,
						"config_published_by": operatorID, "draft_updated_at": nil, "draft_updated_by": 0,
						"created_at": nowValue, "updated_at": nowValue,
					}
					if err := tx.Table("game_games").Create(row).Error; err != nil {
						return created, duplicateError(err)
					}
					if err := tx.Table("game_games").Select("id").Where("tenant_id=? AND site_id=? AND game_code=? AND deleted_at IS NULL", tenantID, siteID, source.GameCode).Take(&gameID).Error; err != nil {
						return created, err
					}
					tenantGameBySource[source.ID] = gameID
					tenantGameIDs = append(tenantGameIDs, gameID)
					created++
				}
			}
			if gameID == 0 || containsUint64(tenantGameCategories[gameID], category.ID) {
				continue
			}
			if err := tx.Table("game_game_category_relations").Create(map[string]any{
				"tenant_id": tenantID, "site_id": siteID, "game_id": gameID, "category_id": category.ID,
				"sort": len(tenantGameCategories[gameID]), "created_at": nowValue, "updated_at": nowValue,
			}).Error; err != nil {
				return created, duplicateError(err)
			}
			tenantGameCategories[gameID] = append(tenantGameCategories[gameID], category.ID)
		}
	}
	for gameID, categoryIDs := range tenantGameCategories {
		categoryDraft := make([]map[string]any, 0, len(categoryIDs))
		for index, categoryID := range categoryIDs {
			categoryDraft = append(categoryDraft, map[string]any{"categoryId": categoryID, "sort": index})
		}
		encoded, err := json.Marshal(categoryDraft)
		if err != nil {
			return created, err
		}
		if err := tx.Table("game_games").Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", gameID, tenantID, siteID).
			Update("category_ids_draft", string(encoded)).Error; err != nil {
			return created, err
		}
	}
	return created, nil
}

func provisionJSONOrDefault(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return value
}

func containsUint64(values []uint64, target uint64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func createProvisionTenantAdmin(tx *gorm.DB, tenantID, siteID uint64, input TenantInput, actor Actor) (uint64, uint64, error) {
	roleID, err := ensureTenantOwnerRole(tx, tenantID)
	if err != nil {
		return 0, 0, err
	}
	if err := createTenantOwnerWithSite(tx, tenantID, siteID, input, actor); err != nil {
		return 0, 0, err
	}
	var adminID uint64
	if err := tx.Table("game_admin_accounts").
		Select("id").
		Where("tenant_id=? AND username=? AND deleted_at IS NULL", tenantID, strings.TrimSpace(input.OwnerUsername)).
		Take(&adminID).Error; err != nil {
		return 0, 0, err
	}
	return adminID, roleID, nil
}

func defaultBrandConfig(input TenantProvisionInput, tenant TenantInput) map[string]any {
	appName := strings.TrimSpace(defaultString(input.AppName, tenant.TenantName))
	return map[string]any{
		"appName": appName, "logo": strings.TrimSpace(input.Logo), "darkLogo": strings.TrimSpace(input.DarkLogo),
		"favicon": strings.TrimSpace(input.Favicon), "loadingLogo": strings.TrimSpace(input.LoadingLogo), "splash": strings.TrimSpace(input.Splash),
		"customerService": strings.TrimSpace(input.CustomerService), "telegram": strings.TrimSpace(input.Telegram), "email": strings.TrimSpace(input.Email),
		"facebook": strings.TrimSpace(input.Facebook), "x": strings.TrimSpace(input.X), "copyright": strings.TrimSpace(input.Copyright),
	}
}

func defaultThemeConfig(input TenantProvisionInput) map[string]any {
	return map[string]any{
		"primary":    defaultString(input.PrimaryColor, "#E53935"),
		"secondary":  defaultString(input.SecondaryColor, "#FFB300"),
		"accent":     defaultString(input.AccentColor, "#00C2A8"),
		"background": defaultString(input.BackgroundColor, "#070A12"),
		"card":       defaultString(input.CardColor, "#121826"),
		"button":     defaultString(input.ButtonColor, defaultString(input.PrimaryColor, "#E53935")),
		"text":       defaultString(input.TextColor, "#FFFFFF"),
	}
}

func defaultFeatureConfig() map[string]any {
	return map[string]any{"vip": true, "promotion": true, "referral": true, "deposit": true, "withdraw": true, "chat": true, "mission": true, "jackpot": true}
}

func defaultSEOConfig(input TenantProvisionInput, tenant TenantInput) map[string]any {
	title := defaultString(input.SEOTitle, tenant.TenantName)
	return map[string]any{"title": title, "description": strings.TrimSpace(input.SEODescription)}
}

func mergeProvisionMaps(base, override map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range override {
		out[key] = value
	}
	return out
}

func normalizeProvisionProviderCodes(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		code := strings.ToLower(strings.TrimSpace(value))
		if code != "" && !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	return out
}

func provisionURL(domain, path string) string {
	if domain == "" {
		return path
	}
	return "https://" + domain + path
}

func defaultProvisionLobbySections() []model.GameLobbySections {
	return []model.GameLobbySections{
		{SectionCode: "top-banner", SectionType: "banner", TitleZh: "顶部 Banner", TitleEn: "Top Banner", DisplayStyle: "carousel", SourceType: "banner", SourceRef: "home", MaxItems: 10, SortOrder: 10, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "jackpot", SectionType: "jackpot", TitleZh: "奖池", TitleEn: "Jackpot", DisplayStyle: "ticker", SourceType: "jackpot", MaxItems: 1, SortOrder: 20, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "hot-games", SectionType: "game-list", TitleZh: "热门游戏", TitleEn: "Hot Games", DisplayStyle: "horizontal", SourceType: "flag", SourceRef: "hot", MaxItems: 12, SortOrder: 30, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "new-games", SectionType: "game-list", TitleZh: "新游戏", TitleEn: "New Games", DisplayStyle: "horizontal", SourceType: "flag", SourceRef: "new", MaxItems: 12, SortOrder: 40, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "recommended-games", SectionType: "game-list", TitleZh: "推荐游戏", TitleEn: "Recommended Games", DisplayStyle: "horizontal", SourceType: "flag", SourceRef: "featured", MaxItems: 12, SortOrder: 50, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "continue-playing", SectionType: "game-list", TitleZh: "继续游戏", TitleEn: "Continue Playing", DisplayStyle: "horizontal", SourceType: "user", SourceRef: "continue-playing", MaxItems: 12, SortOrder: 60, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "recently-played", SectionType: "game-list", TitleZh: "最近玩过", TitleEn: "Recently Played", DisplayStyle: "horizontal", SourceType: "user", SourceRef: "recently-played", MaxItems: 12, SortOrder: 70, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "categories", SectionType: "category-list", TitleZh: "分类", TitleEn: "Categories", DisplayStyle: "grid", SourceType: "category", MaxItems: 20, SortOrder: 80, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "providers", SectionType: "provider-list", TitleZh: "平台", TitleEn: "Providers", DisplayStyle: "grid", SourceType: "provider", MaxItems: 20, SortOrder: 90, Enabled: true, ConfigJSON: "{}"},
		{SectionCode: "footer", SectionType: "footer", TitleZh: "底部", TitleEn: "Footer", DisplayStyle: "footer", SourceType: "config", MaxItems: 1, SortOrder: 100, Enabled: true, ConfigJSON: "{}"},
	}
}
