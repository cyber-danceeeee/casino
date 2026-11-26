package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSystemConfigs business-level http error codes.
// the gameSystemConfigsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSystemConfigsNO       = 41
	gameSystemConfigsName     = "gameSystemConfigs"
	gameSystemConfigsBaseCode = errcode.HCode(gameSystemConfigsNO)

	ErrCreateGameSystemConfigs     = errcode.NewError(gameSystemConfigsBaseCode+1, "failed to create "+gameSystemConfigsName)
	ErrDeleteByIDGameSystemConfigs = errcode.NewError(gameSystemConfigsBaseCode+2, "failed to delete "+gameSystemConfigsName)
	ErrUpdateByIDGameSystemConfigs = errcode.NewError(gameSystemConfigsBaseCode+3, "failed to update "+gameSystemConfigsName)
	ErrGetByIDGameSystemConfigs    = errcode.NewError(gameSystemConfigsBaseCode+4, "failed to get "+gameSystemConfigsName+" details")
	ErrListGameSystemConfigs       = errcode.NewError(gameSystemConfigsBaseCode+5, "failed to list of "+gameSystemConfigsName)

	// error codes are globally unique, adding 1 to the previous error code
)
