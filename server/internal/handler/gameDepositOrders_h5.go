package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"

	"casino/internal/database"
	"casino/internal/depositconfig"
	"casino/internal/model"
)

type gameDepositOrderRequest struct {
	Amount      string `json:"amount"`
	MethodCode  string `json:"methodCode"`
	ChannelCode string `json:"channelCode"`
	VoucherCode string `json:"voucherCode"`
	Currency    string `json:"currency"`
	Country     string `json:"country"`
}

func (h *gameDepositOrdersHandler) WalletDepositConfig(c *gin.Context) {
	query := h5DepositConfigQuery(c)
	data, err := h.configService.GetConfig(query)
	if err != nil {
		h5DepositConfigError(c, err)
		return
	}
	response.Success(c, data)
}

func (h *gameDepositOrdersHandler) CreateWalletDeposit(c *gin.Context) {
	req := &gameDepositOrderRequest{}
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
	configQuery := h5DepositConfigQuery(c)
	configQuery.UserID = user.ID
	configQuery.Scope = scope
	configQuery.MethodCode = strings.TrimSpace(req.MethodCode)
	configQuery.Language = user.Language
	configQuery.Amount = req.Amount
	selection, err := h.configService.ResolveSelection(configQuery, req.ChannelCode, req.VoucherCode)
	if err != nil {
		h5DepositConfigError(c, err)
		return
	}

	voucherID := uint64(0)
	if selection.Voucher != nil {
		voucherID = selection.Voucher.ID
	}
	order := &model.GameDepositOrders{

		SiteID:              scope.SiteID,
		OrderNo:             h5OrderNo("D"),
		UserID:              user.ID,
		Amount:              h5DecimalPtr(amount),
		PaidAmount:          h5DecimalPtr(decimal.Zero),
		Currency:            configQuery.Currency,
		MethodCode:          selection.Method.MethodCode,
		ChannelCode:         selection.Channel.ChannelCode,
		VoucherID:           voucherID,
		VoucherBonusReward:  h5DecimalPtr(selection.BonusReward),
		VoucherCashReward:   h5DecimalPtr(selection.CashReward),
		VoucherRewardAmount: h5DecimalPtr(selection.RewardAmount),
		Status:              1,
		ExpiredAt:           h5TimePtr(time.Now().Add(30 * time.Minute)),
	}
	if err := db.Create(order).Error; err != nil {
		h5Error(c, http.StatusInternalServerError, "failed to create deposit order")
		return
	}
	response.Success(c, gin.H{
		"order":      order,
		"cashierUrl": fmt.Sprintf("/wallet/deposit?orderNo=%s", order.OrderNo),
	})
}

func h5DepositConfigQuery(c *gin.Context) depositconfig.ConfigQuery {
	db := database.GetDB()
	scope := h5ContentScope(c)
	userID := uint64(0)
	if user, _, ok := h5AuthenticatedUser(db, c); ok {
		userID = user.ID
		scope = h5UserContentScope(db, user.ID, scope)
	}
	site := h5BootstrapSite(db, scope.SiteID)
	country := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CountryCode, c.Query("country"), "PK")))
	currency := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CurrencyCode, c.Query("currency"), "PKR")))
	return depositconfig.ConfigQuery{
		UserID: userID, Scope: scope, Country: country, Currency: currency,
		MethodCode: c.Query("methodCode"), Language: c.DefaultQuery("language", "en"), Amount: c.Query("amount"),
	}
}

func h5DepositConfigError(c *gin.Context, err error) {
	var domain *depositconfig.DomainError
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	h5Error(c, http.StatusInternalServerError, "internal error")
}
