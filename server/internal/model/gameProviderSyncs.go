package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameProviderSyncs records the latest catalog sync result for each provider.
type GameProviderSyncs struct {
	sgorm.Model `gorm:"embedded"`

	TenantID                 uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID                   uint64     `gorm:"column:site_id" json:"siteID"`
	ProviderCode             string     `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	LastSuccessfulSyncAt     *time.Time `gorm:"column:last_successful_sync_at;type:datetime(3)" json:"lastSuccessfulSyncAt"`
	LastAttemptAt            *time.Time `gorm:"column:last_attempt_at;type:datetime(3)" json:"lastAttemptAt"`
	LastStatus               string     `gorm:"column:last_status;type:varchar(32);not null" json:"lastStatus"`
	LastError                string     `gorm:"column:last_error;type:varchar(500);not null" json:"lastError"`
	LastSyncedGameCount      int        `gorm:"column:last_synced_game_count;type:int(11);default:0;not null" json:"lastSyncedGameCount"`
	LastInsertedGameCount    int        `gorm:"column:last_inserted_game_count;type:int(11);default:0;not null" json:"lastInsertedGameCount"`
	LastUpdatedGameCount     int        `gorm:"column:last_updated_game_count;type:int(11);default:0;not null" json:"lastUpdatedGameCount"`
	LastUnchangedGameCount   int        `gorm:"column:last_unchanged_game_count;type:int(11);default:0;not null" json:"lastUnchangedGameCount"`
	LastProviderErrorMessage string     `gorm:"column:last_provider_error_message;type:varchar(500);not null" json:"lastProviderErrorMessage"`
}

func (GameProviderSyncs) TableName() string {
	return "game_provider_syncs"
}
