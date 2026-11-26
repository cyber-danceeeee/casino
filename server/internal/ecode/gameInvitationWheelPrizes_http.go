package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameInvitationWheelPrizes business-level http error codes.
// the gameInvitationWheelPrizesNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameInvitationWheelPrizesNO       = 24
	gameInvitationWheelPrizesName     = "gameInvitationWheelPrizes"
	gameInvitationWheelPrizesBaseCode = errcode.HCode(gameInvitationWheelPrizesNO)

	ErrCreateGameInvitationWheelPrizes     = errcode.NewError(gameInvitationWheelPrizesBaseCode+1, "failed to create "+gameInvitationWheelPrizesName)
	ErrDeleteByIDGameInvitationWheelPrizes = errcode.NewError(gameInvitationWheelPrizesBaseCode+2, "failed to delete "+gameInvitationWheelPrizesName)
	ErrUpdateByIDGameInvitationWheelPrizes = errcode.NewError(gameInvitationWheelPrizesBaseCode+3, "failed to update "+gameInvitationWheelPrizesName)
	ErrGetByIDGameInvitationWheelPrizes    = errcode.NewError(gameInvitationWheelPrizesBaseCode+4, "failed to get "+gameInvitationWheelPrizesName+" details")
	ErrListGameInvitationWheelPrizes       = errcode.NewError(gameInvitationWheelPrizesBaseCode+5, "failed to list of "+gameInvitationWheelPrizesName)

	// error codes are globally unique, adding 1 to the previous error code
)
