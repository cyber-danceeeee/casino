package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameAgentCommissionRulesRequest request params
type CreateGameAgentCommissionRulesRequest struct {
	LevelNo        uint   `json:"levelNo" binding:""`
	MinReferrals   uint   `json:"minReferrals" binding:""`
	MinValidBet    string `json:"minValidBet" binding:""`
	CommissionRate string `json:"commissionRate" binding:""`
	BonusRate      string `json:"bonusRate" binding:""`
	Status         int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameAgentCommissionRulesByIDRequest request params
type UpdateGameAgentCommissionRulesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	LevelNo        uint   `json:"levelNo" binding:""`
	MinReferrals   uint   `json:"minReferrals" binding:""`
	MinValidBet    string `json:"minValidBet" binding:""`
	CommissionRate string `json:"commissionRate" binding:""`
	BonusRate      string `json:"bonusRate" binding:""`
	Status         int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameAgentCommissionRulesObjDetail detail
type GameAgentCommissionRulesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	LevelNo        uint       `json:"levelNo"`
	MinReferrals   uint       `json:"minReferrals"`
	MinValidBet    string     `json:"minValidBet"`
	CommissionRate string     `json:"commissionRate"`
	BonusRate      string     `json:"bonusRate"`
	Status         int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt      *time.Time `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

// CreateGameAgentCommissionRulesReply only for api docs
type CreateGameAgentCommissionRulesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameAgentCommissionRulesByIDReply only for api docs
type DeleteGameAgentCommissionRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameAgentCommissionRulesByIDReply only for api docs
type UpdateGameAgentCommissionRulesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameAgentCommissionRulesByIDReply only for api docs
type GetGameAgentCommissionRulesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAgentCommissionRules GameAgentCommissionRulesObjDetail `json:"gameAgentCommissionRules"`
	} `json:"data"` // return data
}

// ListGameAgentCommissionRulessRequest request params
type ListGameAgentCommissionRulessRequest struct {
	query.Params
}

// ListGameAgentCommissionRulessReply only for api docs
type ListGameAgentCommissionRulessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameAgentCommissionRuless []GameAgentCommissionRulesObjDetail `json:"gameAgentCommissionRuless"`
	} `json:"data"` // return data
}
