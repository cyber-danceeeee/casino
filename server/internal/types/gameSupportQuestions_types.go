package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameSupportQuestionsRequest request params
type CreateGameSupportQuestionsRequest struct {
	CategoryID uint64 `json:"categoryID" binding:""`
	Question   string `json:"question" binding:""`
	AnswerHTML string `json:"answerHTML" binding:""`
	Sort       int    `json:"sort" binding:""`
	Status     int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// UpdateGameSupportQuestionsByIDRequest request params
type UpdateGameSupportQuestionsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	CategoryID uint64 `json:"categoryID" binding:""`
	Question   string `json:"question" binding:""`
	AnswerHTML string `json:"answerHTML" binding:""`
	Sort       int    `json:"sort" binding:""`
	Status     int    `json:"status" binding:""` // 1 enabled, 2 disabled
}

// GameSupportQuestionsObjDetail detail
type GameSupportQuestionsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	CategoryID uint64     `json:"categoryID"`
	Question   string     `json:"question"`
	AnswerHTML string     `json:"answerHTML"`
	Sort       int        `json:"sort"`
	Status     int        `json:"status"` // 1 enabled, 2 disabled
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}

// CreateGameSupportQuestionsReply only for api docs
type CreateGameSupportQuestionsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameSupportQuestionsByIDReply only for api docs
type DeleteGameSupportQuestionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameSupportQuestionsByIDReply only for api docs
type UpdateGameSupportQuestionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameSupportQuestionsByIDReply only for api docs
type GetGameSupportQuestionsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportQuestions GameSupportQuestionsObjDetail `json:"gameSupportQuestions"`
	} `json:"data"` // return data
}

// ListGameSupportQuestionssRequest request params
type ListGameSupportQuestionssRequest struct {
	query.Params
}

// ListGameSupportQuestionssReply only for api docs
type ListGameSupportQuestionssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameSupportQuestionss []GameSupportQuestionsObjDetail `json:"gameSupportQuestionss"`
	} `json:"data"` // return data
}
