package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GamePushDeviceToken binds one Firebase app installation token to a player.
// Tokens are intentionally excluded from every admin response.
type GamePushDeviceToken struct {
	sgorm.Model `gorm:"embedded"`

	SiteID                 uint64     `gorm:"column:site_id;type:bigint unsigned;not null;default:0" json:"siteID"`
	UserID                 uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null;index:idx_push_token_user_enabled,priority:1" json:"userID"`
	Platform               string     `gorm:"column:platform;type:varchar(16);not null" json:"platform"`
	TokenHash              string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex:uk_push_token_hash" json:"-"`
	Token                  string     `gorm:"column:token;type:text;not null" json:"-"`
	DeviceID               string     `gorm:"column:device_id;type:varchar(128);not null;default:''" json:"deviceID"`
	FingerprintHash        string     `gorm:"column:fingerprint_hash;type:char(64);not null;default:'';index:idx_push_device_fingerprint" json:"fingerprintHash"`
	Manufacturer           string     `gorm:"column:manufacturer;type:varchar(64);not null;default:''" json:"manufacturer"`
	Brand                  string     `gorm:"column:brand;type:varchar(64);not null;default:''" json:"brand"`
	DeviceModel            string     `gorm:"column:device_model;type:varchar(128);not null;default:''" json:"deviceModel"`
	DeviceName             string     `gorm:"column:device_name;type:varchar(128);not null;default:''" json:"deviceName"`
	OSVersion              string     `gorm:"column:os_version;type:varchar(64);not null;default:''" json:"osVersion"`
	SDKVersion             uint       `gorm:"column:sdk_version;type:int unsigned;not null;default:0" json:"sdkVersion"`
	SecurityPatch          string     `gorm:"column:security_patch;type:varchar(32);not null;default:''" json:"securityPatch"`
	NetworkType            string     `gorm:"column:network_type;type:varchar(24);not null;default:'';index:idx_push_device_network" json:"networkType"`
	NetworkMetered         bool       `gorm:"column:network_metered;type:tinyint(1);not null;default:0" json:"networkMetered"`
	NetworkRoaming         bool       `gorm:"column:network_roaming;type:tinyint(1);not null;default:0" json:"networkRoaming"`
	CarrierName            string     `gorm:"column:carrier_name;type:varchar(128);not null;default:''" json:"carrierName"`
	ScreenWidth            uint       `gorm:"column:screen_width;type:int unsigned;not null;default:0" json:"screenWidth"`
	ScreenHeight           uint       `gorm:"column:screen_height;type:int unsigned;not null;default:0" json:"screenHeight"`
	ScreenDensity          float64    `gorm:"column:screen_density;type:decimal(6,3);not null;default:0" json:"screenDensity"`
	Timezone               string     `gorm:"column:timezone;type:varchar(64);not null;default:''" json:"timezone"`
	Locale                 string     `gorm:"column:locale;type:varchar(32);not null;default:''" json:"locale"`
	BatteryLevel           int        `gorm:"column:battery_level;type:smallint;not null;default:-1" json:"batteryLevel"`
	Charging               bool       `gorm:"column:charging;type:tinyint(1);not null;default:0" json:"charging"`
	NotificationPermission string     `gorm:"column:notification_permission;type:varchar(16);not null;default:'';index:idx_push_device_notification" json:"notificationPermission"`
	AppVersion             string     `gorm:"column:app_version;type:varchar(32);not null;default:''" json:"appVersion"`
	AppVersionCode         uint64     `gorm:"column:app_version_code;type:bigint unsigned;not null;default:0" json:"appVersionCode"`
	Language               string     `gorm:"column:language;type:varchar(16);not null;default:'en'" json:"language"`
	LastIP                 string     `gorm:"column:last_ip;type:varchar(45);not null;default:''" json:"lastIP"`
	UserAgent              string     `gorm:"column:user_agent;type:varchar(500);not null;default:''" json:"userAgent"`
	Enabled                bool       `gorm:"column:enabled;type:tinyint(1);not null;default:1;index:idx_push_token_user_enabled,priority:2" json:"enabled"`
	LastSeenAt             *time.Time `gorm:"column:last_seen_at;type:datetime(3)" json:"lastSeenAt"`
	InvalidatedAt          *time.Time `gorm:"column:invalidated_at;type:datetime(3)" json:"invalidatedAt"`
	LastError              string     `gorm:"column:last_error;type:varchar(500);not null;default:''" json:"lastError"`
}

func (GamePushDeviceToken) TableName() string { return "game_push_device_tokens" }

// GameNotificationDelivery is an immutable per-user delivery result. A retry
// updates the same delivery row instead of creating duplicate pushes.
type GameNotificationDelivery struct {
	sgorm.Model `gorm:"embedded"`

	SiteID            uint64     `gorm:"column:site_id;type:bigint unsigned;not null;default:0" json:"siteID"`
	NotificationID    uint64     `gorm:"column:notification_id;type:bigint(20) unsigned;not null;uniqueIndex:uk_notification_delivery_user,priority:1;index:idx_notification_delivery_status,priority:1" json:"notificationID"`
	JobID             uint64     `gorm:"column:job_id;type:bigint(20) unsigned;not null" json:"jobID"`
	UserID            uint64     `gorm:"column:user_id;type:bigint(20) unsigned;not null;uniqueIndex:uk_notification_delivery_user,priority:2" json:"userID"`
	DeviceTokenID     uint64     `gorm:"column:device_token_id;type:bigint(20) unsigned;not null;default:0" json:"deviceTokenID"`
	Status            string     `gorm:"column:status;type:varchar(24);not null;default:'PENDING';index:idx_notification_delivery_status,priority:2" json:"status"`
	ProviderMessageID string     `gorm:"column:provider_message_id;type:varchar(255);not null;default:''" json:"providerMessageID"`
	ErrorCode         string     `gorm:"column:error_code;type:varchar(64);not null;default:''" json:"errorCode"`
	ErrorMessage      string     `gorm:"column:error_message;type:varchar(500);not null;default:''" json:"errorMessage"`
	AttemptedAt       *time.Time `gorm:"column:attempted_at;type:datetime(3)" json:"attemptedAt"`
	SentAt            *time.Time `gorm:"column:sent_at;type:datetime(3)" json:"sentAt"`
}

func (GameNotificationDelivery) TableName() string { return "game_notification_deliveries" }
