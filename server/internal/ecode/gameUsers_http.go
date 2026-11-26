package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUsers business-level http error codes.
// the gameUsersNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUsersNO       = 52
	gameUsersName     = "gameUsers"
	gameUsersBaseCode = errcode.HCode(gameUsersNO)

	ErrCreateGameUsers     = errcode.NewError(gameUsersBaseCode+1, "failed to create "+gameUsersName)
	ErrDeleteByIDGameUsers = errcode.NewError(gameUsersBaseCode+2, "failed to delete "+gameUsersName)
	ErrUpdateByIDGameUsers = errcode.NewError(gameUsersBaseCode+3, "failed to update "+gameUsersName)
	ErrGetByIDGameUsers    = errcode.NewError(gameUsersBaseCode+4, "failed to get "+gameUsersName+" details")
	ErrListGameUsers       = errcode.NewError(gameUsersBaseCode+5, "failed to list of "+gameUsersName)

	// error codes are globally unique, adding 1 to the previous error code
)
