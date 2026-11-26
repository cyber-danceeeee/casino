package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		h := handler.NewGamePushDeviceHandler()
		group.POST("/push/devices", h.Register)
		group.DELETE("/push/devices", h.Unregister)
	})
}
