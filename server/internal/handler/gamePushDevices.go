package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/database"
	"casino/internal/model"
	"casino/internal/tenancy"
)

type gamePushDeviceRequest struct {
	Token                  string  `json:"token"`
	Platform               string  `json:"platform"`
	DeviceID               string  `json:"deviceId"`
	FingerprintHash        string  `json:"fingerprintHash"`
	Manufacturer           string  `json:"manufacturer"`
	Brand                  string  `json:"brand"`
	Model                  string  `json:"model"`
	DeviceName             string  `json:"deviceName"`
	OSVersion              string  `json:"osVersion"`
	SDKVersion             uint    `json:"sdkVersion"`
	SecurityPatch          string  `json:"securityPatch"`
	NetworkType            string  `json:"networkType"`
	NetworkMetered         bool    `json:"networkMetered"`
	NetworkRoaming         bool    `json:"networkRoaming"`
	CarrierName            string  `json:"carrierName"`
	ScreenWidth            uint    `json:"screenWidth"`
	ScreenHeight           uint    `json:"screenHeight"`
	ScreenDensity          float64 `json:"screenDensity"`
	Timezone               string  `json:"timezone"`
	Locale                 string  `json:"locale"`
	BatteryLevel           int     `json:"batteryLevel"`
	Charging               bool    `json:"charging"`
	NotificationPermission string  `json:"notificationPermission"`
	AppVersion             string  `json:"appVersion"`
	AppVersionCode         uint64  `json:"appVersionCode"`
	Language               string  `json:"language"`
}

type gamePushDeviceHandler struct{}

func NewGamePushDeviceHandler() *gamePushDeviceHandler { return &gamePushDeviceHandler{} }

func (h *gamePushDeviceHandler) Register(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	var request gamePushDeviceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h5Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	request.Token = strings.TrimSpace(request.Token)
	request.Platform = strings.ToUpper(strings.TrimSpace(request.Platform))
	request.DeviceID = pushDeviceText(request.DeviceID, 128)
	request.FingerprintHash = pushDeviceText(request.FingerprintHash, 64)
	request.Manufacturer = pushDeviceText(request.Manufacturer, 64)
	request.Brand = pushDeviceText(request.Brand, 64)
	request.Model = pushDeviceText(request.Model, 128)
	request.DeviceName = pushDeviceText(request.DeviceName, 128)
	request.OSVersion = pushDeviceText(request.OSVersion, 64)
	request.SecurityPatch = pushDeviceText(request.SecurityPatch, 32)
	request.NetworkType = strings.ToUpper(pushDeviceText(request.NetworkType, 24))
	request.CarrierName = pushDeviceText(request.CarrierName, 128)
	request.Timezone = pushDeviceText(request.Timezone, 64)
	request.Locale = pushDeviceText(request.Locale, 32)
	request.NotificationPermission = strings.ToUpper(pushDeviceText(request.NotificationPermission, 16))
	request.AppVersion = pushDeviceText(request.AppVersion, 32)
	request.Language = pushDeviceText(request.Language, 16)
	if request.Token == "" || len(request.Token) > 4096 || request.Platform != "ANDROID" {
		h5Error(c, http.StatusBadRequest, "invalid push device")
		return
	}
	if request.Language == "" {
		request.Language = h5FirstNonEmpty(user.Language, "en")
	}
	if request.BatteryLevel < -1 || request.BatteryLevel > 100 {
		request.BatteryLevel = -1
	}
	now := time.Now()
	tokenHash := pushTokenHash(request.Token)
	scope := tenancy.Scope{TenantID: user.TenantID, SiteID: user.SiteID}
	record := model.GamePushDeviceToken{
		TenantID:               scope.TenantID,
		SiteID:                 scope.SiteID,
		UserID:                 user.ID,
		Platform:               request.Platform,
		TokenHash:              tokenHash,
		Token:                  request.Token,
		DeviceID:               request.DeviceID,
		FingerprintHash:        request.FingerprintHash,
		Manufacturer:           request.Manufacturer,
		Brand:                  request.Brand,
		DeviceModel:            request.Model,
		DeviceName:             request.DeviceName,
		OSVersion:              request.OSVersion,
		SDKVersion:             request.SDKVersion,
		SecurityPatch:          request.SecurityPatch,
		NetworkType:            request.NetworkType,
		NetworkMetered:         request.NetworkMetered,
		NetworkRoaming:         request.NetworkRoaming,
		CarrierName:            request.CarrierName,
		ScreenWidth:            request.ScreenWidth,
		ScreenHeight:           request.ScreenHeight,
		ScreenDensity:          request.ScreenDensity,
		Timezone:               request.Timezone,
		Locale:                 request.Locale,
		BatteryLevel:           request.BatteryLevel,
		Charging:               request.Charging,
		NotificationPermission: request.NotificationPermission,
		AppVersion:             request.AppVersion,
		AppVersionCode:         request.AppVersionCode,
		Language:               request.Language,
		LastIP:                 pushDeviceText(c.ClientIP(), 45),
		UserAgent:              pushDeviceText(c.Request.UserAgent(), 500),
		Enabled:                true,
		LastSeenAt:             &now,
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "token_hash"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"tenant_id", "site_id", "user_id", "platform", "token", "device_id", "fingerprint_hash", "manufacturer", "brand",
				"device_model", "device_name", "os_version", "sdk_version", "security_patch", "network_type",
				"network_metered", "network_roaming", "carrier_name", "screen_width", "screen_height",
				"screen_density", "timezone", "locale", "battery_level", "charging", "notification_permission",
				"app_version", "app_version_code", "language", "last_ip", "user_agent", "enabled", "last_seen_at",
				"invalidated_at", "last_error", "updated_at",
			}),
		}).Create(&record).Error; err != nil {
			return err
		}
		if request.DeviceID == "" {
			return nil
		}
		return tenancy.ApplySite(tx.Model(&model.GamePushDeviceToken{}), "", scope).
			Where("user_id=? AND platform=? AND device_id=? AND token_hash<>? AND deleted_at IS NULL", user.ID, request.Platform, request.DeviceID, tokenHash).
			Updates(map[string]any{"enabled": false, "invalidated_at": &now}).Error
	}); err != nil {
		h5Error(c, http.StatusInternalServerError, "push device registration failed")
		return
	}
	response.Success(c, gin.H{"registered": true})
}

func (h *gamePushDeviceHandler) Unregister(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	var request struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Token) == "" {
		h5Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	scope := tenancy.Scope{TenantID: user.TenantID, SiteID: user.SiteID}
	if err := tenancy.ApplySite(db.Model(&model.GamePushDeviceToken{}), "", scope).
		Where("user_id=? AND token_hash=? AND deleted_at IS NULL", user.ID, pushTokenHash(strings.TrimSpace(request.Token))).
		Updates(map[string]any{"enabled": false, "invalidated_at": &now}).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "push device unregister failed")
		return
	}
	response.Success(c, gin.H{"registered": false})
}

func pushTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func pushDeviceText(value string, maxLength int) string {
	trimmed := strings.TrimSpace(value)
	runes := []rune(trimmed)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return trimmed
}
