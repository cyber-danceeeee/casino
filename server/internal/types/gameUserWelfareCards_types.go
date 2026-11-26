package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserWelfareCardsRequest request params
type CreateGameUserWelfareCardsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	WelfareCardID  uint64 `json:"welfareCardID" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 expired, 3 cancelled
	ActivatedAt  *time.Time `json:"activatedAt" binding:""`
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
}

// UpdateGameUserWelfareCardsByIDRequest request params
type UpdateGameUserWelfareCardsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	WelfareCardID  uint64 `json:"welfareCardID" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 expired, 3 cancelled
	ActivatedAt  *time.Time `json:"activatedAt" binding:""`
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
}

// GameUserWelfareCardsObjDetail detail
type GameUserWelfareCardsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	WelfareCardID  uint64 `json:"welfareCardID"`
	Status  int `json:"status"` // 1 active, 2 expired, 3 cancelled
	ActivatedAt  *time.Time `json:"activatedAt"`
	ExpiredAt  *time.Time `json:"expiredAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserWelfareCardsReply only for api docs
type CreateGameUserWelfareCardsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserWelfareCardsByIDReply only for api docs
type DeleteGameUserWelfareCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserWelfareCardsByIDReply only for api docs
type UpdateGameUserWelfareCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserWelfareCardsByIDReply only for api docs
type GetGameUserWelfareCardsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserWelfareCards GameUserWelfareCardsObjDetail `json:"gameUserWelfareCards"`
	} `json:"data"` // return data
}

// ListGameUserWelfareCardssRequest request params
type ListGameUserWelfareCardssRequest struct {
	query.Params
}

// ListGameUserWelfareCardssReply only for api docs
type ListGameUserWelfareCardssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserWelfareCardss []GameUserWelfareCardsObjDetail `json:"gameUserWelfareCardss"`
	} `json:"data"` // return data
}
