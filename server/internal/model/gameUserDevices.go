package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameUserDevices Member devices
type GameUserDevices struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID        uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID          uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID          uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	DeviceID        string     `gorm:"column:device_id;type:varchar(128);not null" json:"deviceID"`
	Platform        string     `gorm:"column:platform;type:varchar(32);default:h5;not null" json:"platform"`
	DeviceModel     string     `gorm:"column:device_model;type:varchar(128);not null" json:"deviceModel"`
	OS              string     `gorm:"column:os;type:varchar(128);not null" json:"os"`
	Browser         string     `gorm:"column:browser;type:varchar(128);not null" json:"browser"`
	AppVersion      string     `gorm:"column:app_version;type:varchar(32);not null" json:"appVersion"`
	FingerprintHash string     `gorm:"column:fingerprint_hash;type:char(64);not null" json:"fingerprintHash"`
	FirstIP         string     `gorm:"column:first_ip;type:varchar(64);not null" json:"firstIP"`
	LastIP          string     `gorm:"column:last_ip;type:varchar(64);not null" json:"lastIP"`
	ClientIP        string     `gorm:"column:client_ip;type:varchar(64);not null" json:"clientIP"`
	UserAgent       string     `gorm:"column:user_agent;type:varchar(500);not null" json:"userAgent"`
	FirstSeenAt     *time.Time `gorm:"column:first_seen_at;type:datetime(3)" json:"firstSeenAt"`
	LastSeenAt      *time.Time `gorm:"column:last_seen_at;type:datetime(3)" json:"lastSeenAt"`
	Status          int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"`            // 1 active, 2 blocked
	TrustStatus     int        `gorm:"column:trust_status;type:tinyint(4);default:1;not null" json:"trustStatus"` // 1 normal, 2 trusted
	RiskStatus      int        `gorm:"column:risk_status;type:tinyint(4);default:1;not null" json:"riskStatus"`   // 1 normal, 2 observed, 3 risky
	LoginCount      uint       `gorm:"column:login_count;type:int unsigned;default:0;not null" json:"loginCount"`
}

// GameUserDevicesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserDevicesColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"tenant_id":        true,
	"site_id":          true,
	"user_id":          true,
	"device_id":        true,
	"platform":         true,
	"device_model":     true,
	"os":               true,
	"browser":          true,
	"app_version":      true,
	"fingerprint_hash": true,
	"first_ip":         true,
	"last_ip":          true,
	"client_ip":        true,
	"user_agent":       true,
	"first_seen_at":    true,
	"last_seen_at":     true,
	"status":           true,
	"trust_status":     true,
	"risk_status":      true,
	"login_count":      true,
}
