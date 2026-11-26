package onboardingemail

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"casino/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type recordingSESClient struct {
	input *sesv2.SendEmailInput
}

func (c *recordingSESClient) SendEmail(_ context.Context, input *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	c.input = input
	return &sesv2.SendEmailOutput{MessageId: aws.String("message-1")}, nil
}

func TestProvisionedPayloadEncryptionDoesNotExposePassword(t *testing.T) {
	t.Setenv("CASINO_NOTIFICATION_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	payload := ProvisionedPayload{
		BrandName: "Demo", TenantCode: "demo", AdminUsername: "owner",
		TemporaryPassword: "StrongPassword9!", AdminURL: "https://admin.example.com", Language: "en",
	}
	encrypted, err := EncryptPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encrypted, payload.TemporaryPassword) {
		t.Fatal("encrypted payload exposed the temporary password")
	}
	decrypted, err := DecryptPayload(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != payload {
		t.Fatalf("decrypted payload = %#v", decrypted)
	}
}

func TestProvisionedMessageUsesRequestedLanguage(t *testing.T) {
	payload := ProvisionedPayload{BrandName: "Demo", TemporaryPassword: "StrongPassword9!", Language: "zh-CN"}
	message := ProvisionedMessage("owner@example.com", payload)
	if !strings.Contains(message.Subject, "开户信息") || !strings.Contains(message.Body, payload.TemporaryPassword) ||
		!strings.Contains(message.Body, noReplyNoticeEnglish) || !strings.Contains(message.Body, noReplyNoticeChinese) {
		t.Fatalf("unexpected localized message: %#v", message)
	}
}

func TestSESSenderUsesConfiguredIdentityAndNoReplyNotice(t *testing.T) {
	value := config.Email{
		Enabled: true, Provider: "AWS_SES", From: "support@soranopro.com", Region: "ap-northeast-1",
		AccessKeyID: "access-key", SecretAccessKey: "secret-key",
	}
	client := &recordingSESClient{}
	sender := &noticeSender{next: &SESSender{config: value, client: client}}
	if err := sender.Send(context.Background(), Message{To: "owner@example.com", Subject: "Account", Body: "Ready"}); err != nil {
		t.Fatal(err)
	}
	if client.input == nil || aws.ToString(client.input.FromEmailAddress) != value.From {
		t.Fatalf("unexpected SES input: %#v", client.input)
	}
	body := aws.ToString(client.input.Content.Simple.Body.Text.Data)
	if !strings.Contains(body, noReplyNoticeEnglish) || !strings.Contains(body, noReplyNoticeChinese) {
		t.Fatalf("missing no-reply notice: %q", body)
	}
}

func TestSESConfigurationRequiresStaticCredentials(t *testing.T) {
	value := config.Email{Enabled: true, Provider: "AWS_SES", From: "support@soranopro.com", Region: "ap-northeast-1"}
	if IsConfigured(value) {
		t.Fatal("SES configuration should require access credentials")
	}
	value.AccessKeyID = "access-key"
	value.SecretAccessKey = "secret-key"
	if !IsConfigured(value) {
		t.Fatal("complete SES configuration was rejected")
	}
}
