package agentadmin

import (
	"time"

	"casino/internal/content"
)

type ListQuery struct {
	Page, PageSize int
	Values         map[string]string
	Scope          content.Scope
}
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}
type Actor struct {
	Scope                                  content.Scope
	ID                                     uint64
	Username, IP, UserAgent, CorrelationID string
}
type Error struct{ Code, Message string }

func (e *Error) Error() string                   { return e.Message }
func NewError(code string, message string) error { return &Error{Code: code, Message: message} }

type ReasonCommand struct {
	Reason string `json:"reason" binding:"required,max=500"`
}
type StatusCommand struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason" binding:"required,max=500"`
}
type AgentInput struct {
	AgentCode        string `json:"agentCode" binding:"required,max=32"`
	Name             string `json:"name" binding:"required,max=128"`
	UserID           uint64 `json:"userId"`
	ParentAgentID    uint64 `json:"parentAgentId"`
	Country          string `json:"country" binding:"required,max=8"`
	Currency         string `json:"currency" binding:"required,max=8"`
	CommissionPlanID uint64 `json:"commissionPlanId"`
	Reason           string `json:"reason" binding:"required,max=500"`
}
type ChangeParentInput struct {
	ParentAgentID uint64 `json:"parentAgentId"`
	Reason        string `json:"reason" binding:"required,max=500"`
}
type AssignPlanInput struct {
	CommissionPlanID uint64 `json:"commissionPlanId" binding:"required"`
	Reason           string `json:"reason" binding:"required,max=500"`
}
type ChannelInput struct {
	Code               string            `json:"code" binding:"required,max=64"`
	NameZh             string            `json:"nameZh" binding:"required,max=128"`
	NameEn             string            `json:"nameEn" binding:"required,max=128"`
	AgentID            uint64            `json:"agentId" binding:"required"`
	Country            string            `json:"country" binding:"required,max=8"`
	Platform           string            `json:"platform" binding:"required,max=24"`
	RegistrationURL    string            `json:"registrationUrl" binding:"required,max=500"`
	TrackingParameters map[string]string `json:"trackingParameters"`
	Enabled            bool              `json:"enabled"`
	Reason             string            `json:"reason" binding:"required,max=500"`
}
type CommissionPlanInput struct {
	Code                    string              `json:"code" binding:"required,max=64"`
	NameZh                  string              `json:"nameZh" binding:"required,max=128"`
	NameEn                  string              `json:"nameEn" binding:"required,max=128"`
	Model                   string              `json:"model" binding:"required,oneof=CPA REVENUE_SHARE HYBRID TIERED_REVENUE_SHARE"`
	Currency                string              `json:"currency" binding:"required,max=8"`
	CPAAmount               string              `json:"cpaAmount" binding:"required"`
	RevenueShareRate        string              `json:"revenueShareRate" binding:"required"`
	NegativeCarryover       bool                `json:"negativeCarryover"`
	MinimumSettlementAmount string              `json:"minimumSettlementAmount" binding:"required"`
	SettlementCycle         string              `json:"settlementCycle" binding:"required,oneof=DAILY WEEKLY MONTHLY"`
	DeductionRules          map[string]string   `json:"deductionRules"`
	TierRules               []map[string]string `json:"tierRules"`
	EligibleProducts        []string            `json:"eligibleProducts"`
	Enabled                 bool                `json:"enabled"`
	Reason                  string              `json:"reason" binding:"required,max=500"`
}
type GenerateSettlementInput struct {
	AgentID        uint64    `json:"agentId" binding:"required"`
	PeriodStart    time.Time `json:"periodStart" binding:"required"`
	PeriodEnd      time.Time `json:"periodEnd" binding:"required"`
	Currency       string    `json:"currency" binding:"required,max=8"`
	IdempotencyKey string    `json:"idempotencyKey" binding:"required,max=128"`
	Reason         string    `json:"reason" binding:"required,max=500"`
}
type WalletAdjustmentInput struct {
	Type           string `json:"type" binding:"required,oneof=MANUAL_CREDIT MANUAL_DEBIT"`
	Amount         string `json:"amount" binding:"required"`
	IdempotencyKey string `json:"idempotencyKey" binding:"required,max=128"`
	Reason         string `json:"reason" binding:"required,max=500"`
}
type AttributionInput struct {
	UserID    uint64 `json:"userId" binding:"required"`
	AgentID   uint64 `json:"agentId" binding:"required"`
	ChannelID uint64 `json:"channelId"`
	Source    string `json:"source" binding:"required,max=32"`
	Reason    string `json:"reason" binding:"required,max=500"`
}
