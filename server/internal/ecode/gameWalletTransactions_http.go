package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameWalletTransactions business-level http error codes.
// the gameWalletTransactionsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameWalletTransactionsNO       = 55
	gameWalletTransactionsName     = "gameWalletTransactions"
	gameWalletTransactionsBaseCode = errcode.HCode(gameWalletTransactionsNO)

	ErrCreateGameWalletTransactions     = errcode.NewError(gameWalletTransactionsBaseCode+1, "failed to create "+gameWalletTransactionsName)
	ErrDeleteByIDGameWalletTransactions = errcode.NewError(gameWalletTransactionsBaseCode+2, "failed to delete "+gameWalletTransactionsName)
	ErrUpdateByIDGameWalletTransactions = errcode.NewError(gameWalletTransactionsBaseCode+3, "failed to update "+gameWalletTransactionsName)
	ErrGetByIDGameWalletTransactions    = errcode.NewError(gameWalletTransactionsBaseCode+4, "failed to get "+gameWalletTransactionsName+" details")
	ErrListGameWalletTransactions       = errcode.NewError(gameWalletTransactionsBaseCode+5, "failed to list of "+gameWalletTransactionsName)

	// error codes are globally unique, adding 1 to the previous error code
)
