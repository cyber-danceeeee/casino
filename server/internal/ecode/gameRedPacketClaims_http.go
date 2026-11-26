package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gameRedPacketClaims business-level http error codes.
// the gameRedPacketClaimsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gameRedPacketClaimsNO = 34
	gameRedPacketClaimsName     = "gameRedPacketClaims"
	gameRedPacketClaimsBaseCode = errcode.HCode(gameRedPacketClaimsNO)

	ErrCreateGameRedPacketClaims     = errcode.NewError(gameRedPacketClaimsBaseCode+1, "failed to create "+gameRedPacketClaimsName)
	ErrDeleteByIDGameRedPacketClaims = errcode.NewError(gameRedPacketClaimsBaseCode+2, "failed to delete "+gameRedPacketClaimsName)
	ErrUpdateByIDGameRedPacketClaims = errcode.NewError(gameRedPacketClaimsBaseCode+3, "failed to update "+gameRedPacketClaimsName)
	ErrGetByIDGameRedPacketClaims    = errcode.NewError(gameRedPacketClaimsBaseCode+4, "failed to get "+gameRedPacketClaimsName+" details")
	ErrListGameRedPacketClaims       = errcode.NewError(gameRedPacketClaimsBaseCode+5, "failed to list of "+gameRedPacketClaimsName)

	// error codes are globally unique, adding 1 to the previous error code
)
