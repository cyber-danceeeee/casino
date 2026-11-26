package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserLoginRecords stores successful and failed authentication attempts without credentials or tokens.
type GameUserLoginRecords struct {
	sgorm.Model `gorm:"embedded"`

	SiteID          uint64     `gorm:"column:site_id;type:bigint unsigned;not null" json:"siteID"`
	UserID          uint64     `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	LoginIdentifier string     `gorm:"column:login_identifier;type:varchar(190);not null" json:"loginIdentifier"`
	Success         int        `gorm:"column:success;type:tinyint(1);not null" json:"success"`
	FailureReason   string     `gorm:"column:failure_reason;type:varchar(128);not null" json:"failureReason"`
	IPAddress       string     `gorm:"column:ip_address;type:varchar(64);not null" json:"ipAddress"`
	Country         string     `gorm:"column:country;type:varchar(8);not null" json:"country"`
	Region          string     `gorm:"column:region;type:varchar(128);not null" json:"region"`
	DeviceID        string     `gorm:"column:device_id;type:varchar(128);not null" json:"deviceID"`
	DeviceType      string     `gorm:"column:device_type;type:varchar(64);not null" json:"deviceType"`
	Platform        string     `gorm:"column:platform;type:varchar(32);not null" json:"platform"`
	OS              string     `gorm:"column:os;type:varchar(128);not null" json:"os"`
	Browser         string     `gorm:"column:browser;type:varchar(128);not null" json:"browser"`
	AppVersion      string     `gorm:"column:app_version;type:varchar(32);not null" json:"appVersion"`
	UserAgent       string     `gorm:"column:user_agent;type:varchar(500);not null" json:"userAgent"`
	OccurredAt      *time.Time `gorm:"column:occurred_at;type:datetime(3);not null" json:"occurredAt"`
}

func (GameUserLoginRecords) TableName() string { return "game_user_login_records" }
