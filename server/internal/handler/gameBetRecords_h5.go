package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
	"casino/internal/model"
	"casino/internal/tenancy"
)

func (h *gameBetRecordsHandler) PlayerRecords(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope := h5UserTenancyScope(db, user.ID, h5TenancyScope(c))
	tab := strings.TrimSpace(c.Query("tab"))
	if tab == "report" {
		var reports []model.GameBetReports
		q := tenancy.ApplySite(db.Model(&model.GameBetReports{}), "", scope).Where("user_id = ?", user.ID)
		if gameType := strings.TrimSpace(c.Query("gameType")); gameType != "" && gameType != "all" {
			q = q.Where("game_type = ?", gameType)
		}
		_ = q.Order("report_date DESC, id DESC").Limit(h5Limit(c, 50)).Find(&reports).Error
		response.Success(c, gin.H{"tab": "report", "items": reports, "total": len(reports)})
		return
	}

	var records []model.GameBetRecords
	q := tenancy.ApplySite(db.Model(&model.GameBetRecords{}), "", scope).Where("user_id = ?", user.ID)
	if status := strings.TrimSpace(c.Query("status")); status != "" && status != "all" {
		if v, err := strconv.Atoi(status); err == nil {
			q = q.Where("status = ?", v)
		}
	}
	if gameType := strings.TrimSpace(c.Query("gameType")); gameType != "" && gameType != "all" {
		q = q.Where("game_type = ?", gameType)
	}
	_ = q.Order("bet_at DESC, id DESC").Limit(h5Limit(c, 50)).Find(&records).Error
	response.Success(c, gin.H{"tab": "records", "items": records, "total": len(records)})
}
