package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameVipRequirementsRequest request params
type CreateGameVipRequirementsRequest struct {
	VipLevelID      uint64 `json:"vipLevelID" binding:""`
	RequirementType string `json:"requirementType" binding:""` // level_up, retain
	TargetLevelNo   uint   `json:"targetLevelNo" binding:""`
	PeriodDays      uint   `json:"periodDays" binding:""`
	DepositAmount   string `json:"depositAmount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
}

// UpdateGameVipRequirementsByIDRequest request params
type UpdateGameVipRequirementsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	VipLevelID      uint64 `json:"vipLevelID" binding:""`
	RequirementType string `json:"requirementType" binding:""` // level_up, retain
	TargetLevelNo   uint   `json:"targetLevelNo" binding:""`
	PeriodDays      uint   `json:"periodDays" binding:""`
	DepositAmount   string `json:"depositAmount" binding:""`
	ValidBetAmount  string `json:"validBetAmount" binding:""`
}

// GameVipRequirementsObjDetail detail
type GameVipRequirementsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	VipLevelID      uint64     `json:"vipLevelID"`
	RequirementType string     `json:"requirementType"` // level_up, retain
	TargetLevelNo   uint       `json:"targetLevelNo"`
	PeriodDays      uint       `json:"periodDays"`
	DepositAmount   string     `json:"depositAmount"`
	ValidBetAmount  string     `json:"validBetAmount"`
	CreatedAt       *time.Time `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
}

// CreateGameVipRequirementsReply only for api docs
type CreateGameVipRequirementsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameVipRequirementsByIDReply only for api docs
type DeleteGameVipRequirementsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameVipRequirementsByIDReply only for api docs
type UpdateGameVipRequirementsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameVipRequirementsByIDReply only for api docs
type GetGameVipRequirementsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameVipRequirements GameVipRequirementsObjDetail `json:"gameVipRequirements"`
	} `json:"data"` // return data
}

// ListGameVipRequirementssRequest request params
type ListGameVipRequirementssRequest struct {
	query.Params
}

// ListGameVipRequirementssReply only for api docs
type ListGameVipRequirementssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameVipRequirementss []GameVipRequirementsObjDetail `json:"gameVipRequirementss"`
	} `json:"data"` // return data
}
