package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameGames Game list
type GameGames struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID                  uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null;uniqueIndex:uk_game_games_scope_code,priority:2" json:"siteID"`
	GameCode                string     `gorm:"column:game_code;type:varchar(64);not null;uniqueIndex:uk_game_games_scope_code,priority:3" json:"gameCode"`
	ProviderID              uint64     `gorm:"column:provider_id;type:bigint(20) unsigned;default:0;not null" json:"providerID"`
	ProviderGameCode        string     `gorm:"column:provider_game_code;type:varchar(128);not null" json:"providerGameCode"`
	Name                    string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	NameI18n                string     `gorm:"column:name_i18n;type:json" json:"nameI18n"`
	Slug                    string     `gorm:"column:slug;type:varchar(160);not null" json:"slug"`
	Category                string     `gorm:"column:category;type:varchar(128);not null" json:"category"`
	PlatformCategoryID      uint64     `gorm:"column:platform_category_id;type:bigint(20) unsigned;default:0;not null" json:"platformCategoryID"`
	PlatformCategoryCode    string     `gorm:"column:platform_category_code;type:varchar(64);default:'';not null" json:"platformCategoryCode"`
	GameType                string     `gorm:"column:game_type;type:varchar(32);not null" json:"gameType"` // slot, chance, sports, live, fish, table
	RTP                     *float64   `gorm:"column:rtp;type:decimal(6,2)" json:"rtp"`
	Volatility              string     `gorm:"column:volatility;type:varchar(32);not null" json:"volatility"`
	CoverURL                string     `gorm:"column:cover_url;type:varchar(500);not null" json:"coverURL"`
	LocalThumbnailURL       string     `gorm:"column:local_thumbnail_url;type:varchar(500);not null" json:"localThumbnailURL"`
	BackgroundURL           string     `gorm:"column:background_url;type:varchar(500);not null" json:"backgroundURL"`
	LaunchImageURL          string     `gorm:"column:launch_image_url;type:varchar(500);not null" json:"launchImageURL"`
	MaintenanceMessage      string     `gorm:"column:maintenance_message;type:varchar(500);not null" json:"maintenanceMessage"`
	SupportsDemo            bool       `gorm:"column:supports_demo;type:tinyint(1);default:0;not null" json:"supportsDemo"`
	SupportsFreeSpins       bool       `gorm:"column:supports_free_spins;type:tinyint(1);default:0;not null" json:"supportsFreeSpins"`
	ProviderEnabled         bool       `gorm:"column:provider_enabled;type:tinyint(1);default:1;not null" json:"providerEnabled"`
	PlatformEnabled         bool       `gorm:"column:platform_enabled;type:tinyint(1);default:0;not null" json:"platformEnabled"`
	PlatformEnabledDraft    *bool      `gorm:"column:platform_enabled_draft;type:tinyint(1)" json:"platformEnabledDraft"`
	IsHot                   bool       `gorm:"column:is_hot;type:tinyint(1);default:0;not null" json:"isHot"`
	IsHotDraft              *bool      `gorm:"column:is_hot_draft;type:tinyint(1)" json:"isHotDraft"`
	IsFeatured              bool       `gorm:"column:is_featured;type:tinyint(1);default:0;not null" json:"isFeatured"`
	IsFeaturedDraft         *bool      `gorm:"column:is_featured_draft;type:tinyint(1)" json:"isFeaturedDraft"`
	IsNew                   bool       `gorm:"column:is_new;type:tinyint(1);default:0;not null" json:"isNew"`
	IsNewDraft              *bool      `gorm:"column:is_new_draft;type:tinyint(1)" json:"isNewDraft"`
	SupportedCountries      string     `gorm:"column:supported_countries;type:varchar(500);not null" json:"supportedCountries"`
	SupportedCurrencies     string     `gorm:"column:supported_currencies;type:varchar(500);not null" json:"supportedCurrencies"`
	SupportedDevices        string     `gorm:"column:supported_devices;type:varchar(255);not null" json:"supportedDevices"`
	RawMetadata             string     `gorm:"column:raw_metadata;type:json" json:"rawMetadata"`
	MetadataHash            string     `gorm:"column:metadata_hash;type:char(64);not null" json:"metadataHash"`
	LastSyncedAt            *time.Time `gorm:"column:last_synced_at;type:datetime(3)" json:"lastSyncedAt"`
	Badge                   string     `gorm:"column:badge;type:varchar(32);not null" json:"badge"`
	PlayerCount             uint       `gorm:"column:player_count;type:int(11) unsigned;default:0;not null" json:"playerCount"`
	HotScore                int        `gorm:"column:hot_score;type:int(11);default:0;not null" json:"hotScore"`
	LaunchURL               string     `gorm:"column:launch_url;type:varchar(500);not null" json:"launchURL"`
	Status                  int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 online, 2 maintenance, 3 offline
	StatusDraft             *int       `gorm:"column:status_draft;type:tinyint(4)" json:"statusDraft"`
	Sort                    int        `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	SortDraft               *int       `gorm:"column:sort_draft;type:int(11)" json:"sortDraft"`
	MaintenanceMessageDraft string     `gorm:"column:maintenance_message_draft;type:varchar(500)" json:"maintenanceMessageDraft"`
	CategoryIDsDraft        string     `gorm:"column:category_ids_draft;type:json" json:"categoryIDsDraft"`
	ConfigStatus            string     `gorm:"column:config_status;type:varchar(24);default:PUBLISHED;not null" json:"configStatus"`
	ConfigVersion           uint64     `gorm:"column:config_version;type:bigint(20) unsigned;default:1;not null" json:"configVersion"`
	ConfigPublishedAt       *time.Time `gorm:"column:config_published_at;type:datetime(3)" json:"configPublishedAt"`
	ConfigPublishedBy       uint64     `gorm:"column:config_published_by;type:bigint(20) unsigned;default:0;not null" json:"configPublishedBy"`
	DraftUpdatedAt          *time.Time `gorm:"column:draft_updated_at;type:datetime(3)" json:"draftUpdatedAt"`
	DraftUpdatedBy          uint64     `gorm:"column:draft_updated_by;type:bigint(20) unsigned;default:0;not null" json:"draftUpdatedBy"`
}

// GameGamesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameGamesColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":                   true,
	"game_code":                 true,
	"provider_id":               true,
	"provider_game_code":        true,
	"name":                      true,
	"name_i18n":                 true,
	"slug":                      true,
	"category":                  true,
	"platform_category_id":      true,
	"platform_category_code":    true,
	"game_type":                 true,
	"rtp":                       true,
	"volatility":                true,
	"cover_url":                 true,
	"local_thumbnail_url":       true,
	"background_url":            true,
	"launch_image_url":          true,
	"maintenance_message":       true,
	"supports_demo":             true,
	"supports_free_spins":       true,
	"provider_enabled":          true,
	"platform_enabled":          true,
	"platform_enabled_draft":    true,
	"is_hot":                    true,
	"is_hot_draft":              true,
	"is_featured":               true,
	"is_featured_draft":         true,
	"is_new":                    true,
	"is_new_draft":              true,
	"supported_countries":       true,
	"supported_currencies":      true,
	"supported_devices":         true,
	"raw_metadata":              true,
	"metadata_hash":             true,
	"last_synced_at":            true,
	"badge":                     true,
	"player_count":              true,
	"hot_score":                 true,
	"launch_url":                true,
	"status":                    true,
	"status_draft":              true,
	"sort":                      true,
	"sort_draft":                true,
	"maintenance_message_draft": true,
	"category_ids_draft":        true,
	"config_status":             true,
	"config_version":            true,
	"config_published_at":       true,
	"config_published_by":       true,
	"draft_updated_at":          true,
	"draft_updated_by":          true,
}
