package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameGamesRequest request params
type CreateGameGamesRequest struct {
	GameCode         string `json:"gameCode" binding:""`
	ProviderID       uint64 `json:"providerID" binding:""`
	ProviderGameCode string `json:"providerGameCode" binding:""`
	Name             string `json:"name" binding:""`
	GameType         string `json:"gameType" binding:""` // slot, chance, sports, live, fish, table
	CoverURL         string `json:"coverURL" binding:""`
	Badge            string `json:"badge" binding:""`
	PlayerCount      uint   `json:"playerCount" binding:""`
	HotScore         int    `json:"hotScore" binding:""`
	LaunchURL        string `json:"launchURL" binding:""`
	Status           int    `json:"status" binding:""` // 1 online, 2 maintenance, 3 offline
	Sort             int    `json:"sort" binding:""`
}

// UpdateGameGamesByIDRequest request params
type UpdateGameGamesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	GameCode         string `json:"gameCode" binding:""`
	ProviderID       uint64 `json:"providerID" binding:""`
	ProviderGameCode string `json:"providerGameCode" binding:""`
	Name             string `json:"name" binding:""`
	GameType         string `json:"gameType" binding:""` // slot, chance, sports, live, fish, table
	CoverURL         string `json:"coverURL" binding:""`
	Badge            string `json:"badge" binding:""`
	PlayerCount      uint   `json:"playerCount" binding:""`
	HotScore         int    `json:"hotScore" binding:""`
	LaunchURL        string `json:"launchURL" binding:""`
	Status           int    `json:"status" binding:""` // 1 online, 2 maintenance, 3 offline
	Sort             int    `json:"sort" binding:""`
}

// GameGamesObjDetail detail
type GameGamesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	GameCode         string     `json:"gameCode"`
	ProviderID       uint64     `json:"providerID"`
	ProviderGameCode string     `json:"providerGameCode"`
	Name             string     `json:"name"`
	GameType         string     `json:"gameType"` // slot, chance, sports, live, fish, table
	CoverURL         string     `json:"coverURL"`
	Badge            string     `json:"badge"`
	PlayerCount      uint       `json:"playerCount"`
	HotScore         int        `json:"hotScore"`
	LaunchURL        string     `json:"launchURL"`
	Status           int        `json:"status"` // 1 online, 2 maintenance, 3 offline
	Sort             int        `json:"sort"`
	CreatedAt        *time.Time `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
}

// CreateGameGamesReply only for api docs
type CreateGameGamesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameGamesByIDReply only for api docs
type DeleteGameGamesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameGamesByIDReply only for api docs
type UpdateGameGamesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameGamesByIDReply only for api docs
type GetGameGamesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGames GameGamesObjDetail `json:"gameGames"`
	} `json:"data"` // return data
}

// ListGameGamessRequest request params
type ListGameGamessRequest struct {
	query.Params
}

// ListGameGamessReply only for api docs
type ListGameGamessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGamess []GameGamesObjDetail `json:"gameGamess"`
	} `json:"data"` // return data
}
