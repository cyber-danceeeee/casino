package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameAppVersionsHandler) Current(c *gin.Context) {
	platform := strings.TrimSpace(c.Query("platform"))
	if platform == "" {
		platform = "android"
	}
	var app model.GameAppVersions
	err := database.GetDB().Where("status = ? AND platform = ?", 1, platform).
		Order("version_code DESC, id DESC").
		First(&app).Error
	if err != nil {
		response.Success(c, gin.H{
			"platform":    platform,
			"versionName": "1.0.0",
			"versionCode": 1,
			"downloadUrl": "",
			"forceUpdate": 0,
		})
		return
	}
	response.Success(c, app)
}
