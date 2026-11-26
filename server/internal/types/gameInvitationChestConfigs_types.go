package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameInvitationChestConfigsRequest request params
type CreateGameInvitationChestConfigsRequest struct {
	ChestCode  string `json:"chestCode" binding:""`
	Title  string `json:"title" binding:""`
	RequiredInvites  uint `json:"requiredInvites" binding:""`
	RewardCash  string `json:"rewardCash" binding:""`
	RewardBonus  string `json:"rewardBonus" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameInvitationChestConfigsByIDRequest request params
type UpdateGameInvitationChestConfigsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ChestCode  string `json:"chestCode" binding:""`
	Title  string `json:"title" binding:""`
	RequiredInvites  uint `json:"requiredInvites" binding:""`
	RewardCash  string `json:"rewardCash" binding:""`
	RewardBonus  string `json:"rewardBonus" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameInvitationChestConfigsObjDetail detail
type GameInvitationChestConfigsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ChestCode  string `json:"chestCode"`
	Title  string `json:"title"`
	RequiredInvites  uint `json:"requiredInvites"`
	RewardCash  string `json:"rewardCash"`
	RewardBonus  string `json:"rewardBonus"`
	IconURL  string `json:"iconURL"`
	Sort  int `json:"sort"`
	Status  int `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameInvitationChestConfigsReply only for api docs
type CreateGameInvitationChestConfigsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameInvitationChestConfigsByIDReply only for api docs
type DeleteGameInvitationChestConfigsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameInvitationChestConfigsByIDReply only for api docs
type UpdateGameInvitationChestConfigsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameInvitationChestConfigsByIDReply only for api docs
type GetGameInvitationChestConfigsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationChestConfigs GameInvitationChestConfigsObjDetail `json:"gameInvitationChestConfigs"`
	} `json:"data"` // return data
}

// ListGameInvitationChestConfigssRequest request params
type ListGameInvitationChestConfigssRequest struct {
	query.Params
}

// ListGameInvitationChestConfigssReply only for api docs
type ListGameInvitationChestConfigssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameInvitationChestConfigss []GameInvitationChestConfigsObjDetail `json:"gameInvitationChestConfigss"`
	} `json:"data"` // return data
}
