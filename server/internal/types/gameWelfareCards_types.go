package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameWelfareCardsRequest request params
type CreateGameWelfareCardsRequest struct {
	CardCode     string `json:"cardCode" binding:""`
	Title        string `json:"title" binding:""`
	PeriodType   string `json:"periodType" binding:""`
	PriceAmount  string `json:"priceAmount" binding:""`
	RewardAmount string `json:"rewardAmount" binding:""`
	RewardWallet string `json:"rewardWallet" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameWelfareCardsByIDRequest request params
type UpdateGameWelfareCardsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	CardCode     string `json:"cardCode" binding:""`
	Title        string `json:"title" binding:""`
	PeriodType   string `json:"periodType" binding:""`
	PriceAmount  string `json:"priceAmount" binding:""`
	RewardAmount string `json:"rewardAmount" binding:""`
	RewardWallet string `json:"rewardWallet" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameWelfareCardsObjDetail detail
type GameWelfareCardsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	CardCode     string     `json:"cardCode"`
	Title        string     `json:"title"`
	PeriodType   string     `json:"periodType"`
	PriceAmount  string     `json:"priceAmount"`
	RewardAmount string     `json:"rewardAmount"`
	RewardWallet string     `json:"rewardWallet"`
	Status       int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// CreateGameWelfareCardsReply only for api docs
type CreateGameWelfareCardsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameWelfareCardsByIDReply only for api docs
type DeleteGameWelfareCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameWelfareCardsByIDReply only for api docs
type UpdateGameWelfareCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameWelfareCardsByIDReply only for api docs
type GetGameWelfareCardsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWelfareCards GameWelfareCardsObjDetail `json:"gameWelfareCards"`
	} `json:"data"` // return data
}

// ListGameWelfareCardssRequest request params
type ListGameWelfareCardssRequest struct {
	query.Params
}

// ListGameWelfareCardssReply only for api docs
type ListGameWelfareCardssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWelfareCardss []GameWelfareCardsObjDetail `json:"gameWelfareCardss"`
	} `json:"data"` // return data
}
