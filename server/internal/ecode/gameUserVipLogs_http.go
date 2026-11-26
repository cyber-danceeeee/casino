package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserVipLogs business-level http error codes.
// the gameUserVipLogsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserVipLogsNO = 49
	gameUserVipLogsName     = "gameUserVipLogs"
	gameUserVipLogsBaseCode = errcode.HCode(gameUserVipLogsNO)

	ErrCreateGameUserVipLogs     = errcode.NewError(gameUserVipLogsBaseCode+1, "failed to create "+gameUserVipLogsName)
	ErrDeleteByIDGameUserVipLogs = errcode.NewError(gameUserVipLogsBaseCode+2, "failed to delete "+gameUserVipLogsName)
	ErrUpdateByIDGameUserVipLogs = errcode.NewError(gameUserVipLogsBaseCode+3, "failed to update "+gameUserVipLogsName)
	ErrGetByIDGameUserVipLogs    = errcode.NewError(gameUserVipLogsBaseCode+4, "failed to get "+gameUserVipLogsName+" details")
	ErrListGameUserVipLogs       = errcode.NewError(gameUserVipLogsBaseCode+5, "failed to list of "+gameUserVipLogsName)

	// error codes are globally unique, adding 1 to the previous error code
)
