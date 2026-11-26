package systemadmin

import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/model"
)

func TestProvisionValidationAllowsNumericSaaSCodes(t *testing.T) {
	tenant := TenantInput{
		TenantCode: "08abc", TenantName: "Demo", DefaultCurrency: "PKR", DefaultLanguage: "en",
		Timezone: "Asia/Karachi", OwnerUsername: "tenant-owner", OwnerPassword: "strong-password",
	}
	if err := validateTenantInput(tenant, true); err != nil {
		t.Fatalf("numeric tenant code should be valid: %v", err)
	}
	site := SiteInput{TenantID: 1, SiteCode: "88888888-main", SiteName: "Main", Timezone: "Asia/Karachi"}
	if err := validateSiteInput(site, true); err != nil {
		t.Fatalf("numeric site code should be valid: %v", err)
	}
}

func TestNormalizeProvisionInputGeneratesPasswordAndRequiresEmail(t *testing.T) {
	service := &Service{}
	input, err := service.normalizeProvisionInput(TenantProvisionInput{
		TenantCode: "demo", TenantName: "Demo", PrimaryDomain: "demo.example.com",
		OwnerUsername: "tenant-owner", OwnerEmail: "OWNER@EXAMPLE.COM",
	})
	if err != nil {
		t.Fatal(err)
	}
	if input.OwnerEmail != "owner@example.com" {
		t.Fatalf("normalized email = %q", input.OwnerEmail)
	}
	if err := validatePasswordStrength(input.OwnerPassword); err != nil {
		t.Fatalf("generated password is not strong: %v", err)
	}
	if input.ContactEmail != input.OwnerEmail {
		t.Fatalf("contact email = %q", input.ContactEmail)
	}

	_, err = service.normalizeProvisionInput(TenantProvisionInput{
		TenantCode: "missing-email", TenantName: "Demo", PrimaryDomain: "demo.example.com", OwnerUsername: "tenant-owner",
	})
	if ErrorCode(err) != "OWNER_EMAIL_REQUIRED" {
		t.Fatalf("missing email error = %v", err)
	}
}

func TestProvisionUsesConfiguredCountryDimensions(t *testing.T) {
	input := TenantProvisionInput{CountryCode: "PK", CurrencyCode: "MXN", Language: "es-MX", Timezone: "America/Mexico_City"}
	defaults := provisionCountryDefaults{CountryCode: "PK", CurrencyCode: "PKR", DefaultLanguage: "en", Timezone: "Asia/Karachi"}
	tenant := provisionTenantInput(input, defaults)
	if tenant.CountryCode != "PK" || tenant.DefaultCurrency != "PKR" || tenant.DefaultLanguage != "en" || tenant.Timezone != "Asia/Karachi" {
		t.Fatalf("tenant dimensions did not follow country config: %#v", tenant)
	}
	site := provisionSiteInput(input, 1, tenant, defaults)
	if site.CountryCode != "PK" || site.CurrencyCode != "PKR" || site.DefaultLanguage != "en" || site.Timezone != "Asia/Karachi" {
		t.Fatalf("site dimensions did not follow country config: %#v", site)
	}
}

func TestProvisionCountryDefaultsRejectsUnknownCountry(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_country_configs (id INTEGER PRIMARY KEY,country_code TEXT,currency_code TEXT,default_language TEXT,timezone TEXT,status INTEGER,deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_country_configs VALUES (1,'PK','PKR','en','Asia/Karachi',1,NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	service := New(db)
	if _, err := service.provisionCountryDefaults(db, "XX"); ErrorCode(err) != "COUNTRY_NOT_FOUND" {
		t.Fatalf("unknown country error = %v", err)
	}
	defaults, err := service.provisionCountryDefaults(db, "PK")
	if err != nil {
		t.Fatal(err)
	}
	if defaults.CurrencyCode != "PKR" || defaults.DefaultLanguage != "en" || defaults.Timezone != "Asia/Karachi" {
		t.Fatalf("unexpected country defaults: %#v", defaults)
	}
}

func TestProvisionPackageLimitsAreLoaded(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_saas_tenant_packages (id INTEGER PRIMARY KEY,package_code TEXT,game_limit INTEGER,provider_limit INTEGER,deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_tenant_packages (id,package_code,game_limit,provider_limit) VALUES (1,'BASIC',250,3)`).Error; err != nil {
		t.Fatal(err)
	}
	limits, err := provisionLimits(db, "basic")
	if err != nil {
		t.Fatal(err)
	}
	if limits.GameLimit != 250 || limits.ProviderLimit != 3 {
		t.Fatalf("unexpected package limits: %+v", limits)
	}
	if _, err := provisionLimits(db, "UNKNOWN"); ErrorCode(err) != "PACKAGE_NOT_FOUND" {
		t.Fatalf("unknown package was accepted: %v", err)
	}
}

func TestProvisionGamesCreatesDefaultTenantGamesAndCategoryRelations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_game_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,tenant_id INTEGER,site_id INTEGER,provider_code TEXT,name TEXT,status INTEGER,deleted_at DATETIME
		)`,
		`CREATE TABLE game_game_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME,
			tenant_id INTEGER,site_id INTEGER,category_key TEXT,label TEXT,name_zh TEXT,name_en TEXT,short_label TEXT,
			icon_url TEXT,banner_url TEXT,sort INTEGER,status INTEGER
		)`,
		`CREATE TABLE game_game_category_relations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME,
			tenant_id INTEGER,site_id INTEGER,game_id INTEGER,category_id INTEGER,sort INTEGER
		)`,
		`CREATE TABLE game_games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME,
			tenant_id INTEGER,site_id INTEGER,game_code TEXT,provider_id INTEGER,provider_game_code TEXT,name TEXT,name_i18n TEXT,
			slug TEXT,category TEXT,platform_category_id INTEGER,platform_category_code TEXT,game_type TEXT,rtp NUMERIC,
			volatility TEXT,cover_url TEXT,local_thumbnail_url TEXT,background_url TEXT,launch_image_url TEXT,maintenance_message TEXT,
			supports_demo INTEGER,supports_free_spins INTEGER,provider_enabled INTEGER,platform_enabled INTEGER,platform_enabled_draft INTEGER,
			is_hot INTEGER,is_hot_draft INTEGER,is_featured INTEGER,is_featured_draft INTEGER,is_new INTEGER,is_new_draft INTEGER,
			supported_countries TEXT,supported_currencies TEXT,supported_devices TEXT,raw_metadata TEXT,metadata_hash TEXT,last_synced_at DATETIME,
			badge TEXT,player_count INTEGER,hot_score INTEGER,launch_url TEXT,status INTEGER,status_draft INTEGER,sort INTEGER,sort_draft INTEGER,
			maintenance_message_draft TEXT,category_ids_draft TEXT,config_status TEXT,config_version INTEGER,config_published_at DATETIME,
			config_published_by INTEGER,draft_updated_at DATETIME,draft_updated_by INTEGER
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Table("game_game_providers").Create(map[string]any{
		"tenant_id": 0, "site_id": 0, "provider_code": "ag", "name": "AG", "status": 1,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("game_game_providers").Create(map[string]any{
		"tenant_id": 7, "site_id": 9, "provider_code": "ag", "name": "AG", "status": 1,
	}).Error; err != nil {
		t.Fatal(err)
	}
	for _, category := range []map[string]any{
		{"tenant_id": 0, "site_id": 0, "category_key": "fc", "label": "FC", "name_zh": "FC", "name_en": "FC", "short_label": "FC", "sort": 1, "status": 1},
		{"tenant_id": 0, "site_id": 0, "category_key": "jili", "label": "JILI", "name_zh": "JILI", "name_en": "JILI", "short_label": "JILI", "sort": 2, "status": 1},
	} {
		if err := db.Table("game_game_categories").Create(category).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := provisionGameCategories(db, 7, 9); err != nil {
		t.Fatal(err)
	}
	for index, gameCode := range []string{"ag-demo-1", "ag-demo-2"} {
		if err := db.Table("game_games").Create(map[string]any{
			"tenant_id": 0, "site_id": 0, "game_code": gameCode, "provider_id": 1, "provider_game_code": fmt.Sprintf("demo-%d", index+1),
			"name": "Demo", "name_i18n": "{}", "slug": gameCode, "category": "AG_HOT_FC", "game_type": "slot",
			"raw_metadata": "{}", "metadata_hash": "", "provider_enabled": true, "platform_enabled": true, "status": 1,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if count, err := provisionGames(db, 7, 9, []string{"ag"}, 250, 99); err != nil || count != 2 {
		t.Fatalf("provisionGames count=%d err=%v", count, err)
	}
	var tenantGames int64
	if err := db.Table("game_games").Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", 7, 9).Count(&tenantGames).Error; err != nil {
		t.Fatal(err)
	}
	if tenantGames != 2 {
		t.Fatalf("new tenant should receive the default game pool, got %d", tenantGames)
	}
	var relationCount int64
	if err := db.Table("game_game_category_relations").Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", 7, 9).Count(&relationCount).Error; err != nil {
		t.Fatal(err)
	}
	if relationCount != 4 {
		t.Fatalf("each default category should contain the default games, got %d relations", relationCount)
	}
}

func TestProvisionGameCategoriesCopiesPlatformRowsIntoTenantScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_game_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
		tenant_id INTEGER NOT NULL DEFAULT 0, site_id INTEGER NOT NULL DEFAULT 0,
		category_key TEXT NOT NULL, label TEXT NOT NULL, name_zh TEXT NOT NULL, name_en TEXT NOT NULL,
		short_label TEXT NOT NULL, icon_url TEXT NOT NULL, banner_url TEXT NOT NULL,
		sort INTEGER NOT NULL DEFAULT 0, status INTEGER NOT NULL DEFAULT 1
	)`).Error; err != nil {
		t.Fatal(err)
	}
	platform := model.GameGameCategories{
		CategoryKey: "slots", Label: "Slots", NameZh: "老虎机", NameEn: "Slots",
		ShortLabel: "S", IconURL: "slots.png", BannerURL: "slots-banner.png", Sort: 10, Status: 1,
	}
	if err := db.Create(&platform).Error; err != nil {
		t.Fatal(err)
	}
	if err := provisionGameCategories(db, 7, 9); err != nil {
		t.Fatal(err)
	}
	var copied model.GameGameCategories
	if err := db.Where("tenant_id=? AND site_id=? AND category_key=?", 7, 9, "slots").Take(&copied).Error; err != nil {
		t.Fatal(err)
	}
	if copied.ID == platform.ID || copied.TenantID != 7 || copied.SiteID != 9 || copied.NameEn != "Slots" || copied.IconURL != "slots.png" {
		t.Fatalf("unexpected tenant category copy: %+v", copied)
	}
	if err := provisionGameCategories(db, 7, 9); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&model.GameGameCategories{}).Where("tenant_id=? AND site_id=?", 7, 9).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("category provisioning was not idempotent, count=%d", count)
	}
}

func TestProvisionDepositAmountOptionsUsesTenantCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_deposit_amount_options (
		id INTEGER PRIMARY KEY AUTOINCREMENT,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME,
		tenant_id INTEGER,site_id INTEGER,option_code TEXT,currency_code TEXT,amount NUMERIC,bonus_reward NUMERIC,cash_reward NUMERIC,sort INTEGER,status INTEGER
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("game_deposit_amount_options").Create(map[string]any{
		"tenant_id": 0, "site_id": 0, "option_code": "DEFAULT_10", "currency_code": "PKR",
		"amount": 10, "bonus_reward": 1, "cash_reward": 0, "sort": 1, "status": 1,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := provisionDepositAmountOptions(db, 7, 9, "USD"); err != nil {
		t.Fatal(err)
	}
	var row struct {
		TenantID     uint64
		SiteID       uint64
		OptionCode   string
		CurrencyCode string
	}
	if err := db.Table("game_deposit_amount_options").Where("tenant_id=? AND site_id=?", 7, 9).Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.OptionCode != "DEFAULT_10" || row.CurrencyCode != "USD" {
		t.Fatalf("unexpected copied option: %#v", row)
	}
}

func TestPublishRejectsGamesOverPackageLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY,package_code TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_tenant_packages (id INTEGER PRIMARY KEY,package_code TEXT,game_limit INTEGER,deleted_at DATETIME)`,
		`CREATE TABLE game_games (id INTEGER PRIMARY KEY,tenant_id INTEGER,site_id INTEGER,platform_enabled INTEGER,platform_enabled_draft INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_saas_tenants (id,package_code) VALUES (1,'BASIC')`,
		`INSERT INTO game_saas_tenant_packages (id,package_code,game_limit) VALUES (1,'BASIC',1)`,
		`INSERT INTO game_games (id,tenant_id,site_id,platform_enabled,platform_enabled_draft) VALUES (1,1,10,1,NULL),(2,1,10,0,1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := validatePublishedGameLimit(db, 1, 10); ErrorCode(err) != "TENANT_GAME_LIMIT_EXCEEDED" {
		t.Fatalf("expected package limit rejection, got %v", err)
	}
	if err := db.Exec(`UPDATE game_games SET platform_enabled_draft=0 WHERE id=2`).Error; err != nil {
		t.Fatal(err)
	}
	if err := validatePublishedGameLimit(db, 1, 10); err != nil {
		t.Fatalf("valid package selection was rejected: %v", err)
	}
}

func TestPublishGrandfathersExistingGamesAbovePackageLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY,package_code TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_tenant_packages (id INTEGER PRIMARY KEY,package_code TEXT,game_limit INTEGER,deleted_at DATETIME)`,
		`CREATE TABLE game_games (id INTEGER PRIMARY KEY,tenant_id INTEGER,site_id INTEGER,platform_enabled INTEGER,platform_enabled_draft INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_saas_tenants (id,package_code) VALUES (1,'BASIC')`,
		`INSERT INTO game_saas_tenant_packages (id,package_code,game_limit) VALUES (1,'BASIC',1)`,
		`INSERT INTO game_games (id,tenant_id,site_id,platform_enabled,platform_enabled_draft) VALUES (1,1,10,1,NULL),(2,1,10,1,NULL)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := validatePublishedGameLimit(db, 1, 10); err != nil {
		t.Fatalf("unchanged legacy selection was rejected: %v", err)
	}
	if err := db.Exec(`INSERT INTO game_games (id,tenant_id,site_id,platform_enabled,platform_enabled_draft) VALUES (3,1,10,0,1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := validatePublishedGameLimit(db, 1, 10); ErrorCode(err) != "TENANT_GAME_LIMIT_EXCEEDED" {
		t.Fatalf("legacy tenant increased games beyond grandfathered count: %v", err)
	}
}

func TestCreateTenantResolvesInsertedMapID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,tenant_code TEXT UNIQUE,tenant_name TEXT,company_name TEXT,contact_name TEXT,
			contact_email TEXT,contact_phone TEXT,plan_code TEXT,country_code TEXT,package_code TEXT,default_currency TEXT,
			default_language TEXT,timezone TEXT,expires_at DATETIME,enabled INTEGER,status TEXT,maintenance_enabled INTEGER,
			maintenance_message TEXT,remark TEXT,created_by INTEGER,updated_by INTEGER,created_at DATETIME,updated_at DATETIME,deleted_at DATETIME
		)`,
		`CREATE TABLE game_admin_operation_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,operator_id INTEGER,tenant_id INTEGER,site_id INTEGER,operator_username TEXT,
			permission TEXT,module TEXT,action TEXT,result TEXT,target_type TEXT,target_id INTEGER,before_state TEXT,after_state TEXT,
			reason TEXT,ip_address TEXT,user_agent TEXT,correlation_id TEXT,currency TEXT,created_at DATETIME,updated_at DATETIME
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	id, err := New(db).CreateTenant(TenantInput{
		TenantCode: "08abc", TenantName: "Demo", CountryCode: "PK", DefaultCurrency: "PKR",
		DefaultLanguage: "en", Timezone: "Asia/Karachi", Reason: "test",
	}, Actor{ID: 1, Username: "system-admin", Platform: true})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("inserted tenant ID was not resolved")
	}
}

func TestTenantCodeAvailabilityIgnoresDeletedRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY, tenant_code TEXT, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_tenants (id,tenant_code,deleted_at) VALUES (1,'demo',CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := ensureTenantCodeAvailable(db, "demo"); err != nil {
		t.Fatalf("deleted tenant code should be reusable: %v", err)
	}
	if err := db.Exec(`INSERT INTO game_saas_tenants (id,tenant_code) VALUES (2,'demo')`).Error; err != nil {
		t.Fatal(err)
	}
	if ErrorCode(ensureTenantCodeAvailable(db, "demo")) != "TENANT_CODE_EXISTS" {
		t.Fatalf("active tenant code should be rejected")
	}
}

func TestSiteCodeAvailabilityIgnoresDeletedRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY, site_code TEXT, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_sites (id,site_code,deleted_at) VALUES (1,'demo-main',CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := ensureSiteCodeAvailable(db, "demo-main"); err != nil {
		t.Fatalf("deleted site code should be reusable: %v", err)
	}
	if err := db.Exec(`INSERT INTO game_saas_sites (id,site_code) VALUES (2,'demo-main')`).Error; err != nil {
		t.Fatal(err)
	}
	if ErrorCode(ensureSiteCodeAvailable(db, "demo-main")) != "SITE_CODE_EXISTS" {
		t.Fatalf("active site code should be rejected")
	}
}
