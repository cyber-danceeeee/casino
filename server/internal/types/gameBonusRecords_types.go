package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameBonusRecordsRequest request params
type CreateGameBonusRecordsRequest struct {
	OrderNo               string     `json:"orderNo" binding:""`
	UserID                uint64     `json:"userID" binding:""`
	SourceType            string     `json:"sourceType" binding:""`
	SourceID              uint64     `json:"sourceID" binding:""`
	Amount                string     `json:"amount" binding:""`
	RequirementMultiplier string     `json:"requirementMultiplier" binding:""`
	RequiredValidBet      string     `json:"requiredValidBet" binding:""`
	CompletedValidBet     string     `json:"completedValidBet" binding:""`
	Status                int        `json:"status" binding:""` // 1 locked, 2 unlocked, 3 expired, 4 cancelled
	IssuedAt              *time.Time `json:"issuedAt" binding:""`
	UnlockedAt            *time.Time `json:"unlockedAt" binding:""`
}

// UpdateGameBonusRecordsByIDRequest request params
type UpdateGameBonusRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	OrderNo               string     `json:"orderNo" binding:""`
	UserID                uint64     `json:"userID" binding:""`
	SourceType            string     `json:"sourceType" binding:""`
	SourceID              uint64     `json:"sourceID" binding:""`
	Amount                string     `json:"amount" binding:""`
	RequirementMultiplier string     `json:"requirementMultiplier" binding:""`
	RequiredValidBet      string     `json:"requiredValidBet" binding:""`
	CompletedValidBet     string     `json:"completedValidBet" binding:""`
	Status                int        `json:"status" binding:""` // 1 locked, 2 unlocked, 3 expired, 4 cancelled
	IssuedAt              *time.Time `json:"issuedAt" binding:""`
	UnlockedAt            *time.Time `json:"unlockedAt" binding:""`
}

// GameBonusRecordsObjDetail detail
type GameBonusRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	OrderNo               string     `json:"orderNo"`
	UserID                uint64     `json:"userID"`
	SourceType            string     `json:"sourceType"`
	SourceID              uint64     `json:"sourceID"`
	Amount                string     `json:"amount"`
	RequirementMultiplier string     `json:"requirementMultiplier"`
	RequiredValidBet      string     `json:"requiredValidBet"`
	CompletedValidBet     string     `json:"completedValidBet"`
	Status                int        `json:"status"` // 1 locked, 2 unlocked, 3 expired, 4 cancelled
	IssuedAt              *time.Time `json:"issuedAt"`
	UnlockedAt            *time.Time `json:"unlockedAt"`
	CreatedAt             *time.Time `json:"createdAt"`
	UpdatedAt             *time.Time `json:"updatedAt"`
}

// CreateGameBonusRecordsReply only for api docs
type CreateGameBonusRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameBonusRecordsByIDReply only for api docs
type DeleteGameBonusRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameBonusRecordsByIDReply only for api docs
type UpdateGameBonusRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameBonusRecordsByIDReply only for api docs
type GetGameBonusRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBonusRecords GameBonusRecordsObjDetail `json:"gameBonusRecords"`
	} `json:"data"` // return data
}

// ListGameBonusRecordssRequest request params
type ListGameBonusRecordssRequest struct {
	query.Params
}

// ListGameBonusRecordssReply only for api docs
type ListGameBonusRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBonusRecordss []GameBonusRecordsObjDetail `json:"gameBonusRecordss"`
	} `json:"data"` // return data
}
