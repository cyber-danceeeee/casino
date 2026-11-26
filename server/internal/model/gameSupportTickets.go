package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSupportTickets Support tickets
type GameSupportTickets struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TicketNo      string     `gorm:"column:ticket_no;type:varchar(64);not null" json:"ticketNo"`
	UserID        uint64     `gorm:"column:user_id;type:bigint(20) unsigned;default:0;not null" json:"userID"`
	CategoryCode  string     `gorm:"column:category_code;type:varchar(64);not null" json:"categoryCode"`
	Subject       string     `gorm:"column:subject;type:varchar(255);not null" json:"subject"`
	Status        int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 open, 2 pending, 3 closed
	LastMessageAt *time.Time `gorm:"column:last_message_at;type:datetime(3)" json:"lastMessageAt"`
}

// GameSupportTicketsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSupportTicketsColumnNames = map[string]bool{
	"id":              true,
	"created_at":      true,
	"updated_at":      true,
	"deleted_at":      true,
	"ticket_no":       true,
	"user_id":         true,
	"category_code":   true,
	"subject":         true,
	"status":          true,
	"last_message_at": true,
}
