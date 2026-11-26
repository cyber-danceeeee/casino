package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
)

type gameUserSettingsUpdateRequest struct {
	Language     *string `json:"language"`
	MusicEnabled *int    `json:"musicEnabled"`
	PushEnabled  *int    `json:"pushEnabled"`
}

func (h *gameUserSettingsHandler) Current(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(c, h5FindSettings(db, user.ID))
}

func (h *gameUserSettingsHandler) UpdateCurrent(c *gin.Context) {
	req := &gameUserSettingsUpdateRequest{}
	_ = c.ShouldBindJSON(req)

	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	settings := h5FindSettings(db, user.ID)
	if settings.UserID == 0 {
		settings.UserID = user.ID
	}
	if req.Language != nil {
		settings.Language = *req.Language
	}
	if req.MusicEnabled != nil {
		settings.MusicEnabled = *req.MusicEnabled
	}
	if req.PushEnabled != nil {
		settings.PushEnabled = *req.PushEnabled
	}
	if settings.ID == 0 {
		_ = db.Create(&settings).Error
	} else {
		_ = db.Save(&settings).Error
	}
	response.Success(c, settings)
}
