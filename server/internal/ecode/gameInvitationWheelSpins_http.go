package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameInvitationWheelSpins business-level http error codes.
// the gameInvitationWheelSpinsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameInvitationWheelSpinsNO = 26
	gameInvitationWheelSpinsName     = "gameInvitationWheelSpins"
	gameInvitationWheelSpinsBaseCode = errcode.HCode(gameInvitationWheelSpinsNO)

	ErrCreateGameInvitationWheelSpins     = errcode.NewError(gameInvitationWheelSpinsBaseCode+1, "failed to create "+gameInvitationWheelSpinsName)
	ErrDeleteByIDGameInvitationWheelSpins = errcode.NewError(gameInvitationWheelSpinsBaseCode+2, "failed to delete "+gameInvitationWheelSpinsName)
	ErrUpdateByIDGameInvitationWheelSpins = errcode.NewError(gameInvitationWheelSpinsBaseCode+3, "failed to update "+gameInvitationWheelSpinsName)
	ErrGetByIDGameInvitationWheelSpins    = errcode.NewError(gameInvitationWheelSpinsBaseCode+4, "failed to get "+gameInvitationWheelSpinsName+" details")
	ErrListGameInvitationWheelSpins       = errcode.NewError(gameInvitationWheelSpinsBaseCode+5, "failed to list of "+gameInvitationWheelSpinsName)

	// error codes are globally unique, adding 1 to the previous error code
)
