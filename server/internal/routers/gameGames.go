package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameGamesRouter(group, handler.NewGameGamesHandler())
	})
}

func gameGamesRouter(group *gin.RouterGroup, h handler.GameGamesHandler) {
	group.GET("/bootstrap", h.Bootstrap)
	group.GET("/games", h.PublicList)

	g := group.Group("/games")

	// JWT authentication reference: https://go-sponge.com/component/transport/gin.html#jwt-authorization-middleware

	// All the following routes use jwt authentication, you also can use middleware.Auth(middleware.WithExtraVerify(fn))
	//g.Use(middleware.Auth())

	// If jwt authentication is not required for all routes, authentication middleware can be added
	// separately for only certain routes. In this case, g.Use(middleware.Auth()) above should not be used.

	g.POST("/:id/launch", h.Launch) // [post] /api/v1/games/:id/launch
	g.GET("/:id/favorite", h.FavoriteStatus)
	g.POST("/:id/favorite", h.ToggleFavorite)
	g.POST("/", h.Create)          // [post] /api/v1/gameGames
	g.DELETE("/:id", h.DeleteByID) // [delete] /api/v1/gameGames/:id
	g.PUT("/:id", h.UpdateByID)    // [put] /api/v1/gameGames/:id
	g.GET("/:id", h.PublicDetail)  // [get] /api/v1/games/:id
	g.POST("/list", h.List)        // [post] /api/v1/gameGames/list
}
