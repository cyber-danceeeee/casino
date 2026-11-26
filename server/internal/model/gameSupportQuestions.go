package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSupportQuestions FAQ questions and answers
type GameSupportQuestions struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	CategoryID uint64 `gorm:"column:category_id;type:bigint(20) unsigned;default:0;not null" json:"categoryID"`
	Question   string `gorm:"column:question;type:varchar(255);not null" json:"question"`
	AnswerHTML string `gorm:"column:answer_html;type:mediumtext" json:"answerHTML"`
	Sort       int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	Status     int    `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled
}

// GameSupportQuestionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSupportQuestionsColumnNames = map[string]bool{
	"id":          true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
	"category_id": true,
	"question":    true,
	"answer_html": true,
	"sort":        true,
	"status":      true,
}


