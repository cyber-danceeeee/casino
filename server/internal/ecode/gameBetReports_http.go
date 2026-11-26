package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameBetReports business-level http error codes.
// the gameBetReportsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameBetReportsNO       = 11
	gameBetReportsName     = "gameBetReports"
	gameBetReportsBaseCode = errcode.HCode(gameBetReportsNO)

	ErrCreateGameBetReports     = errcode.NewError(gameBetReportsBaseCode+1, "failed to create "+gameBetReportsName)
	ErrDeleteByIDGameBetReports = errcode.NewError(gameBetReportsBaseCode+2, "failed to delete "+gameBetReportsName)
	ErrUpdateByIDGameBetReports = errcode.NewError(gameBetReportsBaseCode+3, "failed to update "+gameBetReportsName)
	ErrGetByIDGameBetReports    = errcode.NewError(gameBetReportsBaseCode+4, "failed to get "+gameBetReportsName+" details")
	ErrListGameBetReports       = errcode.NewError(gameBetReportsBaseCode+5, "failed to list of "+gameBetReportsName)

	// error codes are globally unique, adding 1 to the previous error code
)
