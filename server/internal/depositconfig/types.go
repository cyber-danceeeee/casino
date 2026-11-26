package depositconfig

import (
	"time"

	"casino/internal/content"
)

type Actor struct {
	ID            uint64
	Username      string
	Scope         content.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type ListQuery struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    content.Scope
}

type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type MethodInput struct {
	MethodCode string `json:"methodCode" binding:"required,max=64"`
	MethodType string `json:"methodType" binding:"required"`
	Name       string `json:"name" binding:"required,max=128"`
	IconURL    string `json:"iconURL" binding:"max=500"`
	MinAmount  string `json:"minAmount" binding:"required"`
	MaxAmount  string `json:"maxAmount" binding:"required"`
	Sort       int    `json:"sort"`
	Status     int    `json:"status" binding:"required"`
}

type ChannelInput struct {
	ChannelCode  string `json:"channelCode" binding:"required,max=64"`
	Name         string `json:"name" binding:"required,max=128"`
	CountryCode  string `json:"countryCode" binding:"required,max=8"`
	CurrencyCode string `json:"currencyCode" binding:"required,max=8"`
	MethodID     uint64 `json:"methodId" binding:"required"`
	FeeRate      string `json:"feeRate" binding:"required"`
	MinAmount    string `json:"minAmount" binding:"required"`
	MaxAmount    string `json:"maxAmount" binding:"required"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status" binding:"required"`
}

type AmountOptionInput struct {
	OptionCode          string `json:"optionCode" binding:"required,max=64"`
	CurrencyCode        string `json:"currencyCode" binding:"max=8"`
	Amount              string `json:"amount" binding:"required"`
	DefaultBonusReward  string `json:"defaultBonusReward" binding:"required"`
	DefaultCashReward   string `json:"defaultCashReward" binding:"required"`
	DefaultRewardAmount string `json:"defaultRewardAmount" binding:"required"`
	Sort                int    `json:"sort"`
	Status              int    `json:"status" binding:"required"`
}

type VoucherInput struct {
	VoucherCode      string     `json:"voucherCode" binding:"required,max=64"`
	TitleZh          string     `json:"titleZh" binding:"required,max=128"`
	TitleEn          string     `json:"titleEn" binding:"required,max=128"`
	DescriptionZh    string     `json:"descriptionZh"`
	DescriptionEn    string     `json:"descriptionEn"`
	VoucherType      string     `json:"voucherType" binding:"required"`
	IconURL          string     `json:"iconURL" binding:"max=500"`
	MinAmount        string     `json:"minAmount" binding:"required"`
	MaxAmount        string     `json:"maxAmount" binding:"required"`
	MaxReward        string     `json:"maxReward" binding:"required"`
	Sort             int        `json:"sort"`
	Status           int        `json:"status" binding:"required"`
	StartTime        *time.Time `json:"startTime"`
	EndTime          *time.Time `json:"endTime"`
	NewUsersOnly     bool       `json:"newUsersOnly"`
	FirstDepositOnly bool       `json:"firstDepositOnly"`
	DailyLimit       uint       `json:"dailyLimit"`
	WeeklyLimit      uint       `json:"weeklyLimit"`
	PerUserLimit     uint       `json:"perUserLimit"`
	Countries        string     `json:"countries" binding:"max=500"`
	Currencies       string     `json:"currencies" binding:"max=128"`
	MethodCodes      string     `json:"methodCodes" binding:"max=500"`
}

type RewardTierInput struct {
	ID             uint64 `json:"id"`
	AmountOptionID uint64 `json:"amountOptionID"`
	Amount         string `json:"amount" binding:"required"`
	BonusReward    string `json:"bonusReward" binding:"required"`
	CashReward     string `json:"cashReward" binding:"required"`
	RewardAmount   string `json:"rewardAmount" binding:"required"`
	Sort           int    `json:"sort"`
	Status         int    `json:"status" binding:"required"`
}

type ConfigQuery struct {
	UserID     uint64
	Scope      content.Scope
	Country    string
	Currency   string
	MethodCode string
	Language   string
	Amount     string
}

type DomainError struct{ Code, Message string }

func (e *DomainError) Error() string { return e.Message }
