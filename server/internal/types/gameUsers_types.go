package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUsersRequest request params
type CreateGameUsersRequest struct {
	UserNo  string `json:"userNo" binding:""` // Public user id
	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	PasswordHash  string `json:"passwordHash" binding:""`
	Nickname  string `json:"nickname" binding:""`
	AvatarURL  string `json:"avatarURL" binding:""`
	InviteCode  string `json:"inviteCode" binding:""`
	ReferrerUserID  uint64 `json:"referrerUserID" binding:""`
	VipLevel  uint `json:"vipLevel" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 disabled, 3 risk locked
	RegisteredIP  string `json:"registeredIP" binding:""`
	LastLoginAt  *time.Time `json:"lastLoginAt" binding:""`
}

// UpdateGameUsersByIDRequest request params
type UpdateGameUsersByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserNo  string `json:"userNo" binding:""` // Public user id
	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	PasswordHash  string `json:"passwordHash" binding:""`
	Nickname  string `json:"nickname" binding:""`
	AvatarURL  string `json:"avatarURL" binding:""`
	InviteCode  string `json:"inviteCode" binding:""`
	ReferrerUserID  uint64 `json:"referrerUserID" binding:""`
	VipLevel  uint `json:"vipLevel" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 disabled, 3 risk locked
	RegisteredIP  string `json:"registeredIP" binding:""`
	LastLoginAt  *time.Time `json:"lastLoginAt" binding:""`
}

// GameUsersObjDetail detail
type GameUsersObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserNo  string `json:"userNo"` // Public user id
	PhoneCountryCode  string `json:"phoneCountryCode"`
	Phone  string `json:"phone"`
	PasswordHash  string `json:"passwordHash"`
	Nickname  string `json:"nickname"`
	AvatarURL  string `json:"avatarURL"`
	InviteCode  string `json:"inviteCode"`
	ReferrerUserID  uint64 `json:"referrerUserID"`
	VipLevel  uint `json:"vipLevel"`
	Status  int `json:"status"` // 1 active, 2 disabled, 3 risk locked
	RegisteredIP  string `json:"registeredIP"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUsersReply only for api docs
type CreateGameUsersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUsersByIDReply only for api docs
type DeleteGameUsersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUsersByIDReply only for api docs
type UpdateGameUsersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUsersByIDReply only for api docs
type GetGameUsersByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUsers GameUsersObjDetail `json:"gameUsers"`
	} `json:"data"` // return data
}

// ListGameUserssRequest request params
type ListGameUserssRequest struct {
	query.Params
}

// ListGameUserssReply only for api docs
type ListGameUserssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserss []GameUsersObjDetail `json:"gameUserss"`
	} `json:"data"` // return data
}
