package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type h5NoticeItem struct {
	ID          string     `json:"id"`
	Source      string     `json:"source"`
	Title       string     `json:"title"`
	ContentHTML string     `json:"contentHTML"`
	NoticeType  string     `json:"noticeType"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	Read        bool       `json:"read"`
}

type h5NoticeReadRequest struct {
	IDs []string `json:"ids"`
}

type h5AnnouncementNotice struct {
	ID          uint64
	NoticeType  string
	Title       string
	TitleZh     string
	TitleEn     string
	ContentHTML string
	ContentZh   string
	ContentEn   string
	CreatedAt   time.Time
}

func (h *gameSupportQuestionsHandler) Support(c *gin.Context) {
	db := database.GetDB()
	user, hasUser := h5SupportRequestUser(db, c)
	var categories []model.GameSupportCategories
	_ = db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&categories).Error

	var questions []model.GameSupportQuestions
	q := db.Where("status = ?", 1)
	if categoryID := h5QueryUint64(c, "categoryId"); categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	} else if categoryCode := strings.TrimSpace(c.Query("category")); categoryCode != "" && categoryCode != "all" {
		q = q.Where("category_id IN (?)",
			db.Model(&model.GameSupportCategories{}).
				Select("id").
				Where("category_code = ? AND status = ?", categoryCode, 1),
		)
	}
	_ = q.Order("sort ASC, id ASC").Find(&questions).Error

	notices := h5SupportNotices(db, c, user, hasUser)

	response.Success(c, gin.H{
		"categories": h5SupportCategories(categories),
		"questions":  h5SupportQuestions(questions),
		"notices":    notices,
	})
}

func (h *gameSupportQuestionsHandler) NotificationStatus(c *gin.Context) {
	db := database.GetDB()
	user, ok := h5SupportRequestUser(db, c)
	if !ok {
		response.Success(c, gin.H{"unreadCount": 0})
		return
	}
	response.Success(c, gin.H{"unreadCount": h5UnreadNotificationCount(db, user)})
}

func (h *gameSupportQuestionsHandler) MarkNoticesRead(c *gin.Context) {
	db := database.GetDB()
	user, ok := h5SupportRequestUser(db, c)
	if !ok {
		response.Success(c, gin.H{"unreadCount": 0})
		return
	}

	req := h5NoticeReadRequest{}
	_ = c.ShouldBindJSON(&req)
	allowed := h5EligibleNotifications(db, user)
	allowedByID := make(map[uint64]bool, len(allowed))
	for _, notification := range allowed {
		allowedByID[notification.ID] = true
	}

	ids := h5NotificationIDsFromNoticeKeys(req.IDs)
	if len(ids) == 0 {
		for _, notification := range allowed {
			ids = append(ids, notification.ID)
		}
	}

	now := time.Now()
	records := make([]model.GameUserNotificationReads, 0, len(ids))
	for _, id := range ids {
		if !allowedByID[id] {
			continue
		}
		records = append(records, model.GameUserNotificationReads{
			UserID:         user.ID,
			NotificationID: id,
			ReadAt:         &now,
		})
	}
	if len(records) > 0 {
		_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error
	}

	response.Success(c, gin.H{"unreadCount": h5UnreadNotificationCount(db, user)})
}

func h5SupportCategories(categories []model.GameSupportCategories) []model.GameSupportCategories {
	if len(categories) > 0 {
		return categories
	}
	return []model.GameSupportCategories{
		{CategoryCode: "other", Name: "Other", Sort: 1, Status: 1},
		{CategoryCode: "deposit", Name: "Deposit", Sort: 2, Status: 1},
		{CategoryCode: "withdrawal", Name: "Withdrawal", Sort: 3, Status: 1},
	}
}

func h5SupportQuestions(questions []model.GameSupportQuestions) []gin.H {
	result := make([]gin.H, 0, len(questions))
	for _, question := range questions {
		result = append(result, gin.H{
			"id":         question.ID,
			"categoryId": question.CategoryID,
			"question":   question.Question,
			"answer":     question.AnswerHTML,
			"answerHTML": question.AnswerHTML,
			"sort":       question.Sort,
			"status":     question.Status,
		})
	}
	return result
}

func h5SupportNotices(db *gorm.DB, c *gin.Context, user model.GameUsers, hasUser bool) []h5NoticeItem {
	language := h5SupportLanguage(c)
	result := make([]h5NoticeItem, 0, 32)

	var announcements []h5AnnouncementNotice
	now := time.Now()
	scope := h5ContentScope(c)
	noticeQuery := content.ApplySite(db.Table("game_notices"), "", scope).
		Select("id, notice_type, title, title_zh, title_en, content_html, content_zh, content_en, created_at").
		Where("status = ? AND deleted_at IS NULL", 1).
		Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now)
	if hasUser && !user.CreatedAt.IsZero() {
		noticeQuery = noticeQuery.Where("created_at >= ?", user.CreatedAt)
	}
	_ = noticeQuery.
		Order("sort ASC, id DESC").
		Limit(20).
		Scan(&announcements).Error
	for _, notice := range announcements {
		title := h5LocalizedText(language, notice.TitleZh, notice.TitleEn, notice.Title)
		content := h5LocalizedText(language, notice.ContentZh, notice.ContentEn, notice.ContentHTML)
		if title == "" && content == "" {
			continue
		}
		createdAt := notice.CreatedAt
		result = append(result, h5NoticeItem{
			ID:          "notice:" + h5Uint64String(notice.ID),
			Source:      "notice",
			Title:       title,
			ContentHTML: content,
			NoticeType:  notice.NoticeType,
			CreatedAt:   &createdAt,
			Read:        true,
		})
	}

	if hasUser {
		readIDs := h5ReadNotificationIDs(db, user.ID)
		for _, notification := range h5EligibleNotifications(db, user) {
			title := h5LocalizedText(language, notification.TitleZh, notification.TitleEn, "")
			content := h5LocalizedText(language, notification.ContentZh, notification.ContentEn, "")
			if title == "" && content == "" {
				continue
			}
			createdAt := notification.CreatedAt
			result = append(result, h5NoticeItem{
				ID:          "notification:" + h5Uint64String(notification.ID),
				Source:      "notification",
				Title:       title,
				ContentHTML: content,
				NoticeType:  notification.Channel,
				CreatedAt:   &createdAt,
				Read:        readIDs[notification.ID],
			})
		}
	}
	return result
}

func h5EligibleNotifications(db *gorm.DB, user model.GameUsers) []model.GameNotification {
	if user.ID == 0 {
		return nil
	}
	now := time.Now()
	var rows []model.GameNotification
	_ = db.Where("FIND_IN_SET(?, channel) > 0 AND deleted_at IS NULL", "IN_APP").
		Where("(status IN ? OR (status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)))",
			[]string{"SENDING", "SENT", "PARTIAL_FAILED", "FAILED"}, "SCHEDULED", now).
		Order("scheduled_at DESC, id DESC").
		Limit(100).
		Find(&rows).Error

	result := make([]model.GameNotification, 0, len(rows))
	for _, row := range rows {
		if h5NotificationMatchesUser(row, user) {
			result = append(result, row)
		}
	}
	return result
}

func h5UnreadNotificationCount(db *gorm.DB, user model.GameUsers) int {
	readIDs := h5ReadNotificationIDs(db, user.ID)
	count := 0
	for _, notification := range h5EligibleNotifications(db, user) {
		if !readIDs[notification.ID] {
			count++
		}
	}
	return count
}

func h5ReadNotificationIDs(db *gorm.DB, userID uint64) map[uint64]bool {
	result := map[uint64]bool{}
	if userID == 0 {
		return result
	}
	var rows []model.GameUserNotificationReads
	_ = db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&rows).Error
	for _, row := range rows {
		result[row.NotificationID] = true
	}
	return result
}

func h5NotificationMatchesUser(notification model.GameNotification, user model.GameUsers) bool {
	audienceType := strings.ToUpper(strings.TrimSpace(notification.AudienceType))
	audienceValue := notification.AudienceValue
	switch audienceType {
	case "", "ALL_USERS":
		return true
	case "SELECTED_USERS":
		return h5CSVContains(audienceValue, h5Uint64String(user.ID)) || h5CSVContains(audienceValue, user.UserNo)
	case "VIP_LEVEL":
		return h5CSVContains(audienceValue, h5UintString(user.VipLevel))
	case "COUNTRY":
		return h5CSVContains(audienceValue, user.CountryCode)
	case "REGISTRATION_CHANNEL":
		return h5CSVContains(audienceValue, user.RegistrationChannel)
	default:
		return false
	}
}

func h5NotificationIDsFromNoticeKeys(keys []string) []uint64 {
	ids := make([]uint64, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if strings.HasPrefix(key, "notification:") {
			if id := h5ParseUint64(strings.TrimPrefix(key, "notification:")); id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func h5SupportRequestUser(db *gorm.DB, c *gin.Context) (model.GameUsers, bool) {
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		return model.GameUsers{}, false
	}
	return user, true
}

func h5SupportLanguage(c *gin.Context) string {
	language := strings.ToLower(c.Query("language"))
	if language == "" {
		language = strings.ToLower(c.GetHeader("Accept-Language"))
	}
	if strings.Contains(language, "zh") {
		return "zh"
	}
	return "en"
}

func h5LocalizedText(language string, zh string, en string, fallback string) string {
	if language == "zh" {
		return h5FirstNonEmpty(zh, en, fallback)
	}
	return h5FirstNonEmpty(en, zh, fallback)
}

func h5CSVContains(csv string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, item := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return true
		}
	}
	return false
}

func h5Uint64String(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func h5UintString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

func h5ParseUint64(value string) uint64 {
	var result uint64
	for _, char := range strings.TrimSpace(value) {
		if char < '0' || char > '9' {
			return 0
		}
		result = result*10 + uint64(char-'0')
	}
	return result
}
