package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameInvites Referral records
type GameInvites struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	InviterUserID  uint64     `gorm:"column:inviter_user_id;type:bigint(20) unsigned;not null" json:"inviterUserID"`
	InviteeUserID  uint64     `gorm:"column:invitee_user_id;type:bigint(20) unsigned;default:0;not null" json:"inviteeUserID"`
	InviteCode     string     `gorm:"column:invite_code;type:varchar(32);not null" json:"inviteCode"`
	InviteePhone   string     `gorm:"column:invitee_phone;type:varchar(32);not null" json:"inviteePhone"`
	Status         int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 invited, 2 registered, 3 valid, 4 invalid
	RegisteredAt   *time.Time `gorm:"column:registered_at;type:datetime(3)" json:"registeredAt"`
	FirstDepositAt *time.Time `gorm:"column:first_deposit_at;type:datetime(3)" json:"firstDepositAt"`
	ValidAt        *time.Time `gorm:"column:valid_at;type:datetime(3)" json:"validAt"`
}

// GameInvitesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameInvitesColumnNames = map[string]bool{
	"id":               true,
	"created_at":       true,
	"updated_at":       true,
	"deleted_at":       true,
	"inviter_user_id":  true,
	"invitee_user_id":  true,
	"invite_code":      true,
	"invitee_phone":    true,
	"status":           true,
	"registered_at":    true,
	"first_deposit_at": true,
	"valid_at":         true,
}
