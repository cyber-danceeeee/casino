package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
	"casino/internal/depositconfig"
)

func (h *gamePaymentMethodsHandler) WalletMethods(c *gin.Context) {
	db := database.GetDB()
	data, err := depositconfig.New(db).GetConfig(h5DepositConfigQuery(c))
	if err != nil {
		h5DepositConfigError(c, err)
		return
	}
	response.Success(c, data)
}
