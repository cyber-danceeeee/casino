package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameCashAuditRecordsRequest request params
type CreateGameCashAuditRecordsRequest struct {
	UserID                uint64     `json:"userID" binding:""`
	SourceType            string     `json:"sourceType" binding:""`
	SourceID              uint64     `json:"sourceID" binding:""`
	Amount                string     `json:"amount" binding:""`
	RequirementMultiplier string     `json:"requirementMultiplier" binding:""`
	RequiredValidBet      string     `json:"requiredValidBet" binding:""`
	CompletedValidBet     string     `json:"completedValidBet" binding:""`
	Status                int        `json:"status" binding:""` // 1 active, 2 completed, 3 cancelled
	StartedAt             *time.Time `json:"startedAt" binding:""`
	CompletedAt           *time.Time `json:"completedAt" binding:""`
}

// UpdateGameCashAuditRecordsByIDRequest request params
type UpdateGameCashAuditRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID                uint64     `json:"userID" binding:""`
	SourceType            string     `json:"sourceType" binding:""`
	SourceID              uint64     `json:"sourceID" binding:""`
	Amount                string     `json:"amount" binding:""`
	RequirementMultiplier string     `json:"requirementMultiplier" binding:""`
	RequiredValidBet      string     `json:"requiredValidBet" binding:""`
	CompletedValidBet     string     `json:"completedValidBet" binding:""`
	Status                int        `json:"status" binding:""` // 1 active, 2 completed, 3 cancelled
	StartedAt             *time.Time `json:"startedAt" binding:""`
	CompletedAt           *time.Time `json:"completedAt" binding:""`
}

// GameCashAuditRecordsObjDetail detail
type GameCashAuditRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID                uint64     `json:"userID"`
	SourceType            string     `json:"sourceType"`
	SourceID              uint64     `json:"sourceID"`
	Amount                string     `json:"amount"`
	RequirementMultiplier string     `json:"requirementMultiplier"`
	RequiredValidBet      string     `json:"requiredValidBet"`
	CompletedValidBet     string     `json:"completedValidBet"`
	Status                int        `json:"status"` // 1 active, 2 completed, 3 cancelled
	StartedAt             *time.Time `json:"startedAt"`
	CompletedAt           *time.Time `json:"completedAt"`
	CreatedAt             *time.Time `json:"createdAt"`
	UpdatedAt             *time.Time `json:"updatedAt"`
}

// CreateGameCashAuditRecordsReply only for api docs
type CreateGameCashAuditRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameCashAuditRecordsByIDReply only for api docs
type DeleteGameCashAuditRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameCashAuditRecordsByIDReply only for api docs
type UpdateGameCashAuditRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameCashAuditRecordsByIDReply only for api docs
type GetGameCashAuditRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameCashAuditRecords GameCashAuditRecordsObjDetail `json:"gameCashAuditRecords"`
	} `json:"data"` // return data
}

// ListGameCashAuditRecordssRequest request params
type ListGameCashAuditRecordssRequest struct {
	query.Params
}

// ListGameCashAuditRecordssReply only for api docs
type ListGameCashAuditRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameCashAuditRecordss []GameCashAuditRecordsObjDetail `json:"gameCashAuditRecordss"`
	} `json:"data"` // return data
}
