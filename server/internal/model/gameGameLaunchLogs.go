package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGameLaunchLogs Game launch logs
type GameGameLaunchLogs struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID                 uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserID                 uint64     `gorm:"column:user_id;type:bigint(20) unsigned;default:0;not null" json:"userID"`
	GameID                 uint64     `gorm:"column:game_id;type:bigint(20) unsigned;not null" json:"gameID"`
	ProviderID             uint64     `gorm:"column:provider_id;type:bigint(20) unsigned;default:0;not null" json:"providerID"`
	LaunchNo               string     `gorm:"column:launch_no;type:varchar(64);not null" json:"launchNo"`
	SessionID              string     `gorm:"column:session_id;type:varchar(128);not null" json:"sessionID"`
	LaunchType             string     `gorm:"column:launch_type;type:varchar(16);not null" json:"launchType"`
	Currency               string     `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	Language               string     `gorm:"column:language;type:varchar(16);not null" json:"language"`
	Country                string     `gorm:"column:country;type:varchar(8);not null" json:"country"`
	CountryID              uint64     `gorm:"column:country_id;type:bigint(20) unsigned;default:0;not null" json:"countryID"`
	CountryCodeSnapshot    string     `gorm:"column:country_code_snapshot;type:varchar(8);not null" json:"countryCodeSnapshot"`
	CurrencyCodeSnapshot   string     `gorm:"column:currency_code_snapshot;type:varchar(8);not null" json:"currencyCodeSnapshot"`
	LanguageCodeSnapshot   string     `gorm:"column:language_code_snapshot;type:varchar(16);not null" json:"languageCodeSnapshot"`
	Device                 string     `gorm:"column:device;type:varchar(32);not null" json:"device"`
	ProviderRequest        string     `gorm:"column:provider_request;type:longtext" json:"providerRequest"`
	ProviderLaunchRequest  string     `gorm:"column:provider_launch_request;type:longtext" json:"providerLaunchRequest"`
	ProviderResponse       string     `gorm:"column:provider_response;type:longtext" json:"providerResponse"`
	ProviderLaunchResponse string     `gorm:"column:provider_launch_response;type:longtext" json:"providerLaunchResponse"`
	CorrelationID          string     `gorm:"column:correlation_id;type:varchar(128);not null" json:"correlationID"`
	SessionToken           string     `gorm:"column:session_token;type:varchar(255);not null" json:"sessionToken"`
	LaunchStatus           int        `gorm:"column:launch_status;type:tinyint(4);default:1;not null" json:"launchStatus"` // 1 loading, 2 entered, 3 failed
	SessionStatus          string     `gorm:"column:session_status;type:varchar(32);not null" json:"sessionStatus"`
	LaunchTime             *time.Time `gorm:"column:launch_time;type:datetime(3)" json:"launchTime"`
	LaunchSuccessTime      *time.Time `gorm:"column:launch_success_time;type:datetime(3)" json:"launchSuccessTime"`
	ExitTime               *time.Time `gorm:"column:exit_time;type:datetime(3)" json:"exitTime"`
	CloseTime              *time.Time `gorm:"column:close_time;type:datetime(3)" json:"closeTime"`
	LastHeartbeatTime      *time.Time `gorm:"column:last_heartbeat_time;type:datetime(3)" json:"lastHeartbeatTime"`
	DurationSeconds        uint64     `gorm:"column:duration_seconds;type:bigint unsigned;default:0;not null" json:"durationSeconds"`
	ExitReason             string     `gorm:"column:exit_reason;type:varchar(32);not null" json:"exitReason"`
	ClientIP               string     `gorm:"column:client_ip;type:varchar(64);not null" json:"clientIP"`
	ErrorMessage           string     `gorm:"column:error_message;type:varchar(500);not null" json:"errorMessage"`
}

// GameGameLaunchLogsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGameLaunchLogsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":                  true,
	"user_id":                  true,
	"game_id":                  true,
	"provider_id":              true,
	"launch_no":                true,
	"session_id":               true,
	"launch_type":              true,
	"currency":                 true,
	"language":                 true,
	"country":                  true,
	"country_id":               true,
	"country_code_snapshot":    true,
	"currency_code_snapshot":   true,
	"language_code_snapshot":   true,
	"device":                   true,
	"provider_request":         true,
	"provider_launch_request":  true,
	"provider_response":        true,
	"provider_launch_response": true,
	"correlation_id":           true,
	"session_token":            true,
	"launch_status":            true,
	"session_status":           true,
	"launch_time":              true,
	"launch_success_time":      true,
	"exit_time":                true,
	"close_time":               true,
	"last_heartbeat_time":      true,
	"duration_seconds":         true,
	"exit_reason":              true,
	"client_ip":                true,
	"error_message":            true,
}
