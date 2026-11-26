package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gamePaymentMethods business-level http error codes.
// the gamePaymentMethodsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gamePaymentMethodsNO       = 30
	gamePaymentMethodsName     = "gamePaymentMethods"
	gamePaymentMethodsBaseCode = errcode.HCode(gamePaymentMethodsNO)

	ErrCreateGamePaymentMethods     = errcode.NewError(gamePaymentMethodsBaseCode+1, "failed to create "+gamePaymentMethodsName)
	ErrDeleteByIDGamePaymentMethods = errcode.NewError(gamePaymentMethodsBaseCode+2, "failed to delete "+gamePaymentMethodsName)
	ErrUpdateByIDGamePaymentMethods = errcode.NewError(gamePaymentMethodsBaseCode+3, "failed to update "+gamePaymentMethodsName)
	ErrGetByIDGamePaymentMethods    = errcode.NewError(gamePaymentMethodsBaseCode+4, "failed to get "+gamePaymentMethodsName+" details")
	ErrListGamePaymentMethods       = errcode.NewError(gamePaymentMethodsBaseCode+5, "failed to list of "+gamePaymentMethodsName)

	// error codes are globally unique, adding 1 to the previous error code
)
