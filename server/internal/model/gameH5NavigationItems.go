package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameH5NavigationItems stores tenant-owned H5 operation entry drafts and published values.
type GameH5NavigationItems struct {
	sgorm.Model `gorm:"embedded"`

	TenantID           uint64     `gorm:"column:tenant_id;type:bigint unsigned;not null;default:0;uniqueIndex:uk_h5_navigation_scope_code,priority:1" json:"tenantID"`
	SiteID             uint64     `gorm:"column:site_id;type:bigint unsigned;not null;default:0;uniqueIndex:uk_h5_navigation_scope_code,priority:2" json:"siteID"`
	PlatformSourceID   uint64     `gorm:"column:platform_source_id;type:bigint unsigned;not null;default:0;index:idx_h5_navigation_platform_source" json:"platformSourceID"`
	MenuCode           string     `gorm:"column:menu_code;type:varchar(64);not null;uniqueIndex:uk_h5_navigation_scope_code,priority:3" json:"menuCode"`
	LabelZh            string     `gorm:"column:label_zh;type:varchar(64);not null;default:''" json:"labelZh"`
	LabelEn            string     `gorm:"column:label_en;type:varchar(64);not null;default:''" json:"labelEn"`
	LabelsJSON         string     `gorm:"column:labels_json;type:text" json:"-"`
	IconAssetID        uint64     `gorm:"column:icon_asset_id;type:bigint unsigned;not null;default:0" json:"iconAssetID"`
	LottieAssetID      uint64     `gorm:"column:lottie_asset_id;type:bigint unsigned;not null;default:0" json:"lottieAssetID"`
	SoundAssetID       uint64     `gorm:"column:sound_asset_id;type:bigint unsigned;not null;default:0" json:"soundAssetID"`
	IconAssetIDDraft   uint64     `gorm:"column:icon_asset_id_draft;type:bigint unsigned;not null;default:0" json:"iconAssetIDDraft"`
	LottieAssetIDDraft uint64     `gorm:"column:lottie_asset_id_draft;type:bigint unsigned;not null;default:0" json:"lottieAssetIDDraft"`
	SoundAssetIDDraft  uint64     `gorm:"column:sound_asset_id_draft;type:bigint unsigned;not null;default:0" json:"soundAssetIDDraft"`
	DisplayPosition    string     `gorm:"column:display_position;type:varchar(32);not null;default:'BOTTOM_NAV'" json:"displayPosition"`
	ActionType         string     `gorm:"column:action_type;type:varchar(32);not null;default:'OPEN_INTERNAL_PAGE'" json:"actionType"`
	ActionTarget       string     `gorm:"column:action_target;type:varchar(500);not null;default:''" json:"actionTarget"`
	PageRoute          string     `gorm:"column:page_route;type:varchar(255);not null;default:''" json:"pageRoute"`
	ProviderCode       string     `gorm:"column:provider_code;type:varchar(64);not null;default:''" json:"providerCode"`
	GameID             uint64     `gorm:"column:game_id;type:bigint unsigned;not null;default:0;index:idx_h5_navigation_game" json:"gameID"`
	ProviderGameCode   string     `gorm:"column:provider_game_code;type:varchar(128);not null;default:''" json:"providerGameCode"`
	GameName           string     `gorm:"column:game_name;type:varchar(128);not null;default:''" json:"gameName"`
	SortOrder          int        `gorm:"column:sort_order;type:int;not null;default:0" json:"sortOrder"`
	BadgeType          string     `gorm:"column:badge_type;type:varchar(16);not null;default:'NONE'" json:"badgeType"`
	VisibilityScope    string     `gorm:"column:visibility_scope;type:varchar(16);not null;default:'ALL'" json:"visibilityScope"`
	CountryCodes       string     `gorm:"column:country_codes;type:text" json:"-"`
	LanguageCodes      string     `gorm:"column:language_codes;type:text" json:"-"`
	StartAt            *time.Time `gorm:"column:start_at" json:"startAt"`
	EndAt              *time.Time `gorm:"column:end_at" json:"endAt"`
	Enabled            bool       `gorm:"column:enabled;type:tinyint(1);not null;default:1" json:"enabled"`
	DraftPayload       string     `gorm:"column:draft_payload;type:longtext" json:"-"`
	PublishedPayload   string     `gorm:"column:published_payload;type:longtext" json:"-"`
	PublishStatus      string     `gorm:"column:publish_status;type:varchar(16);not null;default:'DRAFT'" json:"publishStatus"`
	ConfigVersion      uint64     `gorm:"column:config_version;type:bigint unsigned;not null;default:0" json:"configVersion"`
	CreatedBy          uint64     `gorm:"column:created_by;type:bigint unsigned;not null;default:0" json:"createdBy"`
	UpdatedBy          uint64     `gorm:"column:updated_by;type:bigint unsigned;not null;default:0" json:"updatedBy"`
	DraftUpdatedAt     *time.Time `gorm:"column:draft_updated_at" json:"draftUpdatedAt"`
	PublishedAt        *time.Time `gorm:"column:published_at" json:"publishedAt"`
	PublishedBy        uint64     `gorm:"column:published_by;type:bigint unsigned;not null;default:0" json:"publishedBy"`
}

func (GameH5NavigationItems) TableName() string { return "game_h5_navigation_items" }
