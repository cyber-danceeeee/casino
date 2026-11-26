package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameAppVersionsRequest request params
type CreateGameAppVersionsRequest struct {
	Platform    string `json:"platform" binding:""`
	VersionName string `json:"versionName" binding:""`
	VersionCode uint   `json:"versionCode" binding:""`
	DownloadURL string `json:"downloadURL" binding:""`
	ForceUpdate int    `json:"forceUpdate" binding:""`
	ReleaseNote string `json:"releaseNote" binding:""`
	Status      int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameAppVersionsByIDRequest request params
type UpdateGameAppVersionsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	Platform    string `json:"platform" binding:""`
	VersionName string `json:"versionName" binding:""`
	VersionCode uint   `json:"versionCode" binding:""`
	DownloadURL string `json:"downloadURL" binding:""`
	ForceUpdate int    `json:"forceUpdate" binding:""`
	ReleaseNote string `json:"releaseNote" binding:""`
	Status      int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameAppVersionsObjDetail detail
type GameAppVersionsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	Platform    string     `json:"platform"`
	VersionName string     `json:"versionName"`
	VersionCode uint       `json:"versionCode"`
	DownloadURL string     `json:"downloadURL"`
	ForceUpdate int        `json:"forceUpdate"`
	ReleaseNote string     `json:"releaseNote"`
	Status      int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameAppVersionsReply only for api docs
type CreateGameAppVersionsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameAppVersionsByIDReply only for api docs
type DeleteGameAppVersionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameAppVersionsByIDReply only for api docs
type UpdateGameAppVersionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameAppVersionsByIDReply only for api docs
type GetGameAppVersionsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAppVersions GameAppVersionsObjDetail `json:"gameAppVersions"`
	} `json:"data"` // return data
}

// ListGameAppVersionssRequest request params
type ListGameAppVersionssRequest struct {
	query.Params
}

// ListGameAppVersionssReply only for api docs
type ListGameAppVersionssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAppVersionss []GameAppVersionsObjDetail `json:"gameAppVersionss"`
	} `json:"data"` // return data
}
