package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameNoticesRequest request params
type CreateGameNoticesRequest struct {
	NoticeCode  string     `json:"noticeCode" binding:""`
	NoticeType  string     `json:"noticeType" binding:""`
	Title       string     `json:"title" binding:""`
	ContentHTML string     `json:"contentHTML" binding:""`
	ImageURL    string     `json:"imageURL" binding:""`
	Sort        int        `json:"sort" binding:""`
	Status      int        `json:"status" binding:""` // 1 enabled, 2 disabled
	StartAt     *time.Time `json:"startAt" binding:""`
	EndAt       *time.Time `json:"endAt" binding:""`
}

// UpdateGameNoticesByIDRequest request params
type UpdateGameNoticesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	NoticeCode  string     `json:"noticeCode" binding:""`
	NoticeType  string     `json:"noticeType" binding:""`
	Title       string     `json:"title" binding:""`
	ContentHTML string     `json:"contentHTML" binding:""`
	ImageURL    string     `json:"imageURL" binding:""`
	Sort        int        `json:"sort" binding:""`
	Status      int        `json:"status" binding:""` // 1 enabled, 2 disabled
	StartAt     *time.Time `json:"startAt" binding:""`
	EndAt       *time.Time `json:"endAt" binding:""`
}

// GameNoticesObjDetail detail
type GameNoticesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	NoticeCode  string     `json:"noticeCode"`
	NoticeType  string     `json:"noticeType"`
	Title       string     `json:"title"`
	ContentHTML string     `json:"contentHTML"`
	ImageURL    string     `json:"imageURL"`
	Sort        int        `json:"sort"`
	Status      int        `json:"status"` // 1 enabled, 2 disabled
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameNoticesReply only for api docs
type CreateGameNoticesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameNoticesByIDReply only for api docs
type DeleteGameNoticesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameNoticesByIDReply only for api docs
type UpdateGameNoticesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameNoticesByIDReply only for api docs
type GetGameNoticesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameNotices GameNoticesObjDetail `json:"gameNotices"`
	} `json:"data"` // return data
}

// ListGameNoticessRequest request params
type ListGameNoticessRequest struct {
	query.Params
}

// ListGameNoticessReply only for api docs
type ListGameNoticessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameNoticess []GameNoticesObjDetail `json:"gameNoticess"`
	} `json:"data"` // return data
}
