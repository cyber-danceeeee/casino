package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/model"
)

type adminUserDetail struct {
	ID                    uint64     `json:"id" gorm:"column:id"`
	UserNo                string     `json:"userNo" gorm:"column:user_no"`
	Username              string     `json:"username" gorm:"column:username"`
	Email                 string     `json:"email" gorm:"column:email"`
	PhoneCountryCode      string     `json:"phoneCountryCode" gorm:"column:phone_country_code"`
	Phone                 string     `json:"phone" gorm:"column:phone"`
	Nickname              string     `json:"nickname" gorm:"column:nickname"`
	CountryCode           string     `json:"countryCode" gorm:"column:country_code"`
	Language              string     `json:"language" gorm:"column:language"`
	Currency              string     `json:"currency" gorm:"column:currency"`
	RegistrationChannel   string     `json:"registrationChannel" gorm:"column:registration_channel"`
	InviteCode            string     `json:"referralCode" gorm:"column:invite_code"`
	VipLevel              uint       `json:"vipLevel" gorm:"column:vip_level"`
	Status                int        `json:"status" gorm:"column:status"`
	StatusText            string     `json:"statusText" gorm:"-"`
	KYCStatus             string     `json:"kycStatus" gorm:"-"`
	RiskLevel             string     `json:"riskLevel" gorm:"column:risk_level"`
	RiskStatus            string     `json:"riskStatus" gorm:"column:risk_status"`
	InternalTags          string     `json:"internalTags" gorm:"column:internal_tags"`
	RegisteredIP          string     `json:"registeredIP" gorm:"column:registered_ip"`
	RegistrationDevice    string     `json:"registrationDevice" gorm:"column:registration_device"`
	LastLoginIP           string     `json:"lastLoginIP" gorm:"column:last_login_ip"`
	LastLoginDevice       string     `json:"lastLoginDevice" gorm:"column:last_login_device"`
	FailedLoginCount      uint       `json:"failedLoginCount" gorm:"column:failed_login_count"`
	LoginEnabled          int        `json:"loginEnabled" gorm:"column:login_enabled"`
	GameAccessEnabled     int        `json:"gameAccessEnabled" gorm:"column:game_access_enabled"`
	DepositEnabled        int        `json:"depositEnabled" gorm:"column:deposit_enabled"`
	WithdrawalEnabled     int        `json:"withdrawalEnabled" gorm:"column:withdrawal_enabled"`
	RiskReviewRequired    int        `json:"riskReviewRequired" gorm:"column:risk_review_required"`
	CashBalance           string     `json:"availableBalance" gorm:"column:cash_balance"`
	FrozenBalance         string     `json:"frozenBalance" gorm:"column:frozen_balance"`
	BonusBalance          string     `json:"bonusBalance" gorm:"column:bonus_balance"`
	TotalDeposit          string     `json:"totalDeposit" gorm:"column:total_deposit"`
	TotalWithdrawal       string     `json:"totalWithdrawal" gorm:"column:total_withdrawal"`
	TotalBet              string     `json:"totalBet" gorm:"column:total_bet"`
	TotalWin              string     `json:"totalWin" gorm:"column:total_win"`
	RegisteredAtValue     time.Time  `json:"-" gorm:"column:created_at"`
	LastLoginAtValue      *time.Time `json:"-" gorm:"column:last_login_at"`
	LastActiveAtValue     *time.Time `json:"-" gorm:"column:last_active_at"`
	PasswordChangedValue  *time.Time `json:"-" gorm:"column:last_password_changed_at"`
	RegisteredAt          string     `json:"registeredAt" gorm:"-"`
	LastLoginAt           string     `json:"lastLoginAt" gorm:"-"`
	LastActiveAt          string     `json:"lastActiveAt" gorm:"-"`
	LastPasswordChangedAt string     `json:"lastPasswordChangedAt" gorm:"-"`
}

type adminNoteRow struct {
	ID               uint64    `json:"id" gorm:"column:id"`
	OperatorID       uint64    `json:"operatorID" gorm:"column:operator_id"`
	OperatorUsername string    `json:"operatorUsername" gorm:"column:operator_username"`
	Content          string    `json:"content" gorm:"column:content"`
	CreatedAtValue   time.Time `json:"-" gorm:"column:created_at"`
	CreatedAt        string    `json:"createdAt" gorm:"-"`
}

func (h *adminUsersHandler) Me(c *gin.Context) {
	actor, _ := adminauth.FromContext(c)
	response.Success(c, actor)
}

func (h *adminUsersHandler) GetUser(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	detail := adminUserDetail{}
	err := database.GetDB().Table("game_users AS u").
		Select(strings.Join([]string{
			"u.id", "u.user_no", "u.username", "u.email", "u.phone_country_code", "u.phone", "u.nickname",
			"u.country_code", "u.language", "u.registration_channel", "u.invite_code", "u.vip_level", "u.status",
			"u.registered_ip", "u.last_login_ip", "u.failed_login_count", "u.login_enabled", "u.game_access_enabled",
			"u.deposit_enabled", "u.withdrawal_enabled", "u.risk_review_required", "u.created_at", "u.last_login_at",
			"u.last_password_changed_at", "COALESCE(w.currency, 'PKR') AS currency",
			"CAST(COALESCE(w.cash_balance, 0) AS CHAR) AS cash_balance",
			"CAST(COALESCE(w.frozen_balance, 0) AS CHAR) AS frozen_balance",
			"CAST(COALESCE(w.bonus_balance, 0) AS CHAR) AS bonus_balance",
			"COALESCE(rp.risk_level, 'LOW') AS risk_level", "COALESCE(rp.risk_status, 'CLEARED') AS risk_status",
			"COALESCE((SELECT GROUP_CONCAT(t.name ORDER BY t.name SEPARATOR ', ') FROM game_user_tag_relations utr INNER JOIN game_user_tags t ON t.id = utr.tag_id AND t.deleted_at IS NULL WHERE utr.user_id = u.id AND utr.deleted_at IS NULL AND t.tag_type = 'INTERNAL'), '') AS internal_tags",
			"COALESCE((SELECT d.device_id FROM game_user_devices d WHERE d.user_id = u.id AND d.deleted_at IS NULL ORDER BY d.first_seen_at, d.id LIMIT 1), '') AS registration_device",
			"COALESCE((SELECT d.device_id FROM game_user_devices d WHERE d.user_id = u.id AND d.deleted_at IS NULL ORDER BY d.last_seen_at DESC, d.id DESC LIMIT 1), '') AS last_login_device",
			"(SELECT MAX(d.last_seen_at) FROM game_user_devices d WHERE d.user_id = u.id AND d.deleted_at IS NULL) AS last_active_at",
			"CAST(COALESCE((SELECT SUM(o.paid_amount) FROM game_deposit_orders o WHERE o.user_id = u.id AND o.status = 2 AND o.deleted_at IS NULL), 0) AS CHAR) AS total_deposit",
			"CAST(COALESCE((SELECT SUM(o.actual_amount) FROM game_withdraw_orders o WHERE o.user_id = u.id AND o.status = 2 AND o.deleted_at IS NULL), 0) AS CHAR) AS total_withdrawal",
			"CAST(COALESCE((SELECT SUM(b.valid_bet_amount) FROM game_bet_records b WHERE b.user_id = u.id AND b.status = 2 AND b.deleted_at IS NULL), 0) AS CHAR) AS total_bet",
			"CAST(COALESCE((SELECT SUM(b.payout_amount) FROM game_bet_records b WHERE b.user_id = u.id AND b.status = 2 AND b.deleted_at IS NULL), 0) AS CHAR) AS total_win",
		}, ", ")).
		Joins("LEFT JOIN game_user_wallets w ON w.user_id = u.id AND w.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_risk_profiles rp ON rp.user_id = u.id AND rp.deleted_at IS NULL").
		Where("u.id = ? AND u.deleted_at IS NULL", id).
		Take(&detail).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			adminAPIError(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		adminDBError(c, err)
		return
	}
	detail.StatusText = adminUserStatusText(detail.Status)
	detail.KYCStatus = "NOT_SUBMITTED"
	detail.RegisteredAt = adminFormatTime(detail.RegisteredAtValue)
	detail.LastLoginAt = adminFormatTimePtr(detail.LastLoginAtValue)
	detail.LastActiveAt = adminFormatTimePtr(detail.LastActiveAtValue)
	detail.LastPasswordChangedAt = adminFormatTimePtr(detail.PasswordChangedValue)
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		detail.Phone = adminMaskPhone(detail.Phone)
		detail.Email = adminMaskEmail(detail.Email)
		detail.RegisteredIP = adminMaskIP(detail.RegisteredIP)
		detail.LastLoginIP = adminMaskIP(detail.LastLoginIP)
	}
	response.Success(c, detail)
}

func (h *adminUsersHandler) ListUserNotes(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	rows := []adminNoteRow{}
	if err := database.GetDB().Table("game_user_admin_notes").Where("user_id = ? AND deleted_at IS NULL", id).
		Order("created_at DESC, id DESC").Scan(&rows).Error; err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].CreatedAt = adminFormatTime(rows[i].CreatedAtValue)
	}
	response.Success(c, rows)
}

func (h *adminUsersHandler) AddUserNote(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := struct {
		Content string `json:"content"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" || len(req.Content) > 1000 {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "note content is required and must not exceed 1000 characters")
		return
	}
	actor, _ := adminauth.FromContext(c)
	if err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := adminEnsureUser(tx, id); err != nil {
			return err
		}
		note := model.GameUserAdminNotes{UserID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Content: req.Content}
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		return adminWriteAudit(tx, c, actor, adminauth.PermissionUserNoteCreate, "USER_NOTE_CREATED", id, 0, nil, map[string]any{"noteID": note.ID}, "")
	}); err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c)
}

func (h *adminUsersHandler) UpdateUserTags(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	req := struct {
		Tags   []string `json:"tags"`
		Reason string   `json:"reason"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		adminAPIError(c, http.StatusBadRequest, "INVALID_REQUEST", "tags and operator reason are required")
		return
	}
	cleaned := make([]string, 0, len(req.Tags))
	seen := map[string]bool{}
	for _, value := range req.Tags {
		value = strings.TrimSpace(value)
		if value != "" && len(value) <= 64 && !seen[value] {
			seen[value] = true
			cleaned = append(cleaned, value)
		}
	}
	actor, _ := adminauth.FromContext(c)
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := adminEnsureUser(tx, id); err != nil {
			return err
		}
		before := []string{}
		_ = tx.Table("game_user_tag_relations utr").Joins("INNER JOIN game_user_tags t ON t.id = utr.tag_id").Where("utr.user_id = ? AND utr.deleted_at IS NULL AND t.deleted_at IS NULL AND t.tag_type = 'INTERNAL'", id).Pluck("t.name", &before).Error
		now := time.Now()
		if err := tx.Table("game_user_tag_relations").Where("user_id = ? AND tag_id IN (SELECT id FROM game_user_tags WHERE tag_type = 'INTERNAL') AND deleted_at IS NULL", id).Update("deleted_at", now).Error; err != nil {
			return err
		}
		for _, name := range cleaned {
			tag := model.GameUserTags{}
			err := tx.Unscoped().Where("name = ? AND tag_type = 'INTERNAL'", name).Take(&tag).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = model.GameUserTags{Name: name, TagType: "INTERNAL"}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if tag.DeletedAt.Valid {
				if err := tx.Unscoped().Model(&tag).Update("deleted_at", nil).Error; err != nil {
					return err
				}
			}
			relation := model.GameUserTagRelations{}
			err = tx.Unscoped().Where("user_id = ? AND tag_id = ?", id, tag.ID).Take(&relation).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&model.GameUserTagRelations{UserID: id, TagID: tag.ID}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				if err := tx.Unscoped().Model(&relation).Update("deleted_at", nil).Error; err != nil {
					return err
				}
			}
		}
		return adminWriteAudit(tx, c, actor, adminauth.PermissionUserNoteCreate, "USER_TAGS_UPDATED", id, 0, map[string]any{"tags": before}, map[string]any{"tags": cleaned}, strings.TrimSpace(req.Reason))
	})
	if err != nil {
		adminHandleActionError(c, err)
		return
	}
	response.Success(c, gin.H{"tags": cleaned})
}

func (h *adminUsersHandler) ListUserOperationLogs(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	page, pageSize, offset := adminPage(c)
	query := database.GetDB().Table("game_admin_operation_logs").Where("target_user_id = ? AND deleted_at IS NULL", id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		adminDBError(c, err)
		return
	}
	rows := []struct {
		ID               uint64    `json:"id"`
		OperatorUsername string    `json:"operatorUsername" gorm:"column:operator_username"`
		Action           string    `json:"action"`
		BeforeState      string    `json:"beforeState" gorm:"column:before_state"`
		AfterState       string    `json:"afterState" gorm:"column:after_state"`
		Reason           string    `json:"reason"`
		CreatedAtValue   time.Time `json:"-" gorm:"column:created_at"`
		CreatedAt        string    `json:"createdAt" gorm:"-"`
	}{}
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Scan(&rows).Error; err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].CreatedAt = adminFormatTime(rows[i].CreatedAtValue)
	}
	response.Success(c, adminPageResult[any]{List: anySlice(rows), Total: total, Page: page, PageSize: pageSize})
}

func (h *adminUsersHandler) GetDevice(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	c.Request.URL.RawQuery = ""
	row := adminDeviceRow{}
	err := database.GetDB().Table("game_user_devices AS d").
		Select("d.*, u.user_no, u.phone_country_code, u.phone, u.nickname, (SELECT COUNT(DISTINCT d2.user_id) FROM game_user_devices d2 WHERE d2.device_id = d.device_id AND d2.deleted_at IS NULL) AS associated_accounts").
		Joins("INNER JOIN game_users u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Where("d.id = ? AND d.deleted_at IS NULL", id).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			adminAPIError(c, http.StatusNotFound, "DEVICE_NOT_FOUND", "device not found")
			return
		}
		adminDBError(c, err)
		return
	}
	row.StatusText = adminDeviceStatusText(row.Status)
	row.TrustStatusText = adminDeviceTrustStatusText(row.TrustStatus)
	row.RiskStatusText = adminDeviceRiskStatusText(row.RiskStatus)
	row.FirstSeenAt = adminFormatTimePtr(row.FirstSeenValue)
	row.LastSeenAt = adminFormatTimePtr(row.LastSeenValue)
	actor, _ := adminauth.FromContext(c)
	if !actor.Can(adminauth.PermissionSensitiveView) {
		row.Phone = adminMaskPhone(row.Phone)
		row.FingerprintHash = adminMaskHash(row.FingerprintHash)
		row.FirstIP = adminMaskIP(row.FirstIP)
		row.LastIP = adminMaskIP(row.LastIP)
		row.ClientIP = adminMaskIP(row.ClientIP)
	}
	response.Success(c, row)
}

func (h *adminUsersHandler) ListDeviceUsers(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	device := struct {
		DeviceID string `gorm:"column:device_id"`
	}{}
	if err := database.GetDB().Table("game_user_devices").Select("device_id").Where("id = ? AND deleted_at IS NULL", id).Take(&device).Error; err != nil {
		adminHandleDeviceError(c, err)
		return
	}
	rows := []adminUserRow{}
	if err := database.GetDB().Table("game_user_devices d").
		Select("u.id, u.user_no, u.username, u.phone_country_code, u.phone, u.status, u.created_at").
		Joins("INNER JOIN game_users u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Where("d.device_id = ? AND d.deleted_at IS NULL", device.DeviceID).
		Order("d.last_seen_at DESC").Scan(&rows).Error; err != nil {
		adminDBError(c, err)
		return
	}
	for i := range rows {
		rows[i].StatusText = adminUserStatusText(rows[i].Status)
		rows[i].RegisteredAt = adminFormatTime(rows[i].RegisteredAtValue)
	}
	response.Success(c, rows)
}

func (h *adminUsersHandler) GetRiskUser(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	profile := adminRiskUserRow{}
	err := database.GetDB().Table("game_user_risk_profiles rp").
		Select("rp.id, rp.user_id, u.user_no, u.username, u.phone, u.country_code, u.registered_ip, u.last_login_ip, rp.risk_level, rp.risk_status, rp.trigger_count, rp.associated_accounts, rp.assigned_reviewer, rp.latest_reason, rp.updated_at AS latest_risk_at, rp.updated_at").
		Joins("INNER JOIN game_users u ON u.id = rp.user_id AND u.deleted_at IS NULL").
		Where("rp.user_id = ? AND rp.deleted_at IS NULL", id).Take(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			adminAPIError(c, http.StatusNotFound, "RISK_PROFILE_NOT_FOUND", "risk profile not found")
			return
		}
		adminDBError(c, err)
		return
	}
	profile.LatestRiskAt = adminFormatTime(profile.LatestRiskAtValue)
	profile.UpdatedAt = adminFormatTime(profile.UpdatedAtValue)
	events := []struct {
		ID               uint64    `json:"id"`
		EventType        string    `json:"eventType" gorm:"column:event_type"`
		RiskLevel        string    `json:"riskLevel" gorm:"column:risk_level"`
		Source           string    `json:"source"`
		Description      string    `json:"description"`
		OperatorUsername string    `json:"operatorUsername" gorm:"column:operator_username"`
		CreatedAtValue   time.Time `json:"-" gorm:"column:created_at"`
		CreatedAt        string    `json:"createdAt" gorm:"-"`
	}{}
	_ = database.GetDB().Table("game_user_risk_events").Where("user_id = ? AND deleted_at IS NULL", id).Order("created_at DESC, id DESC").Scan(&events).Error
	for i := range events {
		events[i].CreatedAt = adminFormatTime(events[i].CreatedAtValue)
	}
	devices := []adminDeviceRow{}
	_ = database.GetDB().Table("game_user_devices").Where("user_id = ? AND deleted_at IS NULL", id).Order("last_seen_at DESC").Scan(&devices).Error
	response.Success(c, gin.H{"profile": profile, "events": events, "devices": devices})
}

func anySlice[T any](values []T) []any {
	result := make([]any, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}
