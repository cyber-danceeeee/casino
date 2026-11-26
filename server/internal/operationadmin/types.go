package operationadmin

import (
	"time"

	"casino/internal/content"
)

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

type Actor struct {
	ID            uint64
	Username      string
	Scope         content.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type ReasonCommand struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type BannerInput struct {
	TitleZh        string     `json:"titleZh" binding:"required,max=128"`
	TitleEn        string     `json:"titleEn" binding:"required,max=128"`
	ImageURL       string     `json:"imageUrl" binding:"required,max=500"`
	MobileImageURL string     `json:"mobileImageUrl" binding:"max=500"`
	TargetType     string     `json:"targetType" binding:"required"`
	TargetValue    string     `json:"targetValue" binding:"max=500"`
	Position       string     `json:"position" binding:"required,max=64"`
	Platform       string     `json:"platform" binding:"required,max=32"`
	Countries      string     `json:"countries" binding:"max=500"`
	Languages      string     `json:"languages" binding:"required,max=128"`
	SortOrder      int        `json:"sortOrder"`
	Enabled        bool       `json:"enabled"`
	StartTime      *time.Time `json:"startTime"`
	EndTime        *time.Time `json:"endTime"`
}

type AnnouncementInput struct {
	TitleZh   string     `json:"titleZh" binding:"required,max=128"`
	TitleEn   string     `json:"titleEn" binding:"required,max=128"`
	ContentZh string     `json:"contentZh" binding:"required"`
	ContentEn string     `json:"contentEn" binding:"required"`
	Type      string     `json:"type" binding:"required"`
	Priority  int        `json:"priority"`
	Countries string     `json:"countries" binding:"max=500"`
	Platforms string     `json:"platforms" binding:"max=128"`
	Languages string     `json:"languages" binding:"max=128"`
	Enabled   bool       `json:"enabled"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
}

type ParticipationRuleInput struct {
	RuleType           string `json:"ruleType" binding:"required"`
	ComparisonOperator string `json:"comparisonOperator" binding:"required"`
	NumericValue       string `json:"numericValue"`
	TextValue          string `json:"textValue" binding:"max=500"`
	Enabled            bool   `json:"enabled"`
	SortOrder          int    `json:"sortOrder"`
}

type RewardRuleInput struct {
	RewardType         string `json:"rewardType" binding:"required"`
	CalculationType    string `json:"calculationType" binding:"required"`
	RewardAmount       string `json:"rewardAmount" binding:"required"`
	PercentageRate     string `json:"percentageRate"`
	Currency           string `json:"currency" binding:"required,max=8"`
	WageringMultiplier string `json:"wageringMultiplier" binding:"required"`
	MinimumAmount      string `json:"minimumAmount" binding:"required"`
	MaximumReward      string `json:"maximumReward" binding:"required"`
	ExpirationDays     uint   `json:"expirationDays"`
	Enabled            bool   `json:"enabled"`
	SortOrder          int    `json:"sortOrder"`
}

type PromotionInput struct {
	Code               string                   `json:"code" binding:"required,max=64"`
	NameZh             string                   `json:"nameZh" binding:"required,max=128"`
	NameEn             string                   `json:"nameEn" binding:"required,max=128"`
	DescriptionZh      string                   `json:"descriptionZh"`
	DescriptionEn      string                   `json:"descriptionEn"`
	ActionURL          string                   `json:"actionUrl" binding:"max=500"`
	Type               string                   `json:"type" binding:"required"`
	TriggerType        string                   `json:"triggerType" binding:"required"`
	ClaimMode          string                   `json:"claimMode" binding:"required"`
	StackingGroup      string                   `json:"stackingGroup" binding:"max=64"`
	Priority           int                      `json:"priority"`
	Countries          string                   `json:"countries" binding:"max=500"`
	Currencies         string                   `json:"currencies" binding:"required,max=128"`
	Languages          string                   `json:"languages" binding:"required,max=128"`
	Status             string                   `json:"status" binding:"required"`
	TotalBudget        string                   `json:"totalBudget" binding:"required"`
	PerUserLimit       uint                     `json:"perUserLimit"`
	TotalClaimLimit    uint                     `json:"totalClaimLimit"`
	StartTime          *time.Time               `json:"startTime"`
	EndTime            *time.Time               `json:"endTime"`
	ParticipationRules []ParticipationRuleInput `json:"participationRules"`
	RewardRules        []RewardRuleInput        `json:"rewardRules"`
	EligibleGameIDs    []uint64                 `json:"eligibleGameIDs"`
}

type VIPBenefitInput struct {
	BenefitCode  string `json:"benefitCode" binding:"required,max=48"`
	NameZh       string `json:"nameZh" binding:"required,max=128"`
	NameEn       string `json:"nameEn" binding:"required,max=128"`
	BenefitValue string `json:"benefitValue" binding:"required"`
	ValueUnit    string `json:"valueUnit" binding:"required,max=24"`
	Enabled      bool   `json:"enabled"`
	SortOrder    int    `json:"sortOrder"`
}

type VIPLevelInput struct {
	Level              uint              `json:"level" binding:"required"`
	Code               string            `json:"code" binding:"required,max=32"`
	NameZh             string            `json:"nameZh" binding:"max=64"`
	NameEn             string            `json:"nameEn" binding:"max=64"`
	Icon               string            `json:"icon" binding:"max=500"`
	RequiredExperience uint64            `json:"requiredExperience"`
	RequiredDeposit    string            `json:"requiredDeposit" binding:"required"`
	RequiredValidBet   string            `json:"requiredValidBet" binding:"required"`
	WithdrawalLimit    string            `json:"withdrawalLimit" binding:"required"`
	CashbackRate       string            `json:"cashbackRate" binding:"required"`
	BirthdayReward     string            `json:"birthdayReward" binding:"required"`
	MonthlyReward      string            `json:"monthlyReward" binding:"required"`
	SortOrder          int               `json:"sortOrder"`
	Enabled            bool              `json:"enabled"`
	Benefits           []VIPBenefitInput `json:"benefits"`
}

type VIPAdjustCommand struct {
	Level  uint   `json:"level" binding:"required"`
	Reason string `json:"reason" binding:"required,max=500"`
}

type PromoCodeInput struct {
	Code            string     `json:"code" binding:"required,max=64"`
	PromotionID     uint64     `json:"promotionID"`
	RewardType      string     `json:"rewardType" binding:"required"`
	RewardValue     string     `json:"rewardValue" binding:"required"`
	Currency        string     `json:"currency" binding:"required,max=8"`
	TotalUsageLimit uint       `json:"totalUsageLimit"`
	PerUserLimit    uint       `json:"perUserLimit" binding:"required"`
	Countries       string     `json:"countries" binding:"max=500"`
	NewUsersOnly    bool       `json:"newUsersOnly"`
	Enabled         bool       `json:"enabled"`
	StartTime       *time.Time `json:"startTime"`
	ExpiresAt       *time.Time `json:"expiresAt"`
}

type RedeemPromoCodeCommand struct {
	Code           string `json:"code" binding:"required,max=64"`
	UserID         uint64 `json:"-"`
	IdempotencyKey string `json:"-"`
}

type NotificationInput struct {
	TitleZh       string     `json:"titleZh" binding:"required,max=128"`
	TitleEn       string     `json:"titleEn" binding:"required,max=128"`
	ContentZh     string     `json:"contentZh" binding:"required"`
	ContentEn     string     `json:"contentEn" binding:"required"`
	Channel       string     `json:"channel" binding:"required"`
	AudienceType  string     `json:"audienceType" binding:"required"`
	AudienceValue string     `json:"audienceValue" binding:"max=500"`
	ScheduledAt   *time.Time `json:"scheduledAt"`
}

type NotificationAudienceOption struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Phone    string `json:"phone,omitempty"`
	UserNo   string `json:"userNo,omitempty"`
	VIPLevel uint   `json:"vipLevel,omitempty"`
	NameZh   string `json:"nameZh,omitempty"`
	NameEn   string `json:"nameEn,omitempty"`
	TagType  string `json:"tagType,omitempty"`
}

type Error struct{ Code, Message string }

func (e *Error) Error() string                   { return e.Message }
func newError(code string, message string) error { return &Error{Code: code, Message: message} }
