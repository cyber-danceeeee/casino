package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserSessions business-level http error codes.
// the gameUserSessionsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserSessionsNO = 47
	gameUserSessionsName     = "gameUserSessions"
	gameUserSessionsBaseCode = errcode.HCode(gameUserSessionsNO)

	ErrCreateGameUserSessions     = errcode.NewError(gameUserSessionsBaseCode+1, "failed to create "+gameUserSessionsName)
	ErrDeleteByIDGameUserSessions = errcode.NewError(gameUserSessionsBaseCode+2, "failed to delete "+gameUserSessionsName)
	ErrUpdateByIDGameUserSessions = errcode.NewError(gameUserSessionsBaseCode+3, "failed to update "+gameUserSessionsName)
	ErrGetByIDGameUserSessions    = errcode.NewError(gameUserSessionsBaseCode+4, "failed to get "+gameUserSessionsName+" details")
	ErrListGameUserSessions       = errcode.NewError(gameUserSessionsBaseCode+5, "failed to list of "+gameUserSessionsName)

	// error codes are globally unique, adding 1 to the previous error code
)
