package gameadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"casino/internal/model"
	"casino/internal/tenancy"

	"github.com/go-sql-driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNormalizePage(t *testing.T) {
	page, pageSize, offset := normalizePage(ListQuery{Page: 0, PageSize: 0})
	if page != 1 || pageSize != 20 || offset != 0 {
		t.Fatalf("unexpected defaults: %d %d %d", page, pageSize, offset)
	}
	page, pageSize, offset = normalizePage(ListQuery{Page: 3, PageSize: 500})
	if page != 3 || pageSize != 100 || offset != 200 {
		t.Fatalf("unexpected capped pagination: %d %d %d", page, pageSize, offset)
	}
}

func TestValidateCodeAndTimeRange(t *testing.T) {
	if err := validateCode("live-casino"); err != nil {
		t.Fatalf("valid code rejected: %v", err)
	}
	if err := validateCode("Live Casino"); err == nil {
		t.Fatal("invalid code accepted")
	}
	start := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	end := start.Add(-time.Minute)
	if err := validateTimeRange(&start, &end); err == nil {
		t.Fatal("invalid time range accepted")
	}
}

func TestRedactJSON(t *testing.T) {
	redacted := redactJSON(`{"apiPassword":"secret","salt_key":"salt","nested":{"launchUrl":"https://secret","name":"safe"},"accessToken":"token"}`)
	encoded, err := json.Marshal(redacted)
	if err != nil {
		t.Fatal(err)
	}
	value := string(encoded)
	for _, secret := range []string{"secret", "https://secret", "token", "salt\""} {
		if strings.Contains(value, secret) {
			t.Fatalf("sensitive value was not redacted: %s", value)
		}
	}
	if !strings.Contains(value, `"name":"safe"`) {
		t.Fatalf("safe metadata was removed: %s", value)
	}
}

func TestDuplicateCode(t *testing.T) {
	duplicate := &mysql.MySQLError{Number: 1062, Message: "duplicate"}
	err := duplicateCode(duplicate, "CODE_EXISTS", "already exists")
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != "CODE_EXISTS" {
		t.Fatalf("duplicate error not normalized: %v", err)
	}
	original := errors.New("connection lost")
	if got := duplicateCode(original, "CODE_EXISTS", "already exists"); !errors.Is(got, original) {
		t.Fatalf("non-duplicate error was hidden: %v", got)
	}
}

func TestGameListRecordDoesNotExposeRawMetadata(t *testing.T) {
	encoded, err := json.Marshal(gameRecord{ID: 1, Name: "Game", RawMetadata: `{"apiPassword":"secret"}`})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "rawMetadata") || strings.Contains(string(encoded), "secret") {
		t.Fatalf("raw provider metadata leaked: %s", encoded)
	}
}

func TestTenantAdminCannotSyncProvider(t *testing.T) {
	_, err := (&Service{}).SyncProvider(context.Background(), "ag", "manual", Actor{Scope: tenancy.Scope{TenantID: 1, SiteID: 1}})
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != "PLATFORM_SCOPE_REQUIRED" {
		t.Fatalf("expected platform-only error, got %v", err)
	}
}

func TestTenantAdminCannotSyncAllProviders(t *testing.T) {
	_, err := (&Service{}).SyncAll(context.Background(), "manual", Actor{Scope: tenancy.Scope{TenantID: 1, SiteID: 1}})
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != "PLATFORM_SCOPE_REQUIRED" {
		t.Fatalf("expected platform-only error, got %v", err)
	}
}

func TestListProviderGameTypesIsTenantScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_game_providers (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, provider_code TEXT NOT NULL, name TEXT NOT NULL, deleted_at DATETIME)`,
		`CREATE TABLE game_games (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, provider_id INTEGER NOT NULL, category TEXT NOT NULL DEFAULT '', game_type TEXT NOT NULL DEFAULT '', deleted_at DATETIME)`,
		`INSERT INTO game_game_providers (id,tenant_id,site_id,provider_code,name) VALUES (1,1,101,'ag','Tenant A AG'),(2,2,202,'ag','Tenant B AG')`,
		`INSERT INTO game_games (id,tenant_id,site_id,provider_id,category,game_type) VALUES (11,1,101,1,'slots','SLOT'),(12,1,101,1,'slots','SLOT'),(21,2,202,2,'live','LIVE')`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	rows, err := (&Service{db: db}).ListProviderGameTypes("ag", tenancy.Scope{TenantID: 1, SiteID: 101})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Code != "slots" || rows[0].GameCount != 2 {
		t.Fatalf("tenant provider options crossed scope: %+v", rows)
	}

	_, err = (&Service{db: db}).ListProviderGameTypes("ag", tenancy.Scope{TenantID: 3, SiteID: 303})
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != "PROVIDER_NOT_FOUND" {
		t.Fatalf("expected scoped provider rejection, got %v", err)
	}
}

func TestTenantCategoryListContainsOnlyTenantCopiesAndHidesInternalFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_game_categories (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, category_key TEXT NOT NULL, label TEXT NOT NULL, name_zh TEXT NOT NULL, name_en TEXT NOT NULL, short_label TEXT NOT NULL, icon_url TEXT NOT NULL, banner_url TEXT NOT NULL, sort INTEGER NOT NULL, status INTEGER NOT NULL, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_game_category_relations (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, category_id INTEGER NOT NULL, game_id INTEGER NOT NULL, deleted_at DATETIME)`,
		`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key,label,name_zh,name_en,short_label,icon_url,banner_url,sort,status,created_at,updated_at) VALUES (1,0,0,'slots','Slots','老虎机','Slots','S','','',1,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key,label,name_zh,name_en,short_label,icon_url,banner_url,sort,status,created_at,updated_at) VALUES (2,1,101,'vip','VIP','贵宾','VIP','V','','',2,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key,label,name_zh,name_en,short_label,icon_url,banner_url,sort,status,created_at,updated_at) VALUES (3,2,202,'other','Other','其他','Other','O','','',3,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		`INSERT INTO game_game_category_relations (id,tenant_id,site_id,category_id,game_id) VALUES (10,1,101,1,1001),(11,2,202,1,2001),(12,1,101,2,1002)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	result, err := (&Service{db: db}).ListCategories(ListQuery{Page: 1, PageSize: 20, Scope: tenancy.Scope{TenantID: 1, SiteID: 101}})
	if err != nil {
		t.Fatal(err)
	}
	items, ok := result.List.([]tenantCategoryListItem)
	if !ok {
		t.Fatalf("unexpected tenant category list type %T", result.List)
	}
	if len(items) != 1 || items[0].ID != 2 || items[0].GameCount != 1 {
		t.Fatalf("tenant category list did not contain only the tenant copy: %+v", items)
	}
	encoded, err := json.Marshal(result.List)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"code"`) || strings.Contains(string(encoded), `"nameZh"`) || strings.Contains(string(encoded), "老虎机") {
		t.Fatalf("tenant category response leaked internal fields: %s", encoded)
	}
}

func TestPlatformCategoryListExcludesTenantCategories(t *testing.T) {
	db := newCategoryListScopeTestDB(t)
	result, err := (&Service{db: db}).ListCategories(ListQuery{Page: 1, PageSize: 20, Scope: tenancy.Scope{Platform: true}})
	if err != nil {
		t.Fatal(err)
	}
	items, ok := result.List.([]categoryRecord)
	if !ok {
		t.Fatalf("unexpected platform category list type %T", result.List)
	}
	if len(items) != 1 || items[0].ID != 1 || items[0].TenantID != 0 || items[0].SiteID != 0 {
		t.Fatalf("platform category list leaked tenant categories: %+v", items)
	}
}

func TestTenantCategoryWritesAreStrictlyScoped(t *testing.T) {
	db := newCategoryWriteScopeTestDB(t)
	service := &Service{db: db}
	tenantA := Actor{ID: 10, Username: "tenant-a", Scope: tenancy.Scope{TenantID: 1, SiteID: 101}}
	tenantB := Actor{ID: 20, Username: "tenant-b", Scope: tenancy.Scope{TenantID: 2, SiteID: 202}}

	tenantAID, err := service.SaveCategory(0, CategoryInput{Code: "live", NameZh: "LIVE A", NameEn: "LIVE A", Enabled: true, Reason: "create tenant category"}, tenantA)
	if err != nil {
		t.Fatalf("tenant A create failed: %v", err)
	}
	tenantBID, err := service.SaveCategory(0, CategoryInput{Code: "live", NameZh: "LIVE B", NameEn: "LIVE B", Enabled: true, Reason: "create tenant category"}, tenantB)
	if err != nil {
		t.Fatalf("tenant B should be able to reuse category code in its own scope: %v", err)
	}
	var rows []categoryRecord
	if err := db.Table("game_game_categories").Where("id IN ?", []uint64{tenantAID, tenantBID}).Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].TenantID != 1 || rows[0].SiteID != 101 || rows[1].TenantID != 2 || rows[1].SiteID != 202 {
		t.Fatalf("categories were not created in strict tenant/site scope: %+v", rows)
	}

	_, err = service.SaveCategory(100, CategoryInput{Code: "slots", NameZh: "越权", NameEn: "Cross", Enabled: true, Reason: "cross tenant update"}, tenantA)
	assertGameAdminErrorCode(t, err, "CATEGORY_NOT_FOUND")
	_, err = service.SaveCategory(300, CategoryInput{Code: "shared", NameZh: "越权", NameEn: "Cross", Enabled: true, Reason: "cross site update"}, tenantA)
	assertGameAdminErrorCode(t, err, "CATEGORY_NOT_FOUND")
	_, err = service.SaveCategory(tenantBID, CategoryInput{Code: "live", NameZh: "越权", NameEn: "Cross", Enabled: true, Reason: "cross tenant update"}, tenantA)
	assertGameAdminErrorCode(t, err, "CATEGORY_NOT_FOUND")
}

func TestTenantImportsPlatformGamesAsDraft(t *testing.T) {
	db := newGameImportTestDB(t)
	service := &Service{db: db}
	actor := Actor{ID: 10, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 9}}

	result, err := service.ImportGames(GameImportInput{IDs: []uint64{1001}, Reason: "tenant self import"}, actor)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result["importedCount"] != int64(1) {
		t.Fatalf("unexpected import result: %+v", result)
	}
	var row struct {
		TenantID             uint64 `gorm:"column:tenant_id"`
		SiteID               uint64 `gorm:"column:site_id"`
		GameCode             string `gorm:"column:game_code"`
		ProviderID           uint64 `gorm:"column:provider_id"`
		PlatformEnabled      bool   `gorm:"column:platform_enabled"`
		PlatformEnabledDraft bool   `gorm:"column:platform_enabled_draft"`
		Status               int    `gorm:"column:status"`
		StatusDraft          int    `gorm:"column:status_draft"`
		ConfigStatus         string `gorm:"column:config_status"`
		PlatformCategoryID   uint64 `gorm:"column:platform_category_id"`
		PlatformCategoryCode string `gorm:"column:platform_category_code"`
	}
	if err := db.Table("game_games").Where("tenant_id=? AND site_id=? AND provider_game_code=?", 7, 9, "demo-1").Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.ProviderID != 200 || row.PlatformEnabled || row.PlatformEnabledDraft || row.Status != 3 || row.StatusDraft != 3 || row.ConfigStatus != "DRAFT" || row.PlatformCategoryID != 0 || row.PlatformCategoryCode != "" {
		t.Fatalf("imported game was not a tenant draft: %+v", row)
	}

	result, err = service.ImportGames(GameImportInput{IDs: []uint64{1001}, Reason: "repeat import"}, actor)
	if err != nil {
		t.Fatalf("repeat import failed: %v", err)
	}
	if result["importedCount"] != int64(0) || result["skippedCount"] != 1 {
		t.Fatalf("repeat import should be idempotent, got %+v", result)
	}
}

func assertGameAdminErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}

func TestTenantAdminCannotModifyOtherTenantLobbySection(t *testing.T) {
	db := newGameAdminScopeTestDB(t)
	service := &Service{db: db}
	row := model.GameLobbySections{
		TenantID:            2,
		SiteID:              202,
		SectionCode:         "home-featured",
		SectionType:         "game-list",
		TitleZh:             "推荐",
		TitleEn:             "Featured",
		DisplayStyle:        "horizontal",
		SourceType:          "flag",
		SourceRef:           "featured",
		MaxItems:            12,
		SortOrder:           10,
		Enabled:             true,
		SupportedCountries:  "",
		SupportedCurrencies: "",
		ConfigJSON:          "{}",
		PublishStatus:       "PUBLISHED",
		ConfigVersion:       1,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}

	_, err := service.SaveLobbySection(row.ID, LobbySectionInput{
		SectionCode:  "home-featured",
		SectionType:  "game-list",
		TitleZh:      "越权",
		TitleEn:      "Cross Tenant",
		DisplayStyle: "horizontal",
		SourceType:   "flag",
		SourceRef:    "hot",
		MaxItems:     8,
		SortOrder:    1,
		Enabled:      false,
		Config:       map[string]any{},
		Reason:       "attempt cross-tenant lobby update",
	}, Actor{ID: 11, Username: "tenant-a-admin", Scope: tenancy.Scope{TenantID: 1, SiteID: 101}})
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != "LOBBY_SECTION_NOT_FOUND" {
		t.Fatalf("expected lobby scope rejection, got %v", err)
	}
}

func TestPlatformGameUpdateIsImmediateWhileTenantUpdateRemainsDraft(t *testing.T) {
	db := newGameUpdateWorkflowTestDB(t)
	service := &Service{db: db}
	input := GameInput{
		Name:                "Fortune Tiger",
		PlatformEnabled:     true,
		IsHot:               true,
		IsFeatured:          true,
		IsNew:               false,
		SortOrder:           8,
		SupportedCountries:  "PK",
		SupportedCurrencies: "PKR",
		SupportedDevices:    "mobile",
		Reason:              "update game availability",
	}

	if err := service.UpdateGame(1, input, Actor{ID: 10, Username: "platform-admin", Scope: tenancy.Scope{Platform: true}}); err != nil {
		t.Fatal(err)
	}
	var platformRow struct {
		PlatformEnabled      bool   `gorm:"column:platform_enabled"`
		PlatformEnabledDraft bool   `gorm:"column:platform_enabled_draft"`
		IsHot                bool   `gorm:"column:is_hot"`
		ConfigStatus         string `gorm:"column:config_status"`
		ConfigVersion        uint64 `gorm:"column:config_version"`
	}
	if err := db.Table("game_games").Where("id=1").Take(&platformRow).Error; err != nil {
		t.Fatal(err)
	}
	if !platformRow.PlatformEnabled || !platformRow.PlatformEnabledDraft || !platformRow.IsHot || platformRow.ConfigStatus != "PUBLISHED" || platformRow.ConfigVersion != 2 {
		t.Fatalf("platform update was not published immediately: %+v", platformRow)
	}
	if err := service.SetGameCategories(1, RelationInput{IDs: []uint64{100}, Reason: "assign platform category"}, Actor{ID: 10, Username: "platform-admin", Scope: tenancy.Scope{Platform: true}}); err != nil {
		t.Fatal(err)
	}
	var relationCount int64
	if err := db.Table("game_game_category_relations").Where("game_id=1 AND category_id=100 AND deleted_at IS NULL").Count(&relationCount).Error; err != nil {
		t.Fatal(err)
	}
	if relationCount != 1 {
		t.Fatalf("platform category assignment was not published immediately: count=%d", relationCount)
	}
	featured := false
	if _, err := service.BatchUpdateGames(GameBatchInput{
		IDs:         []uint64{1},
		IsFeatured:  &featured,
		CategoryIDs: &[]uint64{100},
		SortOrders:  []GameSortInput{{ID: 1, SortOrder: 12}},
		Reason:      "batch update platform game",
	}, Actor{ID: 10, Username: "platform-admin", Scope: tenancy.Scope{Platform: true}}); err != nil {
		t.Fatal(err)
	}
	var batchRow struct {
		Sort          int    `gorm:"column:sort"`
		IsFeatured    bool   `gorm:"column:is_featured"`
		ConfigStatus  string `gorm:"column:config_status"`
		ConfigVersion uint64 `gorm:"column:config_version"`
	}
	if err := db.Table("game_games").Where("id=1").Take(&batchRow).Error; err != nil {
		t.Fatal(err)
	}
	if batchRow.Sort != 12 || batchRow.IsFeatured || batchRow.ConfigStatus != "PUBLISHED" || batchRow.ConfigVersion != 4 {
		t.Fatalf("platform batch update was not applied as one published revision: %+v", batchRow)
	}

	if err := service.UpdateGame(2, input, Actor{ID: 20, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 9}}); err != nil {
		t.Fatal(err)
	}
	var tenantRow struct {
		PlatformEnabled      bool   `gorm:"column:platform_enabled"`
		PlatformEnabledDraft bool   `gorm:"column:platform_enabled_draft"`
		ConfigStatus         string `gorm:"column:config_status"`
		ConfigVersion        uint64 `gorm:"column:config_version"`
	}
	if err := db.Table("game_games").Where("id=2").Take(&tenantRow).Error; err != nil {
		t.Fatal(err)
	}
	if tenantRow.PlatformEnabled || !tenantRow.PlatformEnabledDraft || tenantRow.ConfigStatus != "DRAFT" || tenantRow.ConfigVersion != 1 {
		t.Fatalf("tenant update bypassed draft workflow: %+v", tenantRow)
	}
}

func TestTenantBatchAddCategoriesPreservesExistingDraft(t *testing.T) {
	db := newGameUpdateWorkflowTestDB(t)
	service := &Service{db: db}
	actor := Actor{ID: 20, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 9}}
	if err := service.SetGameCategories(2, RelationInput{IDs: []uint64{200}, Reason: "initial category"}, actor); err != nil {
		t.Fatal(err)
	}
	additions := []uint64{201}
	result, err := service.BatchUpdateGames(GameBatchInput{IDs: []uint64{2}, AddCategoryIDs: &additions, Reason: "add another category"}, actor)
	if err != nil {
		t.Fatal(err)
	}
	if result["updatedCount"] != int64(1) {
		t.Fatalf("unexpected batch result: %+v", result)
	}
	var row struct {
		CategoryIDsDraft string `gorm:"column:category_ids_draft"`
		ConfigStatus     string `gorm:"column:config_status"`
	}
	if err := db.Table("game_games").Where("id=2").Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	var items []gameCategoryDraftItem
	if err := json.Unmarshal([]byte(row.CategoryIDsDraft), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].CategoryID != 200 || items[1].CategoryID != 201 || row.ConfigStatus != "DRAFT" {
		t.Fatalf("batch category addition replaced or published the draft: row=%+v items=%+v", row, items)
	}
}

func TestTenantGameEnableDraftDoesNotEnforcePackageLimit(t *testing.T) {
	db := newGameUpdateWorkflowTestDB(t)
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY,package_code TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_tenant_packages (id INTEGER PRIMARY KEY,package_code TEXT,game_limit INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_saas_tenants (id,package_code) VALUES (7,'BASIC')`,
		`INSERT INTO game_saas_tenant_packages (id,package_code,game_limit) VALUES (1,'BASIC',1)`,
		`INSERT INTO game_games (id,tenant_id,site_id,provider_enabled,platform_enabled,platform_enabled_draft,config_status,config_version) VALUES (3,7,9,1,1,1,'PUBLISHED',1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	enabled := true
	if _, err := (&Service{db: db}).BatchUpdateGames(GameBatchInput{IDs: []uint64{2}, PlatformEnabled: &enabled, Reason: "save draft beyond package"}, Actor{ID: 20, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 9}}); err != nil {
		t.Fatalf("tenant draft enable should not enforce H5 published game limit: %v", err)
	}
	var row struct {
		PlatformEnabledDraft bool   `gorm:"column:platform_enabled_draft"`
		ConfigStatus         string `gorm:"column:config_status"`
	}
	if err := db.Table("game_games").Where("id=2").Take(&row).Error; err != nil {
		t.Fatal(err)
	}
	if !row.PlatformEnabledDraft || row.ConfigStatus != "DRAFT" {
		t.Fatalf("expected enabled draft, got %+v", row)
	}
}

func TestTenantGameEnableAllowsExistingPublishedCountAbovePackageLimit(t *testing.T) {
	db := newGameUpdateWorkflowTestDB(t)
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY,package_code TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_tenant_packages (id INTEGER PRIMARY KEY,package_code TEXT,game_limit INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_saas_tenants (id,package_code) VALUES (7,'BASIC')`,
		`INSERT INTO game_saas_tenant_packages (id,package_code,game_limit) VALUES (1,'BASIC',1)`,
		`UPDATE game_games SET platform_enabled=1,platform_enabled_draft=1 WHERE id=2`,
		`INSERT INTO game_games (id,tenant_id,site_id,provider_enabled,platform_enabled,platform_enabled_draft,config_status,config_version) VALUES (3,7,9,1,1,1,'PUBLISHED',1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	enabled := true
	if _, err := (&Service{db: db}).BatchUpdateGames(GameBatchInput{IDs: []uint64{2}, PlatformEnabled: &enabled, Reason: "keep legacy selection"}, Actor{ID: 20, Username: "tenant-admin", Scope: tenancy.Scope{TenantID: 7, SiteID: 9}}); err != nil {
		t.Fatalf("existing published count was rejected: %v", err)
	}
}

func newGameUpdateWorkflowTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`CREATE TABLE game_games (
		id INTEGER PRIMARY KEY,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		deleted_at DATETIME,
		provider_enabled INTEGER NOT NULL DEFAULT 1,
		name TEXT NOT NULL DEFAULT '',
		name_i18n TEXT,
		local_thumbnail_url TEXT NOT NULL DEFAULT '',
		background_url TEXT NOT NULL DEFAULT '',
		launch_image_url TEXT NOT NULL DEFAULT '',
		rtp REAL,
		volatility TEXT NOT NULL DEFAULT '',
		supports_demo INTEGER NOT NULL DEFAULT 0,
		supports_free_spins INTEGER NOT NULL DEFAULT 0,
		platform_enabled INTEGER NOT NULL DEFAULT 0,
		platform_enabled_draft INTEGER,
		is_hot INTEGER NOT NULL DEFAULT 0,
		is_hot_draft INTEGER,
		is_featured INTEGER NOT NULL DEFAULT 0,
		is_featured_draft INTEGER,
		is_new INTEGER NOT NULL DEFAULT 0,
		is_new_draft INTEGER,
		sort INTEGER NOT NULL DEFAULT 0,
		sort_draft INTEGER,
		status INTEGER NOT NULL DEFAULT 3,
		status_draft INTEGER,
		maintenance_message TEXT NOT NULL DEFAULT '',
		maintenance_message_draft TEXT,
		category_ids_draft TEXT,
		platform_category_id INTEGER NOT NULL DEFAULT 0,
		platform_category_code TEXT NOT NULL DEFAULT '',
		supported_countries TEXT NOT NULL DEFAULT '',
		supported_currencies TEXT NOT NULL DEFAULT '',
		supported_devices TEXT NOT NULL DEFAULT '',
		config_status TEXT NOT NULL DEFAULT 'PUBLISHED',
		config_version INTEGER NOT NULL DEFAULT 1,
		config_published_at DATETIME,
		config_published_by INTEGER NOT NULL DEFAULT 0,
		draft_updated_at DATETIME,
		draft_updated_by INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.GameAdminOperationLogs{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_game_categories (
		id INTEGER PRIMARY KEY,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		category_key TEXT NOT NULL,
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_game_category_relations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		game_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		sort INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		UNIQUE(game_id,category_id)
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key) VALUES (100,0,0,'slots'),(200,7,9,'featured'),(201,7,9,'live')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_games (id,tenant_id,site_id,provider_enabled,config_status,config_version) VALUES (1,0,0,1,'PUBLISHED',1),(2,7,9,1,'PUBLISHED',1)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func newGameImportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_game_providers (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			site_id INTEGER NOT NULL DEFAULT 0,
			provider_code TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			logo_url TEXT NOT NULL DEFAULT '',
			api_base_url TEXT NOT NULL DEFAULT '',
			environment TEXT NOT NULL DEFAULT '',
			api_status TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			wallet_mode TEXT NOT NULL DEFAULT '',
			callback_url TEXT NOT NULL DEFAULT '',
			supported_countries TEXT NOT NULL DEFAULT '',
			supported_currencies TEXT NOT NULL DEFAULT '',
			maintenance_message TEXT NOT NULL DEFAULT '',
			sort INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE game_games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			site_id INTEGER NOT NULL DEFAULT 0,
			game_code TEXT NOT NULL,
			provider_id INTEGER NOT NULL DEFAULT 0,
			provider_game_code TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			name_i18n TEXT,
			slug TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			platform_category_id INTEGER NOT NULL DEFAULT 0,
			platform_category_code TEXT NOT NULL DEFAULT '',
			game_type TEXT NOT NULL DEFAULT '',
			rtp NUMERIC,
			volatility TEXT NOT NULL DEFAULT '',
			cover_url TEXT NOT NULL DEFAULT '',
			local_thumbnail_url TEXT NOT NULL DEFAULT '',
			background_url TEXT NOT NULL DEFAULT '',
			launch_image_url TEXT NOT NULL DEFAULT '',
			maintenance_message TEXT NOT NULL DEFAULT '',
			supports_demo INTEGER NOT NULL DEFAULT 0,
			supports_free_spins INTEGER NOT NULL DEFAULT 0,
			provider_enabled INTEGER NOT NULL DEFAULT 1,
			platform_enabled INTEGER NOT NULL DEFAULT 0,
			platform_enabled_draft INTEGER,
			is_hot INTEGER NOT NULL DEFAULT 0,
			is_hot_draft INTEGER,
			is_featured INTEGER NOT NULL DEFAULT 0,
			is_featured_draft INTEGER,
			is_new INTEGER NOT NULL DEFAULT 0,
			is_new_draft INTEGER,
			supported_countries TEXT NOT NULL DEFAULT '',
			supported_currencies TEXT NOT NULL DEFAULT '',
			supported_devices TEXT NOT NULL DEFAULT '',
			raw_metadata TEXT,
			metadata_hash TEXT NOT NULL DEFAULT '',
			last_synced_at DATETIME,
			badge TEXT NOT NULL DEFAULT '',
			player_count INTEGER NOT NULL DEFAULT 0,
			hot_score INTEGER NOT NULL DEFAULT 0,
			launch_url TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			status_draft INTEGER,
			sort INTEGER NOT NULL DEFAULT 0,
			sort_draft INTEGER,
			maintenance_message_draft TEXT,
			category_ids_draft TEXT,
			config_status TEXT NOT NULL DEFAULT 'PUBLISHED',
			config_version INTEGER NOT NULL DEFAULT 1,
			config_published_at DATETIME,
			config_published_by INTEGER NOT NULL DEFAULT 0,
			draft_updated_at DATETIME,
			draft_updated_by INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE(tenant_id,site_id,game_code)
		)`,
		`INSERT INTO game_game_providers (id,tenant_id,site_id,provider_code,name,status,created_at,updated_at) VALUES
			(100,0,0,'ag','AG',1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
			(200,7,9,'ag','AG Tenant',1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		`INSERT INTO game_games (
			id,tenant_id,site_id,game_code,provider_id,provider_game_code,name,name_i18n,slug,category,
			game_type,cover_url,local_thumbnail_url,raw_metadata,metadata_hash,provider_enabled,platform_enabled,status,sort,config_status,config_version,created_at,updated_at
		) VALUES (
			1001,0,0,'ag-demo',100,'demo-1','Demo','{}','demo','slots',
			'SLOT','','','{}','hash',1,1,1,5,'PUBLISHED',1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AutoMigrate(&model.GameAdminOperationLogs{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newCategoryWriteScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`CREATE TABLE game_game_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		category_key TEXT NOT NULL,
		label TEXT NOT NULL DEFAULT '',
		name_zh TEXT NOT NULL DEFAULT '',
		name_en TEXT NOT NULL DEFAULT '',
		short_label TEXT NOT NULL DEFAULT '',
		icon_url TEXT NOT NULL DEFAULT '',
		banner_url TEXT NOT NULL DEFAULT '',
		sort INTEGER NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		UNIQUE(tenant_id,site_id,category_key)
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key,label,name_zh,name_en,short_label,sort,status,created_at,updated_at) VALUES
		(100,0,0,'slots','Slots','Slots','Slots','Slots',1,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
		(300,1,0,'shared','Shared','Shared','Shared','Shared',2,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.GameAdminOperationLogs{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newCategoryListScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_game_categories (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, category_key TEXT NOT NULL, label TEXT NOT NULL, name_zh TEXT NOT NULL, name_en TEXT NOT NULL, short_label TEXT NOT NULL, icon_url TEXT NOT NULL, banner_url TEXT NOT NULL, sort INTEGER NOT NULL, status INTEGER NOT NULL, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_game_category_relations (id INTEGER PRIMARY KEY, tenant_id INTEGER NOT NULL, site_id INTEGER NOT NULL, category_id INTEGER NOT NULL, game_id INTEGER NOT NULL, deleted_at DATETIME)`,
		`INSERT INTO game_game_categories (id,tenant_id,site_id,category_key,label,name_zh,name_en,short_label,icon_url,banner_url,sort,status,created_at,updated_at) VALUES
			(1,0,0,'slots','Slots','Slots','Slots','S','','',1,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
			(2,1,101,'vip','VIP','VIP','VIP','V','','',2,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
			(3,2,202,'live','Live','Live','Live','L','','',3,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`,
		`INSERT INTO game_game_category_relations (id,tenant_id,site_id,category_id,game_id) VALUES
			(10,0,0,1,1001),
			(11,1,101,1,2001),
			(12,1,101,2,2002)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func newGameAdminScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
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
	if err := db.Exec(`CREATE TABLE game_lobby_sections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		site_id INTEGER NOT NULL DEFAULT 0,
		section_code TEXT NOT NULL,
		section_type TEXT NOT NULL,
		title_zh TEXT NOT NULL,
		title_en TEXT NOT NULL,
		display_style TEXT NOT NULL,
		source_type TEXT NOT NULL,
		source_ref TEXT NOT NULL,
		max_items INTEGER NOT NULL DEFAULT 12,
		sort_order INTEGER NOT NULL DEFAULT 0,
		enabled INTEGER NOT NULL DEFAULT 1,
		supported_countries TEXT NOT NULL,
		supported_currencies TEXT NOT NULL,
		config_json TEXT,
		draft_payload TEXT,
		publish_status TEXT NOT NULL DEFAULT 'PUBLISHED',
		config_version INTEGER NOT NULL DEFAULT 1,
		published_at DATETIME,
		published_by INTEGER NOT NULL DEFAULT 0,
		draft_updated_at DATETIME,
		draft_updated_by INTEGER NOT NULL DEFAULT 0
	)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}
