package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameSystemConfigsRequest request params
type CreateGameSystemConfigsRequest struct {
	ConfigKey  string `json:"configKey" binding:""`
	ConfigValue  string `json:"configValue" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameSystemConfigsByIDRequest request params
type UpdateGameSystemConfigsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	ConfigKey  string `json:"configKey" binding:""`
	ConfigValue  string `json:"configValue" binding:""`
	Description  string `json:"description" binding:""`
	Status  int `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameSystemConfigsObjDetail detail
type GameSystemConfigsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	ConfigKey  string `json:"configKey"`
	ConfigValue  string `json:"configValue"`
	Description  string `json:"description"`
	Status  int `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameSystemConfigsReply only for api docs
type CreateGameSystemConfigsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSystemConfigsByIDReply only for api docs
type DeleteGameSystemConfigsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSystemConfigsByIDReply only for api docs
type UpdateGameSystemConfigsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSystemConfigsByIDReply only for api docs
type GetGameSystemConfigsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSystemConfigs GameSystemConfigsObjDetail `json:"gameSystemConfigs"`
	} `json:"data"` // return data
}

// ListGameSystemConfigssRequest request params
type ListGameSystemConfigssRequest struct {
	query.Params
}

// ListGameSystemConfigssReply only for api docs
type ListGameSystemConfigssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSystemConfigss []GameSystemConfigsObjDetail `json:"gameSystemConfigss"`
	} `json:"data"` // return data
}
