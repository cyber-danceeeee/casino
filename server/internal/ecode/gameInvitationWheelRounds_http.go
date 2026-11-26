package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameInvitationWheelRounds business-level http error codes.
// the gameInvitationWheelRoundsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameInvitationWheelRoundsNO       = 25
	gameInvitationWheelRoundsName     = "gameInvitationWheelRounds"
	gameInvitationWheelRoundsBaseCode = errcode.HCode(gameInvitationWheelRoundsNO)

	ErrCreateGameInvitationWheelRounds     = errcode.NewError(gameInvitationWheelRoundsBaseCode+1, "failed to create "+gameInvitationWheelRoundsName)
	ErrDeleteByIDGameInvitationWheelRounds = errcode.NewError(gameInvitationWheelRoundsBaseCode+2, "failed to delete "+gameInvitationWheelRoundsName)
	ErrUpdateByIDGameInvitationWheelRounds = errcode.NewError(gameInvitationWheelRoundsBaseCode+3, "failed to update "+gameInvitationWheelRoundsName)
	ErrGetByIDGameInvitationWheelRounds    = errcode.NewError(gameInvitationWheelRoundsBaseCode+4, "failed to get "+gameInvitationWheelRoundsName+" details")
	ErrListGameInvitationWheelRounds       = errcode.NewError(gameInvitationWheelRoundsBaseCode+5, "failed to list of "+gameInvitationWheelRoundsName)

	// error codes are globally unique, adding 1 to the previous error code
)
