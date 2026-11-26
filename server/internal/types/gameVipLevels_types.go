package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameVipLevelsRequest request params
type CreateGameVipLevelsRequest struct {
	LevelNo            uint   `json:"levelNo" binding:""`
	Name               string `json:"name" binding:""`
	IconURL            string `json:"iconURL" binding:""`
	LevelUpReward      string `json:"levelUpReward" binding:""`
	WeeklyReward       string `json:"weeklyReward" binding:""`
	MonthlyReward      string `json:"monthlyReward" binding:""`
	DailyWithdrawCount uint   `json:"dailyWithdrawCount" binding:""`
	Status             int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameVipLevelsByIDRequest request params
type UpdateGameVipLevelsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	LevelNo            uint   `json:"levelNo" binding:""`
	Name               string `json:"name" binding:""`
	IconURL            string `json:"iconURL" binding:""`
	LevelUpReward      string `json:"levelUpReward" binding:""`
	WeeklyReward       string `json:"weeklyReward" binding:""`
	MonthlyReward      string `json:"monthlyReward" binding:""`
	DailyWithdrawCount uint   `json:"dailyWithdrawCount" binding:""`
	Status             int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameVipLevelsObjDetail detail
type GameVipLevelsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	LevelNo            uint       `json:"levelNo"`
	Name               string     `json:"name"`
	IconURL            string     `json:"iconURL"`
	LevelUpReward      string     `json:"levelUpReward"`
	WeeklyReward       string     `json:"weeklyReward"`
	MonthlyReward      string     `json:"monthlyReward"`
	DailyWithdrawCount uint       `json:"dailyWithdrawCount"`
	Status             int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt          *time.Time `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
}

// CreateGameVipLevelsReply only for api docs
type CreateGameVipLevelsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameVipLevelsByIDReply only for api docs
type DeleteGameVipLevelsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameVipLevelsByIDReply only for api docs
type UpdateGameVipLevelsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameVipLevelsByIDReply only for api docs
type GetGameVipLevelsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameVipLevels GameVipLevelsObjDetail `json:"gameVipLevels"`
	} `json:"data"` // return data
}

// ListGameVipLevelssRequest request params
type ListGameVipLevelssRequest struct {
	query.Params
}

// ListGameVipLevelssReply only for api docs
type ListGameVipLevelssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameVipLevelss []GameVipLevelsObjDetail `json:"gameVipLevelss"`
	} `json:"data"` // return data
}
