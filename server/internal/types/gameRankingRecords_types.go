package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameRankingRecordsRequest request params
type CreateGameRankingRecordsRequest struct {
	RankingPeriodID  uint64 `json:"rankingPeriodID" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	RankNo  uint `json:"rankNo" binding:""`
	DisplayName  string `json:"displayName" binding:""`
	AvatarURL  string `json:"avatarURL" binding:""`
	TotalAmount  string `json:"totalAmount" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 rewarded, 3 invalid
}

// UpdateGameRankingRecordsByIDRequest request params
type UpdateGameRankingRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RankingPeriodID  uint64 `json:"rankingPeriodID" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	RankNo  uint `json:"rankNo" binding:""`
	DisplayName  string `json:"displayName" binding:""`
	AvatarURL  string `json:"avatarURL" binding:""`
	TotalAmount  string `json:"totalAmount" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 rewarded, 3 invalid
}

// GameRankingRecordsObjDetail detail
type GameRankingRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RankingPeriodID  uint64 `json:"rankingPeriodID"`
	UserID  uint64 `json:"userID"`
	RankNo  uint `json:"rankNo"`
	DisplayName  string `json:"displayName"`
	AvatarURL  string `json:"avatarURL"`
	TotalAmount  string `json:"totalAmount"`
	RewardAmount  string `json:"rewardAmount"`
	Status  int `json:"status"` // 1 pending, 2 rewarded, 3 invalid
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameRankingRecordsReply only for api docs
type CreateGameRankingRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameRankingRecordsByIDReply only for api docs
type DeleteGameRankingRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameRankingRecordsByIDReply only for api docs
type UpdateGameRankingRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameRankingRecordsByIDReply only for api docs
type GetGameRankingRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRankingRecords GameRankingRecordsObjDetail `json:"gameRankingRecords"`
	} `json:"data"` // return data
}

// ListGameRankingRecordssRequest request params
type ListGameRankingRecordssRequest struct {
	query.Params
}

// ListGameRankingRecordssReply only for api docs
type ListGameRankingRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRankingRecordss []GameRankingRecordsObjDetail `json:"gameRankingRecordss"`
	} `json:"data"` // return data
}
