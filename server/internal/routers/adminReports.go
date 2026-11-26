package routers

import (
	"casino/internal/adminauth"
	"casino/internal/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminReportsRouter(group, handler.NewAdminReportsHandler()) })
}
func adminReportsRouter(group *gin.RouterGroup, h *handler.AdminReportsHandler) {
	g := group.Group("/admin")
	g.Use(adminauth.Middleware())
	g.GET("/reports/financial", adminauth.Require(adminauth.PermissionReportFinancialView), h.Financial)
	g.GET("/reports/financial/detail", adminauth.Require(adminauth.PermissionReportFinancialView), h.FinancialDetail)
	g.GET("/reports/games", adminauth.Require(adminauth.PermissionReportGameView), h.Games)
	g.GET("/reports/games/:id", adminauth.Require(adminauth.PermissionReportGameView), h.GameDetail)
	g.GET("/reports/providers", adminauth.Require(adminauth.PermissionReportProviderView), h.Providers)
	g.GET("/reports/providers/:providerCode", adminauth.Require(adminauth.PermissionReportProviderView), h.ProviderDetail)
	g.GET("/reports/users", adminauth.Require(adminauth.PermissionReportUserView), h.Users)
	g.GET("/reports/users/:id", adminauth.Require(adminauth.PermissionReportUserView), h.UserDetail)
	g.GET("/reports/channels", adminauth.Require(adminauth.PermissionReportChannelView), h.Channels)
	g.GET("/reports/channels/:id", adminauth.Require(adminauth.PermissionReportChannelView), h.ChannelDetail)
	g.GET("/reports/agents", adminauth.Require(adminauth.PermissionReportAgentView), h.Agents)
	g.GET("/reports/agents/:id", adminauth.Require(adminauth.PermissionReportAgentView), h.AgentDetail)
	g.GET("/reports/promotions", adminauth.Require(adminauth.PermissionReportPromotionView), h.Promotions)
	g.GET("/reports/promotions/:id", adminauth.Require(adminauth.PermissionReportPromotionView), h.PromotionDetail)
	g.POST("/report-exports", adminauth.Require(adminauth.PermissionReportExportCreate), h.CreateExport)
	g.GET("/report-exports", adminauth.Require(adminauth.PermissionReportExportList), h.ListExports)
	g.GET("/report-exports/:id", adminauth.Require(adminauth.PermissionReportExportList), h.ExportDetail)
	g.GET("/report-exports/:id/download", adminauth.Require(adminauth.PermissionReportExportDownload), h.DownloadExport)
	g.POST("/report-exports/:id/retry", adminauth.Require(adminauth.PermissionReportExportRetry), h.RetryExport)
	g.POST("/report-exports/:id/cancel", adminauth.Require(adminauth.PermissionReportExportCancel), h.CancelExport)
}
