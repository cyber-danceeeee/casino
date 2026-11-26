package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameRedPacketClaimsRequest request params
type CreateGameRedPacketClaimsRequest struct {
	ClaimNo      string     `json:"claimNo" binding:""`
	UserID       uint64     `json:"userID" binding:""`
	RoundID      uint64     `json:"roundID" binding:""`
	RewardAmount string     `json:"rewardAmount" binding:""`
	Status       int        `json:"status" binding:""` // 1 pending, 2 claimed, 3 failed
	ClaimedAt    *time.Time `json:"claimedAt" binding:""`
}

// UpdateGameRedPacketClaimsByIDRequest request params
type UpdateGameRedPacketClaimsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ClaimNo      string     `json:"claimNo" binding:""`
	UserID       uint64     `json:"userID" binding:""`
	RoundID      uint64     `json:"roundID" binding:""`
	RewardAmount string     `json:"rewardAmount" binding:""`
	Status       int        `json:"status" binding:""` // 1 pending, 2 claimed, 3 failed
	ClaimedAt    *time.Time `json:"claimedAt" binding:""`
}

// GameRedPacketClaimsObjDetail detail
type GameRedPacketClaimsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ClaimNo      string     `json:"claimNo"`
	UserID       uint64     `json:"userID"`
	RoundID      uint64     `json:"roundID"`
	RewardAmount string     `json:"rewardAmount"`
	Status       int        `json:"status"` // 1 pending, 2 claimed, 3 failed
	ClaimedAt    *time.Time `json:"claimedAt"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// CreateGameRedPacketClaimsReply only for api docs
type CreateGameRedPacketClaimsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameRedPacketClaimsByIDReply only for api docs
type DeleteGameRedPacketClaimsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameRedPacketClaimsByIDReply only for api docs
type UpdateGameRedPacketClaimsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameRedPacketClaimsByIDReply only for api docs
type GetGameRedPacketClaimsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRedPacketClaims GameRedPacketClaimsObjDetail `json:"gameRedPacketClaims"`
	} `json:"data"` // return data
}

// ListGameRedPacketClaimssRequest request params
type ListGameRedPacketClaimssRequest struct {
	query.Params
}

// ListGameRedPacketClaimssReply only for api docs
type ListGameRedPacketClaimssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameRedPacketClaimss []GameRedPacketClaimsObjDetail `json:"gameRedPacketClaimss"`
	} `json:"data"` // return data
}
