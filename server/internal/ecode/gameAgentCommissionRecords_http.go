package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameAgentCommissionRecords business-level http error codes.
// the gameAgentCommissionRecordsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameAgentCommissionRecordsNO       = 5
	gameAgentCommissionRecordsName     = "gameAgentCommissionRecords"
	gameAgentCommissionRecordsBaseCode = errcode.HCode(gameAgentCommissionRecordsNO)

	ErrCreateGameAgentCommissionRecords     = errcode.NewError(gameAgentCommissionRecordsBaseCode+1, "failed to create "+gameAgentCommissionRecordsName)
	ErrDeleteByIDGameAgentCommissionRecords = errcode.NewError(gameAgentCommissionRecordsBaseCode+2, "failed to delete "+gameAgentCommissionRecordsName)
	ErrUpdateByIDGameAgentCommissionRecords = errcode.NewError(gameAgentCommissionRecordsBaseCode+3, "failed to update "+gameAgentCommissionRecordsName)
	ErrGetByIDGameAgentCommissionRecords    = errcode.NewError(gameAgentCommissionRecordsBaseCode+4, "failed to get "+gameAgentCommissionRecordsName+" details")
	ErrListGameAgentCommissionRecords       = errcode.NewError(gameAgentCommissionRecordsBaseCode+5, "failed to list of "+gameAgentCommissionRecordsName)

	// error codes are globally unique, adding 1 to the previous error code
)
