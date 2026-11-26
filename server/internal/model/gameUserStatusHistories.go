package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

type GameUserStatusHistories struct {
	sgorm.Model `gorm:"embedded"`

	UserID           uint64 `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	ChangeType       string `gorm:"column:change_type;type:varchar(32);not null" json:"changeType"`
	BeforeValue      string `gorm:"column:before_value;type:varchar(64);not null" json:"beforeValue"`
	AfterValue       string `gorm:"column:after_value;type:varchar(64);not null" json:"afterValue"`
	Reason           string `gorm:"column:reason;type:varchar(500);not null" json:"reason"`
	OperatorID       uint64 `gorm:"column:operator_id;type:bigint unsigned;not null" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username;type:varchar(64);not null" json:"operatorUsername"`
}

func (GameUserStatusHistories) TableName() string { return "game_user_status_histories" }
