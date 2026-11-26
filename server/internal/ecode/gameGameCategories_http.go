package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameGameCategories business-level http error codes.
// the gameGameCategoriesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameGameCategoriesNO       = 16
	gameGameCategoriesName     = "gameGameCategories"
	gameGameCategoriesBaseCode = errcode.HCode(gameGameCategoriesNO)

	ErrCreateGameGameCategories     = errcode.NewError(gameGameCategoriesBaseCode+1, "failed to create "+gameGameCategoriesName)
	ErrDeleteByIDGameGameCategories = errcode.NewError(gameGameCategoriesBaseCode+2, "failed to delete "+gameGameCategoriesName)
	ErrUpdateByIDGameGameCategories = errcode.NewError(gameGameCategoriesBaseCode+3, "failed to update "+gameGameCategoriesName)
	ErrGetByIDGameGameCategories    = errcode.NewError(gameGameCategoriesBaseCode+4, "failed to get "+gameGameCategoriesName+" details")
	ErrListGameGameCategories       = errcode.NewError(gameGameCategoriesBaseCode+5, "failed to list of "+gameGameCategoriesName)

	// error codes are globally unique, adding 1 to the previous error code
)
