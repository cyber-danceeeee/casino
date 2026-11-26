package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"

	"casino/internal/database"
)

type GameBonusCenterHandler interface {
	Records(c *gin.Context)
}

type gameBonusCenterHandler struct{}

func NewGameBonusCenterHandler() GameBonusCenterHandler {
	return &gameBonusCenterHandler{}
}

type h5BonusCenterRecordRow struct {
	ID                  uint64          `json:"id"`
	BonusID             string          `json:"bonusID"`
	OrderNo             string          `json:"orderNo"`
	SourceType          string          `json:"sourceType"`
	SourceName          string          `json:"sourceName"`
	BonusType           string          `json:"bonusType"`
	Amount              decimal.Decimal `json:"amount"`
	RequirementMultiple decimal.Decimal `json:"requirementMultiplier"`
	RequiredValidBet    decimal.Decimal `json:"requiredValidBet"`
	CompletedValidBet   decimal.Decimal `json:"completedValidBet"`
	CurrentWagering     decimal.Decimal `json:"currentWagering"`
	RemainingWagering   decimal.Decimal `json:"remainingWagering"`
	Currency            string          `json:"currency"`
	Status              int             `json:"status"`
	StatusCode          string          `json:"statusCode"`
	IssuedAt            *time.Time      `json:"issuedAt"`
}

func (h *gameBonusCenterHandler) Records(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	page := int(h5QueryUint64(c, "page"))
	if page <= 0 {
		page = 1
	}
	pageSize := int(h5QueryUint64(c, "pageSize"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	category := strings.ToLower(strings.TrimSpace(c.DefaultQuery("category", "all")))
	where, args := h5BonusCategoryWhere(category)

	query := db.Table("game_bonus_records b").
		Joins("LEFT JOIN game_activity_events p ON p.id = b.promotion_id").
		Where("b.user_id = ? AND b.deleted_at IS NULL", user.ID)
	if where != "" {
		query = query.Where(where, args...)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "failed to load bonus records")
		return
	}

	var rows []h5BonusCenterRecordRow
	err := query.Select(`
			b.id,
			COALESCE(NULLIF(b.bonus_id,''), b.order_no) bonus_id,
			b.order_no,
			b.source_type,
			COALESCE(NULLIF(p.title,''), b.source_type) source_name,
			b.bonus_type,
			b.amount,
			b.requirement_multiplier,
			b.required_valid_bet,
			b.completed_valid_bet,
			b.current_wagering,
			b.remaining_wagering,
			b.currency,
			b.status,
			b.status_code,
			b.issued_at
		`).
		Order("COALESCE(b.issued_at, b.created_at) DESC, b.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "failed to load bonus records")
		return
	}

	list := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		code, label := h5BonusCategory(row.SourceType, row.BonusType)
		list = append(list, gin.H{
			"id":                row.ID,
			"orderId":           h5FirstNonEmpty(row.BonusID, row.OrderNo),
			"amount":            row.Amount.StringFixed(2),
			"source":            h5BonusSourceName(row.SourceName, row.SourceType, row.BonusType),
			"category":          label,
			"categoryCode":      code,
			"bonusType":         row.BonusType,
			"issuedTime":        h5FormatTime(row.IssuedAt),
			"status":            h5BonusStatus(row.StatusCode, row.Status),
			"requirement":       h5BonusRequirement(row),
			"requiredValidBet":  row.RequiredValidBet.StringFixed(2),
			"completedValidBet": row.CompletedValidBet.StringFixed(2),
			"currentWagering":   row.CurrentWagering.StringFixed(2),
			"remainingWagering": row.RemainingWagering.StringFixed(2),
			"currency":          h5FirstNonEmpty(row.Currency, "PKR"),
		})
	}

	response.Success(c, gin.H{
		"list":       list,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"categories": h5BonusCategories(),
	})
}

func h5BonusCategories() []gin.H {
	return []gin.H{
		{"code": "all", "label": "All Categories"},
		{"code": "activity", "label": "Activity Bonus"},
		{"code": "vip", "label": "VIP Bonus"},
		{"code": "compensation", "label": "Compensation Bonus"},
		{"code": "incentive", "label": "Incentive Rewards"},
	}
}

func h5BonusCategoryWhere(category string) (string, []any) {
	expr := "UPPER(CONCAT_WS(' ', b.source_type, b.bonus_type))"
	switch category {
	case "vip":
		return expr + " LIKE ?", []any{"%VIP%"}
	case "compensation":
		return "(" + expr + " LIKE ? OR " + expr + " LIKE ?)", []any{"%COMPENSATION%", "%RELIEF%"}
	case "incentive":
		return "(" + expr + " LIKE ? OR " + expr + " LIKE ? OR " + expr + " LIKE ? OR " + expr + " LIKE ?)", []any{"%INCENTIVE%", "%INVITE%", "%REFERRAL%", "%SHARE%"}
	case "activity":
		return "(" + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ? AND " + expr + " NOT LIKE ?)", []any{"%VIP%", "%COMPENSATION%", "%RELIEF%", "%INCENTIVE%", "%INVITE%", "%REFERRAL%", "%SHARE%"}
	default:
		return "", nil
	}
}

func h5BonusCategory(sourceType string, bonusType string) (string, string) {
	value := strings.ToUpper(sourceType + " " + bonusType)
	switch {
	case strings.Contains(value, "VIP"):
		return "vip", "VIP Bonus"
	case strings.Contains(value, "COMPENSATION"), strings.Contains(value, "RELIEF"):
		return "compensation", "Compensation Bonus"
	case strings.Contains(value, "INCENTIVE"), strings.Contains(value, "INVITE"), strings.Contains(value, "REFERRAL"), strings.Contains(value, "SHARE"):
		return "incentive", "Incentive Rewards"
	default:
		return "activity", "Activity Bonus"
	}
}

func h5BonusSourceName(sourceName string, sourceType string, bonusType string) string {
	value := strings.TrimSpace(h5FirstNonEmpty(sourceName, bonusType, sourceType))
	if value == "" {
		return "Bonus"
	}
	return strings.NewReplacer("_", " ", "-", " ").Replace(value)
}

func h5BonusStatus(statusCode string, status int) string {
	if strings.TrimSpace(statusCode) != "" {
		return statusCode
	}
	switch status {
	case 2:
		return "COMPLETED"
	case 3:
		return "EXPIRED"
	case 4:
		return "CANCELLED"
	default:
		return "PENDING"
	}
}

func h5BonusRequirement(row h5BonusCenterRecordRow) string {
	if row.RequirementMultiple.GreaterThan(decimal.Zero) {
		multiplier := strings.TrimRight(strings.TrimRight(row.RequirementMultiple.StringFixed(2), "0"), ".")
		return fmt.Sprintf("Valid Bet %sx", multiplier)
	}
	if row.RequiredValidBet.GreaterThan(decimal.Zero) {
		return fmt.Sprintf("Valid Bet Rs %s", row.RequiredValidBet.StringFixed(2))
	}
	return "Valid Bet 0x"
}

func h5FormatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
