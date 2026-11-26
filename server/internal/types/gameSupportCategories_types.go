package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameSupportCategoriesRequest request params
type CreateGameSupportCategoriesRequest struct {
	CategoryCode string `json:"categoryCode" binding:""`
	Name         string `json:"name" binding:""`
	Sort         int    `json:"sort" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameSupportCategoriesByIDRequest request params
type UpdateGameSupportCategoriesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	CategoryCode string `json:"categoryCode" binding:""`
	Name         string `json:"name" binding:""`
	Sort         int    `json:"sort" binding:""`
	Status       int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameSupportCategoriesObjDetail detail
type GameSupportCategoriesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	CategoryCode string     `json:"categoryCode"`
	Name         string     `json:"name"`
	Sort         int        `json:"sort"`
	Status       int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// CreateGameSupportCategoriesReply only for api docs
type CreateGameSupportCategoriesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSupportCategoriesByIDReply only for api docs
type DeleteGameSupportCategoriesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSupportCategoriesByIDReply only for api docs
type UpdateGameSupportCategoriesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSupportCategoriesByIDReply only for api docs
type GetGameSupportCategoriesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportCategories GameSupportCategoriesObjDetail `json:"gameSupportCategories"`
	} `json:"data"` // return data
}

// ListGameSupportCategoriessRequest request params
type ListGameSupportCategoriessRequest struct {
	query.Params
}

// ListGameSupportCategoriessReply only for api docs
type ListGameSupportCategoriessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportCategoriess []GameSupportCategoriesObjDetail `json:"gameSupportCategoriess"`
	} `json:"data"` // return data
}
