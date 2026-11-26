package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameCmsPagesRequest request params
type CreateGameCmsPagesRequest struct {
	PageCode    string     `json:"pageCode" binding:""`
	Title       string     `json:"title" binding:""`
	ContentHTML string     `json:"contentHTML" binding:""`
	VersionNo   uint       `json:"versionNo" binding:""`
	Status      int        `json:"status" binding:""` // 1 published, 2 draft, 3 disabled
	EffectiveAt *time.Time `json:"effectiveAt" binding:""`
}

// UpdateGameCmsPagesByIDRequest request params
type UpdateGameCmsPagesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	PageCode    string     `json:"pageCode" binding:""`
	Title       string     `json:"title" binding:""`
	ContentHTML string     `json:"contentHTML" binding:""`
	VersionNo   uint       `json:"versionNo" binding:""`
	Status      int        `json:"status" binding:""` // 1 published, 2 draft, 3 disabled
	EffectiveAt *time.Time `json:"effectiveAt" binding:""`
}

// GameCmsPagesObjDetail detail
type GameCmsPagesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	PageCode    string     `json:"pageCode"`
	Title       string     `json:"title"`
	ContentHTML string     `json:"contentHTML"`
	VersionNo   uint       `json:"versionNo"`
	Status      int        `json:"status"` // 1 published, 2 draft, 3 disabled
	EffectiveAt *time.Time `json:"effectiveAt"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameCmsPagesReply only for api docs
type CreateGameCmsPagesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameCmsPagesByIDReply only for api docs
type DeleteGameCmsPagesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameCmsPagesByIDReply only for api docs
type UpdateGameCmsPagesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameCmsPagesByIDReply only for api docs
type GetGameCmsPagesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameCmsPages GameCmsPagesObjDetail `json:"gameCmsPages"`
	} `json:"data"` // return data
}

// ListGameCmsPagessRequest request params
type ListGameCmsPagessRequest struct {
	query.Params
}

// ListGameCmsPagessReply only for api docs
type ListGameCmsPagessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameCmsPagess []GameCmsPagesObjDetail `json:"gameCmsPagess"`
	} `json:"data"` // return data
}
