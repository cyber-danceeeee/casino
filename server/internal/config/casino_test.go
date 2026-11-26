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
