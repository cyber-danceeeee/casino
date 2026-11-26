package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameVipLevelsHandler) Summary(c *gin.Context) {
	db := database.GetDB()
	user, wallet, _ := h5AuthenticatedUser(db, c)
	scope := h5ContentScope(c)
	deposit, err := h5VipDepositTotal(db, user.ID, wallet.Currency)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "VIP progress unavailable")
		return
	}

	var levels []model.GameVipLevels
	_ = content.ApplySite(db.Model(&model.GameVipLevels{}), "", scope).Where("status = ?", 1).Order("level_no ASC").Find(&levels).Error

	var requirements []model.GameVipRequirements
	_ = content.ApplySite(db.Model(&model.GameVipRequirements{}), "", scope).Order("target_level_no ASC, requirement_type ASC").Find(&requirements).Error

	response.Success(c, gin.H{
		"currentLevel":  user.VipLevel,
		"depositAmount": deposit.StringFixed(2),
		"levels":        h5VipLevelsPayload(levels),
		"requirements":  requirements,
		"benefits":      h5ConfigJSONAssets(c, db, "h5.vip.benefits", []gin.H{}),
	})
}

// Only settled deposits belonging to the authenticated member contribute to the
// displayed progress. Pending orders, other currencies and guest IDs never do.
func h5VipDepositTotal(db *gorm.DB, userID uint64, currency string) (decimal.Decimal, error) {
	if userID == 0 || strings.TrimSpace(currency) == "" {
		return decimal.Zero, nil
	}
	var row struct{ Amount decimal.Decimal }
	err := db.Model(&model.GameDepositOrders{}).
		Select("COALESCE(SUM(paid_amount),0) AS amount").
		Where("site_id=1 AND user_id=? AND currency=? AND status=2", userID, strings.ToUpper(strings.TrimSpace(currency))).
		Scan(&row).Error
	return row.Amount, err
}
