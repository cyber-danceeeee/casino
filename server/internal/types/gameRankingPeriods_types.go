package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameRankingPeriodsRequest request params
type CreateGameRankingPeriodsRequest struct {
	RankingType  string `json:"rankingType" binding:""`
	Title  string `json:"title" binding:""`
	RankingDate  *time.Time `json:"rankingDate" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 settled, 3 disabled
}

// UpdateGameRankingPeriodsByIDRequest request params
type UpdateGameRankingPeriodsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RankingType  string `json:"rankingType" binding:""`
	Title  string `json:"title" binding:""`
	RankingDate  *time.Time `json:"rankingDate" binding:""`
	PeriodStart  *time.Time `json:"periodStart" binding:""`
	PeriodEnd  *time.Time `json:"periodEnd" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 settled, 3 disabled
}

// GameRankingPeriodsObjDetail detail
type GameRankingPeriodsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RankingType  string `json:"rankingType"`
	Title  string `json:"title"`
	RankingDate  *time.Time `json:"rankingDate"`
	PeriodStart  *time.Time `json:"periodStart"`
	PeriodEnd  *time.Time `json:"periodEnd"`
	Status  int `json:"status"` // 1 active, 2 settled, 3 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameRankingPeriodsReply only for api docs
type CreateGameRankingPeriodsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameRankingPeriodsByIDReply only for api docs
type DeleteGameRankingPeriodsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameRankingPeriodsByIDReply only for api docs
type UpdateGameRankingPeriodsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameRankingPeriodsByIDReply only for api docs
type GetGameRankingPeriodsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRankingPeriods GameRankingPeriodsObjDetail `json:"gameRankingPeriods"`
	} `json:"data"` // return data
}

// ListGameRankingPeriodssRequest request params
type ListGameRankingPeriodssRequest struct {
	query.Params
}

// ListGameRankingPeriodssReply only for api docs
type ListGameRankingPeriodssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRankingPeriodss []GameRankingPeriodsObjDetail `json:"gameRankingPeriodss"`
	} `json:"data"` // return data
}
