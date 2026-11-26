package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameUserActivityRecordsRequest request params
type CreateGameUserActivityRecordsRequest struct {
	UserID      uint64     `json:"userID" binding:""`
	ActivityID  uint64     `json:"activityID" binding:""`
	ActionType  string     `json:"actionType" binding:""`
	OrderNo     string     `json:"orderNo" binding:""`
	RewardCash  string     `json:"rewardCash" binding:""`
	RewardBonus string     `json:"rewardBonus" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 claimed, 3 rejected, 4 expired
	ClaimedAt   *time.Time `json:"claimedAt" binding:""`
	Remark      string     `json:"remark" binding:""`
}

// UpdateGameUserActivityRecordsByIDRequest request params
type UpdateGameUserActivityRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID      uint64     `json:"userID" binding:""`
	ActivityID  uint64     `json:"activityID" binding:""`
	ActionType  string     `json:"actionType" binding:""`
	OrderNo     string     `json:"orderNo" binding:""`
	RewardCash  string     `json:"rewardCash" binding:""`
	RewardBonus string     `json:"rewardBonus" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 claimed, 3 rejected, 4 expired
	ClaimedAt   *time.Time `json:"claimedAt" binding:""`
	Remark      string     `json:"remark" binding:""`
}

// GameUserActivityRecordsObjDetail detail
type GameUserActivityRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID      uint64     `json:"userID"`
	ActivityID  uint64     `json:"activityID"`
	ActionType  string     `json:"actionType"`
	OrderNo     string     `json:"orderNo"`
	RewardCash  string     `json:"rewardCash"`
	RewardBonus string     `json:"rewardBonus"`
	Status      int        `json:"status"` // 1 pending, 2 claimed, 3 rejected, 4 expired
	ClaimedAt   *time.Time `json:"claimedAt"`
	Remark      string     `json:"remark"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameUserActivityRecordsReply only for api docs
type CreateGameUserActivityRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserActivityRecordsByIDReply only for api docs
type DeleteGameUserActivityRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserActivityRecordsByIDReply only for api docs
type UpdateGameUserActivityRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserActivityRecordsByIDReply only for api docs
type GetGameUserActivityRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserActivityRecords GameUserActivityRecordsObjDetail `json:"gameUserActivityRecords"`
	} `json:"data"` // return data
}

// ListGameUserActivityRecordssRequest request params
type ListGameUserActivityRecordssRequest struct {
	query.Params
}

// ListGameUserActivityRecordssReply only for api docs
type ListGameUserActivityRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserActivityRecordss []GameUserActivityRecordsObjDetail `json:"gameUserActivityRecordss"`
	} `json:"data"` // return data
}
