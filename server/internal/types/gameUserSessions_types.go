package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserSessionsRequest request params
type CreateGameUserSessionsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	TokenHash  string `json:"tokenHash" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	ClientIP  string `json:"clientIP" binding:""`
	UserAgent  string `json:"userAgent" binding:""`
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
	RevokedAt  *time.Time `json:"revokedAt" binding:""`
}

// UpdateGameUserSessionsByIDRequest request params
type UpdateGameUserSessionsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	TokenHash  string `json:"tokenHash" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	ClientIP  string `json:"clientIP" binding:""`
	UserAgent  string `json:"userAgent" binding:""`
	ExpiredAt  *time.Time `json:"expiredAt" binding:""`
	RevokedAt  *time.Time `json:"revokedAt" binding:""`
}

// GameUserSessionsObjDetail detail
type GameUserSessionsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	TokenHash  string `json:"tokenHash"`
	DeviceID  string `json:"deviceID"`
	ClientIP  string `json:"clientIP"`
	UserAgent  string `json:"userAgent"`
	ExpiredAt  *time.Time `json:"expiredAt"`
	RevokedAt  *time.Time `json:"revokedAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserSessionsReply only for api docs
type CreateGameUserSessionsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserSessionsByIDReply only for api docs
type DeleteGameUserSessionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserSessionsByIDReply only for api docs
type UpdateGameUserSessionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserSessionsByIDReply only for api docs
type GetGameUserSessionsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserSessions GameUserSessionsObjDetail `json:"gameUserSessions"`
	} `json:"data"` // return data
}

// ListGameUserSessionssRequest request params
type ListGameUserSessionssRequest struct {
	query.Params
}

// ListGameUserSessionssReply only for api docs
type ListGameUserSessionssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserSessionss []GameUserSessionsObjDetail `json:"gameUserSessionss"`
	} `json:"data"` // return data
}
