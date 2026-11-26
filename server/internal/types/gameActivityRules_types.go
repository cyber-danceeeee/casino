package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameActivityRulesRequest request params
type CreateGameActivityRulesRequest struct {
	ActivityID  uint64 `json:"activityID" binding:""`
	Title       string `json:"title" binding:""`
	ContentHTML string `json:"contentHTML" binding:""`
	Sort        int    `json:"sort" binding:""`
	Status      int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameActivityRulesByIDRequest request params
type UpdateGameActivityRulesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ActivityID  uint64 `json:"activityID" binding:""`
	Title       string `json:"title" binding:""`
	ContentHTML string `json:"contentHTML" binding:""`
	Sort        int    `json:"sort" binding:""`
	Status      int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameActivityRulesObjDetail detail
type GameActivityRulesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ActivityID  uint64     `json:"activityID"`
	Title       string     `json:"title"`
	ContentHTML string     `json:"contentHTML"`
	Sort        int        `json:"sort"`
	Status      int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameActivityRulesReply only for api docs
type CreateGameActivityRulesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameActivityRulesByIDReply only for api docs
type DeleteGameActivityRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameActivityRulesByIDReply only for api docs
type UpdateGameActivityRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameActivityRulesByIDReply only for api docs
type GetGameActivityRulesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityRules GameActivityRulesObjDetail `json:"gameActivityRules"`
	} `json:"data"` // return data
}

// ListGameActivityRulessRequest request params
type ListGameActivityRulessRequest struct {
	query.Params
}

// ListGameActivityRulessReply only for api docs
type ListGameActivityRulessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityRuless []GameActivityRulesObjDetail `json:"gameActivityRuless"`
	} `json:"data"` // return data
}
