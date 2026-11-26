package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameInvitationWheelRoundsRequest request params
type CreateGameInvitationWheelRoundsRequest struct {
	RoundNo       string     `json:"roundNo" binding:""`
	Title         string     `json:"title" binding:""`
	StartAt       *time.Time `json:"startAt" binding:""`
	EndAt         *time.Time `json:"endAt" binding:""`
	TargetAmount  string     `json:"targetAmount" binding:""`
	CurrentAmount string     `json:"currentAmount" binding:""`
	Status        int        `json:"status" binding:""` // 1 active, 2 ended, 3 settled
}

// UpdateGameInvitationWheelRoundsByIDRequest request params
type UpdateGameInvitationWheelRoundsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RoundNo       string     `json:"roundNo" binding:""`
	Title         string     `json:"title" binding:""`
	StartAt       *time.Time `json:"startAt" binding:""`
	EndAt         *time.Time `json:"endAt" binding:""`
	TargetAmount  string     `json:"targetAmount" binding:""`
	CurrentAmount string     `json:"currentAmount" binding:""`
	Status        int        `json:"status" binding:""` // 1 active, 2 ended, 3 settled
}

// GameInvitationWheelRoundsObjDetail detail
type GameInvitationWheelRoundsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RoundNo       string     `json:"roundNo"`
	Title         string     `json:"title"`
	StartAt       *time.Time `json:"startAt"`
	EndAt         *time.Time `json:"endAt"`
	TargetAmount  string     `json:"targetAmount"`
	CurrentAmount string     `json:"currentAmount"`
	Status        int        `json:"status"` // 1 active, 2 ended, 3 settled
	CreatedAt     *time.Time `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

// CreateGameInvitationWheelRoundsReply only for api docs
type CreateGameInvitationWheelRoundsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameInvitationWheelRoundsByIDReply only for api docs
type DeleteGameInvitationWheelRoundsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameInvitationWheelRoundsByIDReply only for api docs
type UpdateGameInvitationWheelRoundsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameInvitationWheelRoundsByIDReply only for api docs
type GetGameInvitationWheelRoundsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelRounds GameInvitationWheelRoundsObjDetail `json:"gameInvitationWheelRounds"`
	} `json:"data"` // return data
}

// ListGameInvitationWheelRoundssRequest request params
type ListGameInvitationWheelRoundssRequest struct {
	query.Params
}

// ListGameInvitationWheelRoundssReply only for api docs
type ListGameInvitationWheelRoundssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelRoundss []GameInvitationWheelRoundsObjDetail `json:"gameInvitationWheelRoundss"`
	} `json:"data"` // return data
}
