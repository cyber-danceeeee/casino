package onboardingemail

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"casino/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

const (
	TemplateTenantProvisioned = "TENANT_PROVISIONED"
	EventEncryptedPayloadKey  = "encryptedOnboardingPayload"
	resendAPIURL              = "https://api.resend.com/emails"
	noReplyNoticeEnglish      = "This is an automated email. The sender address cannot receive replies, so please do not reply to this message."
	noReplyNoticeChinese      = "本邮件由系统自动发送，发件地址无法接收回复，请勿直接回复本邮件。"
)

type Job struct {
	ID               uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	EventID          *string    `gorm:"column:event_id;type:varchar(64);uniqueIndex:uk_tenant_onboarding_email_event"`
	TenantID         uint64     `gorm:"column:tenant_id;type:bigint unsigned;not null;index:idx_tenant_onboarding_email_scope,priority:1"`
	SiteID           uint64     `gorm:"column:site_id;type:bigint unsigned;not null;index:idx_tenant_onboarding_email_scope,priority:2"`
	RecipientEmail   string     `gorm:"column:recipient_email;type:varchar(190);not null"`
	TemplateCode     string     `gorm:"column:template_code;type:varchar(64);not null"`
	EncryptedPayload string     `gorm:"column:encrypted_payload;type:text;not null"`
	Status           string     `gorm:"column:status;type:varchar(24);not null;default:PENDING;index:idx_tenant_onboarding_email_scope,priority:3;index:idx_tenant_onboarding_email_queue,priority:1"`
	AttemptCount     uint       `gorm:"column:attempt_count;type:int unsigned;not null;default:0"`
	LastError        string     `gorm:"column:last_error;type:varchar(500);not null;default:''"`
	IdempotencyKey   string     `gorm:"column:idempotency_key;type:varchar(191);not null;uniqueIndex:uk_tenant_onboarding_email_idempotency"`
	SentAt           *time.Time `gorm:"column:sent_at;type:datetime(3)"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:datetime(3);not null;index:idx_tenant_onboarding_email_scope,priority:4"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:datetime(3);not null;index:idx_tenant_onboarding_email_queue,priority:2"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;type:datetime(3)"`
}

func (Job) TableName() string { return "game_tenant_onboarding_email_jobs" }

type ProvisionedPayload struct {
	BrandName         string `json:"brandName"`
	TenantCode        string `json:"tenantCode"`
	CountryCode       string `json:"countryCode"`
	CurrencyCode      string `json:"currencyCode"`
	AdminUsername     string `json:"adminUsername"`
	TemporaryPassword string `json:"temporaryPassword"`
	AdminURL          string `json:"adminUrl"`
	H5URL             string `json:"h5Url"`
	Language          string `json:"language"`
}

type Message struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(context.Context, Message) error
}

type noticeSender struct {
	next Sender
}

func (s *noticeSender) Send(ctx context.Context, message Message) error {
	if s == nil || s.next == nil {
		return errors.New("email sender is not configured")
	}
	return s.next.Send(ctx, withNoReplyNotice(message))
}

func NewSender(value config.Email) Sender {
	var sender Sender
	switch normalizedProvider(value.Provider) {
	case "AWS_SES":
		sender = NewSESSender(value)
	default:
		sender = NewResendSender(value)
	}
	return &noticeSender{next: sender}
}

func IsConfigured(value config.Email) bool {
	if !value.Enabled || strings.TrimSpace(value.From) == "" {
		return false
	}
	switch normalizedProvider(value.Provider) {
	case "AWS_SES":
		return strings.TrimSpace(value.Region) != "" &&
			strings.TrimSpace(value.AccessKeyID) != "" &&
			strings.TrimSpace(value.SecretAccessKey) != ""
	case "RESEND":
		return strings.TrimSpace(value.APIKey) != ""
	default:
		return false
	}
}

func normalizedProvider(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "SES", "AWS-SES", "AWS_SES":
		return "AWS_SES"
	case "", "RESEND":
		return "RESEND"
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}

type sesEmailClient interface {
	SendEmail(context.Context, *sesv2.SendEmailInput, ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

type SESSender struct {
	config config.Email
	client sesEmailClient
}

func NewSESSender(value config.Email) *SESSender {
	awsConfig := aws.Config{
		Region: strings.TrimSpace(value.Region),
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(value.AccessKeyID), strings.TrimSpace(value.SecretAccessKey), "",
		)),
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
	return &SESSender{config: value, client: sesv2.NewFromConfig(awsConfig)}
}

func (s *SESSender) Send(ctx context.Context, message Message) error {
	if s == nil || !s.config.Enabled {
		return errors.New("email delivery is disabled")
	}
	if normalizedProvider(s.config.Provider) != "AWS_SES" {
		return fmt.Errorf("unsupported email provider %s", strings.TrimSpace(s.config.Provider))
	}
	if !IsConfigured(s.config) || s.client == nil {
		return errors.New("AWS SES email delivery is not configured")
	}
	to := strings.TrimSpace(message.To)
	if _, err := parseEmailAddress(to); err != nil {
		return err
	}
	charset := "UTF-8"
	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(strings.TrimSpace(s.config.From)),
		Destination:      &types.Destination{ToAddresses: []string{to}},
		Content: &types.EmailContent{Simple: &types.Message{
			Subject: &types.Content{Charset: &charset, Data: aws.String(strings.TrimSpace(message.Subject))},
			Body:    &types.Body{Text: &types.Content{Charset: &charset, Data: aws.String(message.Body)}},
		}},
	})
	if err != nil {
		return fmt.Errorf("AWS SES send email: %w", err)
	}
	return nil
}

func parseEmailAddress(value string) (string, error) {
	address, err := mail.ParseAddress(value)
	if err != nil || !strings.EqualFold(strings.TrimSpace(address.Address), strings.TrimSpace(value)) {
		return "", errors.New("recipient email is invalid")
	}
	return address.Address, nil
}

type ResendSender struct {
	config config.Email
	client *http.Client
}

func NewResendSender(value config.Email) *ResendSender {
	return &ResendSender{config: value, client: &http.Client{Timeout: 15 * time.Second}}
}

func (s *ResendSender) Send(ctx context.Context, message Message) error {
	if !s.config.Enabled {
		return errors.New("email delivery is disabled")
	}
	if provider := strings.ToUpper(strings.TrimSpace(s.config.Provider)); provider != "" && provider != "RESEND" {
		return fmt.Errorf("unsupported email provider %s", provider)
	}
	apiKey := strings.TrimSpace(s.config.APIKey)
	from := strings.TrimSpace(s.config.From)
	if apiKey == "" || from == "" {
		return errors.New("resend email delivery is not configured")
	}
	payload, err := json.Marshal(map[string]any{
		"from": from, "to": []string{strings.TrimSpace(message.To)},
		"subject": strings.TrimSpace(message.Subject), "text": message.Body,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("resend email request: %w", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("resend email status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func EncryptPayload(payload ProvisionedPayload) (string, error) {
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	key, err := encryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plain, []byte(TemplateTenantProvisioned))
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func DecryptPayload(value string) (ProvisionedPayload, error) {
	key, err := encryptionKey()
	if err != nil {
		return ProvisionedPayload{}, err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return ProvisionedPayload{}, errors.New("invalid onboarding email payload")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return ProvisionedPayload{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(sealed) < gcm.NonceSize() {
		return ProvisionedPayload{}, errors.New("invalid onboarding email payload")
	}
	plain, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], []byte(TemplateTenantProvisioned))
	if err != nil {
		return ProvisionedPayload{}, errors.New("invalid onboarding email payload")
	}
	var payload ProvisionedPayload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return ProvisionedPayload{}, errors.New("invalid onboarding email payload")
	}
	return payload, nil
}

func ProvisionedMessage(to string, payload ProvisionedPayload) Message {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(payload.Language)), "zh") {
		return withNoReplyNotice(Message{To: to, Subject: "品牌后台开户信息 - " + payload.BrandName, Body: fmt.Sprintf(
			"您的品牌后台已创建。\n\n品牌：%s\n租户编码：%s\n国家/币种：%s / %s\n管理员账号：%s\n一次性临时密码：%s\n品牌后台：%s\nH5 地址：%s\n\n首次登录后必须立即修改临时密码，请勿转发本邮件。",
			payload.BrandName, payload.TenantCode, payload.CountryCode, payload.CurrencyCode, payload.AdminUsername, payload.TemporaryPassword, payload.AdminURL, payload.H5URL,
		)})
	}
	return withNoReplyNotice(Message{To: to, Subject: "Your brand administrator account - " + payload.BrandName, Body: fmt.Sprintf(
		"Your brand administrator account is ready.\n\nBrand: %s\nTenant code: %s\nCountry/Currency: %s / %s\nAdministrator username: %s\nOne-time temporary password: %s\nBrand admin: %s\nH5: %s\n\nYou must change the temporary password after your first sign-in. Do not forward this email.",
		payload.BrandName, payload.TenantCode, payload.CountryCode, payload.CurrencyCode, payload.AdminUsername, payload.TemporaryPassword, payload.AdminURL, payload.H5URL,
	)})
}

func withNoReplyNotice(message Message) Message {
	if strings.Contains(message.Body, noReplyNoticeEnglish) && strings.Contains(message.Body, noReplyNoticeChinese) {
		return message
	}
	message.Body = strings.TrimSpace(message.Body) + "\n\n" + noReplyNoticeEnglish + "\n" + noReplyNoticeChinese
	return message
}

func encryptionKey() ([]byte, error) {
	value := config.NotificationEncryptionKey()
	key, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(key) != 32 {
		return nil, errors.New("notification encryption key is not configured")
	}
	return key, nil
}
