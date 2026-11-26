package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameHomeFeatureCards business-level http error codes.
// the gameHomeFeatureCardsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameHomeFeatureCardsNO = 21
	gameHomeFeatureCardsName     = "gameHomeFeatureCards"
	gameHomeFeatureCardsBaseCode = errcode.HCode(gameHomeFeatureCardsNO)

	ErrCreateGameHomeFeatureCards     = errcode.NewError(gameHomeFeatureCardsBaseCode+1, "failed to create "+gameHomeFeatureCardsName)
	ErrDeleteByIDGameHomeFeatureCards = errcode.NewError(gameHomeFeatureCardsBaseCode+2, "failed to delete "+gameHomeFeatureCardsName)
	ErrUpdateByIDGameHomeFeatureCards = errcode.NewError(gameHomeFeatureCardsBaseCode+3, "failed to update "+gameHomeFeatureCardsName)
	ErrGetByIDGameHomeFeatureCards    = errcode.NewError(gameHomeFeatureCardsBaseCode+4, "failed to get "+gameHomeFeatureCardsName+" details")
	ErrListGameHomeFeatureCards       = errcode.NewError(gameHomeFeatureCardsBaseCode+5, "failed to list of "+gameHomeFeatureCardsName)

	// error codes are globally unique, adding 1 to the previous error code
)
