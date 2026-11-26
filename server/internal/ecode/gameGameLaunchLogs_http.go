package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameGameLaunchLogs business-level http error codes.
// the gameGameLaunchLogsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameGameLaunchLogsNO       = 18
	gameGameLaunchLogsName     = "gameGameLaunchLogs"
	gameGameLaunchLogsBaseCode = errcode.HCode(gameGameLaunchLogsNO)

	ErrCreateGameGameLaunchLogs     = errcode.NewError(gameGameLaunchLogsBaseCode+1, "failed to create "+gameGameLaunchLogsName)
	ErrDeleteByIDGameGameLaunchLogs = errcode.NewError(gameGameLaunchLogsBaseCode+2, "failed to delete "+gameGameLaunchLogsName)
	ErrUpdateByIDGameGameLaunchLogs = errcode.NewError(gameGameLaunchLogsBaseCode+3, "failed to update "+gameGameLaunchLogsName)
	ErrGetByIDGameGameLaunchLogs    = errcode.NewError(gameGameLaunchLogsBaseCode+4, "failed to get "+gameGameLaunchLogsName+" details")
	ErrListGameGameLaunchLogs       = errcode.NewError(gameGameLaunchLogsBaseCode+5, "failed to list of "+gameGameLaunchLogsName)

	// error codes are globally unique, adding 1 to the previous error code
)
