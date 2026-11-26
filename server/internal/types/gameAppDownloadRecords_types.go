package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameAppDownloadRecordsRequest request params
type CreateGameAppDownloadRecordsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	AppVersionID  uint64 `json:"appVersionID" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	RewardWallet  string `json:"rewardWallet" binding:""`
	Status  int `json:"status" binding:""` // 1 downloaded, 2 rewarded, 3 rejected
	RewardedAt  *time.Time `json:"rewardedAt" binding:""`
}

// UpdateGameAppDownloadRecordsByIDRequest request params
type UpdateGameAppDownloadRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	AppVersionID  uint64 `json:"appVersionID" binding:""`
	RewardAmount  string `json:"rewardAmount" binding:""`
	RewardWallet  string `json:"rewardWallet" binding:""`
	Status  int `json:"status" binding:""` // 1 downloaded, 2 rewarded, 3 rejected
	RewardedAt  *time.Time `json:"rewardedAt" binding:""`
}

// GameAppDownloadRecordsObjDetail detail
type GameAppDownloadRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	DeviceID  string `json:"deviceID"`
	AppVersionID  uint64 `json:"appVersionID"`
	RewardAmount  string `json:"rewardAmount"`
	RewardWallet  string `json:"rewardWallet"`
	Status  int `json:"status"` // 1 downloaded, 2 rewarded, 3 rejected
	RewardedAt  *time.Time `json:"rewardedAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameAppDownloadRecordsReply only for api docs
type CreateGameAppDownloadRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameAppDownloadRecordsByIDReply only for api docs
type DeleteGameAppDownloadRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameAppDownloadRecordsByIDReply only for api docs
type UpdateGameAppDownloadRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameAppDownloadRecordsByIDReply only for api docs
type GetGameAppDownloadRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAppDownloadRecords GameAppDownloadRecordsObjDetail `json:"gameAppDownloadRecords"`
	} `json:"data"` // return data
}

// ListGameAppDownloadRecordssRequest request params
type ListGameAppDownloadRecordssRequest struct {
	query.Params
}

// ListGameAppDownloadRecordssReply only for api docs
type ListGameAppDownloadRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAppDownloadRecordss []GameAppDownloadRecordsObjDetail `json:"gameAppDownloadRecordss"`
	} `json:"data"` // return data
}
