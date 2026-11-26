package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameBetRecordsRequest request params
type CreateGameBetRecordsRequest struct {
	BetNo          string     `json:"betNo" binding:""`
	UserID         uint64     `json:"userID" binding:""`
	GameID         uint64     `json:"gameID" binding:""`
	ProviderID     uint64     `json:"providerID" binding:""`
	GameType       string     `json:"gameType" binding:""`
	BetAmount      string     `json:"betAmount" binding:""`
	ValidBetAmount string     `json:"validBetAmount" binding:""`
	PayoutAmount   string     `json:"payoutAmount" binding:""`
	ProfitAmount   string     `json:"profitAmount" binding:""`
	Status         int        `json:"status" binding:""` // 1 not settled, 2 settled, 3 cancelled
	BetAt          *time.Time `json:"betAt" binding:""`
	SettledAt      *time.Time `json:"settledAt" binding:""`
}

// UpdateGameBetRecordsByIDRequest request params
type UpdateGameBetRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	BetNo          string     `json:"betNo" binding:""`
	UserID         uint64     `json:"userID" binding:""`
	GameID         uint64     `json:"gameID" binding:""`
	ProviderID     uint64     `json:"providerID" binding:""`
	GameType       string     `json:"gameType" binding:""`
	BetAmount      string     `json:"betAmount" binding:""`
	ValidBetAmount string     `json:"validBetAmount" binding:""`
	PayoutAmount   string     `json:"payoutAmount" binding:""`
	ProfitAmount   string     `json:"profitAmount" binding:""`
	Status         int        `json:"status" binding:""` // 1 not settled, 2 settled, 3 cancelled
	BetAt          *time.Time `json:"betAt" binding:""`
	SettledAt      *time.Time `json:"settledAt" binding:""`
}

// GameBetRecordsObjDetail detail
type GameBetRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	BetNo          string     `json:"betNo"`
	UserID         uint64     `json:"userID"`
	GameID         uint64     `json:"gameID"`
	ProviderID     uint64     `json:"providerID"`
	GameType       string     `json:"gameType"`
	BetAmount      string     `json:"betAmount"`
	ValidBetAmount string     `json:"validBetAmount"`
	PayoutAmount   string     `json:"payoutAmount"`
	ProfitAmount   string     `json:"profitAmount"`
	Status         int        `json:"status"` // 1 not settled, 2 settled, 3 cancelled
	BetAt          *time.Time `json:"betAt"`
	SettledAt      *time.Time `json:"settledAt"`
	CreatedAt      *time.Time `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

// CreateGameBetRecordsReply only for api docs
type CreateGameBetRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameBetRecordsByIDReply only for api docs
type DeleteGameBetRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameBetRecordsByIDReply only for api docs
type UpdateGameBetRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameBetRecordsByIDReply only for api docs
type GetGameBetRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBetRecords GameBetRecordsObjDetail `json:"gameBetRecords"`
	} `json:"data"` // return data
}

// ListGameBetRecordssRequest request params
type ListGameBetRecordssRequest struct {
	query.Params
}

// ListGameBetRecordssReply only for api docs
type ListGameBetRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBetRecordss []GameBetRecordsObjDetail `json:"gameBetRecordss"`
	} `json:"data"` // return data
}
