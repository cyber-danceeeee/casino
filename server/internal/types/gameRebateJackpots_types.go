package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameRebateJackpotsRequest request params
type CreateGameRebateJackpotsRequest struct {
	PeriodType  string `json:"periodType" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	JackpotAmount  string `json:"jackpotAmount" binding:""`
	DigitText  string `json:"digitText" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 settled, 3 disabled
}

// UpdateGameRebateJackpotsByIDRequest request params
type UpdateGameRebateJackpotsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	PeriodType  string `json:"periodType" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	JackpotAmount  string `json:"jackpotAmount" binding:""`
	DigitText  string `json:"digitText" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 settled, 3 disabled
}

// GameRebateJackpotsObjDetail detail
type GameRebateJackpotsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	PeriodType  string `json:"periodType"`
	PeriodStart  *time.Time `json:"periodStart"`
	PeriodEnd  *time.Time `json:"periodEnd"`
	JackpotAmount  string `json:"jackpotAmount"`
	DigitText  string `json:"digitText"`
	Status  int `json:"status"` // 1 active, 2 settled, 3 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameRebateJackpotsReply only for api docs
type CreateGameRebateJackpotsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameRebateJackpotsByIDReply only for api docs
type DeleteGameRebateJackpotsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameRebateJackpotsByIDReply only for api docs
type UpdateGameRebateJackpotsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameRebateJackpotsByIDReply only for api docs
type GetGameRebateJackpotsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRebateJackpots GameRebateJackpotsObjDetail `json:"gameRebateJackpots"`
	} `json:"data"` // return data
}

// ListGameRebateJackpotssRequest request params
type ListGameRebateJackpotssRequest struct {
	query.Params
}

// ListGameRebateJackpotssReply only for api docs
type ListGameRebateJackpotssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRebateJackpotss []GameRebateJackpotsObjDetail `json:"gameRebateJackpotss"`
	} `json:"data"` // return data
}
