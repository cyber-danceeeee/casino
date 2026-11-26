package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameHomeFeatureCardsRequest request params
type CreateGameHomeFeatureCardsRequest struct {
	FeatureCode     string     `json:"featureCode" binding:""`
	Title           string     `json:"title" binding:""`
	ImageURL        string     `json:"imageURL" binding:""`
	BackgroundStyle string     `json:"backgroundStyle" binding:""`
	LinkType        string     `json:"linkType" binding:""`
	LinkURL         string     `json:"linkURL" binding:""`
	CountdownEndAt  *time.Time `json:"countdownEndAt" binding:""`
	Sort            int        `json:"sort" binding:""`
	Status          int        `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameHomeFeatureCardsByIDRequest request params
type UpdateGameHomeFeatureCardsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	FeatureCode     string     `json:"featureCode" binding:""`
	Title           string     `json:"title" binding:""`
	ImageURL        string     `json:"imageURL" binding:""`
	BackgroundStyle string     `json:"backgroundStyle" binding:""`
	LinkType        string     `json:"linkType" binding:""`
	LinkURL         string     `json:"linkURL" binding:""`
	CountdownEndAt  *time.Time `json:"countdownEndAt" binding:""`
	Sort            int        `json:"sort" binding:""`
	Status          int        `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameHomeFeatureCardsObjDetail detail
type GameHomeFeatureCardsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	FeatureCode     string     `json:"featureCode"`
	Title           string     `json:"title"`
	ImageURL        string     `json:"imageURL"`
	BackgroundStyle string     `json:"backgroundStyle"`
	LinkType        string     `json:"linkType"`
	LinkURL         string     `json:"linkURL"`
	CountdownEndAt  *time.Time `json:"countdownEndAt"`
	Sort            int        `json:"sort"`
	Status          int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt       *time.Time `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
}

// CreateGameHomeFeatureCardsReply only for api docs
type CreateGameHomeFeatureCardsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameHomeFeatureCardsByIDReply only for api docs
type DeleteGameHomeFeatureCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameHomeFeatureCardsByIDReply only for api docs
type UpdateGameHomeFeatureCardsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameHomeFeatureCardsByIDReply only for api docs
type GetGameHomeFeatureCardsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameHomeFeatureCards GameHomeFeatureCardsObjDetail `json:"gameHomeFeatureCards"`
	} `json:"data"` // return data
}

// ListGameHomeFeatureCardssRequest request params
type ListGameHomeFeatureCardssRequest struct {
	query.Params
}

// ListGameHomeFeatureCardssReply only for api docs
type ListGameHomeFeatureCardssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameHomeFeatureCardss []GameHomeFeatureCardsObjDetail `json:"gameHomeFeatureCardss"`
	} `json:"data"` // return data
}
