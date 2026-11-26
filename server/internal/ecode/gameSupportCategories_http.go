package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSupportCategories business-level http error codes.
// the gameSupportCategoriesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSupportCategoriesNO       = 37
	gameSupportCategoriesName     = "gameSupportCategories"
	gameSupportCategoriesBaseCode = errcode.HCode(gameSupportCategoriesNO)

	ErrCreateGameSupportCategories     = errcode.NewError(gameSupportCategoriesBaseCode+1, "failed to create "+gameSupportCategoriesName)
	ErrDeleteByIDGameSupportCategories = errcode.NewError(gameSupportCategoriesBaseCode+2, "failed to delete "+gameSupportCategoriesName)
	ErrUpdateByIDGameSupportCategories = errcode.NewError(gameSupportCategoriesBaseCode+3, "failed to update "+gameSupportCategoriesName)
	ErrGetByIDGameSupportCategories    = errcode.NewError(gameSupportCategoriesBaseCode+4, "failed to get "+gameSupportCategoriesName+" details")
	ErrListGameSupportCategories       = errcode.NewError(gameSupportCategoriesBaseCode+5, "failed to list of "+gameSupportCategoriesName)

	// error codes are globally unique, adding 1 to the previous error code
)
