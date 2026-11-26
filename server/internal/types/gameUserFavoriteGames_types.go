package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserFavoriteGamesRequest request params
type CreateGameUserFavoriteGamesRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	GameID  uint64 `json:"gameID" binding:""`
}

// UpdateGameUserFavoriteGamesByIDRequest request params
type UpdateGameUserFavoriteGamesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	GameID  uint64 `json:"gameID" binding:""`
}

// GameUserFavoriteGamesObjDetail detail
type GameUserFavoriteGamesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	GameID  uint64 `json:"gameID"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserFavoriteGamesReply only for api docs
type CreateGameUserFavoriteGamesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserFavoriteGamesByIDReply only for api docs
type DeleteGameUserFavoriteGamesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserFavoriteGamesByIDReply only for api docs
type UpdateGameUserFavoriteGamesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserFavoriteGamesByIDReply only for api docs
type GetGameUserFavoriteGamesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserFavoriteGames GameUserFavoriteGamesObjDetail `json:"gameUserFavoriteGames"`
	} `json:"data"` // return data
}

// ListGameUserFavoriteGamessRequest request params
type ListGameUserFavoriteGamessRequest struct {
	query.Params
}

// ListGameUserFavoriteGamessReply only for api docs
type ListGameUserFavoriteGamessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserFavoriteGamess []GameUserFavoriteGamesObjDetail `json:"gameUserFavoriteGamess"`
	} `json:"data"` // return data
}
