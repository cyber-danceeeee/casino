package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameInvitationChestConfigs business-level http error codes.
// the gameInvitationChestConfigsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameInvitationChestConfigsNO = 23
	gameInvitationChestConfigsName     = "gameInvitationChestConfigs"
	gameInvitationChestConfigsBaseCode = errcode.HCode(gameInvitationChestConfigsNO)

	ErrCreateGameInvitationChestConfigs     = errcode.NewError(gameInvitationChestConfigsBaseCode+1, "failed to create "+gameInvitationChestConfigsName)
	ErrDeleteByIDGameInvitationChestConfigs = errcode.NewError(gameInvitationChestConfigsBaseCode+2, "failed to delete "+gameInvitationChestConfigsName)
	ErrUpdateByIDGameInvitationChestConfigs = errcode.NewError(gameInvitationChestConfigsBaseCode+3, "failed to update "+gameInvitationChestConfigsName)
	ErrGetByIDGameInvitationChestConfigs    = errcode.NewError(gameInvitationChestConfigsBaseCode+4, "failed to get "+gameInvitationChestConfigsName+" details")
	ErrListGameInvitationChestConfigs       = errcode.NewError(gameInvitationChestConfigsBaseCode+5, "failed to list of "+gameInvitationChestConfigsName)

	// error codes are globally unique, adding 1 to the previous error code
)
