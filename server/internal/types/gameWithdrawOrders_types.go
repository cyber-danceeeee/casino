package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameWithdrawOrdersRequest request params
type CreateGameWithdrawOrdersRequest struct {
	OrderNo  string `json:"orderNo" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	WithdrawAccountID  uint64 `json:"withdrawAccountID" binding:""`
	Amount  string `json:"amount" binding:""`
	FeeAmount  string `json:"feeAmount" binding:""`
	ActualAmount  string `json:"actualAmount" binding:""`
	Currency  string `json:"currency" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 succeeded, 3 failed, 4 returned, 5 cancelled, 6 reviewing
	AuditStatus  int `json:"auditStatus" binding:""` // 1 pending, 2 passed, 3 rejected
	ReviewRemark  string `json:"reviewRemark" binding:""`
	SubmittedAt  *time.Time `json:"submittedAt" binding:""`
	CompletedAt  *time.Time `json:"completedAt" binding:""`
}

// UpdateGameWithdrawOrdersByIDRequest request params
type UpdateGameWithdrawOrdersByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	OrderNo  string `json:"orderNo" binding:""`
	UserID  uint64 `json:"userID" binding:""`
	WithdrawAccountID  uint64 `json:"withdrawAccountID" binding:""`
	Amount  string `json:"amount" binding:""`
	FeeAmount  string `json:"feeAmount" binding:""`
	ActualAmount  string `json:"actualAmount" binding:""`
	Currency  string `json:"currency" binding:""`
	Status  int `json:"status" binding:""` // 1 pending, 2 succeeded, 3 failed, 4 returned, 5 cancelled, 6 reviewing
	AuditStatus  int `json:"auditStatus" binding:""` // 1 pending, 2 passed, 3 rejected
	ReviewRemark  string `json:"reviewRemark" binding:""`
	SubmittedAt  *time.Time `json:"submittedAt" binding:""`
	CompletedAt  *time.Time `json:"completedAt" binding:""`
}

// GameWithdrawOrdersObjDetail detail
type GameWithdrawOrdersObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	OrderNo  string `json:"orderNo"`
	UserID  uint64 `json:"userID"`
	WithdrawAccountID  uint64 `json:"withdrawAccountID"`
	Amount  string `json:"amount"`
	FeeAmount  string `json:"feeAmount"`
	ActualAmount  string `json:"actualAmount"`
	Currency  string `json:"currency"`
	Status  int `json:"status"` // 1 pending, 2 succeeded, 3 failed, 4 returned, 5 cancelled, 6 reviewing
	AuditStatus  int `json:"auditStatus"` // 1 pending, 2 passed, 3 rejected
	ReviewRemark  string `json:"reviewRemark"`
	SubmittedAt  *time.Time `json:"submittedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameWithdrawOrdersReply only for api docs
type CreateGameWithdrawOrdersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameWithdrawOrdersByIDReply only for api docs
type DeleteGameWithdrawOrdersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameWithdrawOrdersByIDReply only for api docs
type UpdateGameWithdrawOrdersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameWithdrawOrdersByIDReply only for api docs
type GetGameWithdrawOrdersByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWithdrawOrders GameWithdrawOrdersObjDetail `json:"gameWithdrawOrders"`
	} `json:"data"` // return data
}

// ListGameWithdrawOrderssRequest request params
type ListGameWithdrawOrderssRequest struct {
	query.Params
}

// ListGameWithdrawOrderssReply only for api docs
type ListGameWithdrawOrderssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWithdrawOrderss []GameWithdrawOrdersObjDetail `json:"gameWithdrawOrderss"`
	} `json:"data"` // return data
}
