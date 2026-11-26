package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameActivityAssets business-level http error codes.
// the gameActivityAssetsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameActivityAssetsNO       = 1
	gameActivityAssetsName     = "gameActivityAssets"
	gameActivityAssetsBaseCode = errcode.HCode(gameActivityAssetsNO)

	ErrCreateGameActivityAssets     = errcode.NewError(gameActivityAssetsBaseCode+1, "failed to create "+gameActivityAssetsName)
	ErrDeleteByIDGameActivityAssets = errcode.NewError(gameActivityAssetsBaseCode+2, "failed to delete "+gameActivityAssetsName)
	ErrUpdateByIDGameActivityAssets = errcode.NewError(gameActivityAssetsBaseCode+3, "failed to update "+gameActivityAssetsName)
	ErrGetByIDGameActivityAssets    = errcode.NewError(gameActivityAssetsBaseCode+4, "failed to get "+gameActivityAssetsName+" details")
	ErrListGameActivityAssets       = errcode.NewError(gameActivityAssetsBaseCode+5, "failed to list of "+gameActivityAssetsName)

	// error codes are globally unique, adding 1 to the previous error code
)
