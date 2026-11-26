package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserCheckins business-level http error codes.
// the gameUserCheckinsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserCheckinsNO       = 43
	gameUserCheckinsName     = "gameUserCheckins"
	gameUserCheckinsBaseCode = errcode.HCode(gameUserCheckinsNO)

	ErrCreateGameUserCheckins     = errcode.NewError(gameUserCheckinsBaseCode+1, "failed to create "+gameUserCheckinsName)
	ErrDeleteByIDGameUserCheckins = errcode.NewError(gameUserCheckinsBaseCode+2, "failed to delete "+gameUserCheckinsName)
	ErrUpdateByIDGameUserCheckins = errcode.NewError(gameUserCheckinsBaseCode+3, "failed to update "+gameUserCheckinsName)
	ErrGetByIDGameUserCheckins    = errcode.NewError(gameUserCheckinsBaseCode+4, "failed to get "+gameUserCheckinsName+" details")
	ErrListGameUserCheckins       = errcode.NewError(gameUserCheckinsBaseCode+5, "failed to list of "+gameUserCheckinsName)

	// error codes are globally unique, adding 1 to the previous error code
)
