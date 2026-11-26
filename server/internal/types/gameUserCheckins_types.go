package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserCheckinsRequest request params
type CreateGameUserCheckinsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	ActivityID  uint64 `json:"activityID" binding:""`
	CheckinDate  *time.Time `json:"checkinDate" binding:""`
	DayNo  uint `json:"dayNo" binding:""`
	StreakCount  uint `json:"streakCount" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	RewardWallet  string `json:"rewardWallet" binding:""`
	Status  int `json:"status" binding:""` // 1 checked, 2 claimed, 3 expired
	ClaimedAt  *time.Time `json:"claimedAt" binding:""`
}

// UpdateGameUserCheckinsByIDRequest request params
type UpdateGameUserCheckinsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	ActivityID  uint64 `json:"activityID" binding:""`
	CheckinDate  *time.Time `json:"checkinDate" binding:""`
	DayNo  uint `json:"dayNo" binding:""`
	StreakCount  uint `json:"streakCount" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	RewardWallet  string `json:"rewardWallet" binding:""`
	Status  int `json:"status" binding:""` // 1 checked, 2 claimed, 3 expired
	ClaimedAt  *time.Time `json:"claimedAt" binding:""`
}

// GameUserCheckinsObjDetail detail
type GameUserCheckinsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	ActivityID  uint64 `json:"activityID"`
	CheckinDate  *time.Time `json:"checkinDate"`
	DayNo  uint `json:"dayNo"`
	StreakCount  uint `json:"streakCount"`
	RewardAmount  string `json:"rewardAmount"`
	RewardWallet  string `json:"rewardWallet"`
	Status  int `json:"status"` // 1 checked, 2 claimed, 3 expired
	ClaimedAt  *time.Time `json:"claimedAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserCheckinsReply only for api docs
type CreateGameUserCheckinsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserCheckinsByIDReply only for api docs
type DeleteGameUserCheckinsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserCheckinsByIDReply only for api docs
type UpdateGameUserCheckinsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserCheckinsByIDReply only for api docs
type GetGameUserCheckinsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserCheckins GameUserCheckinsObjDetail `json:"gameUserCheckins"`
	} `json:"data"` // return data
}

// ListGameUserCheckinssRequest request params
type ListGameUserCheckinssRequest struct {
	query.Params
}

// ListGameUserCheckinssReply only for api docs
type ListGameUserCheckinssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserCheckinss []GameUserCheckinsObjDetail `json:"gameUserCheckinss"`
	} `json:"data"` // return data
}
