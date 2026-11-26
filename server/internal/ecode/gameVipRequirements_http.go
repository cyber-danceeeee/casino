package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameVipRequirements business-level http error codes.
// the gameVipRequirementsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameVipRequirementsNO       = 54
	gameVipRequirementsName     = "gameVipRequirements"
	gameVipRequirementsBaseCode = errcode.HCode(gameVipRequirementsNO)

	ErrCreateGameVipRequirements     = errcode.NewError(gameVipRequirementsBaseCode+1, "failed to create "+gameVipRequirementsName)
	ErrDeleteByIDGameVipRequirements = errcode.NewError(gameVipRequirementsBaseCode+2, "failed to delete "+gameVipRequirementsName)
	ErrUpdateByIDGameVipRequirements = errcode.NewError(gameVipRequirementsBaseCode+3, "failed to update "+gameVipRequirementsName)
	ErrGetByIDGameVipRequirements    = errcode.NewError(gameVipRequirementsBaseCode+4, "failed to get "+gameVipRequirementsName+" details")
	ErrListGameVipRequirements       = errcode.NewError(gameVipRequirementsBaseCode+5, "failed to list of "+gameVipRequirementsName)

	// error codes are globally unique, adding 1 to the previous error code
)
