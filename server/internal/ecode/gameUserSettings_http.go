package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserSettings business-level http error codes.
// the gameUserSettingsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserSettingsNO = 48
	gameUserSettingsName     = "gameUserSettings"
	gameUserSettingsBaseCode = errcode.HCode(gameUserSettingsNO)

	ErrCreateGameUserSettings     = errcode.NewError(gameUserSettingsBaseCode+1, "failed to create "+gameUserSettingsName)
	ErrDeleteByIDGameUserSettings = errcode.NewError(gameUserSettingsBaseCode+2, "failed to delete "+gameUserSettingsName)
	ErrUpdateByIDGameUserSettings = errcode.NewError(gameUserSettingsBaseCode+3, "failed to update "+gameUserSettingsName)
	ErrGetByIDGameUserSettings    = errcode.NewError(gameUserSettingsBaseCode+4, "failed to get "+gameUserSettingsName+" details")
	ErrListGameUserSettings       = errcode.NewError(gameUserSettingsBaseCode+5, "failed to list of "+gameUserSettingsName)

	// error codes are globally unique, adding 1 to the previous error code
)
