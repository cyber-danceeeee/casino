package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameGameCategoriesRequest request params
type CreateGameGameCategoriesRequest struct {
	CategoryKey  string `json:"categoryKey" binding:""`
	Label  string `json:"label" binding:""`
	ShortLabel  string `json:"shortLabel" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameGameCategoriesByIDRequest request params
type UpdateGameGameCategoriesByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	CategoryKey  string `json:"categoryKey" binding:""`
	Label  string `json:"label" binding:""`
	ShortLabel  string `json:"shortLabel" binding:""`
	IconURL  string `json:"iconURL" binding:""`
	Sort  int `json:"sort" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameGameCategoriesObjDetail detail
type GameGameCategoriesObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	CategoryKey  string `json:"categoryKey"`
	Label  string `json:"label"`
	ShortLabel  string `json:"shortLabel"`
	IconURL  string `json:"iconURL"`
	Sort  int `json:"sort"`
	Status  int `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameGameCategoriesReply only for api docs
type CreateGameGameCategoriesReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameGameCategoriesByIDReply only for api docs
type DeleteGameGameCategoriesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameGameCategoriesByIDReply only for api docs
type UpdateGameGameCategoriesByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameGameCategoriesByIDReply only for api docs
type GetGameGameCategoriesByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameCategories GameGameCategoriesObjDetail `json:"gameGameCategories"`
	} `json:"data"` // return data
}

// ListGameGameCategoriessRequest request params
type ListGameGameCategoriessRequest struct {
	query.Params
}

// ListGameGameCategoriessReply only for api docs
type ListGameGameCategoriessReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameCategoriess []GameGameCategoriesObjDetail `json:"gameGameCategoriess"`
	} `json:"data"` // return data
}
