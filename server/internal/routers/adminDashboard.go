package routers

import (
	"casino/internal/adminauth"
	"casino/internal/handler"
	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminDashboardRouter(group, handler.NewAdminDashboardHandler()) })
}

func adminDashboardRouter(group *gin.RouterGroup, h *handler.AdminDashboardHandler) {
	g := group.Group("/admin/dashboard")
	g.Use(adminauth.Middleware())
	g.GET("/filters", adminauth.Require(adminauth.PermissionDashboardView), h.Filters)
	g.GET("/overview", adminauth.Require(adminauth.PermissionDashboardView), h.Overview)
	g.GET("/financial-trend", adminauth.Require(adminauth.PermissionDashboardFinancialView), h.FinancialTrend)
	g.GET("/betting-trend", adminauth.Require(adminauth.PermissionDashboardGameView), h.BettingTrend)
	g.GET("/user-trend", adminauth.Require(adminauth.PermissionDashboardUserView), h.UserTrend)
	g.GET("/user-funnel", adminauth.Require(adminauth.PermissionDashboardUserView), h.UserFunnel)
	g.GET("/order-status-distribution", adminauth.Require(adminauth.PermissionDashboardFinancialView), h.OrderStatus)
	g.GET("/game-category-distribution", adminauth.Require(adminauth.PermissionDashboardGameView), h.GameCategories)
	g.GET("/top-games", adminauth.Require(adminauth.PermissionDashboardGameView), h.TopGames)
	g.GET("/provider-performance", adminauth.Require(adminauth.PermissionDashboardProviderView), h.ProviderPerformance)
	g.GET("/country-distribution", adminauth.Require(adminauth.PermissionDashboardUserView), h.CountryDistribution)
	g.GET("/channel-performance", adminauth.Require(adminauth.PermissionDashboardChannelView), h.ChannelPerformance)
	g.GET("/agent-performance", adminauth.Require(adminauth.PermissionDashboardAgentView), h.AgentPerformance)
	g.GET("/promotion-performance", adminauth.Require(adminauth.PermissionDashboardPromotionView), h.PromotionPerformance)
	g.GET("/risk-overview", adminauth.Require(adminauth.PermissionDashboardRiskView), h.RiskOverview)
	g.GET("/realtime", adminauth.Require(adminauth.PermissionDashboardRealtimeView), h.Realtime)
	g.GET("/alerts", adminauth.Require(adminauth.PermissionDashboardAlertView), h.Alerts)
}
