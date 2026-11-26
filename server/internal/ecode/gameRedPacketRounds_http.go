package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameRedPacketRounds business-level http error codes.
// the gameRedPacketRoundsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameRedPacketRoundsNO       = 35
	gameRedPacketRoundsName     = "gameRedPacketRounds"
	gameRedPacketRoundsBaseCode = errcode.HCode(gameRedPacketRoundsNO)

	ErrCreateGameRedPacketRounds     = errcode.NewError(gameRedPacketRoundsBaseCode+1, "failed to create "+gameRedPacketRoundsName)
	ErrDeleteByIDGameRedPacketRounds = errcode.NewError(gameRedPacketRoundsBaseCode+2, "failed to delete "+gameRedPacketRoundsName)
	ErrUpdateByIDGameRedPacketRounds = errcode.NewError(gameRedPacketRoundsBaseCode+3, "failed to update "+gameRedPacketRoundsName)
	ErrGetByIDGameRedPacketRounds    = errcode.NewError(gameRedPacketRoundsBaseCode+4, "failed to get "+gameRedPacketRoundsName+" details")
	ErrListGameRedPacketRounds       = errcode.NewError(gameRedPacketRoundsBaseCode+5, "failed to list of "+gameRedPacketRoundsName)

	// error codes are globally unique, adding 1 to the previous error code
)
