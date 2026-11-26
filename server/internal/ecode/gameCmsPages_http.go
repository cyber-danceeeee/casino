package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameCmsPages business-level http error codes.
// the gameCmsPagesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameCmsPagesNO       = 14
	gameCmsPagesName     = "gameCmsPages"
	gameCmsPagesBaseCode = errcode.HCode(gameCmsPagesNO)

	ErrCreateGameCmsPages     = errcode.NewError(gameCmsPagesBaseCode+1, "failed to create "+gameCmsPagesName)
	ErrDeleteByIDGameCmsPages = errcode.NewError(gameCmsPagesBaseCode+2, "failed to delete "+gameCmsPagesName)
	ErrUpdateByIDGameCmsPages = errcode.NewError(gameCmsPagesBaseCode+3, "failed to update "+gameCmsPagesName)
	ErrGetByIDGameCmsPages    = errcode.NewError(gameCmsPagesBaseCode+4, "failed to get "+gameCmsPagesName+" details")
	ErrListGameCmsPages       = errcode.NewError(gameCmsPagesBaseCode+5, "failed to list of "+gameCmsPagesName)

	// error codes are globally unique, adding 1 to the previous error code
)
