package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameGameCategoryRelationsRequest request params
type CreateGameGameCategoryRelationsRequest struct {
	GameID  uint64 `json:"gameID" binding:""`
	CategoryID  uint64 `json:"categoryID" binding:""`
	Sort  int `json:"sort" binding:""`
}

// UpdateGameGameCategoryRelationsByIDRequest request params
type UpdateGameGameCategoryRelationsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	GameID  uint64 `json:"gameID" binding:""`
	CategoryID  uint64 `json:"categoryID" binding:""`
	Sort  int `json:"sort" binding:""`
}

// GameGameCategoryRelationsObjDetail detail
type GameGameCategoryRelationsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	GameID  uint64 `json:"gameID"`
	CategoryID  uint64 `json:"categoryID"`
	Sort  int `json:"sort"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameGameCategoryRelationsReply only for api docs
type CreateGameGameCategoryRelationsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameGameCategoryRelationsByIDReply only for api docs
type DeleteGameGameCategoryRelationsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameGameCategoryRelationsByIDReply only for api docs
type UpdateGameGameCategoryRelationsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameGameCategoryRelationsByIDReply only for api docs
type GetGameGameCategoryRelationsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameCategoryRelations GameGameCategoryRelationsObjDetail `json:"gameGameCategoryRelations"`
	} `json:"data"` // return data
}

// ListGameGameCategoryRelationssRequest request params
type ListGameGameCategoryRelationssRequest struct {
	query.Params
}

// ListGameGameCategoryRelationssReply only for api docs
type ListGameGameCategoryRelationssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameCategoryRelationss []GameGameCategoryRelationsObjDetail `json:"gameGameCategoryRelationss"`
	} `json:"data"` // return data
}
