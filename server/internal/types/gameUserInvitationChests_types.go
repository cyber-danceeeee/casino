package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameUserInvitationChestsRequest request params
type CreateGameUserInvitationChestsRequest struct {
	UserID          uint64     `json:"userID" binding:""`
	ChestConfigID   uint64     `json:"chestConfigID" binding:""`
	ProgressInvites uint       `json:"progressInvites" binding:""`
	Status          int        `json:"status" binding:""` // 1 locked, 2 available, 3 opened, 4 expired
	OpenedAt        *time.Time `json:"openedAt" binding:""`
}

// UpdateGameUserInvitationChestsByIDRequest request params
type UpdateGameUserInvitationChestsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID          uint64     `json:"userID" binding:""`
	ChestConfigID   uint64     `json:"chestConfigID" binding:""`
	ProgressInvites uint       `json:"progressInvites" binding:""`
	Status          int        `json:"status" binding:""` // 1 locked, 2 available, 3 opened, 4 expired
	OpenedAt        *time.Time `json:"openedAt" binding:""`
}

// GameUserInvitationChestsObjDetail detail
type GameUserInvitationChestsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID          uint64     `json:"userID"`
	ChestConfigID   uint64     `json:"chestConfigID"`
	ProgressInvites uint       `json:"progressInvites"`
	Status          int        `json:"status"` // 1 locked, 2 available, 3 opened, 4 expired
	OpenedAt        *time.Time `json:"openedAt"`
	CreatedAt       *time.Time `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
}

// CreateGameUserInvitationChestsReply only for api docs
type CreateGameUserInvitationChestsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserInvitationChestsByIDReply only for api docs
type DeleteGameUserInvitationChestsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserInvitationChestsByIDReply only for api docs
type UpdateGameUserInvitationChestsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserInvitationChestsByIDReply only for api docs
type GetGameUserInvitationChestsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserInvitationChests GameUserInvitationChestsObjDetail `json:"gameUserInvitationChests"`
	} `json:"data"` // return data
}

// ListGameUserInvitationChestssRequest request params
type ListGameUserInvitationChestssRequest struct {
	query.Params
}

// ListGameUserInvitationChestssReply only for api docs
type ListGameUserInvitationChestssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserInvitationChestss []GameUserInvitationChestsObjDetail `json:"gameUserInvitationChestss"`
	} `json:"data"` // return data
}
