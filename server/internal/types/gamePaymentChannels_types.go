package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGamePaymentChannelsRequest request params
type CreateGamePaymentChannelsRequest struct {
	MethodID     uint64 `json:"methodID" binding:""`
	ChannelCode  string `json:"channelCode" binding:""`
	Name         string `json:"name" binding:""`
	CountryCode  string `json:"countryCode" binding:""`
	CurrencyCode string `json:"currencyCode" binding:""`
	FeeRate      string `json:"feeRate" binding:""`
	MinAmount    string `json:"minAmount" binding:""`
	MaxAmount    string `json:"maxAmount" binding:""`
	Sort         int    `json:"sort" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGamePaymentChannelsByIDRequest request params
type UpdateGamePaymentChannelsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	MethodID     uint64 `json:"methodID" binding:""`
	ChannelCode  string `json:"channelCode" binding:""`
	Name         string `json:"name" binding:""`
	CountryCode  string `json:"countryCode" binding:""`
	CurrencyCode string `json:"currencyCode" binding:""`
	FeeRate      string `json:"feeRate" binding:""`
	MinAmount    string `json:"minAmount" binding:""`
	MaxAmount    string `json:"maxAmount" binding:""`
	Sort         int    `json:"sort" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GamePaymentChannelsObjDetail detail
type GamePaymentChannelsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	MethodID     uint64     `json:"methodID"`
	ChannelCode  string     `json:"channelCode"`
	Name         string     `json:"name"`
	CountryCode  string     `json:"countryCode"`
	CurrencyCode string     `json:"currencyCode"`
	FeeRate      string     `json:"feeRate"`
	MinAmount    string     `json:"minAmount"`
	MaxAmount    string     `json:"maxAmount"`
	Sort         int        `json:"sort"`
	Status       int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// CreateGamePaymentChannelsReply only for api docs
type CreateGamePaymentChannelsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGamePaymentChannelsByIDReply only for api docs
type DeleteGamePaymentChannelsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGamePaymentChannelsByIDReply only for api docs
type UpdateGamePaymentChannelsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGamePaymentChannelsByIDReply only for api docs
type GetGamePaymentChannelsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GamePaymentChannels GamePaymentChannelsObjDetail `json:"gamePaymentChannels"`
	} `json:"data"` // return data
}

// ListGamePaymentChannelssRequest request params
type ListGamePaymentChannelssRequest struct {
	query.Params
}

// ListGamePaymentChannelssReply only for api docs
type ListGamePaymentChannelssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GamePaymentChannelss []GamePaymentChannelsObjDetail `json:"gamePaymentChannelss"`
	} `json:"data"` // return data
}
