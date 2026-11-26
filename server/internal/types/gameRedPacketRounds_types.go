package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameRedPacketRoundsRequest request params
type CreateGameRedPacketRoundsRequest struct {
	RoundNo     string     `json:"roundNo" binding:""`
	ActivityID  uint64     `json:"activityID" binding:""`
	Title       string     `json:"title" binding:""`
	StartAt     *time.Time `json:"startAt" binding:""`
	EndAt       *time.Time `json:"endAt" binding:""`
	TotalAmount string     `json:"totalAmount" binding:""`
	PacketCount uint       `json:"packetCount" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 active, 3 ended
}

// UpdateGameRedPacketRoundsByIDRequest request params
type UpdateGameRedPacketRoundsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	RoundNo     string     `json:"roundNo" binding:""`
	ActivityID  uint64     `json:"activityID" binding:""`
	Title       string     `json:"title" binding:""`
	StartAt     *time.Time `json:"startAt" binding:""`
	EndAt       *time.Time `json:"endAt" binding:""`
	TotalAmount string     `json:"totalAmount" binding:""`
	PacketCount uint       `json:"packetCount" binding:""`
	Status      int        `json:"status" binding:""` // 1 pending, 2 active, 3 ended
}

// GameRedPacketRoundsObjDetail detail
type GameRedPacketRoundsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	RoundNo     string     `json:"roundNo"`
	ActivityID  uint64     `json:"activityID"`
	Title       string     `json:"title"`
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	TotalAmount string     `json:"totalAmount"`
	PacketCount uint       `json:"packetCount"`
	Status      int        `json:"status"` // 1 pending, 2 active, 3 ended
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// CreateGameRedPacketRoundsReply only for api docs
type CreateGameRedPacketRoundsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameRedPacketRoundsByIDReply only for api docs
type DeleteGameRedPacketRoundsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameRedPacketRoundsByIDReply only for api docs
type UpdateGameRedPacketRoundsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameRedPacketRoundsByIDReply only for api docs
type GetGameRedPacketRoundsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRedPacketRounds GameRedPacketRoundsObjDetail `json:"gameRedPacketRounds"`
	} `json:"data"` // return data
}

// ListGameRedPacketRoundssRequest request params
type ListGameRedPacketRoundssRequest struct {
	query.Params
}

// ListGameRedPacketRoundssReply only for api docs
type ListGameRedPacketRoundssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRedPacketRoundss []GameRedPacketRoundsObjDetail `json:"gameRedPacketRoundss"`
	} `json:"data"` // return data
}
