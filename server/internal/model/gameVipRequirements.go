package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameVipRequirements VIP level-up and retain requirements
type GameVipRequirements struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID        uint64           `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID          uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	VipLevelID      uint64           `gorm:"column:vip_level_id;type:bigint(20) unsigned;not null" json:"vipLevelID"`
	RequirementType string           `gorm:"column:requirement_type;type:varchar(32);not null" json:"requirementType"` // level_up, retain
	TargetLevelNo   uint             `gorm:"column:target_level_no;type:int(11) unsigned;default:0;not null" json:"targetLevelNo"`
	PeriodDays      uint             `gorm:"column:period_days;type:int(11) unsigned;default:0;not null" json:"periodDays"`
	DepositAmount   *decimal.Decimal `gorm:"column:deposit_amount;type:decimal(18,2);default:0.00;not null" json:"depositAmount"`
	ValidBetAmount  *decimal.Decimal `gorm:"column:valid_bet_amount;type:decimal(18,2);default:0.00;not null" json:"validBetAmount"`
}

// GameVipRequirementsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameVipRequirementsColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"tenant_id":        true,
	"site_id":          true,
	"vip_level_id":     true,
	"requirement_type": true,
	"target_level_no":  true,
	"period_days":      true,
	"deposit_amount":   true,
	"valid_bet_amount": true,
}
