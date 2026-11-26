package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/bonusaudit"
	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameUserWalletsHandler) Summary(c *gin.Context) {
	db := database.GetDB()
	user, wallet, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(c, gin.H{
		"user":   h5PlayerPayload(user, wallet),
		"wallet": h5WalletPayload(wallet),
	})
}

func (h *gameUserWalletsHandler) CashAudits(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))

	var claims []model.GameRegisterRewardClaims
	if err := content.ApplySite(db.Model(&model.GameRegisterRewardClaims{}), "", scope).
		Where("user_id = ? AND status = ? AND wallet_type = ? AND deleted_at IS NULL", user.ID, 2, "CASH").
		Order("claimed_at DESC, id DESC").
		Limit(h5Limit(c, 50)).
		Find(&claims).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "failed to load cash audit records")
		return
	}

	items := make([]gin.H, 0, len(claims))
	for _, claim := range claims {
		amount := h5DecimalValue(claim.Amount)
		progress, err := bonusaudit.RegisterRewardProgress(db, scope, claim)
		if err != nil {
			h5Error(c, http.StatusInternalServerError, "failed to calculate cash audit progress")
			return
		}
		timeValue := claim.CreatedAt
		if claim.ClaimedAt != nil {
			timeValue = *claim.ClaimedAt
		}
		status := "Auditing"
		if progress.Remaining.IsZero() {
			status = "Completed"
		}
		items = append(items, gin.H{
			"id":                  claim.ID,
			"walletTransactionID": claim.WalletTransactionID,
			"type":                "Cash Reward",
			"source":              "New User Package",
			"amount":              amount.StringFixed(2),
			"currency":            h5FirstNonEmpty(claim.Currency, "PKR"),
			"createdAt":           h5FormatWalletTime(timeValue),
			"auditMultiplier":     progress.Multiplier.StringFixed(1),
			"auditRequirement":    progress.Multiplier.StringFixed(1) + "x",
			"requiredAmount":      progress.Required.StringFixed(2),
			"progressAmount":      progress.Completed.StringFixed(2),
			"progressPercent":     h5AuditProgressPercent(progress.Completed, progress.Required),
			"status":              status,
		})
	}

	response.Success(c, gin.H{"items": items, "total": len(items)})
}
