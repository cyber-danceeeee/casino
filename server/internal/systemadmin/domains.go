package systemadmin

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DomainRow struct {
	ID                 uint64     `json:"id" gorm:"column:id"`
	TenantID           uint64     `json:"tenantID" gorm:"column:tenant_id"`
	Tenant             string     `json:"tenant" gorm:"column:tenant"`
	SiteID             uint64     `json:"siteID" gorm:"column:site_id"`
	Site               string     `json:"site" gorm:"column:site"`
	Domain             string     `json:"domain" gorm:"column:domain"`
	DomainKey          string     `json:"-" gorm:"column:domain_key"`
	DomainType         string     `json:"domainType" gorm:"column:domain_type"`
	IsPrimary          bool       `json:"isPrimary" gorm:"column:is_primary"`
	Enabled            bool       `json:"enabled" gorm:"column:enabled"`
	Status             string     `json:"status" gorm:"column:status"`
	VerificationStatus string     `json:"verificationStatus" gorm:"column:verification_status"`
	SSLStatus          string     `json:"sslStatus" gorm:"column:ssl_status"`
	SSLExpiresAt       *time.Time `json:"sslExpiresAt" gorm:"column:ssl_expires_at"`
	LastVerifiedAt     *time.Time `json:"lastVerifiedAt" gorm:"column:last_verified_at"`
	VerificationError  string     `json:"verificationError" gorm:"column:verification_error"`
	Remark             string     `json:"remark" gorm:"column:remark"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt          *time.Time `json:"-" gorm:"column:deleted_at"`
}

type DomainInput struct {
	TenantID   uint64 `json:"tenantID"`
	SiteID     uint64 `json:"siteID"`
	Domain     string `json:"domain"`
	DomainType string `json:"domainType"`
	IsPrimary  bool   `json:"isPrimary"`
	Enabled    *bool  `json:"enabled"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
	Reason     string `json:"reason"`
}

type DomainSSLInput struct {
	SSLExpiresAt *time.Time `json:"sslExpiresAt"`
	Reason       string     `json:"reason"`
}

func domainSelect() string {
	return `d.id,d.tenant_id,COALESCE(t.tenant_name,'') tenant,d.site_id,COALESCE(s.site_name,'') site,d.domain,d.domain_key,d.domain_type,d.is_primary,d.enabled,d.status,d.verification_status,d.ssl_status,d.ssl_expires_at,d.last_verified_at,d.verification_error,d.remark,d.created_at,d.updated_at,d.deleted_at`
}

func (s *Service) ListDomains(q Query, actor Actor) (Result, error) {
	rows := []DomainRow{}
	db := s.db.Table("game_saas_domains d").
		Select(domainSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=d.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites s ON s.id=d.site_id AND s.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")
	if !actor.Platform {
		db = db.Where("d.tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("d.site_id=?", actor.SiteID)
		}
	}
	db = exact(db, "d.tenant_id", value(q, "tenantID"))
	db = exact(db, "d.site_id", value(q, "siteID"))
	db = like(db, "d.domain", normalizeDomain(value(q, "domain")))
	db = exact(db, "d.domain_type", strings.ToUpper(value(q, "domainType")))
	db = exact(db, "d.enabled", value(q, "status"))
	db = exact(db, "d.status", value(q, "lifecycleStatus"))
	return paginate(db.Order("d.is_primary DESC,d.created_at DESC,d.id DESC"), q, &rows)
}

func (s *Service) DomainDetail(id uint64, actor Actor) (map[string]any, error) {
	var row DomainRow
	db := s.db.Table("game_saas_domains d").
		Select(domainSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=d.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites s ON s.id=d.site_id AND s.deleted_at IS NULL").
		Where("d.id=? AND d.deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("d.tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("d.site_id=?", actor.SiteID)
		}
	}
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("domain")
	}
	return map[string]any{"overview": row, "operationLogs": s.operationLogs("domain", id)}, nil
}

func (s *Service) CreateDomain(input DomainInput, actor Actor) (uint64, error) {
	if err := s.normalizeDomainInput(&input, actor, true); err != nil {
		return 0, err
	}
	enabled := boolDefault(input.Enabled, true)
	row := domainInputRow(input, enabled, actor.ID)
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if input.IsPrimary {
			if err := tx.Table("game_saas_domains").Where("site_id=? AND domain_type=? AND deleted_at IS NULL", input.SiteID, input.DomainType).Update("is_primary", 0).Error; err != nil {
				return err
			}
		}
		result := tx.Table("game_saas_domains").Create(row)
		if result.Error != nil {
			return duplicateError(result.Error)
		}
		id = row["id"].(uint64)
		return audit(tx, actor, "system:domain:create", "system.domain.create", "domain", id, nil, row, input.Reason)
	})
	return id, err
}

func (s *Service) UpdateDomain(id uint64, input DomainInput, actor Actor) error {
	if err := s.normalizeDomainInput(&input, actor, false); err != nil {
		return err
	}
	var before map[string]any
	db := s.db.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("site_id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("domain")
	}
	enabled := boolDefault(input.Enabled, dbTruthy(before["enabled"]))
	updates := domainInputRow(input, enabled, actor.ID)
	delete(updates, "created_at")
	delete(updates, "created_by")
	return s.db.Transaction(func(tx *gorm.DB) error {
		if input.IsPrimary {
			if err := tx.Table("game_saas_domains").Where("site_id=? AND domain_type=? AND id<>? AND deleted_at IS NULL", input.SiteID, input.DomainType, id).Update("is_primary", 0).Error; err != nil {
				return err
			}
		}
		updateDB := tx.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			updateDB = updateDB.Where("tenant_id=?", actor.TenantID)
			if actor.SiteID > 0 {
				updateDB = updateDB.Where("site_id=?", actor.SiteID)
			}
		}
		if err := updateDB.Updates(updates).Error; err != nil {
			return duplicateError(err)
		}
		return audit(tx, actor, "system:domain:update", "system.domain.update", "domain", id, before, updates, input.Reason)
	})
}

func (s *Service) DomainStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var before DomainRow
	db := s.db.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("site_id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("domain")
	}
	updates := map[string]any{"enabled": boolValue(enabled), "status": normalizeStatus(enabled, "", "ACTIVE", "DISABLED"), "updated_by": actor.ID, "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		updateDB := tx.Table("game_saas_domains").Where("id=?", id)
		if !actor.Platform {
			updateDB = updateDB.Where("tenant_id=?", actor.TenantID)
			if actor.SiteID > 0 {
				updateDB = updateDB.Where("site_id=?", actor.SiteID)
			}
		}
		if err := updateDB.Updates(updates).Error; err != nil {
			return err
		}
		if err := syncBrandDeliveryDomainStatus(tx, before.TenantID, before.SiteID, updates["status"].(string)); err != nil {
			return err
		}
		return audit(tx, actor, "system:domain:status", "system.domain.status", "domain", id, before, updates, reason)
	})
}

func (s *Service) VerifyDomain(ctx context.Context, id uint64, actor Actor) (DomainRow, error) {
	if !actor.Platform {
		return DomainRow{}, NewError("PERMISSION_DENIED", "permission denied")
	}
	var row DomainRow
	if err := s.db.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return row, rowNotFound("domain")
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, row.Domain)
	verifiedAt := now()
	if err != nil || len(addresses) == 0 {
		message := "domain has no public DNS record"
		if err != nil {
			message = truncateDomainError(err)
		}
		updates := map[string]any{"status": "FAILED", "verification_status": "FAILED", "ssl_status": "PENDING", "last_verified_at": verifiedAt, "verification_error": message, "updated_by": actor.ID, "updated_at": verifiedAt}
		if updateErr := s.updateDomainVerification(row, updates, actor); updateErr != nil {
			return row, updateErr
		}
		return row, NewError("DOMAIN_VERIFICATION_FAILED", message)
	}
	for _, address := range addresses {
		if !isPublicDomainIP(address.IP) {
			updates := map[string]any{"status": "FAILED", "verification_status": "FAILED", "ssl_status": "PENDING", "last_verified_at": verifiedAt, "verification_error": "domain resolves to a non-public address", "updated_by": actor.ID, "updated_at": verifiedAt}
			if updateErr := s.updateDomainVerification(row, updates, actor); updateErr != nil {
				return row, updateErr
			}
			return row, NewError("DOMAIN_VERIFICATION_FAILED", "domain resolves to a non-public address")
		}
	}

	sslStatus := "FAILED"
	var sslExpiresAt *time.Time
	verificationError := ""
	connection, dialErr := (&net.Dialer{Timeout: 7 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(row.Domain, "443"))
	if dialErr == nil {
		tlsConnection := tls.Client(connection, &tls.Config{ServerName: row.Domain, MinVersion: tls.VersionTLS12})
		if handshakeErr := tlsConnection.HandshakeContext(ctx); handshakeErr == nil {
			state := tlsConnection.ConnectionState()
			if len(state.PeerCertificates) > 0 {
				expires := state.PeerCertificates[0].NotAfter.UTC()
				sslExpiresAt = &expires
				sslStatus = domainSSLStatus(sslExpiresAt, verifiedAt)
			}
		} else {
			verificationError = "TLS: " + truncateDomainError(handshakeErr)
		}
		_ = tlsConnection.Close()
	} else {
		verificationError = "TLS: " + truncateDomainError(dialErr)
	}
	updates := map[string]any{"status": "ACTIVE", "enabled": 1, "verification_status": "VERIFIED", "ssl_status": sslStatus,
		"ssl_expires_at": sslExpiresAt, "last_verified_at": verifiedAt, "verification_error": verificationError, "updated_by": actor.ID, "updated_at": verifiedAt}
	if err := s.updateDomainVerification(row, updates, actor); err != nil {
		return row, err
	}
	row.Status, row.Enabled, row.VerificationStatus, row.SSLStatus = "ACTIVE", true, "VERIFIED", sslStatus
	row.SSLExpiresAt, row.LastVerifiedAt, row.VerificationError = sslExpiresAt, &verifiedAt, verificationError
	return row, nil
}

func (s *Service) UpdateDomainSSL(id uint64, input DomainSSLInput, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	var before DomainRow
	if err := s.db.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("domain")
	}
	updates := map[string]any{"ssl_expires_at": input.SSLExpiresAt, "ssl_status": domainSSLStatus(input.SSLExpiresAt, now()), "updated_by": actor.ID, "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:domain:update", "system.domain.ssl_expiry", "domain", id, before, updates, input.Reason)
	})
}

func (s *Service) updateDomainVerification(before DomainRow, updates map[string]any, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", before.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := syncBrandDeliveryDomainStatus(tx, before.TenantID, before.SiteID, fmt.Sprint(updates["status"])); err != nil {
			return err
		}
		return audit(tx, actor, "system:domain:update", "system.domain.verify", "domain", before.ID, before, updates, "verify domain DNS and TLS")
	})
}

func syncBrandDeliveryDomainStatus(tx *gorm.DB, tenantID, siteID uint64, status string) error {
	if !tx.Migrator().HasTable("game_saas_brand_deliveries") {
		return nil
	}
	return tx.Table("game_saas_brand_deliveries").Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).
		Updates(map[string]any{"domain_status": status, "updated_at": now()}).Error
}

func domainSSLStatus(expiresAt *time.Time, current time.Time) string {
	if expiresAt == nil {
		return "PENDING"
	}
	if !expiresAt.After(current) {
		return "EXPIRED"
	}
	if expiresAt.Before(current.Add(30 * 24 * time.Hour)) {
		return "EXPIRING"
	}
	return "ACTIVE"
}

func isPublicDomainIP(ip net.IP) bool {
	return ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast()
}

func truncateDomainError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func (s *Service) DeleteDomain(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var before DomainRow
	db := s.db.Table("game_saas_domains").Where("id=? AND deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
		if actor.SiteID > 0 {
			db = db.Where("site_id=?", actor.SiteID)
		}
	}
	if err := db.Take(&before).Error; err != nil {
		return rowNotFound("domain")
	}
	nowValue := now()
	updates := map[string]any{
		"domain_key": before.DomainKey + "#deleted#" + strconv.FormatUint(id, 10),
		"enabled":    0,
		"status":     "DELETED",
		"deleted_at": nowValue,
		"updated_at": nowValue,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		updateDB := tx.Table("game_saas_domains").Where("id=?", id)
		if !actor.Platform {
			updateDB = updateDB.Where("tenant_id=?", actor.TenantID)
			if actor.SiteID > 0 {
				updateDB = updateDB.Where("site_id=?", actor.SiteID)
			}
		}
		if err := updateDB.Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:domain:delete", "system.domain.delete", "domain", id, before, updates, reason)
	})
}

func (s *Service) normalizeDomainInput(input *DomainInput, actor Actor, creating bool) error {
	if !actor.Platform {
		input.TenantID = actor.TenantID
		if actor.SiteID > 0 {
			input.SiteID = actor.SiteID
		}
	}
	if input.TenantID == 0 {
		return NewError("TENANT_REQUIRED", "tenant is required")
	}
	if input.SiteID == 0 {
		return NewError("SITE_REQUIRED", "site is required")
	}
	domain := normalizeDomain(input.Domain)
	if domain == "" {
		return NewError("DOMAIN_REQUIRED", "domain is required")
	}
	input.Domain = domain
	input.DomainType = strings.ToUpper(strings.TrimSpace(defaultString(input.DomainType, "H5")))
	if err := s.ensureTenantSiteAvailable(input.TenantID, input.SiteID); err != nil {
		return err
	}
	if !creating && strings.TrimSpace(input.Reason) == "" {
		return NewError("REASON_REQUIRED", "reason is required")
	}
	return nil
}

func domainInputRow(input DomainInput, enabled bool, actorID uint64) map[string]any {
	return map[string]any{
		"tenant_id": input.TenantID, "site_id": input.SiteID, "domain": input.Domain, "domain_key": input.Domain,
		"domain_type": input.DomainType, "is_primary": boolValue(input.IsPrimary), "enabled": boolValue(enabled),
		"status": normalizeStatus(enabled, input.Status, "ACTIVE", "DISABLED"), "remark": strings.TrimSpace(input.Remark),
		"created_by": actorID, "updated_by": actorID, "created_at": now(), "updated_at": now(),
	}
}

func normalizeDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	value = strings.Split(value, "/")[0]
	value = strings.Split(value, "?")[0]
	value = strings.Split(value, "#")[0]
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	} else if strings.Count(value, ":") == 1 {
		value = strings.Split(value, ":")[0]
	}
	return strings.Trim(value, ". ")
}
