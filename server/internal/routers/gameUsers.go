package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameUsersRouter(group, handler.NewGameUsersHandler())
	})
}

func gameUsersRouter(group *gin.RouterGroup, h handler.GameUsersHandler) {
	auth := group.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/register", h.Register)
	auth.POST("/logout", h.Logout)

	group.GET("/me", h.Me)
	group.GET("/profile", h.Profile)
	group.GET("/rewards/pending", h.PendingRewards)
	group.POST("/rewards/register/claim", h.ClaimRegisterReward)
	group.POST("/rewards/:id/claim", h.ClaimPendingReward)

	g := group.Group("/users")

	// JWT authentication reference: https://go-sponge.com/component/transport/gin.html#jwt-authorization-middleware

	// All the following routes use jwt authentication, you also can use middleware.Auth(middleware.WithExtraVerify(fn))
	//g.Use(middleware.Auth())

	// If jwt authentication is not required for all routes, authentication middleware can be added
	// separately for only certain routes. In this case, g.Use(middleware.Auth()) above should not be used.

	g.POST("/", h.Create)          // [post] /api/v1/gameUsers
	g.DELETE("/:id", h.DeleteByID) // [delete] /api/v1/gameUsers/:id
	g.PUT("/:id", h.UpdateByID)    // [put] /api/v1/gameUsers/:id
	g.GET("/:id", h.GetByID)       // [get] /api/v1/gameUsers/:id
	g.POST("/list", h.List)        // [post] /api/v1/gameUsers/list
}
