package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserDevices business-level http error codes.
// the gameUserDevicesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserDevicesNO       = 44
	gameUserDevicesName     = "gameUserDevices"
	gameUserDevicesBaseCode = errcode.HCode(gameUserDevicesNO)

	ErrCreateGameUserDevices     = errcode.NewError(gameUserDevicesBaseCode+1, "failed to create "+gameUserDevicesName)
	ErrDeleteByIDGameUserDevices = errcode.NewError(gameUserDevicesBaseCode+2, "failed to delete "+gameUserDevicesName)
	ErrUpdateByIDGameUserDevices = errcode.NewError(gameUserDevicesBaseCode+3, "failed to update "+gameUserDevicesName)
	ErrGetByIDGameUserDevices    = errcode.NewError(gameUserDevicesBaseCode+4, "failed to get "+gameUserDevicesName+" details")
	ErrListGameUserDevices       = errcode.NewError(gameUserDevicesBaseCode+5, "failed to list of "+gameUserDevicesName)

	// error codes are globally unique, adding 1 to the previous error code
)
