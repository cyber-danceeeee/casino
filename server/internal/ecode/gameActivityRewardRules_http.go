package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameActivityRewardRules business-level http error codes.
// the gameActivityRewardRulesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameActivityRewardRulesNO       = 3
	gameActivityRewardRulesName     = "gameActivityRewardRules"
	gameActivityRewardRulesBaseCode = errcode.HCode(gameActivityRewardRulesNO)

	ErrCreateGameActivityRewardRules     = errcode.NewError(gameActivityRewardRulesBaseCode+1, "failed to create "+gameActivityRewardRulesName)
	ErrDeleteByIDGameActivityRewardRules = errcode.NewError(gameActivityRewardRulesBaseCode+2, "failed to delete "+gameActivityRewardRulesName)
	ErrUpdateByIDGameActivityRewardRules = errcode.NewError(gameActivityRewardRulesBaseCode+3, "failed to update "+gameActivityRewardRulesName)
	ErrGetByIDGameActivityRewardRules    = errcode.NewError(gameActivityRewardRulesBaseCode+4, "failed to get "+gameActivityRewardRulesName+" details")
	ErrListGameActivityRewardRules       = errcode.NewError(gameActivityRewardRulesBaseCode+5, "failed to list of "+gameActivityRewardRulesName)

	// error codes are globally unique, adding 1 to the previous error code
)
