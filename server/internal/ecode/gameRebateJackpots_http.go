package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameRebateJackpots business-level http error codes.
// the gameRebateJackpotsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameRebateJackpotsNO = 33
	gameRebateJackpotsName     = "gameRebateJackpots"
	gameRebateJackpotsBaseCode = errcode.HCode(gameRebateJackpotsNO)

	ErrCreateGameRebateJackpots     = errcode.NewError(gameRebateJackpotsBaseCode+1, "failed to create "+gameRebateJackpotsName)
	ErrDeleteByIDGameRebateJackpots = errcode.NewError(gameRebateJackpotsBaseCode+2, "failed to delete "+gameRebateJackpotsName)
	ErrUpdateByIDGameRebateJackpots = errcode.NewError(gameRebateJackpotsBaseCode+3, "failed to update "+gameRebateJackpotsName)
	ErrGetByIDGameRebateJackpots    = errcode.NewError(gameRebateJackpotsBaseCode+4, "failed to get "+gameRebateJackpotsName+" details")
	ErrListGameRebateJackpots       = errcode.NewError(gameRebateJackpotsBaseCode+5, "failed to list of "+gameRebateJackpotsName)

	// error codes are globally unique, adding 1 to the previous error code
)
