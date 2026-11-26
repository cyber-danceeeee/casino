package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameNotices business-level http error codes.
// the gameNoticesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameNoticesNO = 28
	gameNoticesName     = "gameNotices"
	gameNoticesBaseCode = errcode.HCode(gameNoticesNO)

	ErrCreateGameNotices     = errcode.NewError(gameNoticesBaseCode+1, "failed to create "+gameNoticesName)
	ErrDeleteByIDGameNotices = errcode.NewError(gameNoticesBaseCode+2, "failed to delete "+gameNoticesName)
	ErrUpdateByIDGameNotices = errcode.NewError(gameNoticesBaseCode+3, "failed to update "+gameNoticesName)
	ErrGetByIDGameNotices    = errcode.NewError(gameNoticesBaseCode+4, "failed to get "+gameNoticesName+" details")
	ErrListGameNotices       = errcode.NewError(gameNoticesBaseCode+5, "failed to list of "+gameNoticesName)

	// error codes are globally unique, adding 1 to the previous error code
)
