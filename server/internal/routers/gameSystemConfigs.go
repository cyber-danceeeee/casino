package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameSystemConfigsRouter(group, handler.NewGameSystemConfigsHandler())
	})
}

func gameSystemConfigsRouter(group *gin.RouterGroup, h handler.GameSystemConfigsHandler) {
	g := group.Group("/system-configs")

	// JWT authentication reference: https://go-sponge.com/component/transport/gin.html#jwt-authorization-middleware

	// All the following routes use jwt authentication, you also can use middleware.Auth(middleware.WithExtraVerify(fn))
	//g.Use(middleware.Auth())

	// If jwt authentication is not required for all routes, authentication middleware can be added
	// separately for only certain routes. In this case, g.Use(middleware.Auth()) above should not be used.

	g.POST("/", h.Create)          // [post] /api/v1/gameSystemConfigs
	g.DELETE("/:id", h.DeleteByID) // [delete] /api/v1/gameSystemConfigs/:id
	g.PUT("/:id", h.UpdateByID)    // [put] /api/v1/gameSystemConfigs/:id
	g.GET("/:id", h.GetByID)       // [get] /api/v1/gameSystemConfigs/:id
	g.POST("/list", h.List)        // [post] /api/v1/gameSystemConfigs/list
}
