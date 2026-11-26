package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameVipLevelsHandler) Summary(c *gin.Context) {
	db := database.GetDB()
	user, _ := h5CurrentUser(db, c)
	scope := h5ContentScope(c)

	var levels []model.GameVipLevels
	_ = content.ApplySite(db.Model(&model.GameVipLevels{}), "", scope).Where("status = ?", 1).Order("level_no ASC").Find(&levels).Error

	var requirements []model.GameVipRequirements
	_ = content.ApplySite(db.Model(&model.GameVipRequirements{}), "", scope).Order("target_level_no ASC, requirement_type ASC").Find(&requirements).Error

	response.Success(c, gin.H{
		"currentLevel": user.VipLevel,
		"levels":       h5VipLevelsPayload(levels),
		"requirements": requirements,
		"benefits":     h5ConfigJSONAssets(c, db, "h5.vip.benefits", []gin.H{}),
	})
}
