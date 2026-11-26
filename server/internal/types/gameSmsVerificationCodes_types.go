package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameSmsVerificationCodesRequest request params
type CreateGameSmsVerificationCodesRequest struct {
	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	CodeHash  string `json:"codeHash" binding:""`
	Scene  string `json:"scene" binding:""` // login, register, bind_phone, reset_password
	Status  int `json:"status" binding:""` // 1 unused, 2 used, 3 expired
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
	UsedAt  *time.Time `json:"usedAt" binding:""`
}

// UpdateGameSmsVerificationCodesByIDRequest request params
type UpdateGameSmsVerificationCodesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	CodeHash  string `json:"codeHash" binding:""`
	Scene  string `json:"scene" binding:""` // login, register, bind_phone, reset_password
	Status  int `json:"status" binding:""` // 1 unused, 2 used, 3 expired
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
	UsedAt  *time.Time `json:"usedAt" binding:""`
}

// GameSmsVerificationCodesObjDetail detail
type GameSmsVerificationCodesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	PhoneCountryCode  string `json:"phoneCountryCode"`
	Phone  string `json:"phone"`
	CodeHash  string `json:"codeHash"`
	Scene  string `json:"scene"` // login, register, bind_phone, reset_password
	Status  int `json:"status"` // 1 unused, 2 used, 3 expired
	ExpiredAt  *time.Time `json:"expiredAt"`
	UsedAt  *time.Time `json:"usedAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameSmsVerificationCodesReply only for api docs
type CreateGameSmsVerificationCodesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSmsVerificationCodesByIDReply only for api docs
type DeleteGameSmsVerificationCodesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSmsVerificationCodesByIDReply only for api docs
type UpdateGameSmsVerificationCodesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSmsVerificationCodesByIDReply only for api docs
type GetGameSmsVerificationCodesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSmsVerificationCodes GameSmsVerificationCodesObjDetail `json:"gameSmsVerificationCodes"`
	} `json:"data"` // return data
}

// ListGameSmsVerificationCodessRequest request params
type ListGameSmsVerificationCodessRequest struct {
	query.Params
}

// ListGameSmsVerificationCodessReply only for api docs
type ListGameSmsVerificationCodessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSmsVerificationCodess []GameSmsVerificationCodesObjDetail `json:"gameSmsVerificationCodess"`
	} `json:"data"` // return data
}
