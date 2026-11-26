package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameWithdrawOrders business-level http error codes.
// the gameWithdrawOrdersNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameWithdrawOrdersNO       = 58
	gameWithdrawOrdersName     = "gameWithdrawOrders"
	gameWithdrawOrdersBaseCode = errcode.HCode(gameWithdrawOrdersNO)

	ErrCreateGameWithdrawOrders     = errcode.NewError(gameWithdrawOrdersBaseCode+1, "failed to create "+gameWithdrawOrdersName)
	ErrDeleteByIDGameWithdrawOrders = errcode.NewError(gameWithdrawOrdersBaseCode+2, "failed to delete "+gameWithdrawOrdersName)
	ErrUpdateByIDGameWithdrawOrders = errcode.NewError(gameWithdrawOrdersBaseCode+3, "failed to update "+gameWithdrawOrdersName)
	ErrGetByIDGameWithdrawOrders    = errcode.NewError(gameWithdrawOrdersBaseCode+4, "failed to get "+gameWithdrawOrdersName+" details")
	ErrListGameWithdrawOrders       = errcode.NewError(gameWithdrawOrdersBaseCode+5, "failed to list of "+gameWithdrawOrdersName)

	// error codes are globally unique, adding 1 to the previous error code
)
