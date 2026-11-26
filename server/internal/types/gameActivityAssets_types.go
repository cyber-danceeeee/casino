package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameActivityAssetsRequest request params
type CreateGameActivityAssetsRequest struct {
	ActivityID uint64 `json:"activityID" binding:""`
	AssetKey   string `json:"assetKey" binding:""`
	AssetURL   string `json:"assetURL" binding:""`
	Sort       int    `json:"sort" binding:""`
}

// UpdateGameActivityAssetsByIDRequest request params
type UpdateGameActivityAssetsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ActivityID uint64 `json:"activityID" binding:""`
	AssetKey   string `json:"assetKey" binding:""`
	AssetURL   string `json:"assetURL" binding:""`
	Sort       int    `json:"sort" binding:""`
}

// GameActivityAssetsObjDetail detail
type GameActivityAssetsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ActivityID uint64     `json:"activityID"`
	AssetKey   string     `json:"assetKey"`
	AssetURL   string     `json:"assetURL"`
	Sort       int        `json:"sort"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}

// CreateGameActivityAssetsReply only for api docs
type CreateGameActivityAssetsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameActivityAssetsByIDReply only for api docs
type DeleteGameActivityAssetsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameActivityAssetsByIDReply only for api docs
type UpdateGameActivityAssetsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameActivityAssetsByIDReply only for api docs
type GetGameActivityAssetsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityAssets GameActivityAssetsObjDetail `json:"gameActivityAssets"`
	} `json:"data"` // return data
}

// ListGameActivityAssetssRequest request params
type ListGameActivityAssetssRequest struct {
	query.Params
}

// ListGameActivityAssetssReply only for api docs
type ListGameActivityAssetssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityAssetss []GameActivityAssetsObjDetail `json:"gameActivityAssetss"`
	} `json:"data"` // return data
}
