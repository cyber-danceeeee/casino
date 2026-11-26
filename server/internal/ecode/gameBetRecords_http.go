package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameBetRecords business-level http error codes.
// the gameBetRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameBetRecordsNO = 10
	gameBetRecordsName     = "gameBetRecords"
	gameBetRecordsBaseCode = errcode.HCode(gameBetRecordsNO)

	ErrCreateGameBetRecords     = errcode.NewError(gameBetRecordsBaseCode+1, "failed to create "+gameBetRecordsName)
	ErrDeleteByIDGameBetRecords = errcode.NewError(gameBetRecordsBaseCode+2, "failed to delete "+gameBetRecordsName)
	ErrUpdateByIDGameBetRecords = errcode.NewError(gameBetRecordsBaseCode+3, "failed to update "+gameBetRecordsName)
	ErrGetByIDGameBetRecords    = errcode.NewError(gameBetRecordsBaseCode+4, "failed to get "+gameBetRecordsName+" details")
	ErrListGameBetRecords       = errcode.NewError(gameBetRecordsBaseCode+5, "failed to list of "+gameBetRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
