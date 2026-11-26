package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameUserInvitationChests business-level http error codes.
// the gameUserInvitationChestsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameUserInvitationChestsNO = 46
	gameUserInvitationChestsName     = "gameUserInvitationChests"
	gameUserInvitationChestsBaseCode = errcode.HCode(gameUserInvitationChestsNO)

	ErrCreateGameUserInvitationChests     = errcode.NewError(gameUserInvitationChestsBaseCode+1, "failed to create "+gameUserInvitationChestsName)
	ErrDeleteByIDGameUserInvitationChests = errcode.NewError(gameUserInvitationChestsBaseCode+2, "failed to delete "+gameUserInvitationChestsName)
	ErrUpdateByIDGameUserInvitationChests = errcode.NewError(gameUserInvitationChestsBaseCode+3, "failed to update "+gameUserInvitationChestsName)
	ErrGetByIDGameUserInvitationChests    = errcode.NewError(gameUserInvitationChestsBaseCode+4, "failed to get "+gameUserInvitationChestsName+" details")
	ErrListGameUserInvitationChests       = errcode.NewError(gameUserInvitationChestsBaseCode+5, "failed to list of "+gameUserInvitationChestsName)

	// error codes are globally unique, adding 1 to the previous error code
)
