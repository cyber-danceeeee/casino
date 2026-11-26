package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameBonusRecords business-level http error codes.
// the gameBonusRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameBonusRecordsNO = 12
	gameBonusRecordsName     = "gameBonusRecords"
	gameBonusRecordsBaseCode = errcode.HCode(gameBonusRecordsNO)

	ErrCreateGameBonusRecords     = errcode.NewError(gameBonusRecordsBaseCode+1, "failed to create "+gameBonusRecordsName)
	ErrDeleteByIDGameBonusRecords = errcode.NewError(gameBonusRecordsBaseCode+2, "failed to delete "+gameBonusRecordsName)
	ErrUpdateByIDGameBonusRecords = errcode.NewError(gameBonusRecordsBaseCode+3, "failed to update "+gameBonusRecordsName)
	ErrGetByIDGameBonusRecords    = errcode.NewError(gameBonusRecordsBaseCode+4, "failed to get "+gameBonusRecordsName+" details")
	ErrListGameBonusRecords       = errcode.NewError(gameBonusRecordsBaseCode+5, "failed to list of "+gameBonusRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
