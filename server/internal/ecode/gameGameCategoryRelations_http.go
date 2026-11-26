package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameGameCategoryRelations business-level http error codes.
// the gameGameCategoryRelationsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameGameCategoryRelationsNO = 17
	gameGameCategoryRelationsName     = "gameGameCategoryRelations"
	gameGameCategoryRelationsBaseCode = errcode.HCode(gameGameCategoryRelationsNO)

	ErrCreateGameGameCategoryRelations     = errcode.NewError(gameGameCategoryRelationsBaseCode+1, "failed to create "+gameGameCategoryRelationsName)
	ErrDeleteByIDGameGameCategoryRelations = errcode.NewError(gameGameCategoryRelationsBaseCode+2, "failed to delete "+gameGameCategoryRelationsName)
	ErrUpdateByIDGameGameCategoryRelations = errcode.NewError(gameGameCategoryRelationsBaseCode+3, "failed to update "+gameGameCategoryRelationsName)
	ErrGetByIDGameGameCategoryRelations    = errcode.NewError(gameGameCategoryRelationsBaseCode+4, "failed to get "+gameGameCategoryRelationsName+" details")
	ErrListGameGameCategoryRelations       = errcode.NewError(gameGameCategoryRelationsBaseCode+5, "failed to list of "+gameGameCategoryRelationsName)

	// error codes are globally unique, adding 1 to the previous error code
)
