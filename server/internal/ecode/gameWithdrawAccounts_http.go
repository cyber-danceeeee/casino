package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameWithdrawAccounts business-level http error codes.
// the gameWithdrawAccountsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameWithdrawAccountsNO       = 57
	gameWithdrawAccountsName     = "gameWithdrawAccounts"
	gameWithdrawAccountsBaseCode = errcode.HCode(gameWithdrawAccountsNO)

	ErrCreateGameWithdrawAccounts     = errcode.NewError(gameWithdrawAccountsBaseCode+1, "failed to create "+gameWithdrawAccountsName)
	ErrDeleteByIDGameWithdrawAccounts = errcode.NewError(gameWithdrawAccountsBaseCode+2, "failed to delete "+gameWithdrawAccountsName)
	ErrUpdateByIDGameWithdrawAccounts = errcode.NewError(gameWithdrawAccountsBaseCode+3, "failed to update "+gameWithdrawAccountsName)
	ErrGetByIDGameWithdrawAccounts    = errcode.NewError(gameWithdrawAccountsBaseCode+4, "failed to get "+gameWithdrawAccountsName+" details")
	ErrListGameWithdrawAccounts       = errcode.NewError(gameWithdrawAccountsBaseCode+5, "failed to list of "+gameWithdrawAccountsName)

	// error codes are globally unique, adding 1 to the previous error code
)
