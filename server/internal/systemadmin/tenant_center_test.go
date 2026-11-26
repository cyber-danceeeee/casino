package systemadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantCenterRejectsTenantAdministrator(t *testing.T) {
	service, _ := newTenantCenterTestService(t)
	actor := Actor{ID: 9, TenantID: 10, SiteID: 20, Platform: false}
	if _, err := service.ListBrandDeliveries(Query{}, actor); ErrorCode(err) != "PERMISSION_DENIED" {
		t.Fatalf("delivery list error = %v", err)
	}
	if _, err := service.TenantPackages(actor); ErrorCode(err) != "PERMISSION_DENIED" {
		t.Fatalf("package list error = %v", err)
	}
	if _, err := service.TenantAudit(10, Query{}, actor); ErrorCode(err) != "PERMISSION_DENIED" {
		t.Fatalf("tenant audit error = %v", err)
	}
}

func TestBrandDeliveryRecordNeverContainsTemporaryPassword(t *testing.T) {
	service, _ := newTenantCenterTestService(t)
	result, err := service.ListBrandDeliveries(Query{}, Actor{ID: 1, Platform: true})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(encoded)), "password") || strings.Contains(string(encoded), "OneTimeSecret") {
		t.Fatalf("delivery response exposed a password: %s", encoded)
	}
}

func TestTenantAuditFindsPlatformActionByTargetOwnership(t *testing.T) {
	service, _ := newTenantCenterTestService(t)
	result, err := service.TenantAudit(10, Query{}, Actor{ID: 1, Platform: true})
	if err != nil {
		t.Fatal(err)
	}
	rows, ok := result.List.(*[]AuditLogRow)
	if !ok || len(*rows) != 1 {
		t.Fatalf("tenant audit rows = %#v", result.List)
	}
	if (*rows)[0].Action != "system.domain.create" || (*rows)[0].TargetID != 30 {
		t.Fatalf("unexpected audit row: %#v", (*rows)[0])
	}
}

func newTenantCenterTestService(t *testing.T) (*Service, *gorm.DB) {
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
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY, tenant_code TEXT, tenant_name TEXT, package_code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_name TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY, tenant_id INTEGER, username TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_domains (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_brand_deliveries (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, admin_id INTEGER, admin_username TEXT, primary_domain TEXT, h5_url TEXT, admin_url TEXT, domain_status TEXT, initial_config_version INTEGER, status TEXT, delivered_by INTEGER, delivered_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY, operator_id INTEGER, tenant_id INTEGER, site_id INTEGER, operator_username TEXT, permission TEXT, module TEXT, action TEXT, result TEXT, target_type TEXT, target_id INTEGER, reason TEXT, ip_address TEXT, correlation_id TEXT, created_at DATETIME, deleted_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	current := time.Now()
	if err := db.Exec(`INSERT INTO game_saas_tenants VALUES (10,'demo-mx','Demo Mexico','PROFESSIONAL',NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_sites VALUES (20,10,'Demo Mexico Main',NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_admin_accounts VALUES (40,10,'owner.mx',NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_domains VALUES (30,10,20,NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_brand_deliveries VALUES (50,10,20,40,'owner.mx','demo.mx','https://demo.mx','https://admin.demo.mx/tenant/login','PENDING',1,'DELIVERED',1,?,NULL)`, current).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_admin_operation_logs VALUES (60,1,0,0,'platform.root','system:domain:create','system','system.domain.create','SUCCESS','domain',30,'provision','','trace-1',?,NULL)`, current).Error; err != nil {
		t.Fatal(err)
	}
	return New(db), db
}
