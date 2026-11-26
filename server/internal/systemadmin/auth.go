package systemadmin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // TOTP uses HMAC-SHA1 for authenticator compatibility.
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/config"
)

const (
	authStatusAuthenticated    = "AUTHENTICATED"
	authStatusMFARequired      = "MFA_REQUIRED"
	authStatusMFASetupRequired = "MFA_SETUP_REQUIRED"
	challengePurposeLogin      = "LOGIN"
	challengePurposeSetup      = "SETUP"
	maxMFAAttempts             = 5
)

type adminCredential struct {
	ID                 uint64 `gorm:"column:id"`
	Username           string `gorm:"column:username"`
	DisplayName        string `gorm:"column:display_name"`
	TenantID           uint64 `gorm:"column:tenant_id"`
	SiteID             uint64 `gorm:"column:site_id"`
	TenantCode         string `gorm:"column:tenant_code"`
	TenantName         string `gorm:"column:tenant_name"`
	SiteCode           string `gorm:"column:site_code"`
	SiteName           string `gorm:"column:site_name"`
	PasswordHash       string `gorm:"column:password_hash"`
	MFAEnabled         bool   `gorm:"column:mfa_enabled"`
	MFASecretEncrypted string `gorm:"column:mfa_secret_encrypted"`
	MustChangePassword bool   `gorm:"column:must_change_password"`
	Status             int    `gorm:"column:status"`
}

type mfaChallenge struct {
	ID                 uint64     `gorm:"column:id"`
	AdminID            uint64     `gorm:"column:admin_id"`
	TenantID           uint64     `gorm:"column:tenant_id"`
	SiteID             uint64     `gorm:"column:site_id"`
	TokenHash          string     `gorm:"column:token_hash"`
	Purpose            string     `gorm:"column:purpose"`
	SecretEncrypted    string     `gorm:"column:secret_encrypted"`
	Attempts           uint       `gorm:"column:attempts"`
	ExpiresAt          time.Time  `gorm:"column:expires_at"`
	ConsumedAt         *time.Time `gorm:"column:consumed_at"`
	IP                 string     `gorm:"column:ip_address"`
	UserAgent          string     `gorm:"column:user_agent"`
	AdminUsername      string     `gorm:"column:admin_username"`
	AdminDisplayName   string     `gorm:"column:admin_display_name"`
	TenantCode         string     `gorm:"column:tenant_code"`
	TenantName         string     `gorm:"column:tenant_name"`
	SiteCode           string     `gorm:"column:site_code"`
	SiteName           string     `gorm:"column:site_name"`
	MFASecretEncrypted string     `gorm:"column:admin_mfa_secret_encrypted"`
	MustChangePassword bool       `gorm:"column:must_change_password"`
}

func authConfig() (issuer string, challengeTTL, sessionTTL time.Duration) {
	cfg := config.Get().AdminAuth
	issuer = strings.TrimSpace(cfg.Issuer)
	if issuer == "" {
		issuer = "Casino Admin"
	}
	challengeMinutes := cfg.ChallengeTTLMinutes
	if challengeMinutes <= 0 {
		challengeMinutes = 5
	}
	sessionHours := cfg.SessionTTLHours
	if sessionHours <= 0 {
		sessionHours = 12
	}
	return issuer, time.Duration(challengeMinutes) * time.Minute, time.Duration(sessionHours) * time.Hour
}

func randomToken(size int) (string, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func mfaEncryptionKey() ([]byte, error) {
	value := strings.TrimSpace(config.Get().AdminAuth.MFAEncryptionKey)
	key, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(key) != 32 {
		return nil, NewError("AUTH_CONFIGURATION_ERROR", "administrator authentication is not configured")
	}
	return key, nil
}

func encryptMFASecret(secret string) (string, error) {
	key, err := mfaEncryptionKey()
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
	sealed := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func decryptMFASecret(value string) (string, error) {
	key, err := mfaEncryptionKey()
	if err != nil {
		return "", err
	}
	payload, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", NewError("MFA_CONFIGURATION_INVALID", "MFA configuration is invalid")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(payload) < gcm.NonceSize() {
		return "", NewError("MFA_CONFIGURATION_INVALID", "MFA configuration is invalid")
	}
	plain, err := gcm.Open(nil, payload[:gcm.NonceSize()], payload[gcm.NonceSize():], nil)
	if err != nil {
		return "", NewError("MFA_CONFIGURATION_INVALID", "MFA configuration is invalid")
	}
	return string(plain), nil
}

func generateMFASecret() (string, error) {
	data := make([]byte, 20)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data), nil
}

func totpCode(secret string, timestamp time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}
	counter := uint64(timestamp.Unix() / 30)
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(message)
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := (uint32(digest[offset])&0x7f)<<24 |
		uint32(digest[offset+1])<<16 |
		uint32(digest[offset+2])<<8 |
		uint32(digest[offset+3])
	return fmt.Sprintf("%06d", value%1000000), nil
}

func verifyTOTP(secret, code string, timestamp time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	for drift := -1; drift <= 1; drift++ {
		expected, err := totpCode(secret, timestamp.Add(time.Duration(drift)*30*time.Second))
		if err == nil && subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func parseUserAgent(value string) (string, string, string) {
	lower := strings.ToLower(value)
	device, osName, browser := "Desktop", "Unknown", "Unknown"
	if strings.Contains(lower, "mobile") {
		device = "Mobile"
	}
	switch {
	case strings.Contains(lower, "windows"):
		osName = "Windows"
	case strings.Contains(lower, "mac os"):
		osName = "macOS"
	case strings.Contains(lower, "android"):
		osName = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		osName = "iOS"
	case strings.Contains(lower, "linux"):
		osName = "Linux"
	}
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/"):
		browser = "Safari"
	}
	return device, osName, browser
}

func (s *Service) writeLoginLog(adminID *uint64, tenantID, siteID uint64, username string, success bool, reason, ip, userAgent, correlationID string) {
	device, osName, browser := parseUserAgent(userAgent)
	row := map[string]any{
		"admin_id": adminID, "tenant_id": tenantID, "site_id": siteID, "username": username, "success": boolValue(success), "failure_reason": reason,
		"ip_address": ip, "device": device, "operating_system": osName, "browser": browser,
		"user_agent": userAgent, "correlation_id": correlationID, "created_at": now(), "updated_at": now(),
	}
	_ = s.db.Table("game_admin_login_logs").Create(row).Error
}

func (s *Service) issueMFAChallenge(account adminCredential, purpose, secret, ip, userAgent string) (string, time.Time, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	_, ttl, _ := authConfig()
	expiresAt := now().Add(ttl)
	secretEncrypted := ""
	if secret != "" {
		secretEncrypted, err = encryptMFASecret(secret)
		if err != nil {
			return "", time.Time{}, err
		}
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_mfa_challenges").Where("admin_id=? AND purpose=? AND consumed_at IS NULL", account.ID, purpose).
			Updates(map[string]any{"consumed_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return tx.Table("game_admin_mfa_challenges").Create(map[string]any{
			"admin_id": account.ID, "token_hash": tokenHash(token), "purpose": purpose,
			"secret_encrypted": secretEncrypted, "expires_at": expiresAt, "ip_address": ip,
			"user_agent": userAgent, "created_at": now(), "updated_at": now(),
		}).Error
	})
	return token, expiresAt, err
}

func (s *Service) AdminLogin(input LoginInput, ip, userAgent, correlationID string) (map[string]any, error) {
	username := strings.TrimSpace(input.Username)
	tenantCode := strings.TrimSpace(input.TenantCode)
	portalType := strings.ToUpper(strings.TrimSpace(input.PortalType))
	if portalType != "" && portalType != "PLATFORM" && portalType != "TENANT" {
		return nil, NewError("INVALID_PORTAL_TYPE", "invalid administrator portal type")
	}

	if portalType == "TENANT" && tenantCode == "" {
		return nil, NewError("TENANT_CODE_REQUIRED", "tenant code is required for tenant administrator login")
	}

	var account adminCredential
	db := s.db.Table("game_admin_accounts a").
		Select("a.id,a.username,a.display_name,a.tenant_id,a.site_id,COALESCE(t.tenant_code,'') tenant_code,COALESCE(t.tenant_name,'') tenant_name,COALESCE(st.site_code,'') site_code,COALESCE(st.site_name,'') site_name,a.password_hash,a.mfa_enabled,a.mfa_secret_encrypted,a.must_change_password,a.status").
		Joins("LEFT JOIN game_saas_tenants t ON t.id=a.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites st ON st.id=a.site_id AND st.deleted_at IS NULL").
		Where("a.username=? AND a.deleted_at IS NULL", username)

	if portalType == "TENANT" {
		if tenantCode == "" {
			return nil, NewError("TENANT_CODE_REQUIRED", "tenant code is required for tenant administrator login")
		} else {
			db = db.Where("t.tenant_code=?", tenantCode)
		}
	}

	if portalType == "PLATFORM" {
		db = db.Where("a.tenant_id=0")
	} else if portalType == "TENANT" {
		db = db.Where("a.tenant_id>0")
	}
	err := db.Take(&account).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query administrator credential: %w", err)
	}
	success := err == nil && account.Status == 1 && account.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)) == nil
	if !success {
		reason := "INVALID_CREDENTIALS"
		var adminID *uint64
		tenantID, siteID := uint64(0), uint64(0)
		if account.ID > 0 {
			adminID = &account.ID
			tenantID = account.TenantID
			siteID = account.SiteID
			if account.Status != 1 {
				reason = "ACCOUNT_DISABLED"
			}
			_ = s.db.Table("game_admin_accounts").Where("id=?", account.ID).UpdateColumn("failed_login_count", gorm.Expr("failed_login_count+1")).Error
		}
		s.writeLoginLog(adminID, tenantID, siteID, username, false, reason, ip, userAgent, correlationID)
		return nil, NewError("INVALID_CREDENTIALS", "invalid username or password")
	}
	accountType := "TENANT"
	if account.TenantID == 0 {
		accountType = "PLATFORM"
	}
	if portalType != "" && portalType != accountType {
		s.writeLoginLog(&account.ID, account.TenantID, account.SiteID, username, false, "PORTAL_TYPE_MISMATCH", ip, userAgent, correlationID)
		return nil, NewError("PORTAL_TYPE_MISMATCH", "this account cannot sign in to the selected administrator portal")
	}

	response := map[string]any{
		"administrator": map[string]any{"id": account.ID, "username": account.Username, "displayName": account.DisplayName, "tenantID": account.TenantID, "siteID": account.SiteID, "defaultSiteID": account.SiteID, "adminType": accountType, "mustChangePassword": account.MustChangePassword},
		"tenant":        map[string]any{"id": account.TenantID, "code": account.TenantCode, "name": account.TenantName},
		"site":          map[string]any{"id": account.SiteID, "code": account.SiteCode, "name": account.SiteName},
	}
	if !account.MFAEnabled {
		secret, err := generateMFASecret()
		if err != nil {
			return nil, err
		}
		token, expiresAt, err := s.issueMFAChallenge(account, challengePurposeSetup, secret, ip, userAgent)
		if err != nil {
			return nil, err
		}
		issuer, _, _ := authConfig()
		label := issuer + ":" + account.Username
		response["status"] = authStatusMFASetupRequired
		response["challengeToken"] = token
		response["expiresAt"] = expiresAt
		response["secret"] = secret
		response["otpauthUri"] = "otpauth://totp/" + url.PathEscape(label) + "?secret=" + url.QueryEscape(secret) + "&issuer=" + url.QueryEscape(issuer) + "&digits=6&period=30"
		return response, nil
	}

	if strings.TrimSpace(account.MFASecretEncrypted) == "" {
		return nil, NewError("MFA_CONFIGURATION_INVALID", "MFA configuration is invalid")
	}
	token, expiresAt, err := s.issueMFAChallenge(account, challengePurposeLogin, "", ip, userAgent)
	if err != nil {
		return nil, err
	}
	response["status"] = authStatusMFARequired
	response["challengeToken"] = token
	response["expiresAt"] = expiresAt
	return response, nil
}

func (s *Service) loadChallenge(tx *gorm.DB, token, purpose string) (mfaChallenge, error) {
	var challenge mfaChallenge
	err := tx.Table("game_admin_mfa_challenges c").
		Select("c.id,c.admin_id,a.tenant_id,a.site_id,c.token_hash,c.purpose,c.secret_encrypted,c.attempts,c.expires_at,c.consumed_at,c.ip_address,c.user_agent,a.username admin_username,a.display_name admin_display_name,COALESCE(t.tenant_code,'') tenant_code,COALESCE(t.tenant_name,'') tenant_name,COALESCE(st.site_code,'') site_code,COALESCE(st.site_name,'') site_name,a.mfa_secret_encrypted admin_mfa_secret_encrypted,a.must_change_password").
		Joins("JOIN game_admin_accounts a ON a.id=c.admin_id AND a.status=1 AND a.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_tenants t ON t.id=a.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites st ON st.id=a.site_id AND st.deleted_at IS NULL").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("c.token_hash=? AND c.purpose=? AND c.deleted_at IS NULL", tokenHash(strings.TrimSpace(token)), purpose).
		Take(&challenge).Error
	if err != nil || challenge.ConsumedAt != nil || !challenge.ExpiresAt.After(now()) || challenge.Attempts >= maxMFAAttempts {
		return mfaChallenge{}, NewError("MFA_CHALLENGE_INVALID", "MFA challenge is invalid or expired")
	}
	return challenge, nil
}

func (s *Service) issueSession(tx *gorm.DB, challenge mfaChallenge, correlationID string) (map[string]any, error) {
	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	_, _, sessionTTL := authConfig()
	expiresAt := now().Add(sessionTTL)
	if err := tx.Table("game_admin_sessions").Create(map[string]any{
		"admin_id": challenge.AdminID, "tenant_id": challenge.TenantID, "site_id": challenge.SiteID, "token_hash": tokenHash(token), "ip_address": challenge.IP,
		"user_agent": challenge.UserAgent, "expired_at": expiresAt, "last_active_at": now(), "created_at": now(), "updated_at": now(),
	}).Error; err != nil {
		return nil, err
	}
	if err := tx.Table("game_admin_accounts").Where("id=?", challenge.AdminID).Updates(map[string]any{
		"last_login_ip": challenge.IP, "last_login_at": now(), "failed_login_count": 0, "updated_at": now(),
	}).Error; err != nil {
		return nil, err
	}
	device, osName, browser := parseUserAgent(challenge.UserAgent)
	if err := tx.Table("game_admin_login_logs").Create(map[string]any{
		"admin_id": challenge.AdminID, "tenant_id": challenge.TenantID, "site_id": challenge.SiteID, "username": challenge.AdminUsername, "success": 1,
		"failure_reason": "", "ip_address": challenge.IP, "device": device, "operating_system": osName,
		"browser": browser, "user_agent": challenge.UserAgent, "correlation_id": correlationID,
		"created_at": now(), "updated_at": now(),
	}).Error; err != nil {
		return nil, err
	}
	adminType := "TENANT"
	if challenge.TenantID == 0 {
		adminType = "PLATFORM"
	}
	return map[string]any{
		"status": authStatusAuthenticated, "token": token, "expiresAt": expiresAt,
		"administrator": map[string]any{"id": challenge.AdminID, "username": challenge.AdminUsername, "displayName": challenge.AdminDisplayName, "tenantID": challenge.TenantID, "siteID": challenge.SiteID, "defaultSiteID": challenge.SiteID, "adminType": adminType, "mustChangePassword": challenge.MustChangePassword},
		"tenant":        map[string]any{"id": challenge.TenantID, "code": challenge.TenantCode, "name": challenge.TenantName},
		"site":          map[string]any{"id": challenge.SiteID, "code": challenge.SiteCode, "name": challenge.SiteName},
	}, nil
}

func (s *Service) ConfirmMFASetup(input MFAChallengeInput, correlationID string) (map[string]any, error) {
	var response map[string]any
	var recoveryCodes []string
	var invalidChallengeID uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		challenge, err := s.loadChallenge(tx, input.ChallengeToken, challengePurposeSetup)
		if err != nil {
			return err
		}
		secret, err := decryptMFASecret(challenge.SecretEncrypted)
		if err != nil {
			return err
		}
		if !verifyTOTP(secret, input.Code, now()) {
			invalidChallengeID = challenge.ID
			return NewError("MFA_CODE_INVALID", "invalid verification code")
		}
		hash := sha256.Sum256([]byte(secret))
		if err := tx.Table("game_admin_accounts").Where("id=? AND mfa_enabled=0", challenge.AdminID).Updates(map[string]any{
			"mfa_enabled": 1, "mfa_secret_hash": hex.EncodeToString(hash[:]), "mfa_secret_encrypted": challenge.SecretEncrypted,
			"version": gorm.Expr("version+1"), "updated_at": now(),
		}).Error; err != nil {
			return err
		}
		consumedAt := now()
		if err := tx.Table("game_admin_mfa_challenges").Where("id=?", challenge.ID).Updates(map[string]any{"consumed_at": consumedAt, "updated_at": consumedAt}).Error; err != nil {
			return err
		}
		recoveryCodes, err = replaceRecoveryCodes(tx, challenge.AdminID)
		if err != nil {
			return err
		}
		if err := audit(tx, Actor{ID: challenge.AdminID, Username: challenge.AdminUsername, TenantID: challenge.TenantID, SiteID: challenge.SiteID, Platform: challenge.TenantID == 0, IP: challenge.IP, UserAgent: challenge.UserAgent, CorrelationID: correlationID}, "system:admin:reset-mfa", "system.admin.mfa.setup", "administrator", challenge.AdminID, map[string]any{"mfaEnabled": false}, map[string]any{"mfaEnabled": true}, "first login MFA enrollment"); err != nil {
			return err
		}
		response, err = s.issueSession(tx, challenge, correlationID)
		if response != nil {
			response["recoveryCodes"] = recoveryCodes
		}
		return err
	})
	if invalidChallengeID > 0 {
		_ = s.db.Table("game_admin_mfa_challenges").Where("id=?", invalidChallengeID).
			UpdateColumn("attempts", gorm.Expr("attempts+1")).Error
	}
	return response, err
}

func (s *Service) VerifyMFA(input MFAChallengeInput, correlationID string) (map[string]any, error) {
	var response map[string]any
	var invalidChallengeID uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		challenge, err := s.loadChallenge(tx, input.ChallengeToken, challengePurposeLogin)
		if err != nil {
			return err
		}
		secret, err := decryptMFASecret(challenge.MFASecretEncrypted)
		if err != nil {
			return err
		}
		if !verifyTOTP(secret, input.Code, now()) && !consumeMFARecoveryCode(tx, challenge.AdminID, input.Code) {
			invalidChallengeID = challenge.ID
			return NewError("MFA_CODE_INVALID", "invalid verification code")
		}
		consumedAt := now()
		if err := tx.Table("game_admin_mfa_challenges").Where("id=?", challenge.ID).Updates(map[string]any{"consumed_at": consumedAt, "updated_at": consumedAt}).Error; err != nil {
			return err
		}
		response, err = s.issueSession(tx, challenge, correlationID)
		return err
	})
	if invalidChallengeID > 0 {
		_ = s.db.Table("game_admin_mfa_challenges").Where("id=?", invalidChallengeID).
			UpdateColumn("attempts", gorm.Expr("attempts+1")).Error
	}
	return response, err
}

func (s *Service) Logout(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return NewError("AUTHENTICATION_REQUIRED", "authentication required")
	}
	result := s.db.Table("game_admin_sessions").Where("token_hash=? AND revoked_at IS NULL AND deleted_at IS NULL", tokenHash(token)).Updates(map[string]any{"revoked_at": now(), "updated_at": now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return NewError("AUTHENTICATION_REQUIRED", "authentication required")
	}
	return nil
}
