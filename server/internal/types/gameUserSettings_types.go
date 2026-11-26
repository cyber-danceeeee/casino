package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameUserSettingsRequest request params
type CreateGameUserSettingsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	Language  string `json:"language" binding:""`
	MusicEnabled  int `json:"musicEnabled" binding:""`
	PushEnabled  int `json:"pushEnabled" binding:""`
}

// UpdateGameUserSettingsByIDRequest request params
type UpdateGameUserSettingsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	Language  string `json:"language" binding:""`
	MusicEnabled  int `json:"musicEnabled" binding:""`
	PushEnabled  int `json:"pushEnabled" binding:""`
}

// GameUserSettingsObjDetail detail
type GameUserSettingsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	Language  string `json:"language"`
	MusicEnabled  int `json:"musicEnabled"`
	PushEnabled  int `json:"pushEnabled"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameUserSettingsReply only for api docs
type CreateGameUserSettingsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserSettingsByIDReply only for api docs
type DeleteGameUserSettingsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserSettingsByIDReply only for api docs
type UpdateGameUserSettingsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserSettingsByIDReply only for api docs
type GetGameUserSettingsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserSettings GameUserSettingsObjDetail `json:"gameUserSettings"`
	} `json:"data"` // return data
}

// ListGameUserSettingssRequest request params
type ListGameUserSettingssRequest struct {
	query.Params
}

// ListGameUserSettingssReply only for api docs
type ListGameUserSettingssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserSettingss []GameUserSettingsObjDetail `json:"gameUserSettingss"`
	} `json:"data"` // return data
}
