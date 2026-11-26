package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameBanners business-level http error codes.
// the gameBannersNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameBannersNO       = 9
	gameBannersName     = "gameBanners"
	gameBannersBaseCode = errcode.HCode(gameBannersNO)

	ErrCreateGameBanners     = errcode.NewError(gameBannersBaseCode+1, "failed to create "+gameBannersName)
	ErrDeleteByIDGameBanners = errcode.NewError(gameBannersBaseCode+2, "failed to delete "+gameBannersName)
	ErrUpdateByIDGameBanners = errcode.NewError(gameBannersBaseCode+3, "failed to update "+gameBannersName)
	ErrGetByIDGameBanners    = errcode.NewError(gameBannersBaseCode+4, "failed to get "+gameBannersName+" details")
	ErrListGameBanners       = errcode.NewError(gameBannersBaseCode+5, "failed to list of "+gameBannersName)

	// error codes are globally unique, adding 1 to the previous error code
)
