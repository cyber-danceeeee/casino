package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameWelfareCards business-level http error codes.
// the gameWelfareCardsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameWelfareCardsNO       = 56
	gameWelfareCardsName     = "gameWelfareCards"
	gameWelfareCardsBaseCode = errcode.HCode(gameWelfareCardsNO)

	ErrCreateGameWelfareCards     = errcode.NewError(gameWelfareCardsBaseCode+1, "failed to create "+gameWelfareCardsName)
	ErrDeleteByIDGameWelfareCards = errcode.NewError(gameWelfareCardsBaseCode+2, "failed to delete "+gameWelfareCardsName)
	ErrUpdateByIDGameWelfareCards = errcode.NewError(gameWelfareCardsBaseCode+3, "failed to update "+gameWelfareCardsName)
	ErrGetByIDGameWelfareCards    = errcode.NewError(gameWelfareCardsBaseCode+4, "failed to get "+gameWelfareCardsName+" details")
	ErrListGameWelfareCards       = errcode.NewError(gameWelfareCardsBaseCode+5, "failed to list of "+gameWelfareCardsName)

	// error codes are globally unique, adding 1 to the previous error code
)
