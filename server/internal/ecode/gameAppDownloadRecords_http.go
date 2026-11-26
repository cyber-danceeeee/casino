package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameAppDownloadRecords business-level http error codes.
// the gameAppDownloadRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameAppDownloadRecordsNO = 7
	gameAppDownloadRecordsName     = "gameAppDownloadRecords"
	gameAppDownloadRecordsBaseCode = errcode.HCode(gameAppDownloadRecordsNO)

	ErrCreateGameAppDownloadRecords     = errcode.NewError(gameAppDownloadRecordsBaseCode+1, "failed to create "+gameAppDownloadRecordsName)
	ErrDeleteByIDGameAppDownloadRecords = errcode.NewError(gameAppDownloadRecordsBaseCode+2, "failed to delete "+gameAppDownloadRecordsName)
	ErrUpdateByIDGameAppDownloadRecords = errcode.NewError(gameAppDownloadRecordsBaseCode+3, "failed to update "+gameAppDownloadRecordsName)
	ErrGetByIDGameAppDownloadRecords    = errcode.NewError(gameAppDownloadRecordsBaseCode+4, "failed to get "+gameAppDownloadRecordsName+" details")
	ErrListGameAppDownloadRecords       = errcode.NewError(gameAppDownloadRecordsBaseCode+5, "failed to list of "+gameAppDownloadRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
