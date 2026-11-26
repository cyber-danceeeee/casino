package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

// GameProviderPlayers maps local users to provider-side player credentials.
type GameProviderPlayers struct {
	sgorm.Model `gorm:"embedded"`

	TenantID               uint64 `gorm:"column:tenant_id" json:"tenantID"`
	SiteID                 uint64 `gorm:"column:site_id" json:"siteID"`
	ProviderCode           string `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	UserID                 uint64 `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	ProviderPlayerID       string `gorm:"column:provider_player_id;type:varchar(128);not null" json:"providerPlayerID"`
	ProviderPlayerPassword string `gorm:"column:provider_player_password;type:varchar(255);not null" json:"-"`
	Currency               string `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	Status                 int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 disabled
}

func (GameProviderPlayers) TableName() string {
	return "game_provider_players"
}

// GameProviderPlayersColumnNames Whitelist for custom query fields to prevent sql injection attacks.
var GameProviderPlayersColumnNames = map[string]bool{
	"id":                       true,
	"created_at":               true,
	"updated_at":               true,
	"deleted_at":               true,
	"tenant_id":                true,
	"site_id":                  true,
	"provider_code":            true,
	"user_id":                  true,
	"provider_player_id":       true,
	"provider_player_password": true,
	"currency":                 true,
	"status":                   true,
}
