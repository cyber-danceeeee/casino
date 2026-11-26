package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameUserFavoriteGames Favorite games
type GameUserFavoriteGames struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	UserID uint64 `gorm:"column:user_id;type:bigint(20) unsigned;not null" json:"userID"`
	GameID uint64 `gorm:"column:game_id;type:bigint(20) unsigned;not null" json:"gameID"`
}

// GameUserFavoriteGamesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUserFavoriteGamesColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
	"user_id":    true,
	"game_id":    true,
}


