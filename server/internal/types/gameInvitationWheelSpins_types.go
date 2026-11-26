package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameInvitationWheelSpinsRequest request params
type CreateGameInvitationWheelSpinsRequest struct {
	SpinNo  string `json:"spinNo" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	RoundID  uint64 `json:"roundID" binding:""`
	PrizeID  uint64 `json:"prizeID" binding:""`
	RewardType  string `json:"rewardType" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 won, 3 failed
	SpunAt  *time.Time `json:"spunAt" binding:""`
}

// UpdateGameInvitationWheelSpinsByIDRequest request params
type UpdateGameInvitationWheelSpinsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	SpinNo  string `json:"spinNo" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	RoundID  uint64 `json:"roundID" binding:""`
	PrizeID  uint64 `json:"prizeID" binding:""`
	RewardType  string `json:"rewardType" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 won, 3 failed
	SpunAt  *time.Time `json:"spunAt" binding:""`
}

// GameInvitationWheelSpinsObjDetail detail
type GameInvitationWheelSpinsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	SpinNo  string `json:"spinNo"`
	UserID  uint64 `json:"userID"`
	RoundID  uint64 `json:"roundID"`
	PrizeID  uint64 `json:"prizeID"`
	RewardType  string `json:"rewardType"`
	RewardAmount  string `json:"rewardAmount"`
	Status  int `json:"status"` // 1 pending, 2 won, 3 failed
	SpunAt  *time.Time `json:"spunAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameInvitationWheelSpinsReply only for api docs
type CreateGameInvitationWheelSpinsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameInvitationWheelSpinsByIDReply only for api docs
type DeleteGameInvitationWheelSpinsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameInvitationWheelSpinsByIDReply only for api docs
type UpdateGameInvitationWheelSpinsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameInvitationWheelSpinsByIDReply only for api docs
type GetGameInvitationWheelSpinsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelSpins GameInvitationWheelSpinsObjDetail `json:"gameInvitationWheelSpins"`
	} `json:"data"` // return data
}

// ListGameInvitationWheelSpinssRequest request params
type ListGameInvitationWheelSpinssRequest struct {
	query.Params
}

// ListGameInvitationWheelSpinssReply only for api docs
type ListGameInvitationWheelSpinssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelSpinss []GameInvitationWheelSpinsObjDetail `json:"gameInvitationWheelSpinss"`
	} `json:"data"` // return data
}
