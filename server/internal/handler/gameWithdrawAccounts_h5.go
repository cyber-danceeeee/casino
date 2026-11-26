package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type gameWithdrawAccountRequest struct {
	AccountType      string `json:"accountType"`
	BankName         string `json:"bankName"`
	WalletType       string `json:"walletType"`
	AccountName      string `json:"accountName"`
	AccountNo        string `json:"accountNo"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	Phone            string `json:"phone"`
	IsDefault        int    `json:"isDefault"`
}

func (h *gameWithdrawAccountsHandler) Current(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(c, gin.H{"items": h5FindWithdrawAccounts(db, user.ID)})
}

func (h *gameWithdrawAccountsHandler) CreateCurrent(c *gin.Context) {
	req := &gameWithdrawAccountRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		h5Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	accountType := req.AccountType
	if accountType == "" {
		accountType = "bank"
	}
	scope := h5UserContentScope(db, user.ID, h5ContentScope(c))
	account := &model.GameWithdrawAccounts{

		SiteID:           scope.SiteID,
		UserID:           user.ID,
		AccountType:      accountType,
		BankName:         req.BankName,
		WalletType:       req.WalletType,
		AccountName:      req.AccountName,
		AccountNo:        req.AccountNo,
		PhoneCountryCode: h5FirstNonEmpty(req.PhoneCountryCode, "+92"),
		Phone:            h5CleanPhone(req.Phone),
		IsDefault:        req.IsDefault,
		Status:           1,
	}
	if account.IsDefault == 1 {
		_ = content.ApplySite(db.Model(&model.GameWithdrawAccounts{}), "", scope).Where("user_id = ?", user.ID).Update("is_default", 0).Error
	}
	_ = db.Create(account).Error
	response.Success(c, gin.H{"account": account})
}
