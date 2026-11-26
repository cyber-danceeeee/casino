package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"

	"casino/internal/database"
	"casino/internal/model"
	"casino/internal/tenancy"
)

func (h *gameWalletTransactionsHandler) WalletRecords(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope := h5UserTenancyScope(db, user.ID, h5TenancyScope(c))
	tab := strings.TrimSpace(c.Query("tab"))
	if tab == "" {
		tab = "statement"
	}
	page := h5Page(c)
	pageSize := h5PageSize(c, 10)
	offset := (page - 1) * pageSize

	switch tab {
	case "deposit":
		var records []model.GameDepositOrders
		query := tenancy.ApplySite(db.Model(&model.GameDepositOrders{}), "", scope).Where("user_id = ?", user.ID)
		var total int64
		_ = query.Count(&total).Error
		_ = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error
		items := make([]gin.H, 0, len(records))
		for _, record := range records {
			items = append(items, h5DepositRecordPayload(record))
		}
		response.Success(c, gin.H{"tab": tab, "items": items, "total": total, "page": page, "pageSize": pageSize})
	case "withdraw":
		var records []model.GameWithdrawOrders
		query := tenancy.ApplySite(db.Model(&model.GameWithdrawOrders{}), "", scope).Where("user_id = ?", user.ID)
		var total int64
		_ = query.Count(&total).Error
		_ = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error
		items := make([]gin.H, 0, len(records))
		for _, record := range records {
			items = append(items, h5WithdrawRecordPayload(record))
		}
		response.Success(c, gin.H{"tab": tab, "items": items, "total": total, "page": page, "pageSize": pageSize})
	default:
		var records []h5WalletRecordRow
		q := tenancy.ApplySite(db.Table("game_wallet_transactions wt"), "wt", scope).
			Where("wt.user_id = ? AND wt.deleted_at IS NULL", user.ID)

		categories, forcedDirection := h5WalletRecordCategoryFilter(c.Query("type"))
		if len(categories) > 0 {
			q = q.Where("UPPER(wt.category) IN ?", categories)
		}
		direction := forcedDirection
		if direction == "" {
			direction = c.Query("direction")
		}
		if directions := h5WalletRecordDirectionFilter(direction); len(directions) > 0 {
			q = q.Where("UPPER(wt.direction) IN ?", directions)
		}

		if raw := strings.TrimSpace(c.Query("startTime")); raw != "" {
			value, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				h5Error(c, http.StatusBadRequest, "invalid start time")
				return
			}
			q = q.Where("wt.created_at >= ?", value)
		}
		if raw := strings.TrimSpace(c.Query("endTime")); raw != "" {
			value, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				h5Error(c, http.StatusBadRequest, "invalid end time")
				return
			}
			q = q.Where("wt.created_at <= ?", value)
		}

		var total int64
		_ = q.Count(&total).Error
		_ = q.Select(`wt.*,
				COALESCE(g.id, 0) AS game_id,
				COALESCE(g.name, '') AS game_name,
				COALESCE(pt.provider_game_id, '') AS provider_game_id`).
			Joins("LEFT JOIN game_provider_transactions pt ON pt.wallet_transaction_id = wt.id AND pt.deleted_at IS NULL").
			Joins("LEFT JOIN game_game_providers gp ON gp.provider_code = wt.provider_code AND gp.deleted_at IS NULL").
			Joins("LEFT JOIN game_games g ON g.provider_id = gp.id AND g.provider_game_code = pt.provider_game_id AND g.deleted_at IS NULL").
			Order("wt.id DESC").Offset(offset).Limit(pageSize).Scan(&records).Error
		items := make([]gin.H, 0, len(records))
		for _, record := range records {
			items = append(items, h5WalletRecordPayload(record))
		}
		response.Success(c, gin.H{"tab": tab, "items": items, "total": total, "page": page, "pageSize": pageSize})
	}
}

type h5WalletRecordRow struct {
	model.GameWalletTransactions
	GameID         uint64 `gorm:"column:game_id"`
	GameName       string `gorm:"column:game_name"`
	ProviderGameID string `gorm:"column:provider_game_id"`
}

func h5WalletRecordCategoryFilter(value string) ([]string, string) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "ALL", "ALL_CATEGORIES":
		return nil, ""
	case "WITHDRAWAL":
		return []string{"WITHDRAW", "WITHDRAWAL", "WITHDRAWAL_FREEZE"}, ""
	case "WITHDRAWAL_REVERSAL":
		return []string{"WITHDRAWAL_REVERSAL", "WITHDRAWAL_RELEASE"}, ""
	case "BET_REFUND":
		return []string{"REFUND", "BET_REFUND"}, ""
	case "PAYOUT_CLAWBACK":
		return []string{"ROLLBACK", "PAYOUT_CLAWBACK"}, ""
	case "CREDIT_ADJUSTMENT":
		return []string{"MANUAL_ADJUSTMENT", "CREDIT_ADJUSTMENT"}, "CREDIT"
	case "DEBIT_ADJUSTMENT":
		return []string{"MANUAL_ADJUSTMENT", "DEBIT_ADJUSTMENT"}, "DEBIT"
	case "CASH_REWARD":
		return []string{"REGISTER_REWARD", "CASH_REWARD"}, ""
	default:
		return []string{strings.ToUpper(strings.TrimSpace(value))}, ""
	}
}

func h5WalletRecordDirectionFilter(value string) []string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "ALL", "ALL_DIRECTIONS":
		return nil
	case "IN", "CREDIT":
		return []string{"IN", "CREDIT"}
	case "OUT", "DEBIT":
		return []string{"OUT", "DEBIT"}
	default:
		return []string{strings.ToUpper(strings.TrimSpace(value))}
	}
}

func h5DepositRecordPayload(record model.GameDepositOrders) gin.H {
	createdAt := record.CreatedAt
	if record.CompletedAt != nil {
		createdAt = *record.CompletedAt
	} else if record.PaidAt != nil {
		createdAt = *record.PaidAt
	}
	return gin.H{
		"id":          record.ID,
		"title":       "Deposit",
		"source":      h5FirstNonEmpty(record.MethodCode, record.ChannelCode, "-"),
		"amount":      h5DecimalValue(record.Amount).Abs().StringFixed(2),
		"currency":    h5FirstNonEmpty(record.Currency, "PKR"),
		"direction":   "IN",
		"sign":        "+",
		"balance":     "",
		"time":        h5FormatWalletTime(createdAt),
		"orderNumber": record.OrderNo,
		"status":      h5OrderRecordStatus(record.Status),
		"category":    "DEPOSIT",
	}
}

func h5WithdrawRecordPayload(record model.GameWithdrawOrders) gin.H {
	createdAt := record.CreatedAt
	if record.CompletedAt != nil {
		createdAt = *record.CompletedAt
	} else if record.SubmittedAt != nil {
		createdAt = *record.SubmittedAt
	}
	return gin.H{
		"id":          record.ID,
		"title":       "Withdraw",
		"source":      h5FirstNonEmpty(record.PaymentMethod, "-"),
		"amount":      h5DecimalValue(record.Amount).Abs().StringFixed(2),
		"currency":    h5FirstNonEmpty(record.Currency, "PKR"),
		"direction":   "OUT",
		"sign":        "-",
		"balance":     "",
		"time":        h5FormatWalletTime(createdAt),
		"orderNumber": record.OrderNo,
		"status":      h5OrderRecordStatus(record.Status),
		"category":    "WITHDRAW",
	}
}

func h5WalletRecordPayload(row h5WalletRecordRow) gin.H {
	record := row.GameWalletTransactions
	createdAt := record.CreatedAt
	if record.CompletedAt != nil {
		createdAt = *record.CompletedAt
	}
	direction := strings.ToUpper(strings.TrimSpace(record.Direction))
	sign := "+"
	if direction == "OUT" || direction == "DEBIT" {
		sign = "-"
	}

	source := h5WalletRecordSource(record)
	if strings.TrimSpace(row.GameName) != "" {
		source = row.GameName
	}

	return gin.H{
		"id":             record.ID,
		"title":          h5WalletRecordTitle(record),
		"source":         source,
		"amount":         h5DecimalValue(record.Amount).Abs().StringFixed(2),
		"currency":       h5FirstNonEmpty(record.Currency, "PKR"),
		"direction":      direction,
		"sign":           sign,
		"balance":        h5DecimalValue(record.BalanceAfter).StringFixed(2),
		"time":           h5FormatWalletTime(createdAt),
		"orderNumber":    h5FirstNonEmpty(record.OrderNo, record.TxnNo),
		"status":         h5WalletRecordStatus(record.Status),
		"category":       record.Category,
		"gameId":         row.GameID,
		"gameName":       row.GameName,
		"providerCode":   record.ProviderCode,
		"providerGameId": row.ProviderGameID,
	}
}

func h5WalletRecordTitle(record model.GameWalletTransactions) string {
	switch strings.ToUpper(strings.TrimSpace(record.Category)) {
	case "REGISTER_REWARD":
		return "Cash Transfer"
	case "DEPOSIT":
		return "Deposit"
	case "WITHDRAW", "WITHDRAWAL":
		return "Withdraw"
	case "BONUS_TRANSFER":
		return "Bonus Transfer"
	case "BET":
		return "Bet"
	case "PAYOUT", "CREDIT":
		return "Payout"
	case "REFUND", "BET_REFUND":
		return "Bet Refund"
	case "ROLLBACK", "PAYOUT_CLAWBACK":
		return "Payout Clawback"
	case "MANUAL_ADJUSTMENT", "CREDIT_ADJUSTMENT", "DEBIT_ADJUSTMENT":
		return "Balance Adjustment"
	default:
		if strings.TrimSpace(record.Remark) != "" {
			return record.Remark
		}
		return strings.Title(strings.ToLower(strings.ReplaceAll(record.Category, "_", " ")))
	}
}

func h5WalletRecordSource(record model.GameWalletTransactions) string {
	switch strings.ToUpper(strings.TrimSpace(record.Category)) {
	case "REGISTER_REWARD":
		return "New User Package"
	default:
		if strings.TrimSpace(record.SourceType) != "" {
			return strings.Title(strings.ToLower(strings.ReplaceAll(record.SourceType, "_", " ")))
		}
		return "-"
	}
}

func h5WalletRecordStatus(status int) string {
	switch status {
	case 1:
		return "Success"
	case 2:
		return "Failed"
	case 0:
		return "Pending"
	default:
		return "Processing"
	}
}

func h5OrderRecordStatus(status int) string {
	switch status {
	case 1:
		return "Pending"
	case 2:
		return "Success"
	case 3:
		return "Failed"
	case 4:
		return "Returned"
	case 5:
		return "Cancelled"
	case 6:
		return "Reviewing"
	default:
		return "Processing"
	}
}

func h5FormatWalletTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func h5AuditProgressPercent(progress decimal.Decimal, required decimal.Decimal) string {
	if !required.GreaterThan(decimal.Zero) {
		return "100"
	}
	percent := progress.Div(required).Mul(decimal.NewFromInt(100))
	if percent.GreaterThan(decimal.NewFromInt(100)) {
		percent = decimal.NewFromInt(100)
	}
	if percent.LessThan(decimal.Zero) {
		percent = decimal.Zero
	}
	return percent.StringFixed(2)
}
