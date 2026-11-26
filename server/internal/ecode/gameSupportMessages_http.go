package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSupportMessages business-level http error codes.
// the gameSupportMessagesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSupportMessagesNO = 38
	gameSupportMessagesName     = "gameSupportMessages"
	gameSupportMessagesBaseCode = errcode.HCode(gameSupportMessagesNO)

	ErrCreateGameSupportMessages     = errcode.NewError(gameSupportMessagesBaseCode+1, "failed to create "+gameSupportMessagesName)
	ErrDeleteByIDGameSupportMessages = errcode.NewError(gameSupportMessagesBaseCode+2, "failed to delete "+gameSupportMessagesName)
	ErrUpdateByIDGameSupportMessages = errcode.NewError(gameSupportMessagesBaseCode+3, "failed to update "+gameSupportMessagesName)
	ErrGetByIDGameSupportMessages    = errcode.NewError(gameSupportMessagesBaseCode+4, "failed to get "+gameSupportMessagesName+" details")
	ErrListGameSupportMessages       = errcode.NewError(gameSupportMessagesBaseCode+5, "failed to list of "+gameSupportMessagesName)

	// error codes are globally unique, adding 1 to the previous error code
)
