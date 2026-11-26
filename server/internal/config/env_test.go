package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDeploymentEnvironment(t *testing.T) {
	previous := config
	t.Cleanup(func() { config = previous })
	dir := t.TempDir()
	filename := filepath.Join(dir, "casino.yml")
	if err := os.WriteFile(filename, []byte("app:\n  env: dev\n  staticDir: ../static\nstorage:\n  oss:\n    accessKeyID: inherited\nnotification:\n  fcm:\n    credentialsFile: old.json\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CASINO_APP_ENV", "prod")
	t.Setenv("CASINO_MYSQL_DSN", "environment-dsn")
	t.Setenv("CASINO_MFA_ENCRYPTION_KEY", "environment-key")
	t.Setenv("CASINO_OSS_ACCESS_KEY_ID", "")
	t.Setenv("CASINO_FCM_ENABLED", "true")
	t.Setenv("CASINO_FCM_CREDENTIALS_FILE", "new.json")
	if err := Init(filename); err != nil {
		t.Fatal(err)
	}
	if Get().App.Env != "prod" || Get().Database.Mysql.Dsn != "environment-dsn" || Get().AdminAuth.MFAEncryptionKey != "environment-key" {
		t.Fatal("deployment overrides not applied")
	}
	if Get().Storage.OSS.AccessKeyID != "" {
		t.Fatal("empty override did not clear the credential")
	}
	if !Get().Notification.FCM.Enabled || Get().Notification.FCM.CredentialsFile != filepath.Join(dir, "new.json") {
		t.Fatal("FCM environment/path resolution failed")
	}
	if Get().App.StaticDir != filepath.Clean(filepath.Join(dir, "../static")) {
		t.Fatal("static path resolution failed")
	}
}

func TestInvalidEnvironmentBooleanFails(t *testing.T) {
	t.Setenv("CASINO_EMAIL_ENABLED", "invalid")
	if err := applyEnvironment(&Config{}); err == nil {
		t.Fatal("invalid boolean accepted")
	}
}
