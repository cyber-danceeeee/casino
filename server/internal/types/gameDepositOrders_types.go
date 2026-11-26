package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameDepositOrdersRequest request params
type CreateGameDepositOrdersRequest struct {
	OrderNo     string     `json:"orderNo" binding:""`
	UserID      uint64     `json:"userID" binding:""`
	Amount      string     `json:"amount" binding:""`
	PaidAmount  string     `json:"paidAmount" binding:""`
	Currency    string     `json:"currency" binding:""`
	MethodCode  string     `json:"methodCode" binding:""`
	ChannelCode string     `json:"channelCode" binding:""`
	VoucherID   uint64     `json:"voucherID" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 succeeded, 3 failed, 4 cancelled
	PaidAt      *time.Time `json:"paidAt" binding:""`
	ExpiredAt   *time.Time `json:"expiredAt" binding:""`
	FailReason  string     `json:"failReason" binding:""`
}

// UpdateGameDepositOrdersByIDRequest request params
type UpdateGameDepositOrdersByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	OrderNo     string     `json:"orderNo" binding:""`
	UserID      uint64     `json:"userID" binding:""`
	Amount      string     `json:"amount" binding:""`
	PaidAmount  string     `json:"paidAmount" binding:""`
	Currency    string     `json:"currency" binding:""`
	MethodCode  string     `json:"methodCode" binding:""`
	ChannelCode string     `json:"channelCode" binding:""`
	VoucherID   uint64     `json:"voucherID" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 succeeded, 3 failed, 4 cancelled
	PaidAt      *time.Time `json:"paidAt" binding:""`
	ExpiredAt   *time.Time `json:"expiredAt" binding:""`
	FailReason  string     `json:"failReason" binding:""`
}

// GameDepositOrdersObjDetail detail
type GameDepositOrdersObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	OrderNo     string     `json:"orderNo"`
	UserID      uint64     `json:"userID"`
	Amount      string     `json:"amount"`
	PaidAmount  string     `json:"paidAmount"`
	Currency    string     `json:"currency"`
	MethodCode  string     `json:"methodCode"`
	ChannelCode string     `json:"channelCode"`
	VoucherID   uint64     `json:"voucherID"`
	Status      int        `json:"status"` // 1 pending, 2 succeeded, 3 failed, 4 cancelled
	PaidAt      *time.Time `json:"paidAt"`
	ExpiredAt   *time.Time `json:"expiredAt"`
	FailReason  string     `json:"failReason"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameDepositOrdersReply only for api docs
type CreateGameDepositOrdersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameDepositOrdersByIDReply only for api docs
type DeleteGameDepositOrdersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameDepositOrdersByIDReply only for api docs
type UpdateGameDepositOrdersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameDepositOrdersByIDReply only for api docs
type GetGameDepositOrdersByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameDepositOrders GameDepositOrdersObjDetail `json:"gameDepositOrders"`
	} `json:"data"` // return data
}

// ListGameDepositOrderssRequest request params
type ListGameDepositOrderssRequest struct {
	query.Params
}

// ListGameDepositOrderssReply only for api docs
type ListGameDepositOrderssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameDepositOrderss []GameDepositOrdersObjDetail `json:"gameDepositOrderss"`
	} `json:"data"` // return data
}
