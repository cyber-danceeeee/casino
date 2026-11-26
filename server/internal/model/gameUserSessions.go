package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameUserSessions Login sessions
type GameUserSessions struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID  uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID    uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID    uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	TokenHash string     `gorm:"column:token_hash;type:varchar(255);not null" json:"tokenHash"`
	DeviceID  string     `gorm:"column:device_id;type:varchar(128);not null" json:"deviceID"`
	ClientIP  string     `gorm:"column:client_ip;type:varchar(64);not null" json:"clientIP"`
	UserAgent string     `gorm:"column:user_agent;type:varchar(500);not null" json:"userAgent"`
	ExpiredAt *time.Time `gorm:"column:expired_at;type:datetime(3)" json:"expiredAt"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:datetime(3)" json:"revokedAt"`
}

// GameUserSessionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserSessionsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
	"tenant_id":  true,
	"site_id":    true,
	"user_id":    true,
	"token_hash": true,
	"device_id":  true,
	"client_ip":  true,
	"user_agent": true,
	"expired_at": true,
	"revoked_at": true,
}
