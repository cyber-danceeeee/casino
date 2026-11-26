package systemadmin

import (
	"casino/internal/singlesite"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func singleSiteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	old, active := singlesite.Scope()
	t.Cleanup(func() { singlesite.Restore(old, active) })
	singlesite.Restore(old, false)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		`CREATE TABLE casino_site_binding(id INTEGER PRIMARY KEY,tenant_id INTEGER,site_id INTEGER)`, `INSERT INTO casino_site_binding VALUES(1,0,0)`,
		`CREATE TABLE game_saas_tenants(id INTEGER PRIMARY KEY,deleted_at DATETIME)`, `INSERT INTO game_saas_tenants VALUES(10,NULL)`,
		`CREATE TABLE game_saas_sites(id INTEGER PRIMARY KEY,tenant_id INTEGER,deleted_at DATETIME)`, `INSERT INTO game_saas_sites VALUES(20,10,NULL)`,
		`CREATE TABLE game_users(id INTEGER PRIMARY KEY,tenant_id INTEGER,site_id INTEGER)`, `CREATE TABLE game_admin_accounts(id INTEGER PRIMARY KEY,username TEXT,deleted_at DATETIME)`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}
func TestSingleSiteAdoptsExistingSiteAcrossRestarts(t *testing.T) {
	db := singleSiteTestDB(t)
	for i := 0; i < 2; i++ {
		if err := New(db).EnsureSingleSite(); err != nil {
			t.Fatal(err)
		}
	}
	scope, ok := singlesite.Scope()
	if !ok || scope.TenantID != 10 || scope.SiteID != 20 || scope.Platform {
		t.Fatalf("wrong binding: %+v", scope)
	}
	var count int64
	db.Table("game_saas_sites").Count(&count)
	if count != 1 {
		t.Fatal("restart created another site")
	}
}
func TestSingleSiteRefusesAmbiguousOrUnsafeConversion(t *testing.T) {
	cases := map[string]string{"multiple": `INSERT INTO game_saas_sites VALUES(21,10,NULL)`, "mismatched": `UPDATE casino_site_binding SET tenant_id=10,site_id=99`, "legacy_players": `INSERT INTO game_users VALUES(1,0,0)`, "ambiguous_login": `INSERT INTO game_admin_accounts VALUES(1,'admin',NULL),(2,'admin',NULL)`}
	for name, sql := range cases {
		t.Run(name, func(t *testing.T) {
			db := singleSiteTestDB(t)
			db.Exec(sql)
			if err := New(db).EnsureSingleSite(); err == nil {
				t.Fatal("unsafe conversion accepted")
			}
			if singlesite.Active() {
				t.Fatal("bound an unsafe database")
			}
		})
	}
}
func TestSingleSiteURL(t *testing.T) {
	for _, value := range []string{"https://h5.example.com", "http://127.0.0.1:18080", "h5.example.com", ""} {
		if _, err := singleSiteURL(value); err != nil {
			t.Fatal(value, err)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "https://user:password@example.com", "https://example.com/subpath", "https://example.com?site=2"} {
		if _, err := singleSiteURL(value); err == nil {
			t.Fatal("invalid URL accepted:", value)
		}
	}
}
func TestSingleSiteNavigationRetainsOperations(t *testing.T) {
	for _, code := range []string{"tenants", "sites", "tenant-packages", "brand-deliveries", "create-brand"} {
		if singleSiteMenuAllowed(code, "") {
			t.Fatal("SaaS menu retained:", code)
		}
	}
	for _, code := range []string{"brand-config", "game-sync", "user-list", "admins", "roles", "country-config"} {
		if !singleSiteMenuAllowed(code, "") {
			t.Fatal("operation removed:", code)
		}
	}
}
