package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSupportMessages Support ticket messages
type GameSupportMessages struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TicketID    uint64     `gorm:"column:ticket_id;type:bigint(20) unsigned;not null" json:"ticketID"`
	SenderType  string     `gorm:"column:sender_type;type:varchar(32);not null" json:"senderType"` // user, admin, system
	SenderID    uint64     `gorm:"column:sender_id;type:bigint(20) unsigned;default:0;not null" json:"senderID"`
	ContentType string     `gorm:"column:content_type;type:varchar(32);default:text;not null" json:"contentType"`
	Content     string     `gorm:"column:content;type:mediumtext" json:"content"`
	ReadAt      *time.Time `gorm:"column:read_at;type:datetime(3)" json:"readAt"`
}

// GameSupportMessagesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSupportMessagesColumnNames = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
	"ticket_id":    true,
	"sender_type":  true,
	"sender_id":    true,
	"content_type": true,
	"content":      true,
	"read_at":      true,
}
