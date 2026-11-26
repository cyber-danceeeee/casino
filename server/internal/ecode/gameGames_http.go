package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameGames business-level http error codes.
// the gameGamesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameGamesNO       = 20
	gameGamesName     = "gameGames"
	gameGamesBaseCode = errcode.HCode(gameGamesNO)

	ErrCreateGameGames     = errcode.NewError(gameGamesBaseCode+1, "failed to create "+gameGamesName)
	ErrDeleteByIDGameGames = errcode.NewError(gameGamesBaseCode+2, "failed to delete "+gameGamesName)
	ErrUpdateByIDGameGames = errcode.NewError(gameGamesBaseCode+3, "failed to update "+gameGamesName)
	ErrGetByIDGameGames    = errcode.NewError(gameGamesBaseCode+4, "failed to get "+gameGamesName+" details")
	ErrListGameGames       = errcode.NewError(gameGamesBaseCode+5, "failed to list of "+gameGamesName)

	// error codes are globally unique, adding 1 to the previous error code
)
