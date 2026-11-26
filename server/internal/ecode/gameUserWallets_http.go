package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserWallets business-level http error codes.
// the gameUserWalletsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserWalletsNO = 50
	gameUserWalletsName     = "gameUserWallets"
	gameUserWalletsBaseCode = errcode.HCode(gameUserWalletsNO)

	ErrCreateGameUserWallets     = errcode.NewError(gameUserWalletsBaseCode+1, "failed to create "+gameUserWalletsName)
	ErrDeleteByIDGameUserWallets = errcode.NewError(gameUserWalletsBaseCode+2, "failed to delete "+gameUserWalletsName)
	ErrUpdateByIDGameUserWallets = errcode.NewError(gameUserWalletsBaseCode+3, "failed to update "+gameUserWalletsName)
	ErrGetByIDGameUserWallets    = errcode.NewError(gameUserWalletsBaseCode+4, "failed to get "+gameUserWalletsName+" details")
	ErrListGameUserWallets       = errcode.NewError(gameUserWalletsBaseCode+5, "failed to list of "+gameUserWalletsName)

	// error codes are globally unique, adding 1 to the previous error code
)
