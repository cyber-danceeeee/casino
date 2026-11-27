package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGames Game list
type GameGames struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID               uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null;uniqueIndex:uk_game_games_scope_code,priority:2" json:"siteID"`
	GameCode             string     `gorm:"column:game_code;type:varchar(64);not null;uniqueIndex:uk_game_games_scope_code,priority:3" json:"gameCode"`
	ProviderID           uint64     `gorm:"column:provider_id;type:bigint(20) unsigned;default:0;not null" json:"providerID"`
	ProviderGameCode     string     `gorm:"column:provider_game_code;type:varchar(128);not null" json:"providerGameCode"`
	Name                 string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	NameI18n             string     `gorm:"column:name_i18n;type:json" json:"nameI18n"`
	Slug                 string     `gorm:"column:slug;type:varchar(160);not null" json:"slug"`
	Category             string     `gorm:"column:category;type:varchar(128);not null" json:"category"`
	PlatformCategoryID   uint64     `gorm:"column:platform_category_id;type:bigint(20) unsigned;default:0;not null" json:"platformCategoryID"`
	PlatformCategoryCode string     `gorm:"column:platform_category_code;type:varchar(64);default:'';not null" json:"platformCategoryCode"`
	GameType             string     `gorm:"column:game_type;type:varchar(32);not null" json:"gameType"` // slot, chance, sports, live, fish, table
	RTP                  *float64   `gorm:"column:rtp;type:decimal(6,2)" json:"rtp"`
	Volatility           string     `gorm:"column:volatility;type:varchar(32);not null" json:"volatility"`
	CoverURL             string     `gorm:"column:cover_url;type:varchar(500);not null" json:"coverURL"`
	LocalThumbnailURL    string     `gorm:"column:local_thumbnail_url;type:varchar(500);not null" json:"localThumbnailURL"`
	BackgroundURL        string     `gorm:"column:background_url;type:varchar(500);not null" json:"backgroundURL"`
	LaunchImageURL       string     `gorm:"column:launch_image_url;type:varchar(500);not null" json:"launchImageURL"`
	MaintenanceMessage   string     `gorm:"column:maintenance_message;type:varchar(500);not null" json:"maintenanceMessage"`
	SupportsDemo         bool       `gorm:"column:supports_demo;type:tinyint(1);default:0;not null" json:"supportsDemo"`
	SupportsFreeSpins    bool       `gorm:"column:supports_free_spins;type:tinyint(1);default:0;not null" json:"supportsFreeSpins"`
	ProviderEnabled      bool       `gorm:"column:provider_enabled;type:tinyint(1);default:1;not null" json:"providerEnabled"`
	PlatformEnabled      bool       `gorm:"column:platform_enabled;type:tinyint(1);default:0;not null" json:"platformEnabled"`
	IsHot                bool       `gorm:"column:is_hot;type:tinyint(1);default:0;not null" json:"isHot"`
	IsFeatured           bool       `gorm:"column:is_featured;type:tinyint(1);default:0;not null" json:"isFeatured"`
	IsNew                bool       `gorm:"column:is_new;type:tinyint(1);default:0;not null" json:"isNew"`
	SupportedCountries   string     `gorm:"column:supported_countries;type:varchar(500);not null" json:"supportedCountries"`
	SupportedCurrencies  string     `gorm:"column:supported_currencies;type:varchar(500);not null" json:"supportedCurrencies"`
	SupportedDevices     string     `gorm:"column:supported_devices;type:varchar(255);not null" json:"supportedDevices"`
	RawMetadata          string     `gorm:"column:raw_metadata;type:json" json:"rawMetadata"`
	MetadataHash         string     `gorm:"column:metadata_hash;type:char(64);not null" json:"metadataHash"`
	LastSyncedAt         *time.Time `gorm:"column:last_synced_at;type:datetime(3)" json:"lastSyncedAt"`
	Badge                string     `gorm:"column:badge;type:varchar(32);not null" json:"badge"`
	PlayerCount          uint       `gorm:"column:player_count;type:int(11) unsigned;default:0;not null" json:"playerCount"`
	HotScore             int        `gorm:"column:hot_score;type:int(11);default:0;not null" json:"hotScore"`
	LaunchURL            string     `gorm:"column:launch_url;type:varchar(500);not null" json:"launchURL"`
	Status               int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 online, 2 maintenance, 3 offline
	Sort                 int        `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	ConfigVersion        uint64     `gorm:"column:config_version;type:bigint(20) unsigned;default:1;not null" json:"configVersion"`
}

// GameGamesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGamesColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":                true,
	"game_code":              true,
	"provider_id":            true,
	"provider_game_code":     true,
	"name":                   true,
	"name_i18n":              true,
	"slug":                   true,
	"category":               true,
	"platform_category_id":   true,
	"platform_category_code": true,
	"game_type":              true,
	"rtp":                    true,
	"volatility":             true,
	"cover_url":              true,
	"local_thumbnail_url":    true,
	"background_url":         true,
	"launch_image_url":       true,
	"maintenance_message":    true,
	"supports_demo":          true,
	"supports_free_spins":    true,
	"provider_enabled":       true,
	"platform_enabled":       true,
	"is_hot":                 true,
	"is_featured":            true,
	"is_new":                 true,
	"supported_countries":    true,
	"supported_currencies":   true,
	"supported_devices":      true,
	"raw_metadata":           true,
	"metadata_hash":          true,
	"last_synced_at":         true,
	"badge":                  true,
	"player_count":           true,
	"hot_score":              true,
	"launch_url":             true,
	"status":                 true,
	"sort":                   true,
	"config_version":         true,
}
