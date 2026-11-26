package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		gameActivityEventsRouter(group, handler.NewGameActivityEventsHandler())
	})
}

func gameActivityEventsRouter(group *gin.RouterGroup, h handler.GameActivityEventsHandler) {
	group.GET("/offers", h.Offers)

	g := group.Group("/activities")
	g.GET("/:id", h.GetByIDOrSlug) // [get] /api/v1/activities/:id or /api/v1/activities/:slug
	g.POST("/:slug/action", h.Action)
}
