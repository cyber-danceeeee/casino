package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSupportQuestions business-level http error codes.
// the gameSupportQuestionsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSupportQuestionsNO       = 39
	gameSupportQuestionsName     = "gameSupportQuestions"
	gameSupportQuestionsBaseCode = errcode.HCode(gameSupportQuestionsNO)

	ErrCreateGameSupportQuestions     = errcode.NewError(gameSupportQuestionsBaseCode+1, "failed to create "+gameSupportQuestionsName)
	ErrDeleteByIDGameSupportQuestions = errcode.NewError(gameSupportQuestionsBaseCode+2, "failed to delete "+gameSupportQuestionsName)
	ErrUpdateByIDGameSupportQuestions = errcode.NewError(gameSupportQuestionsBaseCode+3, "failed to update "+gameSupportQuestionsName)
	ErrGetByIDGameSupportQuestions    = errcode.NewError(gameSupportQuestionsBaseCode+4, "failed to get "+gameSupportQuestionsName+" details")
	ErrListGameSupportQuestions       = errcode.NewError(gameSupportQuestionsBaseCode+5, "failed to list of "+gameSupportQuestionsName)

	// error codes are globally unique, adding 1 to the previous error code
)
