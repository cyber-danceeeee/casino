package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameAdminOperationLogs are append-only audit records.
type GameAdminOperationLogs struct {
	sgorm.Model `gorm:"embedded"`

	SiteID           uint64           `gorm:"column:site_id;type:bigint unsigned;not null" json:"siteID"`
	OperatorID       uint64           `gorm:"column:operator_id;type:bigint unsigned;not null" json:"operatorID"`
	OperatorUsername string           `gorm:"column:operator_username;type:varchar(64);not null" json:"operatorUsername"`
	Permission       string           `gorm:"column:permission;type:varchar(128);not null" json:"permission"`
	Action           string           `gorm:"column:action;type:varchar(128);not null" json:"action"`
	TargetType       string           `gorm:"column:target_type;type:varchar(64);not null" json:"targetType"`
	TargetID         uint64           `gorm:"column:target_id;type:bigint unsigned;not null" json:"targetID"`
	TargetUserID     uint64           `gorm:"column:target_user_id;type:bigint unsigned;not null" json:"targetUserID"`
	TargetDeviceID   uint64           `gorm:"column:target_device_id;type:bigint unsigned;not null" json:"targetDeviceID"`
	BeforeState      string           `gorm:"column:before_state;type:json" json:"beforeState"`
	AfterState       string           `gorm:"column:after_state;type:json" json:"afterState"`
	Amount           *decimal.Decimal `gorm:"column:amount;type:decimal(18,2)" json:"amount"`
	Currency         string           `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	Reason           string           `gorm:"column:reason;type:varchar(500);not null" json:"reason"`
	IPAddress        string           `gorm:"column:ip_address;type:varchar(64);not null" json:"ipAddress"`
	UserAgent        string           `gorm:"column:user_agent;type:varchar(500);not null" json:"userAgent"`
	CorrelationID    string           `gorm:"column:correlation_id;type:varchar(128);not null" json:"correlationID"`
}

func (GameAdminOperationLogs) TableName() string { return "game_admin_operation_logs" }
