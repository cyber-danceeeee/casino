package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameLobbySections controls H5 lobby layout and section ordering.
type GameLobbySections struct {
	sgorm.Model `gorm:"embedded"`

	TenantID            uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID              uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	SectionCode         string     `gorm:"column:section_code;type:varchar(64);not null" json:"sectionCode"`
	SectionType         string     `gorm:"column:section_type;type:varchar(32);not null" json:"sectionType"`
	TitleZh             string     `gorm:"column:title_zh;type:varchar(128);not null" json:"titleZh"`
	TitleEn             string     `gorm:"column:title_en;type:varchar(128);not null" json:"titleEn"`
	DisplayStyle        string     `gorm:"column:display_style;type:varchar(64);not null" json:"displayStyle"`
	SourceType          string     `gorm:"column:source_type;type:varchar(32);not null" json:"sourceType"`
	SourceRef           string     `gorm:"column:source_ref;type:varchar(128);not null" json:"sourceRef"`
	MaxItems            int        `gorm:"column:max_items;type:int(11);default:12;not null" json:"maxItems"`
	SortOrder           int        `gorm:"column:sort_order;type:int(11);default:0;not null" json:"sortOrder"`
	Enabled             bool       `gorm:"column:enabled;type:tinyint(1);default:1;not null" json:"enabled"`
	SupportedCountries  string     `gorm:"column:supported_countries;type:varchar(500);not null" json:"supportedCountries"`
	SupportedCurrencies string     `gorm:"column:supported_currencies;type:varchar(500);not null" json:"supportedCurrencies"`
	ConfigJSON          string     `gorm:"column:config_json;type:json" json:"configJSON"`
	DraftPayload        string     `gorm:"column:draft_payload;type:json" json:"draftPayload"`
	PublishStatus       string     `gorm:"column:publish_status;type:varchar(24);default:PUBLISHED;not null" json:"publishStatus"`
	ConfigVersion       uint64     `gorm:"column:config_version;type:bigint(20) unsigned;default:1;not null" json:"configVersion"`
	PublishedAt         *time.Time `gorm:"column:published_at" json:"publishedAt"`
	PublishedBy         uint64     `gorm:"column:published_by;type:bigint(20) unsigned;default:0;not null" json:"publishedBy"`
	DraftUpdatedAt      *time.Time `gorm:"column:draft_updated_at" json:"draftUpdatedAt"`
	DraftUpdatedBy      uint64     `gorm:"column:draft_updated_by;type:bigint(20) unsigned;default:0;not null" json:"draftUpdatedBy"`
}

func (GameLobbySections) TableName() string { return "game_lobby_sections" }
