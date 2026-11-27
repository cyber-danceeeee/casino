package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameProviderAggregatorRouter(group, handler.NewGameProviderAggregatorHandler())
	})
}

func gameProviderAggregatorRouter(group *gin.RouterGroup, h handler.GameProviderAggregatorHandler) {
	group.GET("/game/providers/:providerCode/gamelist", h.GameList)
	group.POST("/games/:id/:gameId/launch", h.Launch)
	group.POST("/games/:id/:gameId/launch/demo", h.LaunchDemo)
	group.POST("/game/session/:sessionId/heartbeat", h.HeartbeatSession)
	group.POST("/game/session/:sessionId/exit", h.ExitSession)
	group.GET("/game/providers/:providerCode/callback", h.Callback)
	group.GET("/game/providers/:providerCode/callbacks/:operation", h.Callback)
	group.POST("/game/providers/:providerCode/callback", h.Callback)
	group.POST("/game/providers/:providerCode/callback/api/:operation", h.Callback)
	group.POST("/game/providers/:providerCode/callbacks/:operation", h.Callback)
	group.POST("/game/providers/:providerCode/callbacks/wallet/:operation", h.Callback)
	group.POST("/game/providers/:providerCode/callbacks/player/info", h.Callback)

	admin := group.Group("/admin/game/providers", adminauth.Middleware(), adminauth.Require(adminauth.PermissionGameProviderSync))
	admin.POST("/:providerCode/sync", h.SyncProviderGames)
}
