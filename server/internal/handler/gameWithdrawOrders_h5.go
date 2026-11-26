package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"

	"casino/internal/database"
	"casino/internal/fundadmin"
)

type gameWithdrawOrderRequest struct {
	Amount            string `json:"amount"`
	WithdrawAccountID uint64 `json:"withdrawAccountId"`
}

func (h *gameWithdrawOrdersHandler) CreateWalletWithdraw(c *gin.Context) {
	req := &gameWithdrawOrderRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		h5Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	amount, err := h5ParseMoney(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		h5Error(c, http.StatusBadRequest, "invalid amount")
		return
	}

	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))
	if blocked, code, message := h5MaintenanceBlocked(c, db, scope); blocked {
		h5MachineError(c, http.StatusServiceUnavailable, code, message)
		return
	}
	order, err := fundadmin.New(db).CreateWithdrawal(user.ID, req.WithdrawAccountID, amount, fundadmin.Actor{Scope: scope, IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent")})
	if err != nil {
		h5Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"order": order})
}
