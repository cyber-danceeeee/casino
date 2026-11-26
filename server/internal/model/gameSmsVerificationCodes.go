package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameSmsVerificationCodes SMS verification codes
type GameSmsVerificationCodes struct {
	sgorm.Model `gorm:"embedded"` // embed id and time

	PhoneCountryCode string     `gorm:"column:phone_country_code;type:varchar(8);default:+92;not null" json:"phoneCountryCode"`
	Phone            string     `gorm:"column:phone;type:varchar(32);not null" json:"phone"`
	CodeHash         string     `gorm:"column:code_hash;type:varchar(255);not null" json:"codeHash"`
	Scene            string     `gorm:"column:scene;type:varchar(32);not null" json:"scene"`            // login, register, bind_phone, reset_password
	Status           int        `gorm:"column:status;type:tinyint(4);default:1;not null" json:"status"` // 1 unused, 2 used, 3 expired
	ExpiredAt        *time.Time `gorm:"column:expired_at;type:datetime(3);not null" json:"expiredAt"`
	UsedAt           *time.Time `gorm:"column:used_at;type:datetime(3)" json:"usedAt"`
}

// GameSmsVerificationCodesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var GameSmsVerificationCodesColumnNames = map[string]bool{
	"id":                 true,
	"created_at":         true,
	"updated_at":         true,
	"deleted_at":         true,
	"phone_country_code": true,
	"phone":              true,
	"code_hash":          true,
	"scene":              true,
	"status":             true,
	"expired_at":         true,
	"used_at":            true,
}
