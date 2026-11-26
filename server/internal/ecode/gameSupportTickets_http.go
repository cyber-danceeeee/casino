package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameSupportTickets business-level http error codes.
// the gameSupportTicketsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameSupportTicketsNO       = 40
	gameSupportTicketsName     = "gameSupportTickets"
	gameSupportTicketsBaseCode = errcode.HCode(gameSupportTicketsNO)

	ErrCreateGameSupportTickets     = errcode.NewError(gameSupportTicketsBaseCode+1, "failed to create "+gameSupportTicketsName)
	ErrDeleteByIDGameSupportTickets = errcode.NewError(gameSupportTicketsBaseCode+2, "failed to delete "+gameSupportTicketsName)
	ErrUpdateByIDGameSupportTickets = errcode.NewError(gameSupportTicketsBaseCode+3, "failed to update "+gameSupportTicketsName)
	ErrGetByIDGameSupportTickets    = errcode.NewError(gameSupportTicketsBaseCode+4, "failed to get "+gameSupportTicketsName+" details")
	ErrListGameSupportTickets       = errcode.NewError(gameSupportTicketsBaseCode+5, "failed to list of "+gameSupportTicketsName)

	// error codes are globally unique, adding 1 to the previous error code
)
