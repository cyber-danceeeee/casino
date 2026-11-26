package systemadmin

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"casino/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdminPasswordStrength(t *testing.T) {
	for _, value := range []string{"short", "alllowercase1!", "ALLUPPERCASE1!", "NoNumberHere!", "NoSymbolHere1"} {
		if validatePasswordStrength(value) == nil {
			t.Fatalf("weak password %q was accepted", value)
		}
	}
	if err := validatePasswordStrength("StrongPassword9!"); err != nil {
		t.Fatalf("strong password was rejected: %v", err)
	}
}

func TestRecoveryCodeHashNormalizesFormatting(t *testing.T) {
	formatted := recoveryCodeHash("ABCD-EF12-3456")
	if formatted != recoveryCodeHash("abcd ef12 3456") {
		t.Fatal("recovery code formatting changed its hash")
	}
	if formatted == recoveryCodeHash("ABCD-EF12-3457") {
		t.Fatal("different recovery codes produced the same hash")
	}
}

func TestNavigationPortalScopesAreDisjoint(t *testing.T) {
	parentID := uint64(1)
	rows := []NavigationRow{
		{ID: 1, Code: "system"},
		{ID: 2, ParentID: &parentID, Code: "tenants", RoutePath: "/system/tenants"},
		{ID: 3, ParentID: &parentID, Code: "user-list", RoutePath: "/users/list"},
		{ID: 4, ParentID: &parentID, Code: "sites", RoutePath: "/system/sites"},
	}
	contains := func(values []NavigationRow, code string) bool {
		for _, value := range values {
			if value.Code == code {
				return true
			}
		}
		return false
	}
	platform := filterNavigationPortal(rows, true)
	tenant := filterNavigationPortal(rows, false)
	if !contains(platform, "tenants") || contains(tenant, "tenants") {
		t.Fatal("tenant management portal scope is invalid")
	}
	if contains(platform, "user-list") || !contains(tenant, "user-list") {
		t.Fatal("tenant operations portal scope is invalid")
	}
	if !contains(platform, "sites") || !contains(tenant, "sites") {
		t.Fatal("scoped site management must be shared by both portals")
	}
}

func TestAdminLoginDoesNotCrossPortal(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY,username TEXT,display_name TEXT,tenant_id INTEGER,site_id INTEGER,password_hash TEXT,mfa_enabled INTEGER,mfa_secret_encrypted TEXT,must_change_password INTEGER,status INTEGER,failed_login_count INTEGER,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY,tenant_code TEXT,tenant_name TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY,site_code TEXT,site_name TEXT,deleted_at DATETIME)`,
		`CREATE TABLE game_admin_login_logs (id INTEGER PRIMARY KEY AUTOINCREMENT,admin_id INTEGER,tenant_id INTEGER,site_id INTEGER,username TEXT,success INTEGER,failure_reason TEXT,ip_address TEXT,device TEXT,operating_system TEXT,browser TEXT,user_agent TEXT,correlation_id TEXT,created_at DATETIME,updated_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPassword9!"), bcrypt.MinCost)
	if err := db.Exec(`INSERT INTO game_saas_tenants(id,tenant_code,tenant_name) VALUES(10,'tenant-a','Tenant A')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_saas_sites(id,site_code,site_name) VALUES(20,'main','Main')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO game_admin_accounts(id,username,display_name,tenant_id,site_id,password_hash,mfa_enabled,must_change_password,status,failed_login_count) VALUES(1,'owner','Owner',10,20,?,1,0,1,0)`, string(hash)).Error; err != nil {
		t.Fatal(err)
	}
	service := New(db)
	_, err = service.AdminLogin(LoginInput{Username: "owner", Password: "StrongPassword9!", TenantCode: "tenant-a", PortalType: "PLATFORM"}, "127.0.0.1", "test", "trace")
	if ErrorCode(err) != "INVALID_CREDENTIALS" {
		t.Fatalf("portal mismatch error = %v (%s)", err, ErrorCode(err))
	}
}

func TestTenantAdminLoginRequiresTenantCode(t *testing.T) {
	service := New(nil)
	_, err := service.AdminLogin(LoginInput{Username: "owner", Password: "StrongPassword9!", PortalType: "TENANT"}, "127.0.0.1", "test", "trace")
	if ErrorCode(err) != "TENANT_CODE_REQUIRED" {
		t.Fatalf("missing tenant code error = %v (%s)", err, ErrorCode(err))
	}
}

func TestAdminLoginDoesNotMaskSchemaFailureAsInvalidCredentials(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY, username TEXT, password_hash TEXT, status INTEGER, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	_, err = New(db).AdminLogin(LoginInput{Username: "system-admin", Password: "StrongPassword9!", PortalType: "PLATFORM"}, "127.0.0.1", "test", "trace")
	if ErrorCode(err) != "INTERNAL_ERROR" {
		t.Fatalf("schema failure = %v (%s), want INTERNAL_ERROR", err, ErrorCode(err))
	}
}

func TestTOTPUsesRFC6238Vector(t *testing.T) {
	const secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	code, err := totpCode(secret, time.Unix(59, 0))
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" {
		t.Fatalf("TOTP code = %s, want 287082", code)
	}
	if !verifyTOTP(secret, code, time.Unix(59, 0)) {
		t.Fatal("generated TOTP code was not accepted")
	}
}

func TestMFASecretEncryptionRoundTrip(t *testing.T) {
	t.Cleanup(func() { config.Set(&config.Config{}) })
	key := []byte("0123456789abcdef0123456789abcdef")
	config.Set(&config.Config{AdminAuth: config.AdminAuth{MFAEncryptionKey: base64.StdEncoding.EncodeToString(key)}})

	const secret = "JBSWY3DPEHPK3PXP"
	encrypted, err := encryptMFASecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == secret {
		t.Fatal("MFA secret was stored as plaintext")
	}
	decrypted, err := decryptMFASecret(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != secret {
		t.Fatalf("decrypted secret = %q, want %q", decrypted, secret)
	}
}
