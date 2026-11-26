package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameRankingRecords business-level http error codes.
// the gameRankingRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameRankingRecordsNO       = 32
	gameRankingRecordsName     = "gameRankingRecords"
	gameRankingRecordsBaseCode = errcode.HCode(gameRankingRecordsNO)

	ErrCreateGameRankingRecords     = errcode.NewError(gameRankingRecordsBaseCode+1, "failed to create "+gameRankingRecordsName)
	ErrDeleteByIDGameRankingRecords = errcode.NewError(gameRankingRecordsBaseCode+2, "failed to delete "+gameRankingRecordsName)
	ErrUpdateByIDGameRankingRecords = errcode.NewError(gameRankingRecordsBaseCode+3, "failed to update "+gameRankingRecordsName)
	ErrGetByIDGameRankingRecords    = errcode.NewError(gameRankingRecordsBaseCode+4, "failed to get "+gameRankingRecordsName+" details")
	ErrListGameRankingRecords       = errcode.NewError(gameRankingRecordsBaseCode+5, "failed to list of "+gameRankingRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
