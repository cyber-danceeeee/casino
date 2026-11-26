package riskadmin

import (
	"time"

	"casino/internal/tenancy"
)

type Actor struct {
	ID            uint64
	Username      string
	Scope         tenancy.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type ListQuery struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    tenancy.Scope
}

type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type ConditionInput struct {
	FieldName       string `json:"fieldName" binding:"required,max=64"`
	Operator        string `json:"operator" binding:"required,max=32"`
	ComparisonValue string `json:"comparisonValue" binding:"max=1000"`
	WindowSeconds   uint   `json:"windowSeconds"`
	SortOrder       int    `json:"sortOrder"`
}

type ActionInput struct {
	ActionType     string `json:"actionType" binding:"required,max=32"`
	ParameterValue string `json:"parameterValue" binding:"max=500"`
	SortOrder      int    `json:"sortOrder"`
}

type RuleInput struct {
	Code          string           `json:"code" binding:"required,max=64"`
	NameZh        string           `json:"nameZh" binding:"required,max=128"`
	NameEn        string           `json:"nameEn" binding:"required,max=128"`
	DescriptionZh string           `json:"descriptionZh" binding:"max=1000"`
	DescriptionEn string           `json:"descriptionEn" binding:"max=1000"`
	RuleType      string           `json:"ruleType" binding:"required"`
	RiskLevel     string           `json:"riskLevel" binding:"required"`
	Priority      int              `json:"priority"`
	Enabled       bool             `json:"enabled"`
	Countries     string           `json:"countries" binding:"max=500"`
	Currencies    string           `json:"currencies" binding:"max=128"`
	Channels      string           `json:"channels" binding:"max=500"`
	StartTime     *time.Time       `json:"startTime"`
	EndTime       *time.Time       `json:"endTime"`
	Conditions    []ConditionInput `json:"conditions" binding:"required,min=1,dive"`
	Actions       []ActionInput    `json:"actions" binding:"required,min=1,dive"`
}

type RuleStatusCommand struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason" binding:"required,max=500"`
}

type RuleTestCommand struct {
	Facts map[string]any `json:"facts" binding:"required"`
}

type AssignmentCommand struct {
	ReviewerID   uint64 `json:"reviewerID" binding:"required"`
	ReviewerName string `json:"reviewerName" binding:"required,max=64"`
	Reason       string `json:"reason" binding:"required,max=500"`
}

type EventActionCommand struct {
	Decision       string `json:"decision" binding:"required,max=32"`
	Reason         string `json:"reason" binding:"required,max=500"`
	InternalRemark string `json:"internalRemark" binding:"max=1000"`
	RiskTags       string `json:"riskTags" binding:"max=500"`
}

type NoteCommand struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type ReviewDecisionCommand struct {
	Decision       string `json:"decision" binding:"required,max=32"`
	Reason         string `json:"reason" binding:"required,max=500"`
	InternalRemark string `json:"internalRemark" binding:"required,max=1000"`
	RiskTags       string `json:"riskTags" binding:"required,max=500"`
}

type BlacklistInput struct {
	Type      string     `json:"type" binding:"required,max=32"`
	Value     string     `json:"value" binding:"required,max=500"`
	Reason    string     `json:"reason" binding:"required,max=500"`
	Enabled   bool       `json:"enabled"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type BlacklistUpdateCommand struct {
	Enabled   bool       `json:"enabled"`
	Reason    string     `json:"reason" binding:"required,max=500"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type WithdrawalDecisionCommand struct {
	Decision       string `json:"decision" binding:"required,max=32"`
	Reason         string `json:"reason" binding:"required,max=500"`
	InternalRemark string `json:"internalRemark" binding:"max=1000"`
	RiskTags       string `json:"riskTags" binding:"max=500"`
}

type Error struct{ Code, Message string }

func (e *Error) Error() string            { return e.Message }
func newError(code, message string) error { return &Error{Code: code, Message: message} }
