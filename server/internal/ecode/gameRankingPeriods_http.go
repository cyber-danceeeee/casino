package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameRankingPeriods business-level http error codes.
// the gameRankingPeriodsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameRankingPeriodsNO = 31
	gameRankingPeriodsName     = "gameRankingPeriods"
	gameRankingPeriodsBaseCode = errcode.HCode(gameRankingPeriodsNO)

	ErrCreateGameRankingPeriods     = errcode.NewError(gameRankingPeriodsBaseCode+1, "failed to create "+gameRankingPeriodsName)
	ErrDeleteByIDGameRankingPeriods = errcode.NewError(gameRankingPeriodsBaseCode+2, "failed to delete "+gameRankingPeriodsName)
	ErrUpdateByIDGameRankingPeriods = errcode.NewError(gameRankingPeriodsBaseCode+3, "failed to update "+gameRankingPeriodsName)
	ErrGetByIDGameRankingPeriods    = errcode.NewError(gameRankingPeriodsBaseCode+4, "failed to get "+gameRankingPeriodsName+" details")
	ErrListGameRankingPeriods       = errcode.NewError(gameRankingPeriodsBaseCode+5, "failed to list of "+gameRankingPeriodsName)

	// error codes are globally unique, adding 1 to the previous error code
)
