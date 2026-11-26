package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameInvitationWheelPrizesRequest request params
type CreateGameInvitationWheelPrizesRequest struct {
	PrizeCode  string `json:"prizeCode" binding:""`
	Label  string `json:"label" binding:""`
	RewardType  string `json:"rewardType" binding:""` // cash, bonus, spin, chest, empty
	MinAmount  string `json:"minAmount" binding:""`
	MaxAmount  string `json:"maxAmount" binding:""`
	Weight  uint `json:"weight" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameInvitationWheelPrizesByIDRequest request params
type UpdateGameInvitationWheelPrizesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	PrizeCode  string `json:"prizeCode" binding:""`
	Label  string `json:"label" binding:""`
	RewardType  string `json:"rewardType" binding:""` // cash, bonus, spin, chest, empty
	MinAmount  string `json:"minAmount" binding:""`
	MaxAmount  string `json:"maxAmount" binding:""`
	Weight  uint `json:"weight" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameInvitationWheelPrizesObjDetail detail
type GameInvitationWheelPrizesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	PrizeCode  string `json:"prizeCode"`
	Label  string `json:"label"`
	RewardType  string `json:"rewardType"` // cash, bonus, spin, chest, empty
	MinAmount  string `json:"minAmount"`
	MaxAmount  string `json:"maxAmount"`
	Weight  uint `json:"weight"`
	IconURL  string `json:"iconURL"`
	Sort  int `json:"sort"`
	Status  int `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameInvitationWheelPrizesReply only for api docs
type CreateGameInvitationWheelPrizesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameInvitationWheelPrizesByIDReply only for api docs
type DeleteGameInvitationWheelPrizesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameInvitationWheelPrizesByIDReply only for api docs
type UpdateGameInvitationWheelPrizesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameInvitationWheelPrizesByIDReply only for api docs
type GetGameInvitationWheelPrizesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelPrizes GameInvitationWheelPrizesObjDetail `json:"gameInvitationWheelPrizes"`
	} `json:"data"` // return data
}

// ListGameInvitationWheelPrizessRequest request params
type ListGameInvitationWheelPrizessRequest struct {
	query.Params
}

// ListGameInvitationWheelPrizessReply only for api docs
type ListGameInvitationWheelPrizessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationWheelPrizess []GameInvitationWheelPrizesObjDetail `json:"gameInvitationWheelPrizess"`
	} `json:"data"` // return data
}
