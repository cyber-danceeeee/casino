package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameGameProviders business-level http error codes.
// the gameGameProvidersNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameGameProvidersNO       = 19
	gameGameProvidersName     = "gameGameProviders"
	gameGameProvidersBaseCode = errcode.HCode(gameGameProvidersNO)

	ErrCreateGameGameProviders     = errcode.NewError(gameGameProvidersBaseCode+1, "failed to create "+gameGameProvidersName)
	ErrDeleteByIDGameGameProviders = errcode.NewError(gameGameProvidersBaseCode+2, "failed to delete "+gameGameProvidersName)
	ErrUpdateByIDGameGameProviders = errcode.NewError(gameGameProvidersBaseCode+3, "failed to update "+gameGameProvidersName)
	ErrGetByIDGameGameProviders    = errcode.NewError(gameGameProvidersBaseCode+4, "failed to get "+gameGameProvidersName+" details")
	ErrListGameGameProviders       = errcode.NewError(gameGameProvidersBaseCode+5, "failed to list of "+gameGameProvidersName)

	// error codes are globally unique, adding 1 to the previous error code
)
