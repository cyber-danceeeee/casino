package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGamePaymentMethodsRequest request params
type CreateGamePaymentMethodsRequest struct {
	MethodCode string `json:"methodCode" binding:""`
	Name       string `json:"name" binding:""`
	MethodType string `json:"methodType" binding:""` // deposit, withdraw, both
	IconURL    string `json:"iconURL" binding:""`
	MinAmount  string `json:"minAmount" binding:""`
	MaxAmount  string `json:"maxAmount" binding:""`
	Sort       int    `json:"sort" binding:""`
	Status     int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGamePaymentMethodsByIDRequest request params
type UpdateGamePaymentMethodsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	MethodCode string `json:"methodCode" binding:""`
	Name       string `json:"name" binding:""`
	MethodType string `json:"methodType" binding:""` // deposit, withdraw, both
	IconURL    string `json:"iconURL" binding:""`
	MinAmount  string `json:"minAmount" binding:""`
	MaxAmount  string `json:"maxAmount" binding:""`
	Sort       int    `json:"sort" binding:""`
	Status     int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GamePaymentMethodsObjDetail detail
type GamePaymentMethodsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	MethodCode string     `json:"methodCode"`
	Name       string     `json:"name"`
	MethodType string     `json:"methodType"` // deposit, withdraw, both
	IconURL    string     `json:"iconURL"`
	MinAmount  string     `json:"minAmount"`
	MaxAmount  string     `json:"maxAmount"`
	Sort       int        `json:"sort"`
	Status     int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}

// CreateGamePaymentMethodsReply only for api docs
type CreateGamePaymentMethodsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGamePaymentMethodsByIDReply only for api docs
type DeleteGamePaymentMethodsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGamePaymentMethodsByIDReply only for api docs
type UpdateGamePaymentMethodsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGamePaymentMethodsByIDReply only for api docs
type GetGamePaymentMethodsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GamePaymentMethods GamePaymentMethodsObjDetail `json:"gamePaymentMethods"`
	} `json:"data"` // return data
}

// ListGamePaymentMethodssRequest request params
type ListGamePaymentMethodssRequest struct {
	query.Params
}

// ListGamePaymentMethodssReply only for api docs
type ListGamePaymentMethodssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GamePaymentMethodss []GamePaymentMethodsObjDetail `json:"gamePaymentMethodss"`
	} `json:"data"` // return data
}
