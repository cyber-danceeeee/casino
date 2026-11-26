package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		group.GET("/announcements", handler.PublicAnnouncements)
		group.GET("/announcements/:id", handler.PublicAnnouncement)
		group.GET("/popups", handler.PublicPopups)
		group.POST("/popups/:id/impression", handler.TrackPopupImpression)
		group.POST("/popups/:id/click", handler.TrackPopupClick)
		group.POST("/popups/:id/dismiss", handler.TrackPopupDismiss)
	})
}
