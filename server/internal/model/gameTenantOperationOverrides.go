package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

// GameTenantOperationOverrides stores tenant visibility choices for platform-owned operation defaults.
type GameTenantOperationOverrides struct {
	sgorm.Model `gorm:"embedded"`

	TenantID     uint64 `gorm:"column:tenant_id;type:bigint unsigned;not null;uniqueIndex:uk_tenant_operation_override,priority:1" json:"tenantID"`
	SiteID       uint64 `gorm:"column:site_id;type:bigint unsigned;not null;uniqueIndex:uk_tenant_operation_override,priority:2" json:"siteID"`
	ResourceType string `gorm:"column:resource_type;type:varchar(32);not null;uniqueIndex:uk_tenant_operation_override,priority:3" json:"resourceType"`
	ResourceID   uint64 `gorm:"column:resource_id;type:bigint unsigned;not null;uniqueIndex:uk_tenant_operation_override,priority:4" json:"resourceID"`
	Enabled      bool   `gorm:"column:enabled;type:tinyint(1);not null;default:1" json:"enabled"`
	CreatedBy    uint64 `gorm:"column:created_by;type:bigint unsigned;not null;default:0" json:"createdBy"`
	UpdatedBy    uint64 `gorm:"column:updated_by;type:bigint unsigned;not null;default:0" json:"updatedBy"`
}

func (GameTenantOperationOverrides) TableName() string {
	return "game_tenant_operation_overrides"
}
