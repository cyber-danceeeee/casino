package config

import (
	"fmt"
	"os"
	"strconv"
)

// applyEnvironment overrides deployment settings without writing secrets into YAML.
// LookupEnv deliberately allows empty values to clear optional credentials.
func applyEnvironment(c *Config) error {
	strings := map[string]*string{
		"CASINO_APP_ENV":                 &c.App.Env,
		"CASINO_STATIC_DIR":              &c.App.StaticDir,
		"CASINO_MYSQL_DSN":               &c.Database.Mysql.Dsn,
		"CASINO_REDIS_DSN":               &c.Redis.Dsn,
		"CASINO_MFA_ENCRYPTION_KEY":      &c.AdminAuth.MFAEncryptionKey,
		"CASINO_STORAGE_DRIVER":          &c.Storage.Driver,
		"CASINO_OSS_REGION":              &c.Storage.OSS.Region,
		"CASINO_OSS_ENDPOINT":            &c.Storage.OSS.Endpoint,
		"CASINO_OSS_BUCKET":              &c.Storage.OSS.Bucket,
		"CASINO_OSS_ACCESS_KEY_ID":       &c.Storage.OSS.AccessKeyID,
		"CASINO_OSS_ACCESS_KEY_SECRET":   &c.Storage.OSS.AccessKeySecret,
		"CASINO_OSS_SECURITY_TOKEN":      &c.Storage.OSS.SecurityToken,
		"CASINO_OSS_PUBLIC_BASE_URL":     &c.Storage.OSS.PublicBaseURL,
		"CASINO_EMAIL_PROVIDER":          &c.Notification.Email.Provider,
		"CASINO_EMAIL_FROM":              &c.Notification.Email.From,
		"CASINO_EMAIL_REGION":            &c.Notification.Email.Region,
		"CASINO_EMAIL_ACCESS_KEY_ID":     &c.Notification.Email.AccessKeyID,
		"CASINO_EMAIL_SECRET_ACCESS_KEY": &c.Notification.Email.SecretAccessKey,
		"CASINO_RESEND_API_KEY":          &c.Notification.Email.APIKey,
		"CASINO_FCM_PROJECT_ID":          &c.Notification.FCM.ProjectID,
		"CASINO_FCM_CREDENTIALS_FILE":    &c.Notification.FCM.CredentialsFile,
		"GAME_AG_MERCHANT_CODE":          &c.Game.AG.MerchantCode,
		"GAME_AG_SECRET_KEY":             &c.Game.AG.SecretKey,
		"GAME_AG_API_BASE_URL":           &c.Game.AG.APIBaseURL,
	}
	for name, target := range strings {
		if value, ok := os.LookupEnv(name); ok {
			*target = value
		}
	}
	bools := map[string]*bool{
		"CASINO_EMAIL_ENABLED": &c.Notification.Email.Enabled,
		"CASINO_FCM_ENABLED":   &c.Notification.FCM.Enabled,
		"GAME_AG_ENABLED":      &c.Game.AG.Enabled,
	}
	for name, target := range bools {
		if value, ok := os.LookupEnv(name); ok {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("%s must be a boolean", name)
			}
			*target = parsed
		}
	}
	return nil
}
