package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserActivityRecords business-level http error codes.
// the gameUserActivityRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserActivityRecordsNO       = 42
	gameUserActivityRecordsName     = "gameUserActivityRecords"
	gameUserActivityRecordsBaseCode = errcode.HCode(gameUserActivityRecordsNO)

	ErrCreateGameUserActivityRecords     = errcode.NewError(gameUserActivityRecordsBaseCode+1, "failed to create "+gameUserActivityRecordsName)
	ErrDeleteByIDGameUserActivityRecords = errcode.NewError(gameUserActivityRecordsBaseCode+2, "failed to delete "+gameUserActivityRecordsName)
	ErrUpdateByIDGameUserActivityRecords = errcode.NewError(gameUserActivityRecordsBaseCode+3, "failed to update "+gameUserActivityRecordsName)
	ErrGetByIDGameUserActivityRecords    = errcode.NewError(gameUserActivityRecordsBaseCode+4, "failed to get "+gameUserActivityRecordsName+" details")
	ErrListGameUserActivityRecords       = errcode.NewError(gameUserActivityRecordsBaseCode+5, "failed to list of "+gameUserActivityRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
