package systemadmin

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidateSafeURL(t *testing.T) {
	for _, value := range []string{"", "https://admin.example.com/path", "http://127.0.0.1:8080"} {
		if err := validateSafeURL(value); err != nil {
			t.Fatalf("expected safe URL %q: %v", value, err)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "file:///etc/passwd", "//missing-scheme.example.com"} {
		if err := validateSafeURL(value); err == nil {
			t.Fatalf("expected unsafe URL %q to fail", value)
		}
	}
}

func TestRedactMasksNestedSecrets(t *testing.T) {
	value := redact(map[string]any{
		"username": "operator",
		"password": "plain-text",
		"nested":   map[string]any{"accessToken": "token", "visible": "ok"},
	}).(map[string]any)
	if value["password"] != "******" {
		t.Fatalf("password was not redacted: %#v", value)
	}
	nested := value["nested"].(map[string]any)
	if nested["accessToken"] != "******" || nested["visible"] != "ok" {
		t.Fatalf("nested redaction mismatch: %#v", nested)
	}
}

func TestNextCronTime(t *testing.T) {
	after := time.Date(2026, 7, 16, 8, 2, 45, 0, time.UTC)
	next, err := nextCronTime("*/5 * * * *", after)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 7, 16, 8, 5, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("next run = %s, want %s", next, want)
	}
	if _, err := nextCronTime("not a cron expression", after); err == nil {
		t.Fatal("invalid cron expression was accepted")
	}
}

func TestValidateTaskRejectsUnregisteredHandler(t *testing.T) {
	enabled := true
	input := TaskInput{
		TaskCode:       "unsafe-task",
		NameZh:         "不安全任务",
		NameEn:         "Unsafe task",
		TaskType:       "SYSTEM",
		HandlerCode:    "shell.execute",
		CronExpression: "0 * * * *",
		TimeoutSeconds: 30,
		Enabled:        &enabled,
	}
	if err := validateTask(input, true); err == nil {
		t.Fatal("unregistered task handler was accepted")
	}
}

func TestProviderSyncIsNotRegisteredAsScheduledTaskHandler(t *testing.T) {
	for _, handler := range []string{"game.provider_sync", "provider.sync", "game:provider:sync"} {
		if registeredTaskHandlers()[handler] {
			t.Fatalf("provider sync handler %q must stay platform-only", handler)
		}
	}
}

func TestCountryResourceVersionChangesWithRuntimeAssets(t *testing.T) {
	updatedAt := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	record := countryConfigRecord{
		CountryCode: "PK", CurrencyCode: "PKR", CurrencySymbol: "Rs",
		Timezone: "Asia/Karachi", DefaultLanguage: "en", SupportedLanguages: `["en","ur"]`,
		BrandingConfig: `{"logoUrl":"h5/logo-a.avif"}`,
		ThemeConfig:    `{"primaryColor":"#003f77"}`,
		UpdatedAt:      updatedAt,
	}
	version := countryResourceVersion(record)
	if version == "" || version != countryResourceVersion(record) {
		t.Fatal("resource version must be stable for identical runtime configuration")
	}
	record.BrandingConfig = `{"logoUrl":"h5/logo-b.avif"}`
	if version == countryResourceVersion(record) {
		t.Fatal("resource version did not change after the branding asset changed")
	}
	record.BrandingConfig = `{"logoUrl":"h5/logo-a.avif"}`
	record.ThemeConfig = `{"primaryColor":"#ffcc00"}`
	if version == countryResourceVersion(record) {
		t.Fatal("resource version did not change after the theme changed")
	}
}

func newSystemAdminScopeTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatal(err)
		}
	})
	for _, statement := range []string{

		`CREATE TABLE game_site_config (id INTEGER PRIMARY KEY, site_code TEXT, site_name TEXT, primary_domain TEXT, admin_domain TEXT, h5_domain TEXT, android_package_name TEXT, ios_bundle_id TEXT, country_code TEXT, currency_code TEXT, default_language TEXT, timezone TEXT, brand_config TEXT, theme_config TEXT, feature_config TEXT, maintenance_config TEXT, seo_config TEXT, site_name_draft TEXT, country_code_draft TEXT, currency_code_draft TEXT, default_language_draft TEXT, timezone_draft TEXT, brand_config_draft TEXT, theme_config_draft TEXT, feature_config_draft TEXT, maintenance_config_draft TEXT, seo_config_draft TEXT, config_status TEXT, config_version INTEGER, config_published_at DATETIME, config_draft_updated_at DATETIME, is_initialized INTEGER, enabled INTEGER, status TEXT, remark TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	nowValue := time.Now()
	for _, item := range []struct {
		id     int
		siteID int
		code   string
	}{
		{id: 101, code: "site-a"},
		{id: 202, code: "site-b"},
	} {
		if err := db.Exec("INSERT INTO game_site_config (id, site_code, site_name, primary_domain, admin_domain, h5_domain, android_package_name, ios_bundle_id, country_code, currency_code, default_language, timezone, brand_config, theme_config, feature_config, maintenance_config, seo_config, site_name_draft, country_code_draft, currency_code_draft, default_language_draft, timezone_draft, brand_config_draft, theme_config_draft, feature_config_draft, maintenance_config_draft, seo_config_draft, config_status, config_version, is_initialized, enabled, status, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, '', '', 'MX', 'MXN', 'es', 'America/Mexico_City', '{}', '{}', '{}', '{}', '{}', '', '', '', '', '', '', '', '', '', '', 'PUBLISHED', 1, 1, 1, 'ONLINE', '', ?, ?)", item.id, item.code, item.code, item.code+".example.com", "admin."+item.code+".example.com", "h5."+item.code+".example.com", nowValue, nowValue).Error; err != nil {
			t.Fatal(err)
		}
	}
	return New(db), db
}

func requireSystemErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}
