package systemadmin

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateTenantUpdatesPackageAndDomains(t *testing.T) {
	service, db := newTenantManagementTestService(t)
	primary := "new.example.com"
	h5 := "casino.new.example.com"
	admin := "admin.new.example.com"

	if err := service.UpdateTenant(10, TenantInput{
		TenantName:      "Demo Updated",
		CompanyName:     "Demo Co",
		PackageCode:     "ENTERPRISE",
		CountryCode:     "PK",
		DefaultCurrency: "PKR",
		DefaultLanguage: "en",
		Timezone:        "Asia/Karachi",
		PrimaryDomain:   &primary,
		H5Domain:        &h5,
		AdminDomain:     &admin,
		Reason:          "platform tenant maintenance",
	}, Actor{ID: 7, Username: "system-admin", Platform: true}); err != nil {
		t.Fatalf("UpdateTenant error = %v", err)
	}

	var tenant struct {
		PackageCode string
		PlanCode    string
	}
	if err := db.Table("game_saas_tenants").Select("package_code,plan_code").Where("id=10").Take(&tenant).Error; err != nil {
		t.Fatal(err)
	}
	if tenant.PackageCode != "ENTERPRISE" || tenant.PlanCode != "ENTERPRISE" {
		t.Fatalf("unexpected package/plan: %#v", tenant)
	}

	var site struct {
		PrimaryDomain string
		H5Domain      string
		AdminDomain   string
	}
	if err := db.Table("game_saas_sites").Select("primary_domain,h5_domain,admin_domain").Where("id=20").Take(&site).Error; err != nil {
		t.Fatal(err)
	}
	if site.PrimaryDomain != primary || site.H5Domain != h5 || site.AdminDomain != admin {
		t.Fatalf("site domains not updated: %#v", site)
	}

	var primaryDomain struct {
		PrimaryDomain string `gorm:"column:primary_domain"`
	}
	if err := db.Table("game_saas_tenants t").Select(tenantSelect()).Where("t.id=10 AND t.deleted_at IS NULL").Take(&primaryDomain).Error; err != nil {
		t.Fatal(err)
	}
	if primaryDomain.PrimaryDomain != primary {
		t.Fatalf("tenant list primary domain = %q, want %q", primaryDomain.PrimaryDomain, primary)
	}

	var activeDomains int64
	if err := db.Table("game_saas_domains").Where("tenant_id=10 AND deleted_at IS NULL").Count(&activeDomains).Error; err != nil {
		t.Fatal(err)
	}
	if activeDomains != 3 {
		t.Fatalf("expected 3 active domain records, got %d", activeDomains)
	}
	var oldDeleted int64
	if err := db.Table("game_saas_domains").Where("domain='old.example.com' AND deleted_at IS NOT NULL").Count(&oldDeleted).Error; err != nil {
		t.Fatal(err)
	}
	if oldDeleted != 1 {
		t.Fatalf("old primary domain was not soft deleted")
	}
	var stalePrimary int64
	if err := db.Table("game_saas_domains").Where("tenant_id=10 AND domain_type='PRIMARY' AND domain<>? AND deleted_at IS NULL", primary).Count(&stalePrimary).Error; err != nil {
		t.Fatal(err)
	}
	if stalePrimary != 0 {
		t.Fatalf("stale primary domains are still active")
	}
}

func TestDeleteTenantSoftDeletesDeliveryResources(t *testing.T) {
	service, db := newTenantManagementTestService(t)
	if err := service.DeleteTenant(10, "tenant contract terminated", Actor{ID: 7, Username: "system-admin", Platform: true}); err != nil {
		t.Fatalf("DeleteTenant error = %v", err)
	}

	for table, where := range map[string]string{
		"game_saas_tenants":          "id=10",
		"game_saas_sites":            "tenant_id=10",
		"game_saas_domains":          "tenant_id=10",
		"game_admin_accounts":        "tenant_id=10",
		"game_saas_brand_deliveries": "tenant_id=10",
	} {
		var count int64
		if err := db.Table(table).Where(where + " AND deleted_at IS NOT NULL").Count(&count).Error; err != nil {
			t.Fatalf("%s count error = %v", table, err)
		}
		if count == 0 {
			t.Fatalf("%s was not soft deleted", table)
		}
	}

	var revoked int64
	if err := db.Table("game_admin_sessions").Where("admin_id=30 AND revoked_at IS NOT NULL").Count(&revoked).Error; err != nil {
		t.Fatal(err)
	}
	if revoked != 1 {
		t.Fatalf("admin sessions were not revoked")
	}
}

func newTenantManagementTestService(t *testing.T) (*Service, *gorm.DB) {
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
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY, tenant_code TEXT, tenant_name TEXT, company_name TEXT, contact_name TEXT, contact_email TEXT, contact_phone TEXT, plan_code TEXT, country_code TEXT, package_code TEXT, default_currency TEXT, default_language TEXT, timezone TEXT, expires_at DATETIME, enabled INTEGER, status TEXT, remark TEXT, maintenance_enabled INTEGER, maintenance_message TEXT, version INTEGER, updated_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY, tenant_id INTEGER, primary_domain TEXT, h5_domain TEXT, admin_domain TEXT, enabled INTEGER, status TEXT, updated_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_domains (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, domain TEXT, domain_key TEXT, domain_type TEXT, is_primary INTEGER, enabled INTEGER, status TEXT, remark TEXT, created_by INTEGER, updated_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY, tenant_id INTEGER, username TEXT, status INTEGER, last_login_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_sessions (id INTEGER PRIMARY KEY, admin_id INTEGER, revoked_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE game_games (id INTEGER PRIMARY KEY, tenant_id INTEGER, platform_enabled INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE game_users (id INTEGER PRIMARY KEY, tenant_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_brand_deliveries (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, status TEXT, domain_status TEXT, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY, operator_id INTEGER, tenant_id INTEGER, site_id INTEGER, operator_username TEXT, permission TEXT, module TEXT, action TEXT, result TEXT, target_type TEXT, target_id INTEGER, before_state TEXT, after_state TEXT, reason TEXT, ip_address TEXT, user_agent TEXT, correlation_id TEXT, currency TEXT, created_at DATETIME, updated_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	nowValue := time.Now()
	if err := db.Exec(`INSERT INTO game_saas_tenants (id,tenant_code,tenant_name,company_name,plan_code,country_code,package_code,default_currency,default_language,timezone,enabled,status,maintenance_enabled,maintenance_message,version,created_at,updated_at) VALUES (10,'demo','Demo','Demo Co','STANDARD','MX','STANDARD','MXN','es','America/Mexico_City',1,'ACTIVE',0,'',1,?,?)`, nowValue, nowValue).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_sites (id,tenant_id,primary_domain,h5_domain,admin_domain,enabled,status,created_at,updated_at) VALUES (20,10,'old.example.com','casino.old.example.com','admin.old.example.com',1,'ONLINE',?,?)`, nowValue, nowValue).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id      int
		domain  string
		kind    string
		primary int
	}{
		{100, "old.example.com", "PRIMARY", 1},
		{101, "casino.old.example.com", "H5", 1},
		{102, "admin.old.example.com", "ADMIN", 0},
	} {
		if err := db.Exec(`INSERT INTO game_saas_domains (id,tenant_id,site_id,domain,domain_key,domain_type,is_primary,enabled,status,created_at,updated_at) VALUES (?,10,20,?,?,?,?,1,'ACTIVE',?,?)`, row.id, row.domain, row.domain, row.kind, row.primary, nowValue, nowValue).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec(`INSERT INTO game_admin_accounts (id,tenant_id,username,status,updated_at) VALUES (30,10,'owner',1,?)`, nowValue).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_admin_sessions (id,admin_id,updated_at) VALUES (40,30,?)`, nowValue).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_brand_deliveries (id,tenant_id,site_id,status,domain_status,updated_at) VALUES (50,10,20,'DELIVERED','PENDING',?)`, nowValue).Error; err != nil {
		t.Fatal(err)
	}
	return New(db), db
}
