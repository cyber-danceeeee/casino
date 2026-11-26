package model

import "github.com/go-dev-frame/sponge/pkg/sgorm"

type GameUserRiskProfiles struct {
	sgorm.Model `gorm:"embedded"`

	UserID             uint64 `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	RiskLevel          string `gorm:"column:risk_level;type:varchar(16);not null" json:"riskLevel"`
	RiskStatus         string `gorm:"column:risk_status;type:varchar(32);not null" json:"riskStatus"`
	TriggerCount       uint   `gorm:"column:trigger_count;type:int unsigned;not null" json:"triggerCount"`
	AssociatedAccounts uint   `gorm:"column:associated_accounts;type:int unsigned;not null" json:"associatedAccounts"`
	AssignedReviewerID uint64 `gorm:"column:assigned_reviewer_id;type:bigint unsigned;not null" json:"assignedReviewerID"`
	AssignedReviewer   string `gorm:"column:assigned_reviewer;type:varchar(64);not null" json:"assignedReviewer"`
	LatestReason       string `gorm:"column:latest_reason;type:varchar(500);not null" json:"latestReason"`
}

func (GameUserRiskProfiles) TableName() string { return "game_user_risk_profiles" }
