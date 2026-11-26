package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameCashAuditRecords business-level http error codes.
// the gameCashAuditRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameCashAuditRecordsNO       = 13
	gameCashAuditRecordsName     = "gameCashAuditRecords"
	gameCashAuditRecordsBaseCode = errcode.HCode(gameCashAuditRecordsNO)

	ErrCreateGameCashAuditRecords     = errcode.NewError(gameCashAuditRecordsBaseCode+1, "failed to create "+gameCashAuditRecordsName)
	ErrDeleteByIDGameCashAuditRecords = errcode.NewError(gameCashAuditRecordsBaseCode+2, "failed to delete "+gameCashAuditRecordsName)
	ErrUpdateByIDGameCashAuditRecords = errcode.NewError(gameCashAuditRecordsBaseCode+3, "failed to update "+gameCashAuditRecordsName)
	ErrGetByIDGameCashAuditRecords    = errcode.NewError(gameCashAuditRecordsBaseCode+4, "failed to get "+gameCashAuditRecordsName+" details")
	ErrListGameCashAuditRecords       = errcode.NewError(gameCashAuditRecordsBaseCode+5, "failed to list of "+gameCashAuditRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
