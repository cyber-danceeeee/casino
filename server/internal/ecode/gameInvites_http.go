package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameInvites business-level http error codes.
// the gameInvitesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameInvitesNO       = 27
	gameInvitesName     = "gameInvites"
	gameInvitesBaseCode = errcode.HCode(gameInvitesNO)

	ErrCreateGameInvites     = errcode.NewError(gameInvitesBaseCode+1, "failed to create "+gameInvitesName)
	ErrDeleteByIDGameInvites = errcode.NewError(gameInvitesBaseCode+2, "failed to delete "+gameInvitesName)
	ErrUpdateByIDGameInvites = errcode.NewError(gameInvitesBaseCode+3, "failed to update "+gameInvitesName)
	ErrGetByIDGameInvites    = errcode.NewError(gameInvitesBaseCode+4, "failed to get "+gameInvitesName+" details")
	ErrListGameInvites       = errcode.NewError(gameInvitesBaseCode+5, "failed to list of "+gameInvitesName)

	// error codes are globally unique, adding 1 to the previous error code
)
