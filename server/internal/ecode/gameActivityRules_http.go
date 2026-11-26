package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameActivityRules business-level http error codes.
// the gameActivityRulesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameActivityRulesNO       = 4
	gameActivityRulesName     = "gameActivityRules"
	gameActivityRulesBaseCode = errcode.HCode(gameActivityRulesNO)

	ErrCreateGameActivityRules     = errcode.NewError(gameActivityRulesBaseCode+1, "failed to create "+gameActivityRulesName)
	ErrDeleteByIDGameActivityRules = errcode.NewError(gameActivityRulesBaseCode+2, "failed to delete "+gameActivityRulesName)
	ErrUpdateByIDGameActivityRules = errcode.NewError(gameActivityRulesBaseCode+3, "failed to update "+gameActivityRulesName)
	ErrGetByIDGameActivityRules    = errcode.NewError(gameActivityRulesBaseCode+4, "failed to get "+gameActivityRulesName+" details")
	ErrListGameActivityRules       = errcode.NewError(gameActivityRulesBaseCode+5, "failed to list of "+gameActivityRulesName)

	// error codes are globally unique, adding 1 to the previous error code
)
