package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

// GameUserAdminNotes are immutable notes visible only to administrators.
type GameUserAdminNotes struct {
	sgorm.Model `gorm:"embedded"`

	UserID           uint64 `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	OperatorID       uint64 `gorm:"column:operator_id;type:bigint unsigned;not null" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username;type:varchar(64);not null" json:"operatorUsername"`
	Content          string `gorm:"column:content;type:varchar(1000);not null" json:"content"`
}

func (GameUserAdminNotes) TableName() string { return "game_user_admin_notes" }
