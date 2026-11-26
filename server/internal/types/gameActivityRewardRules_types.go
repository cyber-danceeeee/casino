package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameActivityRewardRulesRequest request params
type CreateGameActivityRewardRulesRequest struct {
	ActivityID            uint64 `json:"activityID" binding:""`
	Label                 string `json:"label" binding:""`
	RewardValue           string `json:"rewardValue" binding:""`
	BonusText             string `json:"bonusText" binding:""`
	MinAmount             string `json:"minAmount" binding:""`
	MaxAmount             string `json:"maxAmount" binding:""`
	RewardCash            string `json:"rewardCash" binding:""`
	RewardBonus           string `json:"rewardBonus" binding:""`
	RequirementMultiplier string `json:"requirementMultiplier" binding:""`
	Sort                  int    `json:"sort" binding:""`
	Status                int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameActivityRewardRulesByIDRequest request params
type UpdateGameActivityRewardRulesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ActivityID            uint64 `json:"activityID" binding:""`
	Label                 string `json:"label" binding:""`
	RewardValue           string `json:"rewardValue" binding:""`
	BonusText             string `json:"bonusText" binding:""`
	MinAmount             string `json:"minAmount" binding:""`
	MaxAmount             string `json:"maxAmount" binding:""`
	RewardCash            string `json:"rewardCash" binding:""`
	RewardBonus           string `json:"rewardBonus" binding:""`
	RequirementMultiplier string `json:"requirementMultiplier" binding:""`
	Sort                  int    `json:"sort" binding:""`
	Status                int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameActivityRewardRulesObjDetail detail
type GameActivityRewardRulesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ActivityID            uint64     `json:"activityID"`
	Label                 string     `json:"label"`
	RewardValue           string     `json:"rewardValue"`
	BonusText             string     `json:"bonusText"`
	MinAmount             string     `json:"minAmount"`
	MaxAmount             string     `json:"maxAmount"`
	RewardCash            string     `json:"rewardCash"`
	RewardBonus           string     `json:"rewardBonus"`
	RequirementMultiplier string     `json:"requirementMultiplier"`
	Sort                  int        `json:"sort"`
	Status                int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt             *time.Time `json:"createdAt"`
	UpdatedAt             *time.Time `json:"updatedAt"`
}

// CreateGameActivityRewardRulesReply only for api docs
type CreateGameActivityRewardRulesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameActivityRewardRulesByIDReply only for api docs
type DeleteGameActivityRewardRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameActivityRewardRulesByIDReply only for api docs
type UpdateGameActivityRewardRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameActivityRewardRulesByIDReply only for api docs
type GetGameActivityRewardRulesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityRewardRules GameActivityRewardRulesObjDetail `json:"gameActivityRewardRules"`
	} `json:"data"` // return data
}

// ListGameActivityRewardRulessRequest request params
type ListGameActivityRewardRulessRequest struct {
	query.Params
}

// ListGameActivityRewardRulessReply only for api docs
type ListGameActivityRewardRulessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityRewardRuless []GameActivityRewardRulesObjDetail `json:"gameActivityRewardRuless"`
	} `json:"data"` // return data
}
