package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameInvitesRequest request params
type CreateGameInvitesRequest struct {
	InviterUserID  uint64     `json:"inviterUserID" binding:""`
	InviteeUserID  uint64     `json:"inviteeUserID" binding:""`
	InviteCode     string     `json:"inviteCode" binding:""`
	InviteePhone   string     `json:"inviteePhone" binding:""`
	Status         int        `json:"status" binding:""` // 1 invited, 2 registered, 3 valid, 4 invalid
	RegisteredAt   *time.Time `json:"registeredAt" binding:""`
	FirstDepositAt *time.Time `json:"firstDepositAt" binding:""`
	ValidAt        *time.Time `json:"validAt" binding:""`
}

// UpdateGameInvitesByIDRequest request params
type UpdateGameInvitesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	InviterUserID  uint64     `json:"inviterUserID" binding:""`
	InviteeUserID  uint64     `json:"inviteeUserID" binding:""`
	InviteCode     string     `json:"inviteCode" binding:""`
	InviteePhone   string     `json:"inviteePhone" binding:""`
	Status         int        `json:"status" binding:""` // 1 invited, 2 registered, 3 valid, 4 invalid
	RegisteredAt   *time.Time `json:"registeredAt" binding:""`
	FirstDepositAt *time.Time `json:"firstDepositAt" binding:""`
	ValidAt        *time.Time `json:"validAt" binding:""`
}

// GameInvitesObjDetail detail
type GameInvitesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	InviterUserID  uint64     `json:"inviterUserID"`
	InviteeUserID  uint64     `json:"inviteeUserID"`
	InviteCode     string     `json:"inviteCode"`
	InviteePhone   string     `json:"inviteePhone"`
	Status         int        `json:"status"` // 1 invited, 2 registered, 3 valid, 4 invalid
	RegisteredAt   *time.Time `json:"registeredAt"`
	FirstDepositAt *time.Time `json:"firstDepositAt"`
	ValidAt        *time.Time `json:"validAt"`
	CreatedAt      *time.Time `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

// CreateGameInvitesReply only for api docs
type CreateGameInvitesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameInvitesByIDReply only for api docs
type DeleteGameInvitesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameInvitesByIDReply only for api docs
type UpdateGameInvitesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameInvitesByIDReply only for api docs
type GetGameInvitesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvites GameInvitesObjDetail `json:"gameInvites"`
	} `json:"data"` // return data
}

// ListGameInvitessRequest request params
type ListGameInvitessRequest struct {
	query.Params
}

// ListGameInvitessReply only for api docs
type ListGameInvitessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitess []GameInvitesObjDetail `json:"gameInvitess"`
	} `json:"data"` // return data
}
