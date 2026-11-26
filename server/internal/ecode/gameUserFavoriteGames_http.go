package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserFavoriteGames business-level http error codes.
// the gameUserFavoriteGamesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserFavoriteGamesNO       = 45
	gameUserFavoriteGamesName     = "gameUserFavoriteGames"
	gameUserFavoriteGamesBaseCode = errcode.HCode(gameUserFavoriteGamesNO)

	ErrCreateGameUserFavoriteGames     = errcode.NewError(gameUserFavoriteGamesBaseCode+1, "failed to create "+gameUserFavoriteGamesName)
	ErrDeleteByIDGameUserFavoriteGames = errcode.NewError(gameUserFavoriteGamesBaseCode+2, "failed to delete "+gameUserFavoriteGamesName)
	ErrUpdateByIDGameUserFavoriteGames = errcode.NewError(gameUserFavoriteGamesBaseCode+3, "failed to update "+gameUserFavoriteGamesName)
	ErrGetByIDGameUserFavoriteGames    = errcode.NewError(gameUserFavoriteGamesBaseCode+4, "failed to get "+gameUserFavoriteGamesName+" details")
	ErrListGameUserFavoriteGames       = errcode.NewError(gameUserFavoriteGamesBaseCode+5, "failed to list of "+gameUserFavoriteGamesName)

	// error codes are globally unique, adding 1 to the previous error code
)
