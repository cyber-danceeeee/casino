package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameActivityEvents business-level http error codes.
// the gameActivityEventsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameActivityEventsNO = 2
	gameActivityEventsName     = "gameActivityEvents"
	gameActivityEventsBaseCode = errcode.HCode(gameActivityEventsNO)

	ErrCreateGameActivityEvents     = errcode.NewError(gameActivityEventsBaseCode+1, "failed to create "+gameActivityEventsName)
	ErrDeleteByIDGameActivityEvents = errcode.NewError(gameActivityEventsBaseCode+2, "failed to delete "+gameActivityEventsName)
	ErrUpdateByIDGameActivityEvents = errcode.NewError(gameActivityEventsBaseCode+3, "failed to update "+gameActivityEventsName)
	ErrGetByIDGameActivityEvents    = errcode.NewError(gameActivityEventsBaseCode+4, "failed to get "+gameActivityEventsName+" details")
	ErrListGameActivityEvents       = errcode.NewError(gameActivityEventsBaseCode+5, "failed to list of "+gameActivityEventsName)

	// error codes are globally unique, adding 1 to the previous error code
)
