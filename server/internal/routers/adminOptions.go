package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		adminOptionsRouter(group, handler.NewAdminOptionsHandler())
	})
}

func adminOptionsRouter(group *gin.RouterGroup, h handler.AdminOptionsHandler) {
	g := group.Group("/admin")
	g.Use(adminauth.Middleware())
	g.GET("/options", h.List)
}
