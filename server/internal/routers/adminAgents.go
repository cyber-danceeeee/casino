package routers

import (
	"casino/internal/adminauth"
	"casino/internal/handler"
	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminAgentsRouter(group, handler.NewAdminAgentsHandler()) })
}
func adminAgentsRouter(group *gin.RouterGroup, h *handler.AdminAgentsHandler) {
	g := group.Group("/admin")
	g.Use(adminauth.Middleware())
	g.GET("/agents", adminauth.Require(adminauth.PermissionAgentList), h.ListAgents)
	g.GET("/agents/:id", adminauth.Require(adminauth.PermissionAgentView), h.GetAgent)
	g.POST("/agents", adminauth.Require(adminauth.PermissionAgentCreate), h.CreateAgent)
	g.PATCH("/agents/:id", adminauth.Require(adminauth.PermissionAgentUpdate), h.UpdateAgent)
	g.PATCH("/agents/:id/status", adminauth.Require(adminauth.PermissionAgentStatus), h.AgentStatus)
	g.GET("/agents/:id/users", adminauth.Require(adminauth.PermissionAgentView), h.AgentUsers)
	g.GET("/agents/:id/children", adminauth.Require(adminauth.PermissionAgentHierarchyView), h.AgentChildren)
	g.GET("/agent-hierarchy", adminauth.Require(adminauth.PermissionAgentHierarchyView), h.Hierarchy)
	g.POST("/agents/:id/change-parent", adminauth.Require(adminauth.PermissionAgentHierarchyUpdate), h.ChangeParent)
	g.POST("/agents/:id/assign-plan", adminauth.Require(adminauth.PermissionAgentPlanAssign), h.AssignPlan)
	g.POST("/agent-attributions/change", adminauth.Require(adminauth.PermissionAgentAttributionUpdate), h.ChangeAttribution)
	g.GET("/agent-channels", adminauth.Require(adminauth.PermissionAgentChannelList), h.ListChannels)
	g.GET("/agent-channels/:id", adminauth.Require(adminauth.PermissionAgentChannelList), h.GetChannel)
	g.POST("/agent-channels", adminauth.Require(adminauth.PermissionAgentChannelCreate), h.CreateChannel)
	g.PATCH("/agent-channels/:id", adminauth.Require(adminauth.PermissionAgentChannelUpdate), h.UpdateChannel)
	g.PATCH("/agent-channels/:id/status", adminauth.Require(adminauth.PermissionAgentChannelStatus), h.ChannelStatus)
	g.GET("/commission-plans", adminauth.Require(adminauth.PermissionAgentCommissionPlanList), h.ListPlans)
	g.GET("/commission-plans/:id", adminauth.Require(adminauth.PermissionAgentCommissionPlanList), h.GetPlan)
	g.POST("/commission-plans", adminauth.Require(adminauth.PermissionAgentCommissionPlanCreate), h.CreatePlan)
	g.PATCH("/commission-plans/:id", adminauth.Require(adminauth.PermissionAgentCommissionPlanUpdate), h.UpdatePlan)
	g.PATCH("/commission-plans/:id/status", adminauth.Require(adminauth.PermissionAgentCommissionPlanUpdate), h.PlanStatus)
	g.DELETE("/commission-plans/:id", adminauth.Require(adminauth.PermissionAgentCommissionPlanDelete), h.DeletePlan)
	g.GET("/commission-records", adminauth.Require(adminauth.PermissionAgentCommissionRecordList), h.ListCommissions)
	g.GET("/commission-records/:id", adminauth.Require(adminauth.PermissionAgentCommissionRecordView), h.GetCommission)
	g.GET("/agent-settlements", adminauth.Require(adminauth.PermissionAgentSettlementList), h.ListSettlements)
	g.GET("/agent-settlements/:id", adminauth.Require(adminauth.PermissionAgentSettlementView), h.GetSettlement)
	g.GET("/agent-settlements/:id/export", adminauth.Require(adminauth.PermissionAgentSettlementExport), h.ExportSettlement)
	g.POST("/agent-settlements/generate", adminauth.Require(adminauth.PermissionAgentSettlementGenerate), h.GenerateSettlement)
	g.POST("/agent-settlements/:id/submit", adminauth.Require(adminauth.PermissionAgentSettlementSubmit), h.SettlementAction("submit"))
	g.POST("/agent-settlements/:id/approve", adminauth.Require(adminauth.PermissionAgentSettlementApprove), h.SettlementAction("approve"))
	g.POST("/agent-settlements/:id/reject", adminauth.Require(adminauth.PermissionAgentSettlementReject), h.SettlementAction("reject"))
	g.POST("/agent-settlements/:id/mark-paid", adminauth.Require(adminauth.PermissionAgentSettlementPay), h.SettlementAction("mark-paid"))
	g.POST("/agent-settlements/:id/cancel", adminauth.Require(adminauth.PermissionAgentSettlementCancel), h.SettlementAction("cancel"))
	g.GET("/agent-wallets", adminauth.Require(adminauth.PermissionAgentWalletList), h.ListWallets)
	g.GET("/agent-wallets/:id", adminauth.Require(adminauth.PermissionAgentWalletView), h.GetWallet)
	g.GET("/agent-wallet-transactions", adminauth.Require(adminauth.PermissionAgentWalletView), h.ListWalletTransactions)
	g.POST("/agent-wallets/:id/freeze", adminauth.Require(adminauth.PermissionAgentWalletFreeze), h.WalletFrozen(true))
	g.POST("/agent-wallets/:id/unfreeze", adminauth.Require(adminauth.PermissionAgentWalletFreeze), h.WalletFrozen(false))
	g.POST("/agent-wallets/:id/adjustments", adminauth.Require(adminauth.PermissionAgentWalletAdjust), h.WalletAdjustment)
	g.GET("/agent-operation-logs", adminauth.Require(adminauth.PermissionAgentOperationLogView), h.ListLogs)
	g.GET("/agent-operation-logs/:id", adminauth.Require(adminauth.PermissionAgentOperationLogView), h.GetLog)
}
