package systemadmin

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"casino/internal/config"

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

func TestAdminLoginDoesNotMaskSchemaFailureAsInvalidCredentials(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE game_admin_accounts (id INTEGER PRIMARY KEY, username TEXT, password_hash TEXT, status INTEGER, deleted_at DATETIME)").Error; err != nil {
		t.Fatal(err)
	}
	_, err = New(db).AdminLogin(LoginInput{Username: "system-admin", Password: "StrongPassword9!"}, "127.0.0.1", "test", "trace")
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
