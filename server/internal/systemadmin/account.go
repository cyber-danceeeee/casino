package systemadmin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	challengePurposeAccountSetup = "ACCOUNT_SETUP"
	passwordHistoryLimit         = 5
)

type accountProfileRow struct {
	ID                 uint64     `json:"id"`
	Username           string     `json:"username"`
	DisplayName        string     `json:"displayName" gorm:"column:display_name"`
	Email              string     `json:"email"`
	Mobile             string     `json:"mobile"`
	TenantID           uint64     `json:"tenantID" gorm:"column:tenant_id"`
	TenantName         string     `json:"tenantName" gorm:"column:tenant_name"`
	SiteID             uint64     `json:"defaultSiteID" gorm:"column:site_id"`
	SiteName           string     `json:"defaultSiteName" gorm:"column:site_name"`
	PreferredLanguage  string     `json:"preferredLanguage" gorm:"column:preferred_language"`
	Timezone           string     `json:"timezone"`
	MFAEnabled         bool       `json:"mfaEnabled" gorm:"column:mfa_enabled"`
	MustChangePassword bool       `json:"mustChangePassword" gorm:"column:must_change_password"`
	LastLoginAt        *time.Time `json:"lastLoginAt" gorm:"column:last_login_at"`
	LastPasswordAt     *time.Time `json:"lastPasswordChangedAt" gorm:"column:last_password_changed_at"`
}

func (s *Service) AccountProfile(actor Actor) (map[string]any, error) {
	var row accountProfileRow
	err := s.db.Table("game_admin_accounts a").
		Select("a.id,a.username,a.display_name,a.email,a.mobile,a.tenant_id,COALESCE(t.tenant_name,'') tenant_name,a.site_id,COALESCE(st.site_name,'') site_name,a.preferred_language,a.timezone,a.mfa_enabled,a.must_change_password,a.last_login_at,a.last_password_changed_at").
		Joins("LEFT JOIN game_saas_tenants t ON t.id=a.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites st ON st.id=a.site_id AND st.deleted_at IS NULL").
		Where("a.id=? AND a.deleted_at IS NULL", actor.ID).Take(&row).Error
	if err != nil {
		return nil, rowNotFound("administrator")
	}
	roles := []string{}
	if err := s.db.Table("game_admin_account_roles ar").Distinct("r.code").
		Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.status=1 AND r.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", actor.ID).Order("r.code").Pluck("r.code", &roles).Error; err != nil {
		return nil, err
	}
	adminType := "TENANT"
	if row.TenantID == 0 {
		adminType = "PLATFORM"
	}
	return map[string]any{"profile": row, "adminType": adminType, "roles": roles}, nil
}

func validateAccountProfile(input AccountProfileInput) error {
	if strings.TrimSpace(input.DisplayName) == "" {
		return NewError("DISPLAY_NAME_REQUIRED", "display name is required")
	}
	if value := strings.TrimSpace(input.Email); value != "" {
		if _, err := mail.ParseAddress(value); err != nil {
			return NewError("INVALID_EMAIL", "invalid email address")
		}
	}
	language := strings.TrimSpace(input.PreferredLanguage)
	if language != "zh-CN" && language != "en-US" {
		return NewError("INVALID_LANGUAGE", "unsupported administrator language")
	}
	if strings.TrimSpace(input.Timezone) == "" || len(input.Timezone) > 64 {
		return NewError("INVALID_TIMEZONE", "invalid timezone")
	}
	return nil
}

func (s *Service) UpdateAccountProfile(input AccountProfileInput, actor Actor) error {
	if err := validateAccountProfile(input); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_admin_accounts").Select("display_name,email,mobile,preferred_language,timezone").Where("id=? AND deleted_at IS NULL", actor.ID).Take(&before).Error; err != nil {
		return rowNotFound("administrator")
	}
	updates := map[string]any{
		"display_name": strings.TrimSpace(input.DisplayName), "email": strings.TrimSpace(input.Email),
		"mobile": strings.TrimSpace(input.Mobile), "preferred_language": strings.TrimSpace(input.PreferredLanguage),
		"timezone": strings.TrimSpace(input.Timezone), "version": gorm.Expr("version+1"), "updated_at": now(),
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", actor.ID).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "account:profile:update", "admin.account.profile.update", "administrator", actor.ID, before, updates, "self-service profile update")
	})
}

func validatePasswordStrength(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return NewError("WEAK_PASSWORD", "password must contain 10 to 128 characters")
	}
	var upper, lower, digit, symbol bool
	for _, value := range password {
		upper = upper || unicode.IsUpper(value)
		lower = lower || unicode.IsLower(value)
		digit = digit || unicode.IsDigit(value)
		symbol = symbol || (!unicode.IsLetter(value) && !unicode.IsDigit(value))
	}
	if !upper || !lower || !digit || !symbol {
		return NewError("WEAK_PASSWORD", "password must include uppercase, lowercase, number and symbol")
	}
	return nil
}

func (s *Service) ChangeAccountPassword(input ChangePasswordInput, actor Actor) error {
	if input.NewPassword != input.ConfirmPassword {
		return NewError("PASSWORD_CONFIRMATION_MISMATCH", "password confirmation does not match")
	}
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return err
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var account struct{ PasswordHash string }
		if err := tx.Table("game_admin_accounts").Select("password_hash").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", actor.ID).Take(&account).Error; err != nil {
			return rowNotFound("administrator")
		}
		if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.CurrentPassword)) != nil {
			return NewError("CURRENT_PASSWORD_INVALID", "current password is invalid")
		}
		if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.NewPassword)) == nil {
			return NewError("PASSWORD_REUSED", "new password was used recently")
		}
		history := []string{}
		if err := tx.Table("game_admin_password_history").Where("admin_id=?", actor.ID).Order("created_at DESC,id DESC").Limit(passwordHistoryLimit).Pluck("password_hash", &history).Error; err != nil {
			return err
		}
		for _, hash := range history {
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.NewPassword)) == nil {
				return NewError("PASSWORD_REUSED", "new password was used recently")
			}
		}
		if err := tx.Table("game_admin_password_history").Create(map[string]any{"admin_id": actor.ID, "password_hash": account.PasswordHash, "created_at": now()}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", actor.ID).Updates(map[string]any{
			"password_hash": string(newHash), "must_change_password": 0, "last_password_changed_at": now(),
			"version": gorm.Expr("version+1"), "updated_at": now(),
		}).Error; err != nil {
			return err
		}
		if input.RevokeOtherSessions {
			if err := tx.Table("game_admin_sessions").Where("admin_id=? AND id<>? AND revoked_at IS NULL", actor.ID, actor.SessionID).
				Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, "account:password:update", "admin.account.password.update", "administrator", actor.ID, nil, map[string]any{"passwordChanged": true, "otherSessionsRevoked": input.RevokeOtherSessions}, "self-service password change")
	})
}

func verifyAccountPassword(db *gorm.DB, adminID uint64, password string) (adminCredential, error) {
	var account adminCredential
	err := db.Table("game_admin_accounts a").
		Select("a.id,a.username,a.display_name,a.tenant_id,a.site_id,a.password_hash,a.mfa_enabled,a.mfa_secret_encrypted,a.must_change_password,a.status").
		Where("a.id=? AND a.status=1 AND a.deleted_at IS NULL", adminID).Take(&account).Error
	if err != nil || bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) != nil {
		return adminCredential{}, NewError("CURRENT_PASSWORD_INVALID", "current password is invalid")
	}
	return account, nil
}

func (s *Service) AccountMFAStatus(actor Actor) (map[string]any, error) {
	var row struct{ MFAEnabled bool }
	if err := s.db.Table("game_admin_accounts").Select("mfa_enabled").Where("id=? AND deleted_at IS NULL", actor.ID).Take(&row).Error; err != nil {
		return nil, rowNotFound("administrator")
	}
	var remaining int64
	if err := s.db.Table("game_admin_mfa_recovery_codes").Where("admin_id=? AND used_at IS NULL AND deleted_at IS NULL", actor.ID).Count(&remaining).Error; err != nil {
		return nil, err
	}
	return map[string]any{"enabled": row.MFAEnabled, "required": actor.Platform, "recoveryCodesRemaining": remaining}, nil
}

func (s *Service) StartAccountMFASetup(input AccountMFAInput, actor Actor) (map[string]any, error) {
	account, err := verifyAccountPassword(s.db, actor.ID, input.CurrentPassword)
	if err != nil {
		return nil, err
	}
	secret, err := generateMFASecret()
	if err != nil {
		return nil, err
	}
	token, expiresAt, err := s.issueMFAChallenge(account, challengePurposeAccountSetup, secret, actor.IP, actor.UserAgent)
	if err != nil {
		return nil, err
	}
	issuer, _, _ := authConfig()
	label := issuer + ":" + account.Username
	return map[string]any{
		"challengeToken": token, "expiresAt": expiresAt, "secret": secret,
		"otpauthUri": "otpauth://totp/" + url.PathEscape(label) + "?secret=" + url.QueryEscape(secret) + "&issuer=" + url.QueryEscape(issuer) + "&digits=6&period=30",
	}, nil
}

func generateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for index := range codes {
		value := make([]byte, 6)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		raw := strings.ToUpper(hex.EncodeToString(value))
		codes[index] = raw[:4] + "-" + raw[4:8] + "-" + raw[8:]
	}
	return codes, nil
}

func recoveryCodeHash(code string) string {
	normalized := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code)))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

func replaceRecoveryCodes(tx *gorm.DB, adminID uint64) ([]string, error) {
	codes, err := generateRecoveryCodes()
	if err != nil {
		return nil, err
	}
	if err := tx.Table("game_admin_mfa_recovery_codes").Where("admin_id=? AND deleted_at IS NULL", adminID).Update("deleted_at", now()).Error; err != nil {
		return nil, err
	}
	for _, code := range codes {
		if err := tx.Table("game_admin_mfa_recovery_codes").Create(map[string]any{"admin_id": adminID, "code_hash": recoveryCodeHash(code), "created_at": now()}).Error; err != nil {
			return nil, err
		}
	}
	return codes, nil
}

func consumeMFARecoveryCode(tx *gorm.DB, adminID uint64, code string) bool {
	if len(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code))) != 12 {
		return false
	}
	result := tx.Table("game_admin_mfa_recovery_codes").Where("admin_id=? AND code_hash=? AND used_at IS NULL AND deleted_at IS NULL", adminID, recoveryCodeHash(code)).
		Updates(map[string]any{"used_at": now()})
	return result.Error == nil && result.RowsAffected == 1
}

func (s *Service) ConfirmAccountMFASetup(input MFAChallengeInput, actor Actor) (map[string]any, error) {
	var codes []string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		challenge, err := s.loadChallenge(tx, input.ChallengeToken, challengePurposeAccountSetup)
		if err != nil || challenge.AdminID != actor.ID {
			return NewError("MFA_CHALLENGE_INVALID", "MFA challenge is invalid or expired")
		}
		secret, err := decryptMFASecret(challenge.SecretEncrypted)
		if err != nil {
			return err
		}
		if !verifyTOTP(secret, input.Code, now()) {
			return NewError("MFA_CODE_INVALID", "invalid verification code")
		}
		hash := sha256.Sum256([]byte(secret))
		if err := tx.Table("game_admin_accounts").Where("id=?", actor.ID).Updates(map[string]any{
			"mfa_enabled": 1, "mfa_secret_hash": hex.EncodeToString(hash[:]), "mfa_secret_encrypted": challenge.SecretEncrypted,
			"version": gorm.Expr("version+1"), "updated_at": now(),
		}).Error; err != nil {
			return err
		}
		codes, err = replaceRecoveryCodes(tx, actor.ID)
		if err != nil {
			return err
		}
		if err := tx.Table("game_admin_mfa_challenges").Where("id=?", challenge.ID).Updates(map[string]any{"consumed_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "account:mfa:update", "admin.account.mfa.enable", "administrator", actor.ID, nil, map[string]any{"mfaEnabled": true}, "self-service MFA enrollment")
	})
	return map[string]any{"recoveryCodes": codes}, err
}

func verifyCurrentMFA(tx *gorm.DB, account adminCredential, code string) bool {
	secret, err := decryptMFASecret(account.MFASecretEncrypted)
	if err == nil && verifyTOTP(secret, code, now()) {
		return true
	}
	return consumeMFARecoveryCode(tx, account.ID, code)
}

func (s *Service) DisableAccountMFA(input AccountMFAInput, actor Actor) error {
	account, err := verifyAccountPassword(s.db, actor.ID, input.CurrentPassword)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if !account.MFAEnabled || !verifyCurrentMFA(tx, account, input.Code) {
			return NewError("MFA_CODE_INVALID", "invalid verification code")
		}
		if err := tx.Table("game_admin_accounts").Where("id=?", actor.ID).Updates(map[string]any{"mfa_enabled": 0, "mfa_secret_hash": "", "mfa_secret_encrypted": "", "version": gorm.Expr("version+1"), "updated_at": now()}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_admin_mfa_recovery_codes").Where("admin_id=? AND deleted_at IS NULL", actor.ID).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		if err := tx.Table("game_admin_sessions").Where("admin_id=? AND id<>? AND revoked_at IS NULL", actor.ID, actor.SessionID).Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "account:mfa:update", "admin.account.mfa.disable", "administrator", actor.ID, map[string]any{"mfaEnabled": true}, map[string]any{"mfaEnabled": false}, "self-service MFA disable")
	})
}

func (s *Service) RegenerateAccountRecoveryCodes(input AccountMFAInput, actor Actor) (map[string]any, error) {
	account, err := verifyAccountPassword(s.db, actor.ID, input.CurrentPassword)
	if err != nil {
		return nil, err
	}
	var codes []string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if !account.MFAEnabled || !verifyCurrentMFA(tx, account, input.Code) {
			return NewError("MFA_CODE_INVALID", "invalid verification code")
		}
		var err error
		codes, err = replaceRecoveryCodes(tx, actor.ID)
		if err != nil {
			return err
		}
		return audit(tx, actor, "account:mfa:update", "admin.account.mfa.recovery.regenerate", "administrator", actor.ID, nil, map[string]any{"recoveryCodesRegenerated": true}, "self-service recovery code regeneration")
	})
	return map[string]any{"recoveryCodes": codes}, err
}

type accountSessionRow struct {
	ID           uint64     `gorm:"column:id"`
	IPAddress    string     `gorm:"column:ip_address"`
	Country      string     `gorm:"column:country"`
	UserAgent    string     `gorm:"column:user_agent"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	LastActiveAt *time.Time `gorm:"column:last_active_at"`
	ExpiredAt    time.Time  `gorm:"column:expired_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
}

func (s *Service) AccountSessions(actor Actor) ([]map[string]any, error) {
	rows := []accountSessionRow{}
	if err := s.db.Table("game_admin_sessions s").Select("s.id,s.ip_address,s.user_agent,s.created_at,s.last_active_at,s.expired_at,s.revoked_at,COALESCE((SELECT l.country FROM game_admin_login_logs l WHERE l.admin_id=s.admin_id AND l.ip_address=s.ip_address AND l.deleted_at IS NULL ORDER BY l.created_at DESC LIMIT 1),'') country").Where("s.admin_id=? AND s.deleted_at IS NULL", actor.ID).Order("s.created_at DESC").Limit(100).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		device, osName, browser := parseUserAgent(row.UserAgent)
		status := "ACTIVE"
		if row.RevokedAt != nil {
			status = "REVOKED"
		} else if !row.ExpiredAt.After(now()) {
			status = "EXPIRED"
		}
		out = append(out, map[string]any{
			"id": row.ID, "device": device, "operatingSystem": osName, "browser": browser,
			"ipAddress": row.IPAddress, "country": row.Country, "userAgent": row.UserAgent,
			"createdAt": row.CreatedAt, "lastActiveAt": row.LastActiveAt, "expiredAt": row.ExpiredAt,
			"status": status, "current": row.ID == actor.SessionID,
		})
	}
	return out, nil
}

func (s *Service) RevokeAccountSession(id uint64, actor Actor) error {
	if id == actor.SessionID {
		return NewError("CURRENT_SESSION_REQUIRES_LOGOUT", "use logout to revoke the current session")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("game_admin_sessions").Where("id=? AND admin_id=? AND revoked_at IS NULL AND deleted_at IS NULL", id, actor.ID).Updates(map[string]any{"revoked_at": now(), "updated_at": now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return rowNotFound("session")
		}
		return audit(tx, actor, "account:session:revoke", "admin.account.session.revoke", "admin_session", id, nil, map[string]any{"revoked": true}, "self-service session revoke")
	})
}

func (s *Service) RevokeOtherAccountSessions(actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_sessions").Where("admin_id=? AND id<>? AND revoked_at IS NULL AND deleted_at IS NULL", actor.ID, actor.SessionID).Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "account:session:revoke", "admin.account.session.revoke_others", "administrator", actor.ID, nil, map[string]any{"otherSessionsRevoked": true}, "self-service session revoke")
	})
}

func (s *Service) AccountLoginRecords(actor Actor) ([]map[string]any, error) {
	rows := []map[string]any{}
	err := s.db.Table("game_admin_login_logs").Select("id,success,failure_reason failureReason,ip_address ipAddress,country,device,operating_system operatingSystem,browser,user_agent userAgent,created_at createdAt").
		Where("admin_id=? AND deleted_at IS NULL", actor.ID).Order("created_at DESC").Limit(100).Find(&rows).Error
	return rows, err
}
