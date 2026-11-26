package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameGameProvidersRequest request params
type CreateGameGameProvidersRequest struct {
	ProviderCode  string `json:"providerCode" binding:""`
	Name  string `json:"name" binding:""`
	LogoURL  string `json:"logoURL" binding:""`
	ApiBaseURL  string `json:"apiBaseURL" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 maintenance, 3 disabled
	Sort  int `json:"sort" binding:""`
}

// UpdateGameGameProvidersByIDRequest request params
type UpdateGameGameProvidersByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ProviderCode  string `json:"providerCode" binding:""`
	Name  string `json:"name" binding:""`
	LogoURL  string `json:"logoURL" binding:""`
	ApiBaseURL  string `json:"apiBaseURL" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 maintenance, 3 disabled
	Sort  int `json:"sort" binding:""`
}

// GameGameProvidersObjDetail detail
type GameGameProvidersObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ProviderCode  string `json:"providerCode"`
	Name  string `json:"name"`
	LogoURL  string `json:"logoURL"`
	ApiBaseURL  string `json:"apiBaseURL"`
	Status  int `json:"status"` // 1 enabled, 2 maintenance, 3 disabled
	Sort  int `json:"sort"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameGameProvidersReply only for api docs
type CreateGameGameProvidersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameGameProvidersByIDReply only for api docs
type DeleteGameGameProvidersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameGameProvidersByIDReply only for api docs
type UpdateGameGameProvidersByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameGameProvidersByIDReply only for api docs
type GetGameGameProvidersByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameProviders GameGameProvidersObjDetail `json:"gameGameProviders"`
	} `json:"data"` // return data
}

// ListGameGameProviderssRequest request params
type ListGameGameProviderssRequest struct {
	query.Params
}

// ListGameGameProviderssReply only for api docs
type ListGameGameProviderssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameGameProviderss []GameGameProvidersObjDetail `json:"gameGameProviderss"`
	} `json:"data"` // return data
}
