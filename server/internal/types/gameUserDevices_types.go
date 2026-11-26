package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserDevicesRequest request params
type CreateGameUserDevicesRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	Platform  string `json:"platform" binding:""`
	AppVersion  string `json:"appVersion" binding:""`
	ClientIP  string `json:"clientIP" binding:""`
	UserAgent  string `json:"userAgent" binding:""`
	FirstSeenAt  *time.Time `json:"firstSeenAt" binding:""`
	LastSeenAt  *time.Time `json:"lastSeenAt" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 blocked
}

// UpdateGameUserDevicesByIDRequest request params
type UpdateGameUserDevicesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	DeviceID  string `json:"deviceID" binding:""`
	Platform  string `json:"platform" binding:""`
	AppVersion  string `json:"appVersion" binding:""`
	ClientIP  string `json:"clientIP" binding:""`
	UserAgent  string `json:"userAgent" binding:""`
	FirstSeenAt  *time.Time `json:"firstSeenAt" binding:""`
	LastSeenAt  *time.Time `json:"lastSeenAt" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 blocked
}

// GameUserDevicesObjDetail detail
type GameUserDevicesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	DeviceID  string `json:"deviceID"`
	Platform  string `json:"platform"`
	AppVersion  string `json:"appVersion"`
	ClientIP  string `json:"clientIP"`
	UserAgent  string `json:"userAgent"`
	FirstSeenAt  *time.Time `json:"firstSeenAt"`
	LastSeenAt  *time.Time `json:"lastSeenAt"`
	Status  int `json:"status"` // 1 active, 2 blocked
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserDevicesReply only for api docs
type CreateGameUserDevicesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserDevicesByIDReply only for api docs
type DeleteGameUserDevicesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserDevicesByIDReply only for api docs
type UpdateGameUserDevicesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserDevicesByIDReply only for api docs
type GetGameUserDevicesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserDevices GameUserDevicesObjDetail `json:"gameUserDevices"`
	} `json:"data"` // return data
}

// ListGameUserDevicessRequest request params
type ListGameUserDevicessRequest struct {
	query.Params
}

// ListGameUserDevicessReply only for api docs
type ListGameUserDevicessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserDevicess []GameUserDevicesObjDetail `json:"gameUserDevicess"`
	} `json:"data"` // return data
}
