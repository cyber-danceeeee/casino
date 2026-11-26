package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameAgentCommissionRules business-level http error codes.
// the gameAgentCommissionRulesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameAgentCommissionRulesNO = 6
	gameAgentCommissionRulesName     = "gameAgentCommissionRules"
	gameAgentCommissionRulesBaseCode = errcode.HCode(gameAgentCommissionRulesNO)

	ErrCreateGameAgentCommissionRules     = errcode.NewError(gameAgentCommissionRulesBaseCode+1, "failed to create "+gameAgentCommissionRulesName)
	ErrDeleteByIDGameAgentCommissionRules = errcode.NewError(gameAgentCommissionRulesBaseCode+2, "failed to delete "+gameAgentCommissionRulesName)
	ErrUpdateByIDGameAgentCommissionRules = errcode.NewError(gameAgentCommissionRulesBaseCode+3, "failed to update "+gameAgentCommissionRulesName)
	ErrGetByIDGameAgentCommissionRules    = errcode.NewError(gameAgentCommissionRulesBaseCode+4, "failed to get "+gameAgentCommissionRulesName+" details")
	ErrListGameAgentCommissionRules       = errcode.NewError(gameAgentCommissionRulesBaseCode+5, "failed to list of "+gameAgentCommissionRulesName)

	// error codes are globally unique, adding 1 to the previous error code
)
