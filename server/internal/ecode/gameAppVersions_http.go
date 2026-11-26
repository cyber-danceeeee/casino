package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameAppVersions business-level http error codes.
// the gameAppVersionsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameAppVersionsNO       = 8
	gameAppVersionsName     = "gameAppVersions"
	gameAppVersionsBaseCode = errcode.HCode(gameAppVersionsNO)

	ErrCreateGameAppVersions     = errcode.NewError(gameAppVersionsBaseCode+1, "failed to create "+gameAppVersionsName)
	ErrDeleteByIDGameAppVersions = errcode.NewError(gameAppVersionsBaseCode+2, "failed to delete "+gameAppVersionsName)
	ErrUpdateByIDGameAppVersions = errcode.NewError(gameAppVersionsBaseCode+3, "failed to update "+gameAppVersionsName)
	ErrGetByIDGameAppVersions    = errcode.NewError(gameAppVersionsBaseCode+4, "failed to get "+gameAppVersionsName+" details")
	ErrListGameAppVersions       = errcode.NewError(gameAppVersionsBaseCode+5, "failed to list of "+gameAppVersionsName)

	// error codes are globally unique, adding 1 to the previous error code
)
