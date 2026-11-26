package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

// GameUsers Members
type GameUsers struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	TenantID              uint64     `gorm:"column:tenant_id;type:bigint(20) unsigned;default:0;not null" json:"tenantID"`
	SiteID                uint64     `gorm:"column:site_id;type:bigint(20) unsigned;default:0;not null" json:"siteID"`
	UserNo                string     `gorm:"column:user_no;type:varchar(32);not null" json:"userNo"` // Public user id
	Username              string     `gorm:"column:username;type:varchar(64);not null" json:"username"`
	Email                 string     `gorm:"column:email;type:varchar(190);not null" json:"email"`
	PhoneCountryCode      string     `gorm:"column:phone_country_code;type:varchar(8);default:+92;not null" json:"phoneCountryCode"`
	Phone                 string     `gorm:"column:phone;type:varchar(32);not null" json:"phone"`
	PasswordHash          string     `gorm:"column:password_hash;type:varchar(255);not null" json:"passwordHash"`
	Nickname              string     `gorm:"column:nickname;type:varchar(64);not null" json:"nickname"`
	AvatarURL             string     `gorm:"column:avatar_url;type:varchar(500);not null" json:"avatarURL"`
	InviteCode            string     `gorm:"column:invite_code;type:varchar(32);not null" json:"inviteCode"`
	ReferrerUserID        uint64     `gorm:"column:referrer_user_id;type:bigint(20) unsigned;default:0;not null" json:"referrerUserID"`
	VipLevel              uint       `gorm:"column:vip_level;type:int(11) unsigned;default:0;not null" json:"vipLevel"`
	Status                int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 active, 2 disabled, 3 frozen, 4 locked, 5 closed
	CountryCode           string     `gorm:"column:country_code;type:varchar(8);not null" json:"countryCode"`
	Language              string     `gorm:"column:language;type:varchar(16);default:en;not null" json:"language"`
	RegistrationChannel   string     `gorm:"column:registration_channel;type:varchar(64);not null" json:"registrationChannel"`
	RegisteredIP          string     `gorm:"column:registered_ip;type:varchar(64);not null" json:"registeredIP"`
	LastLoginIP           string     `gorm:"column:last_login_ip;type:varchar(64);not null" json:"lastLoginIP"`
	LoginEnabled          int        `gorm:"column:login_enabled;type:tinyint(1);default:1;not null" json:"loginEnabled"`
	GameAccessEnabled     int        `gorm:"column:game_access_enabled;type:tinyint(1);default:1;not null" json:"gameAccessEnabled"`
	DepositEnabled        int        `gorm:"column:deposit_enabled;type:tinyint(1);default:1;not null" json:"depositEnabled"`
	WithdrawalEnabled     int        `gorm:"column:withdrawal_enabled;type:tinyint(1);default:1;not null" json:"withdrawalEnabled"`
	RiskReviewRequired    int        `gorm:"column:risk_review_required;type:tinyint(1);default:0;not null" json:"riskReviewRequired"`
	FailedLoginCount      uint       `gorm:"column:failed_login_count;type:int unsigned;default:0;not null" json:"failedLoginCount"`
	LastPasswordChangedAt *time.Time `gorm:"column:last_password_changed_at;type:datetime(3)" json:"lastPasswordChangedAt"`
	LastLoginAt           *time.Time `gorm:"column:last_login_at;type:datetime(3)" json:"lastLoginAt"`
}

// GameUsersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameUsersColumnNames = map[string]bool{
	"id":                       true,
	"created_at":               true,
	"updated_at":               true,
	"deleted_at":               true,
	"tenant_id":                true,
	"site_id":                  true,
	"user_no":                  true,
	"username":                 true,
	"email":                    true,
	"phone_country_code":       true,
	"phone":                    true,
	"password_hash":            true,
	"nickname":                 true,
	"avatar_url":               true,
	"invite_code":              true,
	"referrer_user_id":         true,
	"vip_level":                true,
	"status":                   true,
	"country_code":             true,
	"language":                 true,
	"registration_channel":     true,
	"registered_ip":            true,
	"last_login_ip":            true,
	"login_enabled":            true,
	"game_access_enabled":      true,
	"deposit_enabled":          true,
	"withdrawal_enabled":       true,
	"risk_review_required":     true,
	"failed_login_count":       true,
	"last_password_changed_at": true,
	"last_login_at":            true,
}
