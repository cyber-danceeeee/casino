package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSmsVerificationCodes business-level http error codes.
// the gameSmsVerificationCodesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSmsVerificationCodesNO = 36
	gameSmsVerificationCodesName     = "gameSmsVerificationCodes"
	gameSmsVerificationCodesBaseCode = errcode.HCode(gameSmsVerificationCodesNO)

	ErrCreateGameSmsVerificationCodes     = errcode.NewError(gameSmsVerificationCodesBaseCode+1, "failed to create "+gameSmsVerificationCodesName)
	ErrDeleteByIDGameSmsVerificationCodes = errcode.NewError(gameSmsVerificationCodesBaseCode+2, "failed to delete "+gameSmsVerificationCodesName)
	ErrUpdateByIDGameSmsVerificationCodes = errcode.NewError(gameSmsVerificationCodesBaseCode+3, "failed to update "+gameSmsVerificationCodesName)
	ErrGetByIDGameSmsVerificationCodes    = errcode.NewError(gameSmsVerificationCodesBaseCode+4, "failed to get "+gameSmsVerificationCodesName+" details")
	ErrListGameSmsVerificationCodes       = errcode.NewError(gameSmsVerificationCodesBaseCode+5, "failed to list of "+gameSmsVerificationCodesName)

	// error codes are globally unique, adding 1 to the previous error code
)
