package config

import (
	"strings"
	"testing"
)

func TestShowRedactsEmailCredentials(t *testing.T) {
	previous := config
	t.Cleanup(func() { config = previous })
	config = &Config{Notification: Notification{Email: Email{
		Enabled: true, Provider: "AWS_SES", From: "support@example.com", Region: "ap-northeast-1",
		AccessKeyID: "visible-access-key", SecretAccessKey: "visible-secret-key", APIKey: "visible-api-key",
	}}}

	shown := Show()
	for _, secret := range []string{"visible-access-key", "visible-secret-key", "visible-api-key"} {
		if strings.Contains(shown, secret) {
			t.Fatalf("configuration output exposed %q", secret)
		}
	}
}

func TestShowRedactsProviderCredentials(t *testing.T) {
	previous := config
	t.Cleanup(func() { config = previous })
	config = &Config{}
	config.Game.SoftAggregator.APILogin = "sensitive-login"
	config.Game.SoftAggregator.APIPassword = "sensitive-api-password"
	config.Game.SoftAggregator.SaltKey = "sensitive-signature-key"
	config.Game.AG.SecretKey = "sensitive-ag-key"
	config.Game.AllInTop.ClientID = "sensitive-client-id"
	config.Game.AllInTop.ClientSecret = "sensitive-client-secret"
	shown := Show()
	for _, value := range []string{"sensitive-login", "sensitive-api-password", "sensitive-signature-key", "sensitive-ag-key", "sensitive-client-id", "sensitive-client-secret"} {
		if strings.Contains(shown, value) {
			t.Fatal("configuration output exposed a provider credential")
		}
	}
}
