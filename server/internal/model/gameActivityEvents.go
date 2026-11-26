package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/shopspring/decimal"
)

// GameActivityEvents Offer and activity list
type GameActivityEvents struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	SiteID          uint64           `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	ActivityCode    string           `gorm:"column:activity_code;type:varchar(64);not null" json:"activityCode"`
	Slug            string           `gorm:"column:slug;type:varchar(128);not null" json:"slug"`
	Kind            string           `gorm:"column:kind;type:varchar(32);not null" json:"kind"` // deposit, spin, chest, red_packet, checkin, card
	Title           string           `gorm:"column:title;type:varchar(128);not null" json:"title"`
	NameZh          string           `gorm:"column:name_zh;type:varchar(128);not null" json:"nameZh"`
	NameEn          string           `gorm:"column:name_en;type:varchar(128);not null" json:"nameEn"`
	Subtitle        string           `gorm:"column:subtitle;type:varchar(255);not null" json:"subtitle"`
	DescriptionZh   string           `gorm:"column:description_zh;type:text" json:"descriptionZh"`
	DescriptionEn   string           `gorm:"column:description_en;type:text" json:"descriptionEn"`
	PromotionType   string           `gorm:"column:promotion_type;type:varchar(48);not null" json:"promotionType"`
	TriggerType     string           `gorm:"column:trigger_type;type:varchar(48);not null" json:"triggerType"`
	ClaimMode       string           `gorm:"column:claim_mode;type:varchar(24);not null" json:"claimMode"`
	StackingGroup   string           `gorm:"column:stacking_group;type:varchar(64);not null" json:"stackingGroup"`
	Priority        int              `gorm:"column:priority;type:int;not null" json:"priority"`
	Countries       string           `gorm:"column:countries;type:varchar(500);not null" json:"countries"`
	Currencies      string           `gorm:"column:currencies;type:varchar(128);not null" json:"currencies"`
	Languages       string           `gorm:"column:languages;type:varchar(128);not null" json:"languages"`
	RewardText      string           `gorm:"column:reward_text;type:varchar(128);not null" json:"rewardText"`
	MaxRewardText   string           `gorm:"column:max_reward_text;type:varchar(128);not null" json:"maxRewardText"`
	Note            string           `gorm:"column:note;type:varchar(255);not null" json:"note"`
	CtaText         string           `gorm:"column:cta_text;type:varchar(64);not null" json:"ctaText"`
	ActionURL       string           `gorm:"column:action_url;type:varchar(500);not null" json:"actionURL"`
	HeroURL         string           `gorm:"column:hero_url;type:varchar(500);not null" json:"heroURL"`
	CardImageURL    string           `gorm:"column:card_image_url;type:varchar(500);not null" json:"cardImageURL"`
	Status          int              `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 enabled, 2 disabled, 3 ended
	StatusCode      string           `gorm:"column:status_code;type:varchar(32);not null" json:"statusCode"`
	TotalBudget     *decimal.Decimal `gorm:"column:total_budget;type:decimal(18,2);not null" json:"totalBudget"`
	PerUserLimit    uint             `gorm:"column:per_user_limit" json:"perUserLimit"`
	TotalClaimLimit uint             `gorm:"column:total_claim_limit" json:"totalClaimLimit"`
	ClaimCount      uint             `gorm:"column:claim_count" json:"claimCount"`
	RewardAmount    *decimal.Decimal `gorm:"column:reward_amount;type:decimal(18,2);not null" json:"rewardAmount"`
	Sort            int              `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"`
	StartAt         *time.Time       `gorm:"column:start_at;type:datetime(3)" json:"startAt"`
	EndAt           *time.Time       `gorm:"column:end_at;type:datetime(3)" json:"endAt"`
}

// GameActivityEventsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameActivityEventsColumnNames = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,

	"site_id":           true,
	"activity_code":     true,
	"slug":              true,
	"kind":              true,
	"title":             true,
	"name_zh":           true,
	"name_en":           true,
	"subtitle":          true,
	"description_zh":    true,
	"description_en":    true,
	"promotion_type":    true,
	"trigger_type":      true,
	"claim_mode":        true,
	"stacking_group":    true,
	"priority":          true,
	"countries":         true,
	"currencies":        true,
	"languages":         true,
	"reward_text":       true,
	"max_reward_text":   true,
	"note":              true,
	"cta_text":          true,
	"action_url":        true,
	"hero_url":          true,
	"card_image_url":    true,
	"status":            true,
	"status_code":       true,
	"total_budget":      true,
	"per_user_limit":    true,
	"total_claim_limit": true,
	"claim_count":       true,
	"reward_amount":     true,
	"sort":              true,
	"start_at":          true,
	"end_at":            true,
}
