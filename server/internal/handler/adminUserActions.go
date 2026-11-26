package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/model"
)

var (
	errAdminUserNotFound       = errors.New("admin user target not found")
	errAdminDeviceNotFound     = errors.New("admin device target not found")
	errInvalidStatusTransition = errors.New("invalid user status transition")
	errInvalidDeviceStatus     = errors.New("invalid device status")
	errRiskProfileNotFound     = errors.New("risk profile not found")
)

type adminStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type adminRestrictionRequest struct {
	LoginEnabled       *bool  `json:"loginEnabled"`
	GameAccessEnabled  *bool  `json:"gameAccessEnabled"`
	DepositEnabled     *bool  `json:"depositEnabled"`
	WithdrawalEnabled  *bool  `json:"withdrawalEnabled"`
	RiskReviewRequired *bool  `json:"riskReviewRequired"`
	Reason             string `json:"reason"`
}

type adminRiskRequest struct {
	RiskLevel        string `json:"riskLevel"`
	RiskStatus       string `json:"riskStatus"`
	AssignedReviewer string `json:"assignedReviewer"`
	Reason           string `json:"reason"`
}

type adminDeviceStatusRequest struct {
	Status      string `json:"status"`
	TrustStatus string `json:"trustStatus"`
	Reason      string `json:"reason"`
}

func (h *adminUsersHandler) UpdateUserStatus(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := adminStatusRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	req.Reason = strings.TrimSpace(req.Reason)
	targetStatus, valid := adminUserStatusValue(req.Status)
	if !valid || req.Reason == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "status and operator reason are required")
		return
	}
	actor, _ := adminauth.FromContext(c)
	var updated map[string]any
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		user := model.GameUsers{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errAdminUserNotFound
			}
			return err
		}
		if !adminValidUserStatusTransition(user.Status, targetStatus) {
			return errInvalidStatusTransition
		}
		before := map[string]any{"status": adminUserStatusText(user.Status)}
		after := map[string]any{"status": adminUserStatusText(targetStatus)}
		if err := tx.Model(&model.GameUsers{}).Where("id = ? AND status = ?", id, user.Status).Update("status", targetStatus).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.GameUserStatusHistories{
			UserID: id, ChangeType: "ACCOUNT_STATUS", BeforeValue: before["status"].(string), AfterValue: after["status"].(string),
			Reason: req.Reason, OperatorID: actor.ID, OperatorUsername: actor.Username,
		}).Error; err != nil {
			return err
		}
		if err := adminWriteAudit(tx, c, actor, adminauth.PermissionUserStatusUpdate, "USER_STATUS_UPDATED", id, 0, before, after, req.Reason); err != nil {
			return err
		}
		updated = after
		return nil
	})
	if err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *adminUsersHandler) UpdateUserRestrictions(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := adminRestrictionRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	if req.LoginEnabled == nil && req.GameAccessEnabled == nil && req.DepositEnabled == nil && req.WithdrawalEnabled == nil && req.RiskReviewRequired == nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "at least one restriction is required")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "operator reason is required")
		return
	}
	actor, _ := adminauth.FromContext(c)
	var updated map[string]any
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		user := model.GameUsers{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errAdminUserNotFound
			}
			return err
		}
		before := adminRestrictionState(user)
		updates := map[string]any{}
		applyRestrictionUpdate(updates, "login_enabled", req.LoginEnabled)
		applyRestrictionUpdate(updates, "game_access_enabled", req.GameAccessEnabled)
		applyRestrictionUpdate(updates, "deposit_enabled", req.DepositEnabled)
		applyRestrictionUpdate(updates, "withdrawal_enabled", req.WithdrawalEnabled)
		applyRestrictionUpdate(updates, "risk_review_required", req.RiskReviewRequired)
		if err := tx.Model(&model.GameUsers{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Take(&user).Error; err != nil {
			return err
		}
		after := adminRestrictionState(user)
		if err := tx.Create(&model.GameUserStatusHistories{
			UserID: id, ChangeType: "RESTRICTIONS", BeforeValue: compactState(before), AfterValue: compactState(after),
			Reason: req.Reason, OperatorID: actor.ID, OperatorUsername: actor.Username,
		}).Error; err != nil {
			return err
		}
		if err := adminWriteAudit(tx, c, actor, adminauth.PermissionUserRestrictionUpdate, "USER_RESTRICTIONS_UPDATED", id, 0, before, after, req.Reason); err != nil {
			return err
		}
		updated = after
		return nil
	})
	if err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *adminUsersHandler) UpdateUserRisk(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := adminRiskRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	req.RiskLevel = strings.ToUpper(strings.TrimSpace(req.RiskLevel))
	req.RiskStatus = strings.ToUpper(strings.TrimSpace(req.RiskStatus))
	req.AssignedReviewer = strings.TrimSpace(req.AssignedReviewer)
	req.Reason = strings.TrimSpace(req.Reason)
	if !adminValidRiskLevel(req.RiskLevel) || !adminValidRiskStatus(req.RiskStatus) || req.Reason == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "valid risk level, risk status and reason are required")
		return
	}
	actor, _ := adminauth.FromContext(c)
	var updated map[string]any
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := adminEnsureUser(tx, id); err != nil {
			return err
		}
		profile := model.GameUserRiskProfiles{}
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND deleted_at IS NULL", id).Take(&profile).Error
		before := map[string]any{"riskLevel": "LOW", "riskStatus": "CLEARED", "assignedReviewer": ""}
		if err == nil {
			before = map[string]any{"riskLevel": profile.RiskLevel, "riskStatus": profile.RiskStatus, "assignedReviewer": profile.AssignedReviewer}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		after := map[string]any{"riskLevel": req.RiskLevel, "riskStatus": req.RiskStatus, "assignedReviewer": req.AssignedReviewer}
		if profile.ID == 0 {
			profile = model.GameUserRiskProfiles{UserID: id, RiskLevel: req.RiskLevel, RiskStatus: req.RiskStatus, AssignedReviewer: req.AssignedReviewer, AssignedReviewerID: actor.ID, LatestReason: req.Reason, TriggerCount: 1}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&profile).Updates(map[string]any{
				"risk_level": req.RiskLevel, "risk_status": req.RiskStatus, "assigned_reviewer": req.AssignedReviewer,
				"assigned_reviewer_id": actor.ID, "latest_reason": req.Reason, "trigger_count": gorm.Expr("trigger_count + 1"),
			}).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&model.GameUserRiskEvents{
			UserID: id, EventType: "MANUAL_STATUS_CHANGE", RiskLevel: req.RiskLevel, Source: "MANUAL", Description: req.Reason,
			OperatorID: actor.ID, OperatorUsername: actor.Username,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.GameUserStatusHistories{
			UserID: id, ChangeType: "RISK_STATUS", BeforeValue: compactState(before), AfterValue: compactState(after), Reason: req.Reason,
			OperatorID: actor.ID, OperatorUsername: actor.Username,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.GameUsers{}).Where("id = ?", id).Update("risk_review_required", req.RiskStatus != "CLEARED").Error; err != nil {
			return err
		}
		if err := adminWriteAudit(tx, c, actor, adminauth.PermissionUserRiskUpdate, "USER_RISK_UPDATED", id, 0, before, after, req.Reason); err != nil {
			return err
		}
		updated = after
		return nil
	})
	if err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *adminUsersHandler) UpdateDeviceStatus(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := adminDeviceStatusRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	req.TrustStatus = strings.ToUpper(strings.TrimSpace(req.TrustStatus))
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Status == "" && req.TrustStatus == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "device status or trust status is required")
		return
	}
	if req.Status != "" && req.Status != "ACTIVE" && req.Status != "BLOCKED" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_DEVICE_STATUS", "invalid device status")
		return
	}
	if req.TrustStatus != "" && req.TrustStatus != "NORMAL" && req.TrustStatus != "TRUSTED" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_DEVICE_STATUS", "invalid device trust status")
		return
	}
	if req.Reason == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "operator reason is required")
		return
	}
	actor, _ := adminauth.FromContext(c)
	var updated map[string]any
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		device := model.GameUserDevices{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", id).Take(&device).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errAdminDeviceNotFound
			}
			return err
		}
		before := map[string]any{"status": adminDeviceStatusText(device.Status), "trustStatus": adminDeviceTrustStatusText(device.TrustStatus)}
		updates := map[string]any{}
		if req.Status == "ACTIVE" {
			updates["status"] = 1
		} else if req.Status == "BLOCKED" {
			updates["status"] = 2
		}
		if req.TrustStatus == "NORMAL" {
			updates["trust_status"] = 1
		} else if req.TrustStatus == "TRUSTED" {
			updates["trust_status"] = 2
		}
		if err := tx.Model(&device).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Take(&device).Error; err != nil {
			return err
		}
		after := map[string]any{"status": adminDeviceStatusText(device.Status), "trustStatus": adminDeviceTrustStatusText(device.TrustStatus)}
		if err := adminWriteAudit(tx, c, actor, adminauth.PermissionDeviceManage, "USER_DEVICE_UPDATED", device.UserID, id, before, after, req.Reason); err != nil {
			return err
		}
		updated = after
		return nil
	})
	if err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c, updated)
}

func adminValidUserStatusTransition(from int, to int) bool {
	if from == to || from == 5 {
		return false
	}
	if to == 5 {
		return true
	}
	return from >= 1 && from <= 4 && to >= 1 && to <= 4
}

func adminValidRiskLevel(value string) bool {
	return value == "LOW" || value == "MEDIUM" || value == "HIGH" || value == "CRITICAL"
}

func adminValidRiskStatus(value string) bool {
	return value == "OBSERVED" || value == "UNDER_REVIEW" || value == "RESTRICTED" || value == "CLEARED" || value == "BLOCKED"
}

func adminRestrictionState(user model.GameUsers) map[string]any {
	return map[string]any{
		"loginEnabled": user.LoginEnabled == 1, "gameAccessEnabled": user.GameAccessEnabled == 1,
		"depositEnabled": user.DepositEnabled == 1, "withdrawalEnabled": user.WithdrawalEnabled == 1,
		"riskReviewRequired": user.RiskReviewRequired == 1,
	}
}

func applyRestrictionUpdate(updates map[string]any, column string, value *bool) {
	if value != nil {
		updates[column] = *value
	}
}

func compactState(value any) string {
	data, _ := json.Marshal(value)
	result := string(data)
	if len(result) > 64 {
		return result[:64]
	}
	return result
}

func adminEnsureUser(tx *gorm.DB, id uint64) error {
	var count int64
	if err := tx.Model(&model.GameUsers{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errAdminUserNotFound
	}
	return nil
}

func adminWriteAudit(tx *gorm.DB, c *gin.Context, actor adminauth.Actor, permission string, action string, userID uint64, deviceID uint64, before any, after any, reason string) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	correlationID := strings.TrimSpace(c.GetHeader("X-Correlation-ID"))
	if correlationID == "" {
		correlationID = fmt.Sprintf("admin-%d-%d", actor.ID, time.Now().UnixNano())
	}
	return tx.Create(&model.GameAdminOperationLogs{
		TenantID: actor.TenantID, SiteID: actor.SiteID,
		OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action,
		TargetUserID: userID, TargetDeviceID: deviceID, BeforeState: string(beforeJSON), AfterState: string(afterJSON),
		Reason: reason, IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(), CorrelationID: correlationID,
	}).Error
}

func adminAPIError(c *gin.Context, status int, code string, message string) {
	c.JSON(http.StatusOK, gin.H{"code": status, "msg": message, "data": gin.H{"errorCode": code}})
}

func adminHandleActionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errAdminUserNotFound):
		adminAPIError(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	case errors.Is(err, errAdminDeviceNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		adminAPIError(c, http.StatusNotFound, "DEVICE_NOT_FOUND", "device not found")
	case errors.Is(err, errInvalidStatusTransition):
		adminAPIError(c, http.StatusConflict, "INVALID_USER_STATUS_TRANSITION", "invalid user status transition")
	case errors.Is(err, errInvalidDeviceStatus):
		adminAPIError(c, http.StatusBadRequest, "INVALID_DEVICE_STATUS", "invalid device status")
	case errors.Is(err, errRiskProfileNotFound):
		adminAPIError(c, http.StatusNotFound, "RISK_PROFILE_NOT_FOUND", "risk profile not found")
	default:
		adminDBError(c, err)
	}
}

func adminHandleDeviceError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		adminAPIError(c, http.StatusNotFound, "DEVICE_NOT_FOUND", "device not found")
		return
	}
	adminDBError(c, err)
}
