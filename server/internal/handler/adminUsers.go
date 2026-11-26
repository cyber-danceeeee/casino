package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/adminauth"
	"casino/internal/content"
	"casino/internal/database"
)

type AdminUsersHandler interface {
	Me(c *gin.Context)
	ListUsers(c *gin.Context)
	GetUser(c *gin.Context)
	ListUserNotes(c *gin.Context)
	AddUserNote(c *gin.Context)
	UpdateUserTags(c *gin.Context)
	ListUserOperationLogs(c *gin.Context)
	ListLoginRecords(c *gin.Context)
	ListDevices(c *gin.Context)
	ListPushDevices(c *gin.Context)
	GetDevice(c *gin.Context)
	ListDeviceUsers(c *gin.Context)
	ListRiskUsers(c *gin.Context)
	GetRiskUser(c *gin.Context)
	UpdateUserStatus(c *gin.Context)
	UpdateUserRestrictions(c *gin.Context)
	UpdateUserRisk(c *gin.Context)
	UpdateDeviceStatus(c *gin.Context)
}

type adminUsersHandler struct{}

func NewAdminUsersHandler() AdminUsersHandler {
	return &adminUsersHandler{}
}

type adminPageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type adminUserRow struct {
	ID                  uint64     `json:"id" gorm:"column:id"`
	UserNo              string     `json:"userNo" gorm:"column:user_no"`
	Username            string     `json:"username" gorm:"column:username"`
	Email               string     `json:"email" gorm:"column:email"`
	PhoneCountryCode    string     `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone               string     `json:"phone" gorm:"column:phone"`
	Nickname            string     `json:"nickname" gorm:"column:nickname"`
	VipLevel            uint       `json:"vipLevel" gorm:"column:vip_level"`
	Status              int        `json:"status" gorm:"column:status"`
	StatusText          string     `json:"statusText" gorm:"-"`
	CountryCode         string     `json:"countryCode" gorm:"column:country_code"`
	Language            string     `json:"language" gorm:"column:language"`
	RegistrationChannel string     `json:"registrationChannel" gorm:"column:registration_channel"`
	RegisteredIP        string     `json:"registeredIP" gorm:"column:registered_ip"`
	LastLoginIP         string     `json:"lastLoginIP" gorm:"column:last_login_ip"`
	KYCStatus           string     `json:"kycStatus" gorm:"-"`
	RiskLevel           string     `json:"riskLevel" gorm:"column:risk_level"`
	RiskStatus          string     `json:"riskStatus" gorm:"column:risk_status"`
	LoginEnabled        int        `json:"loginEnabled" gorm:"column:login_enabled"`
	GameAccessEnabled   int        `json:"gameAccessEnabled" gorm:"column:game_access_enabled"`
	LastLoginAtValue    *time.Time `json:"-" gorm:"column:last_login_at"`
	RegisteredAtValue   time.Time  `json:"-" gorm:"column:created_at"`
	UpdatedAtValue      time.Time  `json:"-" gorm:"column:updated_at"`
	LastLoginAt         string     `json:"lastLoginAt" gorm:"-"`
	RegisteredAt        string     `json:"registeredAt" gorm:"-"`
	UpdatedAt           string     `json:"updatedAt" gorm:"-"`
	Currency            string     `json:"currency" gorm:"column:currency"`
	CashBalance         string     `json:"cashBalance" gorm:"column:cash_balance"`
	BonusBalance        string     `json:"bonusBalance" gorm:"column:bonus_balance"`
	WithdrawableBalance string     `json:"withdrawableBalance" gorm:"column:withdrawable_balance"`
	FrozenBalance       string     `json:"frozenBalance" gorm:"column:frozen_balance"`
}

type adminLoginRecordRow struct {
	ID               uint64    `json:"id" gorm:"column:id"`
	UserID           uint64    `json:"userID" gorm:"column:user_id"`
	UserNo           string    `json:"userNo" gorm:"column:user_no"`
	PhoneCountryCode string    `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone            string    `json:"phone" gorm:"column:phone"`
	Nickname         string    `json:"nickname" gorm:"column:nickname"`
	Username         string    `json:"username" gorm:"column:username"`
	Success          int       `json:"success" gorm:"column:success"`
	LoginResult      string    `json:"loginResult" gorm:"-"`
	FailureReason    string    `json:"failureReason" gorm:"column:failure_reason"`
	DeviceID         string    `json:"deviceID" gorm:"column:device_id"`
	ClientIP         string    `json:"clientIP" gorm:"column:ip_address"`
	Country          string    `json:"country" gorm:"column:country"`
	Region           string    `json:"region" gorm:"column:region"`
	DeviceType       string    `json:"deviceType" gorm:"column:device_type"`
	Platform         string    `json:"platform" gorm:"column:platform"`
	OS               string    `json:"os" gorm:"column:os"`
	Browser          string    `json:"browser" gorm:"column:browser"`
	AppVersion       string    `json:"appVersion" gorm:"column:app_version"`
	UserAgent        string    `json:"userAgent" gorm:"column:user_agent"`
	LoginAtValue     time.Time `json:"-" gorm:"column:occurred_at"`
	LoginAt          string    `json:"loginAt" gorm:"-"`
}

type adminDeviceRow struct {
	ID                 uint64     `json:"id" gorm:"column:id"`
	UserID             uint64     `json:"userID" gorm:"column:user_id"`
	UserNo             string     `json:"userNo" gorm:"column:user_no"`
	PhoneCountryCode   string     `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone              string     `json:"phone" gorm:"column:phone"`
	Nickname           string     `json:"nickname" gorm:"column:nickname"`
	DeviceID           string     `json:"deviceID" gorm:"column:device_id"`
	Platform           string     `json:"platform" gorm:"column:platform"`
	DeviceModel        string     `json:"deviceModel" gorm:"column:device_model"`
	OS                 string     `json:"os" gorm:"column:os"`
	Browser            string     `json:"browser" gorm:"column:browser"`
	AppVersion         string     `json:"appVersion" gorm:"column:app_version"`
	FingerprintHash    string     `json:"fingerprintHash" gorm:"column:fingerprint_hash"`
	FirstIP            string     `json:"firstIP" gorm:"column:first_ip"`
	LastIP             string     `json:"lastIP" gorm:"column:last_ip"`
	ClientIP           string     `json:"clientIP" gorm:"column:client_ip"`
	UserAgent          string     `json:"userAgent" gorm:"column:user_agent"`
	Status             int        `json:"status" gorm:"column:status"`
	StatusText         string     `json:"statusText" gorm:"-"`
	TrustStatus        int        `json:"trustStatus" gorm:"column:trust_status"`
	TrustStatusText    string     `json:"trustStatusText" gorm:"-"`
	RiskStatus         int        `json:"riskStatus" gorm:"column:risk_status"`
	RiskStatusText     string     `json:"riskStatusText" gorm:"-"`
	LoginCount         uint       `json:"loginCount" gorm:"column:login_count"`
	AssociatedAccounts uint       `json:"associatedAccounts" gorm:"column:associated_accounts"`
	FirstSeenValue     *time.Time `json:"-" gorm:"column:first_seen_at"`
	LastSeenValue      *time.Time `json:"-" gorm:"column:last_seen_at"`
	FirstSeenAt        string     `json:"firstSeenAt" gorm:"-"`
	LastSeenAt         string     `json:"lastSeenAt" gorm:"-"`
}

type adminPushDeviceRow struct {
	ID                     uint64     `json:"id" gorm:"column:id"`
	UserID                 uint64     `json:"userID" gorm:"column:user_id"`
	UserNo                 string     `json:"userNo" gorm:"column:user_no"`
	PhoneCountryCode       string     `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone                  string     `json:"phone" gorm:"column:phone"`
	Nickname               string     `json:"nickname" gorm:"column:nickname"`
	Platform               string     `json:"platform" gorm:"column:platform"`
	DeviceID               string     `json:"deviceID" gorm:"column:device_id"`
	FingerprintHash        string     `json:"fingerprintHash" gorm:"column:fingerprint_hash"`
	Manufacturer           string     `json:"manufacturer" gorm:"column:manufacturer"`
	Brand                  string     `json:"brand" gorm:"column:brand"`
	DeviceModel            string     `json:"deviceModel" gorm:"column:device_model"`
	DeviceName             string     `json:"deviceName" gorm:"column:device_name"`
	OSVersion              string     `json:"osVersion" gorm:"column:os_version"`
	SDKVersion             uint       `json:"sdkVersion" gorm:"column:sdk_version"`
	SecurityPatch          string     `json:"securityPatch" gorm:"column:security_patch"`
	NetworkType            string     `json:"networkType" gorm:"column:network_type"`
	NetworkMetered         bool       `json:"networkMetered" gorm:"column:network_metered"`
	NetworkRoaming         bool       `json:"networkRoaming" gorm:"column:network_roaming"`
	CarrierName            string     `json:"carrierName" gorm:"column:carrier_name"`
	ScreenWidth            uint       `json:"screenWidth" gorm:"column:screen_width"`
	ScreenHeight           uint       `json:"screenHeight" gorm:"column:screen_height"`
	ScreenDensity          float64    `json:"screenDensity" gorm:"column:screen_density"`
	Timezone               string     `json:"timezone" gorm:"column:timezone"`
	Locale                 string     `json:"locale" gorm:"column:locale"`
	BatteryLevel           int        `json:"batteryLevel" gorm:"column:battery_level"`
	Charging               bool       `json:"charging" gorm:"column:charging"`
	NotificationPermission string     `json:"notificationPermission" gorm:"column:notification_permission"`
	AppVersion             string     `json:"appVersion" gorm:"column:app_version"`
	AppVersionCode         uint64     `json:"appVersionCode" gorm:"column:app_version_code"`
	Language               string     `json:"language" gorm:"column:language"`
	LastIP                 string     `json:"lastIP" gorm:"column:last_ip"`
	UserAgent              string     `json:"userAgent" gorm:"column:user_agent"`
	Enabled                bool       `json:"enabled" gorm:"column:enabled"`
	Status                 string     `json:"status" gorm:"-"`
	LastSeenValue          *time.Time `json:"-" gorm:"column:last_seen_at"`
	InvalidatedValue       *time.Time `json:"-" gorm:"column:invalidated_at"`
	CreatedAtValue         time.Time  `json:"-" gorm:"column:created_at"`
	UpdatedAtValue         time.Time  `json:"-" gorm:"column:updated_at"`
	LastSeenAt             string     `json:"lastSeenAt" gorm:"-"`
	InvalidatedAt          string     `json:"invalidatedAt" gorm:"-"`
	CreatedAt              string     `json:"createdAt" gorm:"-"`
	UpdatedAt              string     `json:"updatedAt" gorm:"-"`
	LastError              string     `json:"lastError" gorm:"column:last_error"`
}

type adminRiskUserRow struct {
	ID                 uint64    `json:"id" gorm:"column:id"`
	UserID             uint64    `json:"userID" gorm:"column:user_id"`
	UserNo             string    `json:"userNo" gorm:"column:user_no"`
	Username           string    `json:"username" gorm:"column:username"`
	Phone              string    `json:"phone" gorm:"column:phone"`
	CountryCode        string    `json:"countryCode" gorm:"column:country_code"`
	RiskLevel          string    `json:"riskLevel" gorm:"column:risk_level"`
	RiskStatus         string    `json:"riskStatus" gorm:"column:risk_status"`
	RiskTags           string    `json:"riskTags" gorm:"column:risk_tags"`
	TriggerCount       uint      `json:"triggerCount" gorm:"column:trigger_count"`
	RegisteredIP       string    `json:"registeredIP" gorm:"column:registered_ip"`
	LastLoginIP        string    `json:"lastLoginIP" gorm:"column:last_login_ip"`
	AssociatedAccounts uint      `json:"associatedAccounts" gorm:"column:associated_accounts"`
	AssignedReviewer   string    `json:"assignedReviewer" gorm:"column:assigned_reviewer"`
	LatestReason       string    `json:"latestReason" gorm:"column:latest_reason"`
	LatestRiskAtValue  time.Time `json:"-" gorm:"column:latest_risk_at"`
	UpdatedAtValue     time.Time `json:"-" gorm:"column:updated_at"`
	LatestRiskAt       string    `json:"latestRiskAt" gorm:"-"`
	UpdatedAt          string    `json:"updatedAt" gorm:"-"`
}

func (h *adminUsersHandler) ListUsers(c *gin.Context) {
	page, pageSize, offset := adminPage(c)
	db := database.GetDB()
	query := db.Table("game_users AS u").
		Joins("LEFT JOIN game_user_wallets AS w ON w.user_id = u.id AND w.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_risk_profiles AS rp ON rp.user_id = u.id AND rp.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")

	query = adminApplyUserFilters(c, query)
	query = adminApplyWalletFilters(c, query)

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}

	rows := []adminUserRow{}
	err := query.Session(&gorm.Session{}).
		Select(strings.Join([]string{
			"u.id",
			"u.user_no",
			"u.username",
			"u.email",
			"u.phone_country_code",
			"u.phone",
			"u.nickname",
			"u.vip_level",
			"u.status",
			"u.country_code",
			"u.language",
			"u.registration_channel",
			"u.registered_ip",
			"u.last_login_ip",
			"u.login_enabled",
			"u.game_access_enabled",
			"u.last_login_at",
			"u.created_at",
			"u.updated_at",
			"COALESCE(rp.risk_level, 'LOW') AS risk_level",
			"COALESCE(rp.risk_status, 'CLEARED') AS risk_status",
			"COALESCE(w.currency, 'PKR') AS currency",
			"CAST(COALESCE(w.cash_balance, 0.00) AS CHAR) AS cash_balance",
			"CAST(COALESCE(w.bonus_balance, 0.00) AS CHAR) AS bonus_balance",
			"CAST(COALESCE(w.withdrawable_balance, 0.00) AS CHAR) AS withdrawable_balance",
			"CAST(COALESCE(w.frozen_balance, 0.00) AS CHAR) AS frozen_balance",
		}, ", ")).
		Order(adminUserOrder(c)).
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].StatusText = adminUserStatusText(rows[i].Status)
		rows[i].KYCStatus = "NOT_SUBMITTED"
		rows[i].LastLoginAt = adminFormatTimePtr(rows[i].LastLoginAtValue)
		rows[i].RegisteredAt = adminFormatTime(rows[i].RegisteredAtValue)
		rows[i].UpdatedAt = adminFormatTime(rows[i].UpdatedAtValue)
	}
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		for i := range rows {
			rows[i].Phone = adminMaskPhone(rows[i].Phone)
			rows[i].Email = adminMaskEmail(rows[i].Email)
			rows[i].RegisteredIP = adminMaskIP(rows[i].RegisteredIP)
			rows[i].LastLoginIP = adminMaskIP(rows[i].LastLoginIP)
		}
	}

	response.Success(c, adminPageResult[adminUserRow]{List: rows, Total: total, Page: page, PageSize: pageSize})
}

func (h *adminUsersHandler) ListLoginRecords(c *gin.Context) {
	page, pageSize, offset := adminPage(c)
	db := database.GetDB()
	query := db.Table("game_user_login_records AS l").
		Joins("LEFT JOIN game_users AS u ON u.id = l.user_id AND u.deleted_at IS NULL").
		Where("l.deleted_at IS NULL")

	query = adminApplyAccountFilters(c, query)
	if value := adminQuery(c, "clientIP"); value != "" {
		query = query.Where("l.ip_address LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "deviceID"); value != "" {
		query = query.Where("l.device_id LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "country"); value != "" {
		query = query.Where("l.country = ?", value)
	}
	if value := adminQuery(c, "platform"); value != "" {
		query = query.Where("l.platform = ?", value)
	}
	if value := adminQuery(c, "browser"); value != "" {
		query = query.Where("l.browser LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "os"); value != "" {
		query = query.Where("l.os LIKE ?", "%"+value+"%")
	}
	if result := strings.ToUpper(adminQuery(c, "loginResult")); result != "" {
		if result == "SUCCESS" {
			query = query.Where("l.success = 1")
		} else if result == "FAILED" {
			query = query.Where("l.success = 0")
		}
	}
	query = adminApplyTimeRange(c, query, "l.occurred_at", "loginFrom", "loginTo")

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}

	rows := []adminLoginRecordRow{}
	err := query.Session(&gorm.Session{}).
		Select(strings.Join([]string{
			"l.id",
			"l.user_id",
			"u.user_no",
			"u.username",
			"u.phone_country_code",
			"u.phone",
			"u.nickname",
			"l.success",
			"l.failure_reason",
			"l.device_id",
			"l.ip_address",
			"l.country",
			"l.region",
			"l.device_type",
			"l.platform",
			"l.os",
			"l.browser",
			"l.app_version",
			"l.user_agent",
			"l.occurred_at",
		}, ", ")).
		Order("l.occurred_at DESC, l.id DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].LoginAt = adminFormatTime(rows[i].LoginAtValue)
		if rows[i].Success == 1 {
			rows[i].LoginResult = "SUCCESS"
		} else {
			rows[i].LoginResult = "FAILED"
		}
	}
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		for i := range rows {
			rows[i].Phone = adminMaskPhone(rows[i].Phone)
			rows[i].ClientIP = adminMaskIP(rows[i].ClientIP)
		}
	}

	response.Success(c, adminPageResult[adminLoginRecordRow]{List: rows, Total: total, Page: page, PageSize: pageSize})
}

func (h *adminUsersHandler) ListDevices(c *gin.Context) {
	page, pageSize, offset := adminPage(c)
	db := database.GetDB()
	query := db.Table("game_user_devices AS d").
		Joins("INNER JOIN game_users AS u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")

	query = adminApplyAccountFilters(c, query)
	if value := adminQuery(c, "platform"); value != "" {
		query = query.Where("d.platform = ?", value)
	}
	if value := adminQuery(c, "appVersion"); value != "" {
		query = query.Where("d.app_version LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "clientIP"); value != "" {
		query = query.Where("d.client_ip LIKE ?", "%"+value+"%")
	}
	if status, ok := adminIntQuery(c, "status"); ok {
		query = query.Where("d.status = ?", status)
	}
	if value := adminQuery(c, "deviceKeyword"); value != "" {
		like := "%" + value + "%"
		query = query.Where("(d.device_id LIKE ? OR d.user_agent LIKE ?)", like, like)
	}
	if value := adminQuery(c, "fingerprint"); value != "" {
		query = query.Where("d.fingerprint_hash LIKE ?", "%"+value+"%")
	}
	if trustStatus, ok := adminIntQuery(c, "trustStatus"); ok {
		query = query.Where("d.trust_status = ?", trustStatus)
	}
	if riskStatus, ok := adminIntQuery(c, "riskStatus"); ok {
		query = query.Where("d.risk_status = ?", riskStatus)
	}
	query = adminApplyTimeRange(c, query, "d.first_seen_at", "firstSeenFrom", "firstSeenTo")
	query = adminApplyTimeRange(c, query, "d.last_seen_at", "lastSeenFrom", "lastSeenTo")

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}

	rows := []adminDeviceRow{}
	err := query.Session(&gorm.Session{}).
		Select(strings.Join([]string{
			"d.id",
			"d.user_id",
			"u.user_no",
			"u.phone_country_code",
			"u.phone",
			"u.nickname",
			"d.device_id",
			"d.platform",
			"d.device_model",
			"d.os",
			"d.browser",
			"d.app_version",
			"d.fingerprint_hash",
			"d.first_ip",
			"d.last_ip",
			"d.client_ip",
			"d.user_agent",
			"d.status",
			"d.trust_status",
			"d.risk_status",
			"d.login_count",
			"(SELECT COUNT(DISTINCT d2.user_id) FROM game_user_devices d2 WHERE d2.device_id = d.device_id AND d2.deleted_at IS NULL) AS associated_accounts",
			"d.first_seen_at",
			"d.last_seen_at",
		}, ", ")).
		Order("d.last_seen_at DESC, d.id DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].StatusText = adminDeviceStatusText(rows[i].Status)
		rows[i].TrustStatusText = adminDeviceTrustStatusText(rows[i].TrustStatus)
		rows[i].RiskStatusText = adminDeviceRiskStatusText(rows[i].RiskStatus)
		rows[i].FirstSeenAt = adminFormatTimePtr(rows[i].FirstSeenValue)
		rows[i].LastSeenAt = adminFormatTimePtr(rows[i].LastSeenValue)
	}
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		for i := range rows {
			rows[i].Phone = adminMaskPhone(rows[i].Phone)
			rows[i].FingerprintHash = adminMaskHash(rows[i].FingerprintHash)
			rows[i].FirstIP = adminMaskIP(rows[i].FirstIP)
			rows[i].LastIP = adminMaskIP(rows[i].LastIP)
			rows[i].ClientIP = adminMaskIP(rows[i].ClientIP)
		}
	}

	response.Success(c, adminPageResult[adminDeviceRow]{List: rows, Total: total, Page: page, PageSize: pageSize})
}

func (h *adminUsersHandler) ListPushDevices(c *gin.Context) {
	page, pageSize, offset := adminPage(c)
	db := database.GetDB()
	scope := adminContentScope(c)
	query := content.ApplySite(db.Table("game_push_device_tokens AS t"), "t", scope).
		Joins("INNER JOIN game_users AS u ON u.id=t.user_id AND u.deleted_at IS NULL").
		Where("t.deleted_at IS NULL")
	query = content.ApplySite(query, "u", scope)
	query = adminApplyAccountFilters(c, query)
	if value := strings.ToUpper(adminQuery(c, "platform")); value != "" {
		query = query.Where("t.platform=?", value)
	}
	if value := adminQuery(c, "appVersion"); value != "" {
		query = query.Where("t.app_version LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "deviceKeyword"); value != "" {
		query = query.Where("(t.device_id LIKE ? OR t.fingerprint_hash LIKE ? OR t.device_model LIKE ?)", "%"+value+"%", "%"+value+"%", "%"+value+"%")
	}
	if value := strings.ToUpper(adminQuery(c, "networkType")); value != "" {
		query = query.Where("t.network_type=?", value)
	}
	if value := strings.ToUpper(adminQuery(c, "notificationPermission")); value != "" {
		query = query.Where("t.notification_permission=?", value)
	}
	switch strings.ToUpper(adminQuery(c, "pushStatus")) {
	case "ENABLED":
		query = query.Where("t.enabled=1 AND t.invalidated_at IS NULL")
	case "INVALIDATED":
		query = query.Where("(t.enabled=0 OR t.invalidated_at IS NOT NULL)")
	}
	query = adminApplyTimeRange(c, query, "t.last_seen_at", "lastSeenFrom", "lastSeenTo")

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}
	rows := []adminPushDeviceRow{}
	err := query.Session(&gorm.Session{}).
		Select(strings.Join([]string{
			"t.id", "t.user_id", "u.user_no", "u.phone_country_code", "u.phone", "u.nickname",
			"t.platform", "t.device_id", "t.fingerprint_hash", "t.manufacturer", "t.brand", "t.device_model",
			"t.device_name", "t.os_version", "t.sdk_version", "t.security_patch", "t.network_type",
			"t.network_metered", "t.network_roaming", "t.carrier_name", "t.screen_width", "t.screen_height",
			"t.screen_density", "t.timezone", "t.locale", "t.battery_level", "t.charging",
			"t.notification_permission", "t.app_version", "t.app_version_code", "t.language", "t.last_ip",
			"t.user_agent", "t.enabled",
			"t.last_seen_at", "t.invalidated_at", "t.last_error", "t.created_at", "t.updated_at",
		}, ", ")).
		Order("t.last_seen_at DESC, t.id DESC").Offset(offset).Limit(pageSize).Scan(&rows).Error
	if err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].Status = "ENABLED"
		if !rows[i].Enabled || rows[i].InvalidatedValue != nil {
			rows[i].Status = "INVALIDATED"
		}
		rows[i].LastSeenAt = adminFormatTimePtr(rows[i].LastSeenValue)
		rows[i].InvalidatedAt = adminFormatTimePtr(rows[i].InvalidatedValue)
		rows[i].CreatedAt = adminFormatTime(rows[i].CreatedAtValue)
		rows[i].UpdatedAt = adminFormatTime(rows[i].UpdatedAtValue)
	}
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		for i := range rows {
			rows[i].Phone = adminMaskPhone(rows[i].Phone)
			rows[i].FingerprintHash = adminMaskHash(rows[i].FingerprintHash)
			rows[i].LastIP = adminMaskIP(rows[i].LastIP)
		}
	}
	response.Success(c, adminPageResult[adminPushDeviceRow]{List: rows, Total: total, Page: page, PageSize: pageSize})
}

func (h *adminUsersHandler) ListRiskUsers(c *gin.Context) {
	page, pageSize, offset := adminPage(c)
	db := database.GetDB()
	query := db.Table("game_user_risk_profiles AS rp").
		Joins("INNER JOIN game_users AS u ON u.id = rp.user_id AND u.deleted_at IS NULL").
		Where("rp.deleted_at IS NULL")

	query = adminApplyAccountFilters(c, query)
	if value := strings.ToUpper(adminQuery(c, "riskLevel")); value != "" {
		query = query.Where("rp.risk_level = ?", value)
	}
	if value := strings.ToUpper(adminQuery(c, "riskStatus")); value != "" {
		query = query.Where("rp.risk_status = ?", value)
	}
	if value := adminQuery(c, "country"); value != "" {
		query = query.Where("u.country_code = ?", value)
	}
	if value := adminQuery(c, "registrationIP"); value != "" {
		query = query.Where("u.registered_ip LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "lastLoginIP"); value != "" {
		query = query.Where("u.last_login_ip LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "assignedReviewer"); value != "" {
		query = query.Where("rp.assigned_reviewer LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "riskTag"); value != "" {
		query = query.Where("EXISTS (SELECT 1 FROM game_user_tag_relations utr INNER JOIN game_user_tags t ON t.id = utr.tag_id AND t.deleted_at IS NULL WHERE utr.user_id = u.id AND utr.deleted_at IS NULL AND t.tag_type = 'RISK' AND t.name LIKE ?)", "%"+value+"%")
	}
	query = adminApplyTimeRange(c, query, "rp.created_at", "createdFrom", "createdTo")
	query = adminApplyTimeRange(c, query, "rp.updated_at", "updatedFrom", "updatedTo")

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}

	rows := []adminRiskUserRow{}
	err := query.Session(&gorm.Session{}).
		Select(strings.Join([]string{
			"rp.id",
			"rp.user_id",
			"u.user_no",
			"u.username",
			"u.phone",
			"u.country_code",
			"u.registered_ip",
			"u.last_login_ip",
			"rp.risk_level",
			"rp.risk_status",
			"rp.trigger_count",
			"rp.associated_accounts",
			"rp.assigned_reviewer",
			"rp.latest_reason",
			"rp.updated_at AS latest_risk_at",
			"rp.updated_at",
			"COALESCE((SELECT GROUP_CONCAT(t.name ORDER BY t.name SEPARATOR ', ') FROM game_user_tag_relations utr INNER JOIN game_user_tags t ON t.id = utr.tag_id AND t.deleted_at IS NULL WHERE utr.user_id = u.id AND utr.deleted_at IS NULL AND t.tag_type = 'RISK'), '') AS risk_tags",
		}, ", ")).
		Order("rp.updated_at DESC, rp.id DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].LatestRiskAt = adminFormatTime(rows[i].LatestRiskAtValue)
		rows[i].UpdatedAt = adminFormatTime(rows[i].UpdatedAtValue)
	}
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		for i := range rows {
			rows[i].Phone = adminMaskPhone(rows[i].Phone)
			rows[i].RegisteredIP = adminMaskIP(rows[i].RegisteredIP)
			rows[i].LastLoginIP = adminMaskIP(rows[i].LastLoginIP)
		}
	}

	response.Success(c, adminPageResult[adminRiskUserRow]{List: rows, Total: total, Page: page, PageSize: pageSize})
}

func adminApplyUserFilters(c *gin.Context, query *gorm.DB) *gorm.DB {
	query = adminApplyAccountFilters(c, query)
	if vip, ok := adminIntQuery(c, "vipLevel"); ok {
		query = query.Where("u.vip_level = ?", vip)
	}
	if status := adminQuery(c, "status"); status != "" {
		if value, ok := adminUserStatusValue(status); ok {
			query = query.Where("u.status = ?", value)
		}
	}
	if value := strings.ToUpper(adminQuery(c, "kycStatus")); value != "" && value != "NOT_SUBMITTED" {
		query = query.Where("1 = 0")
	}
	if value := adminQuery(c, "country"); value != "" {
		query = query.Where("u.country_code = ?", value)
	}
	if value := adminQuery(c, "channel"); value != "" {
		query = query.Where("u.registration_channel = ?", value)
	}
	if value := adminQuery(c, "registeredIP"); value != "" {
		query = query.Where("u.registered_ip LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "lastLoginIP"); value != "" {
		query = query.Where("u.last_login_ip LIKE ?", "%"+value+"%")
	}
	if value := strings.ToUpper(adminQuery(c, "riskStatus")); value != "" {
		query = query.Where("COALESCE(rp.risk_status, 'CLEARED') = ?", value)
	}
	query = adminApplyTimeRange(c, query, "u.created_at", "registeredFrom", "registeredTo")
	query = adminApplyTimeRange(c, query, "u.last_login_at", "lastLoginFrom", "lastLoginTo")
	return query
}

func adminApplyAccountFilters(c *gin.Context, query *gorm.DB) *gorm.DB {
	if value := adminQuery(c, "userId"); value != "" {
		if numericID, err := strconv.ParseUint(value, 10, 64); err == nil && numericID > 0 {
			query = query.Where("(u.id = ? OR u.user_no = ?)", numericID, value)
		} else {
			query = query.Where("u.user_no = ?", value)
		}
	}
	if value := adminQuery(c, "account"); value != "" {
		like := "%" + value + "%"
		query = query.Where("(u.user_no LIKE ? OR u.username LIKE ? OR u.email LIKE ? OR u.phone LIKE ? OR u.nickname LIKE ?)", like, like, like, like, like)
	}
	if value := adminQuery(c, "username"); value != "" {
		query = query.Where("u.username LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "email"); value != "" {
		query = query.Where("u.email LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "phoneCountryCode"); value != "" {
		query = query.Where("u.phone_country_code = ?", value)
	}
	if value := firstNonEmpty(adminQuery(c, "mobile"), adminQuery(c, "phone")); value != "" {
		query = query.Where("u.phone LIKE ?", "%"+value+"%")
	}
	if value := adminQuery(c, "nickname"); value != "" {
		query = query.Where("u.nickname LIKE ?", "%"+value+"%")
	}
	return query
}

func adminApplyWalletFilters(c *gin.Context, query *gorm.DB) *gorm.DB {
	if value := adminQuery(c, "currency"); value != "" {
		query = query.Where("w.currency = ?", value)
	}
	if value := adminQuery(c, "cashBalanceMin"); value != "" {
		query = query.Where("w.cash_balance >= ?", value)
	}
	if value := adminQuery(c, "cashBalanceMax"); value != "" {
		query = query.Where("w.cash_balance <= ?", value)
	}
	return query
}

func adminApplyTimeRange(c *gin.Context, query *gorm.DB, column string, fromKey string, toKey string) *gorm.DB {
	if value := adminQuery(c, fromKey); value != "" {
		if t, ok := adminParseDate(value, false); ok {
			query = query.Where(column+" >= ?", t)
		}
	}
	if value := adminQuery(c, toKey); value != "" {
		if t, ok := adminParseDate(value, true); ok {
			query = query.Where(column+" <= ?", t)
		}
	}
	return query
}

func adminParseDate(value string, endOfDay bool) (time.Time, bool) {
	value = strings.TrimSpace(value)
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			if layout == "2006-01-02" && endOfDay {
				t = t.Add(24*time.Hour - time.Millisecond)
			}
			return t, true
		}
	}
	return time.Time{}, false
}

func adminPage(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, (page - 1) * pageSize
}

func adminPathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		adminBadRequest(c, "invalid id")
		return 0, false
	}
	return id, true
}

func adminQuery(c *gin.Context, key string) string {
	return strings.TrimSpace(c.Query(key))
}

func adminIntQuery(c *gin.Context, key string) (int, bool) {
	value := adminQuery(c, key)
	if value == "" {
		return 0, false
	}
	intValue, err := strconv.Atoi(value)
	return intValue, err == nil
}

func adminUintQuery(c *gin.Context, key string) (uint64, bool) {
	value := adminQuery(c, key)
	if value == "" {
		return 0, false
	}
	intValue, err := strconv.ParseUint(value, 10, 64)
	return intValue, err == nil && intValue > 0
}

func adminBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": msg, "data": nil})
}

func adminDBError(c *gin.Context, err error) {
	_ = err
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

func adminFormatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func adminFormatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return adminFormatTime(*value)
}

func adminUserStatusText(status int) string {
	switch status {
	case 1:
		return "ACTIVE"
	case 2:
		return "DISABLED"
	case 3:
		return "FROZEN"
	case 4:
		return "LOCKED"
	case 5:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

func adminUserStatusValue(status string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "1", "ACTIVE":
		return 1, true
	case "2", "DISABLED":
		return 2, true
	case "3", "FROZEN":
		return 3, true
	case "4", "LOCKED":
		return 4, true
	case "5", "CLOSED":
		return 5, true
	default:
		return 0, false
	}
}

func adminDeviceStatusText(status int) string {
	if status == 2 {
		return "blocked"
	}
	return "active"
}

func adminDeviceTrustStatusText(status int) string {
	if status == 2 {
		return "TRUSTED"
	}
	return "NORMAL"
}

func adminDeviceRiskStatusText(status int) string {
	switch status {
	case 2:
		return "OBSERVED"
	case 3:
		return "RISKY"
	default:
		return "NORMAL"
	}
}

func adminUserOrder(c *gin.Context) string {
	fields := map[string]string{
		"registeredAt": "u.created_at",
		"lastLoginAt":  "u.last_login_at",
		"vipLevel":     "u.vip_level",
		"username":     "u.username",
	}
	column, ok := fields[adminQuery(c, "sortField")]
	if !ok {
		return "u.created_at DESC, u.id DESC"
	}
	order := "ASC"
	if strings.EqualFold(adminQuery(c, "sortOrder"), "desc") {
		order = "DESC"
	}
	return column + " " + order + ", u.id DESC"
}

func adminMaskPhone(value string) string {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

func adminMaskEmail(value string) string {
	parts := strings.SplitN(value, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return value
	}
	return parts[0][:1] + "***@" + parts[1]
}

func adminMaskIP(value string) string {
	parts := strings.Split(value, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + ".*.*"
	}
	if value == "" {
		return ""
	}
	return "***"
}

func adminMaskHash(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:6] + "..." + value[len(value)-6:]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func adminSessionStatus(expiredAt *time.Time, revokedAt *time.Time, now time.Time) string {
	if revokedAt != nil {
		return "revoked"
	}
	if expiredAt != nil && !expiredAt.After(now) {
		return "expired"
	}
	return "active"
}

func adminRiskLevel(status int) string {
	switch status {
	case 3:
		return "high"
	case 2:
		return "medium"
	default:
		return "low"
	}
}

func adminRiskReason(status int) string {
	switch status {
	case 3:
		return "Risk locked by platform rule"
	case 2:
		return "Account disabled, pending manual review"
	default:
		return "No current risk lock"
	}
}
