package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameSupportMessagesRequest request params
type CreateGameSupportMessagesRequest struct {
	TicketID    uint64     `json:"ticketID" binding:""`
	SenderType  string     `json:"senderType" binding:""` // user, admin, system
	SenderID    uint64     `json:"senderID" binding:""`
	ContentType string     `json:"contentType" binding:""`
	Content     string     `json:"content" binding:""`
	ReadAt      *time.Time `json:"readAt" binding:""`
}

// UpdateGameSupportMessagesByIDRequest request params
type UpdateGameSupportMessagesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	TicketID    uint64     `json:"ticketID" binding:""`
	SenderType  string     `json:"senderType" binding:""` // user, admin, system
	SenderID    uint64     `json:"senderID" binding:""`
	ContentType string     `json:"contentType" binding:""`
	Content     string     `json:"content" binding:""`
	ReadAt      *time.Time `json:"readAt" binding:""`
}

// GameSupportMessagesObjDetail detail
type GameSupportMessagesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	TicketID    uint64     `json:"ticketID"`
	SenderType  string     `json:"senderType"` // user, admin, system
	SenderID    uint64     `json:"senderID"`
	ContentType string     `json:"contentType"`
	Content     string     `json:"content"`
	ReadAt      *time.Time `json:"readAt"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameSupportMessagesReply only for api docs
type CreateGameSupportMessagesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSupportMessagesByIDReply only for api docs
type DeleteGameSupportMessagesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSupportMessagesByIDReply only for api docs
type UpdateGameSupportMessagesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSupportMessagesByIDReply only for api docs
type GetGameSupportMessagesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportMessages GameSupportMessagesObjDetail `json:"gameSupportMessages"`
	} `json:"data"` // return data
}

// ListGameSupportMessagessRequest request params
type ListGameSupportMessagessRequest struct {
	query.Params
}

// ListGameSupportMessagessReply only for api docs
type ListGameSupportMessagessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportMessagess []GameSupportMessagesObjDetail `json:"gameSupportMessagess"`
	} `json:"data"` // return data
}
