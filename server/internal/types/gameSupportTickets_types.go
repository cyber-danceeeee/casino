package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameSupportTicketsRequest request params
type CreateGameSupportTicketsRequest struct {
	TicketNo      string     `json:"ticketNo" binding:""`
	UserID        uint64     `json:"userID" binding:""`
	CategoryCode  string     `json:"categoryCode" binding:""`
	Subject       string     `json:"subject" binding:""`
	Status        int        `json:"status" binding:""` // 1 open, 2 pending, 3 closed
	LastMessageAt *time.Time `json:"lastMessageAt" binding:""`
}

// UpdateGameSupportTicketsByIDRequest request params
type UpdateGameSupportTicketsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	TicketNo      string     `json:"ticketNo" binding:""`
	UserID        uint64     `json:"userID" binding:""`
	CategoryCode  string     `json:"categoryCode" binding:""`
	Subject       string     `json:"subject" binding:""`
	Status        int        `json:"status" binding:""` // 1 open, 2 pending, 3 closed
	LastMessageAt *time.Time `json:"lastMessageAt" binding:""`
}

// GameSupportTicketsObjDetail detail
type GameSupportTicketsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	TicketNo      string     `json:"ticketNo"`
	UserID        uint64     `json:"userID"`
	CategoryCode  string     `json:"categoryCode"`
	Subject       string     `json:"subject"`
	Status        int        `json:"status"` // 1 open, 2 pending, 3 closed
	LastMessageAt *time.Time `json:"lastMessageAt"`
	CreatedAt     *time.Time `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

// CreateGameSupportTicketsReply only for api docs
type CreateGameSupportTicketsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSupportTicketsByIDReply only for api docs
type DeleteGameSupportTicketsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSupportTicketsByIDReply only for api docs
type UpdateGameSupportTicketsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSupportTicketsByIDReply only for api docs
type GetGameSupportTicketsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportTickets GameSupportTicketsObjDetail `json:"gameSupportTickets"`
	} `json:"data"` // return data
}

// ListGameSupportTicketssRequest request params
type ListGameSupportTicketssRequest struct {
	query.Params
}

// ListGameSupportTicketssReply only for api docs
type ListGameSupportTicketssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportTicketss []GameSupportTicketsObjDetail `json:"gameSupportTicketss"`
	} `json:"data"` // return data
}
