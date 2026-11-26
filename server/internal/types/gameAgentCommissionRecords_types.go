package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameAgentCommissionRecordsRequest request params
type CreateGameAgentCommissionRecordsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	ReferralsCount  uint `json:"referralsCount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
	CommissionAmount  string `json:"commissionAmount" binding:""`
	BonusAmount  string `json:"bonusAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 settled, 3 rejected
	SettledAt  *time.Time `json:"settledAt" binding:""`
}

// UpdateGameAgentCommissionRecordsByIDRequest request params
type UpdateGameAgentCommissionRecordsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	ReferralsCount  uint `json:"referralsCount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
	CommissionAmount  string `json:"commissionAmount" binding:""`
	BonusAmount  string `json:"bonusAmount" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 settled, 3 rejected
	SettledAt  *time.Time `json:"settledAt" binding:""`
}

// GameAgentCommissionRecordsObjDetail detail
type GameAgentCommissionRecordsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	PeriodStart  *time.Time `json:"periodStart"`
	PeriodEnd  *time.Time `json:"periodEnd"`
	ReferralsCount  uint `json:"referralsCount"`
	ValidBetAmount  string `json:"validBetAmount"`
	CommissionAmount  string `json:"commissionAmount"`
	BonusAmount  string `json:"bonusAmount"`
	Status  int `json:"status"` // 1 pending, 2 settled, 3 rejected
	SettledAt  *time.Time `json:"settledAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameAgentCommissionRecordsReply only for api docs
type CreateGameAgentCommissionRecordsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameAgentCommissionRecordsByIDReply only for api docs
type DeleteGameAgentCommissionRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameAgentCommissionRecordsByIDReply only for api docs
type UpdateGameAgentCommissionRecordsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameAgentCommissionRecordsByIDReply only for api docs
type GetGameAgentCommissionRecordsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAgentCommissionRecords GameAgentCommissionRecordsObjDetail `json:"gameAgentCommissionRecords"`
	} `json:"data"` // return data
}

// ListGameAgentCommissionRecordssRequest request params
type ListGameAgentCommissionRecordssRequest struct {
	query.Params
}

// ListGameAgentCommissionRecordssReply only for api docs
type ListGameAgentCommissionRecordssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAgentCommissionRecordss []GameAgentCommissionRecordsObjDetail `json:"gameAgentCommissionRecordss"`
	} `json:"data"` // return data
}
