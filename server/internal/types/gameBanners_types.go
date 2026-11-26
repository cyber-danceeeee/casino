package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameBannersRequest request params
type CreateGameBannersRequest struct {
	BannerCode  string `json:"bannerCode" binding:""`
	Position  string `json:"position" binding:""`
	Title  string `json:"title" binding:""`
	ImageURL  string `json:"imageURL" binding:""`
	LinkType  string `json:"linkType" binding:""` // none, route, activity, external
	LinkURL  string `json:"linkURL" binding:""`
	ActivitySlug  string `json:"activitySlug" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
	StartAt  *time.Time `json:"startAt" binding:""`
	EndAt  *time.Time `json:"endAt" binding:""`
}

// UpdateGameBannersByIDRequest request params
type UpdateGameBannersByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	BannerCode  string `json:"bannerCode" binding:""`
	Position  string `json:"position" binding:""`
	Title  string `json:"title" binding:""`
	ImageURL  string `json:"imageURL" binding:""`
	LinkType  string `json:"linkType" binding:""` // none, route, activity, external
	LinkURL  string `json:"linkURL" binding:""`
	ActivitySlug  string `json:"activitySlug" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
	StartAt  *time.Time `json:"startAt" binding:""`
	EndAt  *time.Time `json:"endAt" binding:""`
}

// GameBannersObjDetail detail
type GameBannersObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	BannerCode  string `json:"bannerCode"`
	Position  string `json:"position"`
	Title  string `json:"title"`
	ImageURL  string `json:"imageURL"`
	LinkType  string `json:"linkType"` // none, route, activity, external
	LinkURL  string `json:"linkURL"`
	ActivitySlug  string `json:"activitySlug"`
	Sort  int `json:"sort"`
	Status  int `json:"status"` // 1 enabled, 2 disabled
	StartAt  *time.Time `json:"startAt"`
	EndAt  *time.Time `json:"endAt"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameBannersReply only for api docs
type CreateGameBannersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameBannersByIDReply only for api docs
type DeleteGameBannersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameBannersByIDReply only for api docs
type UpdateGameBannersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameBannersByIDReply only for api docs
type GetGameBannersByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBanners GameBannersObjDetail `json:"gameBanners"`
	} `json:"data"` // return data
}

// ListGameBannerssRequest request params
type ListGameBannerssRequest struct {
	query.Params
}

// ListGameBannerssReply only for api docs
type ListGameBannerssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameBannerss []GameBannersObjDetail `json:"gameBannerss"`
	} `json:"data"` // return data
}
