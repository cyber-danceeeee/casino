package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// gamePaymentChannels business-level http error codes.
// the gamePaymentChannelsNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	gamePaymentChannelsNO = 29
	gamePaymentChannelsName     = "gamePaymentChannels"
	gamePaymentChannelsBaseCode = errcode.HCode(gamePaymentChannelsNO)

	ErrCreateGamePaymentChannels     = errcode.NewError(gamePaymentChannelsBaseCode+1, "failed to create "+gamePaymentChannelsName)
	ErrDeleteByIDGamePaymentChannels = errcode.NewError(gamePaymentChannelsBaseCode+2, "failed to delete "+gamePaymentChannelsName)
	ErrUpdateByIDGamePaymentChannels = errcode.NewError(gamePaymentChannelsBaseCode+3, "failed to update "+gamePaymentChannelsName)
	ErrGetByIDGamePaymentChannels    = errcode.NewError(gamePaymentChannelsBaseCode+4, "failed to get "+gamePaymentChannelsName+" details")
	ErrListGamePaymentChannels       = errcode.NewError(gamePaymentChannelsBaseCode+5, "failed to list of "+gamePaymentChannelsName)

	// error codes are globally unique, adding 1 to the previous error code
)
