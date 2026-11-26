package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameVipLevels business-level http error codes.
// the gameVipLevelsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameVipLevelsNO       = 53
	gameVipLevelsName     = "gameVipLevels"
	gameVipLevelsBaseCode = errcode.HCode(gameVipLevelsNO)

	ErrCreateGameVipLevels     = errcode.NewError(gameVipLevelsBaseCode+1, "failed to create "+gameVipLevelsName)
	ErrDeleteByIDGameVipLevels = errcode.NewError(gameVipLevelsBaseCode+2, "failed to delete "+gameVipLevelsName)
	ErrUpdateByIDGameVipLevels = errcode.NewError(gameVipLevelsBaseCode+3, "failed to update "+gameVipLevelsName)
	ErrGetByIDGameVipLevels    = errcode.NewError(gameVipLevelsBaseCode+4, "failed to get "+gameVipLevelsName+" details")
	ErrListGameVipLevels       = errcode.NewError(gameVipLevelsBaseCode+5, "failed to list of "+gameVipLevelsName)

	// error codes are globally unique, adding 1 to the previous error code
)
