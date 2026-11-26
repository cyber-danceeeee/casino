package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

type GameUserRiskEvents struct {
	sgorm.Model `gorm:"embedded"`

	UserID           uint64 `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	EventType        string `gorm:"column:event_type;type:varchar(64);not null" json:"eventType"`
	RiskLevel        string `gorm:"column:risk_level;type:varchar(16);not null" json:"riskLevel"`
	Source           string `gorm:"column:source;type:varchar(32);not null" json:"source"`
	Description      string `gorm:"column:description;type:varchar(1000);not null" json:"description"`
	OperatorID       uint64 `gorm:"column:operator_id;type:bigint unsigned;not null" json:"operatorID"`
	OperatorUsername string `gorm:"column:operator_username;type:varchar(64);not null" json:"operatorUsername"`
}

func (GameUserRiskEvents) TableName() string { return "game_user_risk_events" }
