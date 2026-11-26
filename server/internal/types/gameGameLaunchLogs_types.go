package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameGameLaunchLogsRequest request params
type CreateGameGameLaunchLogsRequest struct {
	UserID       uint64 `json:"userID" binding:""`
	GameID       uint64 `json:"gameID" binding:""`
	ProviderID   uint64 `json:"providerID" binding:""`
	SessionToken string `json:"sessionToken" binding:""`
	LaunchStatus int    `json:"launchStatus" binding:""` // 1 loading, 2 entered, 3 failed
	ClientIP     string `json:"clientIP" binding:""`
	ErrorMessage string `json:"errorMessage" binding:""`
}

// UpdateGameGameLaunchLogsByIDRequest request params
type UpdateGameGameLaunchLogsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID       uint64 `json:"userID" binding:""`
	GameID       uint64 `json:"gameID" binding:""`
	ProviderID   uint64 `json:"providerID" binding:""`
	SessionToken string `json:"sessionToken" binding:""`
	LaunchStatus int    `json:"launchStatus" binding:""` // 1 loading, 2 entered, 3 failed
	ClientIP     string `json:"clientIP" binding:""`
	ErrorMessage string `json:"errorMessage" binding:""`
}

// GameGameLaunchLogsObjDetail detail
type GameGameLaunchLogsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID       uint64     `json:"userID"`
	GameID       uint64     `json:"gameID"`
	ProviderID   uint64     `json:"providerID"`
	SessionToken string     `json:"sessionToken"`
	LaunchStatus int        `json:"launchStatus"` // 1 loading, 2 entered, 3 failed
	ClientIP     string     `json:"clientIP"`
	ErrorMessage string     `json:"errorMessage"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// CreateGameGameLaunchLogsReply only for api docs
type CreateGameGameLaunchLogsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameGameLaunchLogsByIDReply only for api docs
type DeleteGameGameLaunchLogsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameGameLaunchLogsByIDReply only for api docs
type UpdateGameGameLaunchLogsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameGameLaunchLogsByIDReply only for api docs
type GetGameGameLaunchLogsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameLaunchLogs GameGameLaunchLogsObjDetail `json:"gameGameLaunchLogs"`
	} `json:"data"` // return data
}

// ListGameGameLaunchLogssRequest request params
type ListGameGameLaunchLogssRequest struct {
	query.Params
}

// ListGameGameLaunchLogssReply only for api docs
type ListGameGameLaunchLogssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameLaunchLogss []GameGameLaunchLogsObjDetail `json:"gameGameLaunchLogss"`
	} `json:"data"` // return data
}
