package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameActivityEventsRequest request params
type CreateGameActivityEventsRequest struct {
	ActivityCode  string     `json:"activityCode" binding:""`
	Slug          string     `json:"slug" binding:""`
	Kind          string     `json:"kind" binding:""` // deposit, spin, chest, red_packet, checkin, card
	Title         string     `json:"title" binding:""`
	Subtitle      string     `json:"subtitle" binding:""`
	RewardText    string     `json:"rewardText" binding:""`
	MaxRewardText string     `json:"maxRewardText" binding:""`
	Note          string     `json:"note" binding:""`
	CtaText       string     `json:"ctaText" binding:""`
	HeroURL       string     `json:"heroURL" binding:""`
	CardImageURL  string     `json:"cardImageURL" binding:""`
	Status        int        `json:"status" binding:""` // 1 enabled, 2 disabled, 3 ended
	Sort          int        `json:"sort" binding:""`
	StartAt       *time.Time `json:"startAt" binding:""`
	EndAt         *time.Time `json:"endAt" binding:""`
}

// UpdateGameActivityEventsByIDRequest request params
type UpdateGameActivityEventsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ActivityCode  string     `json:"activityCode" binding:""`
	Slug          string     `json:"slug" binding:""`
	Kind          string     `json:"kind" binding:""` // deposit, spin, chest, red_packet, checkin, card
	Title         string     `json:"title" binding:""`
	Subtitle      string     `json:"subtitle" binding:""`
	RewardText    string     `json:"rewardText" binding:""`
	MaxRewardText string     `json:"maxRewardText" binding:""`
	Note          string     `json:"note" binding:""`
	CtaText       string     `json:"ctaText" binding:""`
	HeroURL       string     `json:"heroURL" binding:""`
	CardImageURL  string     `json:"cardImageURL" binding:""`
	Status        int        `json:"status" binding:""` // 1 enabled, 2 disabled, 3 ended
	Sort          int        `json:"sort" binding:""`
	StartAt       *time.Time `json:"startAt" binding:""`
	EndAt         *time.Time `json:"endAt" binding:""`
}

// GameActivityEventsObjDetail detail
type GameActivityEventsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ActivityCode  string     `json:"activityCode"`
	Slug          string     `json:"slug"`
	Kind          string     `json:"kind"` // deposit, spin, chest, red_packet, checkin, card
	Title         string     `json:"title"`
	Subtitle      string     `json:"subtitle"`
	RewardText    string     `json:"rewardText"`
	MaxRewardText string     `json:"maxRewardText"`
	Note          string     `json:"note"`
	CtaText       string     `json:"ctaText"`
	HeroURL       string     `json:"heroURL"`
	CardImageURL  string     `json:"cardImageURL"`
	Status        int        `json:"status"` // 1 enabled, 2 disabled, 3 ended
	Sort          int        `json:"sort"`
	StartAt       *time.Time `json:"startAt"`
	EndAt         *time.Time `json:"endAt"`
	CreatedAt     *time.Time `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

// CreateGameActivityEventsReply only for api docs
type CreateGameActivityEventsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameActivityEventsByIDReply only for api docs
type DeleteGameActivityEventsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameActivityEventsByIDReply only for api docs
type UpdateGameActivityEventsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameActivityEventsByIDReply only for api docs
type GetGameActivityEventsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityEvents GameActivityEventsObjDetail `json:"gameActivityEvents"`
	} `json:"data"` // return data
}

// ListGameActivityEventssRequest request params
type ListGameActivityEventssRequest struct {
	query.Params
}

// ListGameActivityEventssReply only for api docs
type ListGameActivityEventssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameActivityEventss []GameActivityEventsObjDetail `json:"gameActivityEventss"`
	} `json:"data"` // return data
}
