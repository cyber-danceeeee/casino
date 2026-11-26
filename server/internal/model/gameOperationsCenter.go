package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type GamePromotionParticipationRule struct {
	ID                 uint64           `gorm:"primaryKey" json:"id"`
	TenantID           uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID             uint64           `gorm:"column:site_id" json:"siteID"`
	PromotionID        uint64           `gorm:"column:promotion_id" json:"promotionID"`
	RuleType           string           `json:"ruleType"`
	ComparisonOperator string           `json:"comparisonOperator"`
	NumericValue       *decimal.Decimal `json:"numericValue"`
	TextValue          string           `json:"textValue"`
	Enabled            bool             `json:"enabled"`
	SortOrder          int              `json:"sortOrder"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          *time.Time       `json:"-"`
}

func (GamePromotionParticipationRule) TableName() string { return "game_promotion_participation_rules" }

type GamePromotionRewardRule struct {
	ID                 uint64           `gorm:"primaryKey" json:"id"`
	TenantID           uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID             uint64           `gorm:"column:site_id" json:"siteID"`
	PromotionID        uint64           `gorm:"column:promotion_id" json:"promotionID"`
	RewardType         string           `json:"rewardType"`
	CalculationType    string           `json:"calculationType"`
	RewardAmount       *decimal.Decimal `json:"rewardAmount"`
	PercentageRate     *decimal.Decimal `json:"percentageRate"`
	Currency           string           `json:"currency"`
	WageringMultiplier *decimal.Decimal `json:"wageringMultiplier"`
	MinimumAmount      *decimal.Decimal `json:"minimumAmount"`
	MaximumReward      *decimal.Decimal `json:"maximumReward"`
	ExpirationDays     uint             `json:"expirationDays"`
	Enabled            bool             `json:"enabled"`
	SortOrder          int              `json:"sortOrder"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          *time.Time       `json:"-"`
}

func (GamePromotionRewardRule) TableName() string { return "game_promotion_reward_rules" }

type GameBonusStatusHistory struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TenantID         uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID           uint64     `gorm:"column:site_id" json:"siteID"`
	BonusRecordID    uint64     `gorm:"column:bonus_record_id" json:"bonusRecordID"`
	BeforeStatus     string     `json:"beforeStatus"`
	AfterStatus      string     `json:"afterStatus"`
	OperatorID       uint64     `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string     `json:"operatorUsername"`
	Reason           string     `json:"reason"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
}

func (GameBonusStatusHistory) TableName() string { return "game_bonus_status_histories" }

type GameVIPBenefit struct {
	ID           uint64           `gorm:"primaryKey" json:"id"`
	TenantID     uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID       uint64           `gorm:"column:site_id" json:"siteID"`
	VIPLevelID   uint64           `gorm:"column:vip_level_id" json:"vipLevelID"`
	BenefitCode  string           `json:"benefitCode"`
	NameZh       string           `json:"nameZh"`
	NameEn       string           `json:"nameEn"`
	BenefitValue *decimal.Decimal `json:"benefitValue"`
	ValueUnit    string           `json:"valueUnit"`
	Enabled      bool             `json:"enabled"`
	SortOrder    int              `json:"sortOrder"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	DeletedAt    *time.Time       `json:"-"`
}

func (GameVIPBenefit) TableName() string { return "game_vip_benefits" }

type GameVIPAdjustmentLog struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TenantID         uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID           uint64     `gorm:"column:site_id" json:"siteID"`
	UserID           uint64     `gorm:"column:user_id" json:"userID"`
	BeforeLevel      uint       `json:"beforeLevel"`
	AfterLevel       uint       `json:"afterLevel"`
	OperatorID       uint64     `gorm:"column:operator_id" json:"operatorID"`
	OperatorUsername string     `json:"operatorUsername"`
	Reason           string     `json:"reason"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
}

func (GameVIPAdjustmentLog) TableName() string { return "game_vip_adjustment_logs" }

type GamePromoCode struct {
	ID              uint64           `gorm:"primaryKey" json:"id"`
	TenantID        uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID          uint64           `gorm:"column:site_id" json:"siteID"`
	Code            string           `json:"code"`
	PromotionID     uint64           `gorm:"column:promotion_id" json:"promotionID"`
	RewardType      string           `json:"rewardType"`
	RewardValue     *decimal.Decimal `json:"rewardValue"`
	Currency        string           `json:"currency"`
	TotalUsageLimit uint             `json:"totalUsageLimit"`
	UsedCount       uint             `json:"usedCount"`
	PerUserLimit    uint             `json:"perUserLimit"`
	Countries       string           `json:"countries"`
	NewUsersOnly    bool             `json:"newUsersOnly"`
	Enabled         bool             `json:"enabled"`
	StartAt         *time.Time       `json:"startAt"`
	ExpiresAt       *time.Time       `json:"expiresAt"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	DeletedAt       *time.Time       `json:"-"`
}

func (GamePromoCode) TableName() string { return "game_promo_codes" }

type GamePromoCodeRedemption struct {
	ID                  uint64           `gorm:"primaryKey" json:"id"`
	TenantID            uint64           `gorm:"column:tenant_id" json:"tenantID"`
	SiteID              uint64           `gorm:"column:site_id" json:"siteID"`
	PromoCodeID         uint64           `gorm:"column:promo_code_id" json:"promoCodeID"`
	UserID              uint64           `gorm:"column:user_id" json:"userID"`
	IdempotencyKey      string           `json:"idempotencyKey"`
	RewardType          string           `json:"rewardType"`
	RewardValue         *decimal.Decimal `json:"rewardValue"`
	Currency            string           `json:"currency"`
	WalletTransactionID uint64           `gorm:"column:wallet_transaction_id" json:"walletTransactionID"`
	Status              string           `json:"status"`
	CreatedAt           time.Time        `json:"createdAt"`
	UpdatedAt           time.Time        `json:"updatedAt"`
	DeletedAt           *time.Time       `json:"-"`
}

func (GamePromoCodeRedemption) TableName() string { return "game_promo_code_redemptions" }

type GameNotification struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	TenantID        uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID          uint64     `gorm:"column:site_id" json:"siteID"`
	NotificationID  string     `gorm:"column:notification_id" json:"notificationID"`
	TitleZh         string     `json:"titleZh"`
	TitleEn         string     `json:"titleEn"`
	ContentZh       string     `json:"contentZh"`
	ContentEn       string     `json:"contentEn"`
	Channel         string     `json:"channel"`
	AudienceType    string     `json:"audienceType"`
	AudienceValue   string     `json:"audienceValue"`
	Status          string     `json:"status"`
	TargetUserCount uint       `json:"targetUserCount"`
	SuccessCount    uint       `json:"successCount"`
	FailureCount    uint       `json:"failureCount"`
	ScheduledAt     *time.Time `json:"scheduledAt"`
	SentAt          *time.Time `json:"sentAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

func (GameNotification) TableName() string { return "game_notifications" }

type GameNotificationJob struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	TenantID       uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID         uint64     `gorm:"column:site_id" json:"siteID"`
	NotificationID uint64     `gorm:"column:notification_id" json:"notificationID"`
	JobType        string     `json:"jobType"`
	Status         string     `json:"status"`
	ScheduledAt    *time.Time `json:"scheduledAt"`
	Attempts       uint       `json:"attempts"`
	LastError      string     `json:"lastError"`
	IdempotencyKey string     `json:"idempotencyKey"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"-"`
}

func (GameNotificationJob) TableName() string { return "game_notification_jobs" }
