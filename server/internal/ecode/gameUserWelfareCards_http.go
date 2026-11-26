package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserWelfareCards business-level http error codes.
// the gameUserWelfareCardsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserWelfareCardsNO       = 51
	gameUserWelfareCardsName     = "gameUserWelfareCards"
	gameUserWelfareCardsBaseCode = errcode.HCode(gameUserWelfareCardsNO)

	ErrCreateGameUserWelfareCards     = errcode.NewError(gameUserWelfareCardsBaseCode+1, "failed to create "+gameUserWelfareCardsName)
	ErrDeleteByIDGameUserWelfareCards = errcode.NewError(gameUserWelfareCardsBaseCode+2, "failed to delete "+gameUserWelfareCardsName)
	ErrUpdateByIDGameUserWelfareCards = errcode.NewError(gameUserWelfareCardsBaseCode+3, "failed to update "+gameUserWelfareCardsName)
	ErrGetByIDGameUserWelfareCards    = errcode.NewError(gameUserWelfareCardsBaseCode+4, "failed to get "+gameUserWelfareCardsName+" details")
	ErrListGameUserWelfareCards       = errcode.NewError(gameUserWelfareCardsBaseCode+5, "failed to list of "+gameUserWelfareCardsName)

	// error codes are globally unique, adding 1 to the previous error code
)
