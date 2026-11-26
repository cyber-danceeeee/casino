package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameBonusCenterRouter(group, handler.NewGameBonusCenterHandler())
	})
}

func gameBonusCenterRouter(group *gin.RouterGroup, h handler.GameBonusCenterHandler) {
	group.GET("/bonus-center/records", h.Records)
}
