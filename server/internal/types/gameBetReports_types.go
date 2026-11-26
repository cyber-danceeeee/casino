package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameBetReportsRequest request params
type CreateGameBetReportsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	ReportDate  *time.Time `json:"reportDate" binding:""`
	GameType  string `json:"gameType" binding:""`
	BetCount  uint `json:"betCount" binding:""`
	BetAmount  string `json:"betAmount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
	PayoutAmount  string `json:"payoutAmount" binding:""`
	ProfitAmount  string `json:"profitAmount" binding:""`
}

// UpdateGameBetReportsByIDRequest request params
type UpdateGameBetReportsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	ReportDate  *time.Time `json:"reportDate" binding:""`
	GameType  string `json:"gameType" binding:""`
	BetCount  uint `json:"betCount" binding:""`
	BetAmount  string `json:"betAmount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
	PayoutAmount  string `json:"payoutAmount" binding:""`
	ProfitAmount  string `json:"profitAmount" binding:""`
}

// GameBetReportsObjDetail detail
type GameBetReportsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	ReportDate  *time.Time `json:"reportDate"`
	GameType  string `json:"gameType"`
	BetCount  uint `json:"betCount"`
	BetAmount  string `json:"betAmount"`
	ValidBetAmount  string `json:"validBetAmount"`
	PayoutAmount  string `json:"payoutAmount"`
	ProfitAmount  string `json:"profitAmount"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameBetReportsReply only for api docs
type CreateGameBetReportsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameBetReportsByIDReply only for api docs
type DeleteGameBetReportsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameBetReportsByIDReply only for api docs
type UpdateGameBetReportsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameBetReportsByIDReply only for api docs
type GetGameBetReportsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBetReports GameBetReportsObjDetail `json:"gameBetReports"`
	} `json:"data"` // return data
}

// ListGameBetReportssRequest request params
type ListGameBetReportssRequest struct {
	query.Params
}

// ListGameBetReportssReply only for api docs
type ListGameBetReportssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBetReportss []GameBetReportsObjDetail `json:"gameBetReportss"`
	} `json:"data"` // return data
}
