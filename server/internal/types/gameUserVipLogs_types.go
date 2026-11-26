package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameUserVipLogsRequest request params
type CreateGameUserVipLogsRequest struct {
	UserID         uint64     `json:"userID" binding:""`
	OldLevelNo     uint       `json:"oldLevelNo" binding:""`
	NewLevelNo     uint       `json:"newLevelNo" binding:""`
	Reason         string     `json:"reason" binding:""`
	DepositAmount  string     `json:"depositAmount" binding:""`
	ValidBetAmount string     `json:"validBetAmount" binding:""`
	RewardAmount   string     `json:"rewardAmount" binding:""`
	ChangedAt      *time.Time `json:"changedAt" binding:""`
}

// UpdateGameUserVipLogsByIDRequest request params
type UpdateGameUserVipLogsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID         uint64     `json:"userID" binding:""`
	OldLevelNo     uint       `json:"oldLevelNo" binding:""`
	NewLevelNo     uint       `json:"newLevelNo" binding:""`
	Reason         string     `json:"reason" binding:""`
	DepositAmount  string     `json:"depositAmount" binding:""`
	ValidBetAmount string     `json:"validBetAmount" binding:""`
	RewardAmount   string     `json:"rewardAmount" binding:""`
	ChangedAt      *time.Time `json:"changedAt" binding:""`
}

// GameUserVipLogsObjDetail detail
type GameUserVipLogsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID         uint64     `json:"userID"`
	OldLevelNo     uint       `json:"oldLevelNo"`
	NewLevelNo     uint       `json:"newLevelNo"`
	Reason         string     `json:"reason"`
	DepositAmount  string     `json:"depositAmount"`
	ValidBetAmount string     `json:"validBetAmount"`
	RewardAmount   string     `json:"rewardAmount"`
	ChangedAt      *time.Time `json:"changedAt"`
	CreatedAt      *time.Time `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

// CreateGameUserVipLogsReply only for api docs
type CreateGameUserVipLogsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserVipLogsByIDReply only for api docs
type DeleteGameUserVipLogsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserVipLogsByIDReply only for api docs
type UpdateGameUserVipLogsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserVipLogsByIDReply only for api docs
type GetGameUserVipLogsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserVipLogs GameUserVipLogsObjDetail `json:"gameUserVipLogs"`
	} `json:"data"` // return data
}

// ListGameUserVipLogssRequest request params
type ListGameUserVipLogssRequest struct {
	query.Params
}

// ListGameUserVipLogssReply only for api docs
type ListGameUserVipLogssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserVipLogss []GameUserVipLogsObjDetail `json:"gameUserVipLogss"`
	} `json:"data"` // return data
}
