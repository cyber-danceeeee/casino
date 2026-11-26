package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameDepositOrders business-level http error codes.
// the gameDepositOrdersNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameDepositOrdersNO       = 15
	gameDepositOrdersName     = "gameDepositOrders"
	gameDepositOrdersBaseCode = errcode.HCode(gameDepositOrdersNO)

	ErrCreateGameDepositOrders     = errcode.NewError(gameDepositOrdersBaseCode+1, "failed to create "+gameDepositOrdersName)
	ErrDeleteByIDGameDepositOrders = errcode.NewError(gameDepositOrdersBaseCode+2, "failed to delete "+gameDepositOrdersName)
	ErrUpdateByIDGameDepositOrders = errcode.NewError(gameDepositOrdersBaseCode+3, "failed to update "+gameDepositOrdersName)
	ErrGetByIDGameDepositOrders    = errcode.NewError(gameDepositOrdersBaseCode+4, "failed to get "+gameDepositOrdersName+" details")
	ErrListGameDepositOrders       = errcode.NewError(gameDepositOrdersBaseCode+5, "failed to list of "+gameDepositOrdersName)

	// error codes are globally unique, adding 1 to the previous error code
)
