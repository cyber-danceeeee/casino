package systemadmin

import (
	"errors"
	"strings"
	"time"
)

type Query struct {
	Page     int
	PageSize int
	Values   map[string]string
}

type Result struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type Actor struct {
	ID            uint64
	SessionID     uint64
	Username      string
	TenantID      uint64
	SiteID        uint64
	Platform      bool
	IP            string
	UserAgent     string
	CorrelationID string
	Sensitive     bool
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string            { return e.Message }
func NewError(code, message string) error { return &Error{Code: code, Message: message} }
func ErrorCode(err error) string {
	var target *Error
	if errors.As(err, &target) {
		return target.Code
	}
	return "INTERNAL_ERROR"
}

type AdminInput struct {
	Username     string   `json:"username"`
	DisplayName  string   `json:"displayName"`
	Password     string   `json:"password"`
	TenantID     uint64   `json:"tenantID"`
	SiteID       uint64   `json:"siteID"`
	DepartmentID *uint64  `json:"departmentID"`
	Mobile       string   `json:"mobile"`
	Email        string   `json:"email"`
	RoleIDs      []uint64 `json:"roleIDs"`
	Status       *bool    `json:"status"`
}

type RoleInput struct {
	Code          string `json:"code"`
	NameZh        string `json:"nameZh"`
	NameEn        string `json:"nameEn"`
	DescriptionZh string `json:"descriptionZh"`
	DescriptionEn string `json:"descriptionEn"`
	DataScope     string `json:"dataScope"`
	Enabled       *bool  `json:"enabled"`
	Reason        string `json:"reason"`
}

type PermissionInput struct {
	Code           string `json:"code"`
	NameZh         string `json:"nameZh"`
	NameEn         string `json:"nameEn"`
	Module         string `json:"module"`
	PermissionType string `json:"permissionType"`
	APIPath        string `json:"apiPath"`
	Enabled        *bool  `json:"enabled"`
}

type DepartmentInput struct {
	ParentID  *uint64 `json:"parentID"`
	Code      string  `json:"code"`
	NameZh    string  `json:"nameZh"`
	NameEn    string  `json:"nameEn"`
	LeaderID  *uint64 `json:"leaderID"`
	SortOrder int     `json:"sortOrder"`
	Enabled   *bool   `json:"enabled"`
}

type MenuInput struct {
	ParentID       *uint64 `json:"parentID"`
	MenuType       string  `json:"menuType"`
	Code           string  `json:"code"`
	NameZh         string  `json:"nameZh"`
	NameEn         string  `json:"nameEn"`
	RoutePath      string  `json:"routePath"`
	ComponentPath  string  `json:"componentPath"`
	Icon           string  `json:"icon"`
	PermissionCode string  `json:"permissionCode"`
	SortOrder      int     `json:"sortOrder"`
	Visible        *bool   `json:"visible"`
	Enabled        *bool   `json:"enabled"`
	Cacheable      *bool   `json:"cacheable"`
	ExternalLink   string  `json:"externalLink"`
	Reason         string  `json:"reason"`
}

type ConfigInput struct {
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type CountryInput struct {
	CountryCode           string `json:"countryCode"`
	ISOCountryCode        string `json:"isoCountryCode"`
	CountryNameZh         string `json:"countryNameZh"`
	CountryNameEn         string `json:"countryNameEn"`
	CurrencyCode          string `json:"currencyCode"`
	CurrencyNameZh        string `json:"currencyNameZh"`
	CurrencyNameEn        string `json:"currencyNameEn"`
	CurrencySymbol        string `json:"currencySymbol"`
	CurrencyDecimalPlaces int    `json:"currencyDecimalPlaces"`
	Timezone              string `json:"timezone"`
	DefaultLanguage       string `json:"defaultLanguage"`
	SupportedLanguages    string `json:"supportedLanguages"`
	PhoneCountryCode      string `json:"phoneCountryCode"`
	PhoneLength           int    `json:"phoneLength"`
	Enabled               *bool  `json:"enabled"`
	LogoURL               string `json:"logoUrl"`
	FaviconURL            string `json:"faviconUrl"`
	LoginLogoURL          string `json:"loginLogoUrl"`
	HeaderLogoURL         string `json:"headerLogoUrl"`
	LoadingLogoURL        string `json:"loadingLogoUrl"`
	AppIconURL            string `json:"appIconUrl"`
	LoadingAnimationURL   string `json:"loadingAnimationUrl"`
	LoadingBackgroundURL  string `json:"loadingBackgroundUrl"`
	PrimaryColor          string `json:"primaryColor"`
	SecondaryColor        string `json:"secondaryColor"`
	ButtonColor           string `json:"buttonColor"`
	BackgroundColor       string `json:"backgroundColor"`
	TextColor             string `json:"textColor"`
	Reason                string `json:"reason"`
}

type TaskInput struct {
	TaskCode       string         `json:"taskCode"`
	NameZh         string         `json:"nameZh"`
	NameEn         string         `json:"nameEn"`
	TaskType       string         `json:"taskType"`
	HandlerCode    string         `json:"handlerCode"`
	CronExpression string         `json:"cronExpression"`
	Payload        map[string]any `json:"payload"`
	TimeoutSeconds uint           `json:"timeoutSeconds"`
	Enabled        *bool          `json:"enabled"`
}

type ReasonInput struct {
	Reason string `json:"reason"`
}

type IDsInput struct {
	IDs    []uint64 `json:"ids"`
	Reason string   `json:"reason"`
}

type LoginInput struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TenantCode string `json:"tenantCode"`
	PortalType string `json:"portalType"`
}

type MFAChallengeInput struct {
	ChallengeToken string `json:"challengeToken"`
	Code           string `json:"code"`
}

type AccountProfileInput struct {
	DisplayName       string `json:"displayName"`
	Email             string `json:"email"`
	Mobile            string `json:"mobile"`
	PreferredLanguage string `json:"preferredLanguage"`
	Timezone          string `json:"timezone"`
}

type ChangePasswordInput struct {
	CurrentPassword     string `json:"currentPassword"`
	NewPassword         string `json:"newPassword"`
	ConfirmPassword     string `json:"confirmPassword"`
	RevokeOtherSessions bool   `json:"revokeOtherSessions"`
}

type AccountMFAInput struct {
	CurrentPassword string `json:"currentPassword"`
	Code            string `json:"code"`
}

func value(q Query, key string) string { return strings.TrimSpace(q.Values[key]) }
func page(q Query) (int, int) {
	p, size := q.Page, q.PageSize
	if p < 1 {
		p = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return p, size
}
func boolValue(value bool) int {
	if value {
		return 1
	}
	return 0
}
func boolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
func now() time.Time { return time.Now() }
