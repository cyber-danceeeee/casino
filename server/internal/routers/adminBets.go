package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminBetsRouter(group, handler.NewAdminBetsHandler()) })
}

func adminBetsRouter(group *gin.RouterGroup, h handler.AdminBetsHandler) {
	g := group.Group("/admin/betting")
	g.Use(adminauth.Middleware())
	g.GET("/bets", adminauth.Require(adminauth.PermissionBetList), h.ListBets)
	g.GET("/bets/:id", adminauth.Require(adminauth.PermissionBetView), h.GetBet)
	g.POST("/bets/:id/review", adminauth.Require(adminauth.PermissionBetReview), h.MarkBetForReview)
	g.GET("/rounds", adminauth.Require(adminauth.PermissionBetRoundList), h.ListRounds)
	g.GET("/rounds/:id", adminauth.Require(adminauth.PermissionBetRoundView), h.GetRound)
	g.GET("/rollbacks", adminauth.Require(adminauth.PermissionBetRollbackList), h.ListRollbacks)
	g.GET("/rollbacks/:id", adminauth.Require(adminauth.PermissionBetRollbackView), h.GetRollback)
	g.POST("/rollbacks/:id/investigate", adminauth.Require(adminauth.PermissionBetRollbackReview), h.MarkRollbackForInvestigation)
	g.GET("/abnormal-bets", adminauth.Require(adminauth.PermissionBetAbnormalList), h.ListAbnormal)
	g.GET("/abnormal-bets/:id", adminauth.Require(adminauth.PermissionBetAbnormalView), h.GetAbnormal)
	g.POST("/abnormal-bets/:id/assign", adminauth.Require(adminauth.PermissionBetAbnormalAssign), h.AssignAbnormal)
	g.POST("/abnormal-bets/:id/review", adminauth.Require(adminauth.PermissionBetAbnormalReview), h.ReviewAbnormal)
	g.POST("/abnormal-bets/:id/clear", adminauth.Require(adminauth.PermissionBetAbnormalClear), h.ClearAbnormal)
	g.POST("/abnormal-bets/:id/restrict", adminauth.Require(adminauth.PermissionBetAbnormalRestrict), h.RestrictAbnormal)
	g.POST("/abnormal-bets/:id/notes", adminauth.Require(adminauth.PermissionBetAbnormalNote), h.AddAbnormalNote)
	g.GET("/provider-callback-logs", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetCallbackLogList), h.ListCallbacks)
	g.GET("/provider-callback-logs/:id", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetCallbackLogView), h.GetCallback)
	g.POST("/provider-callback-logs/:id/investigate", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetCallbackLogView), h.MarkCallbackForInvestigation)
	g.GET("/provider-reconciliation/runs", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetReconciliationList), h.ListReconciliationRuns)
	g.POST("/provider-reconciliation/runs", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetReconciliationGenerate), h.GenerateReconciliationRun)
	g.GET("/provider-reconciliation/runs/:id", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetReconciliationView), h.GetReconciliationRun)
	g.GET("/provider-reconciliation/items", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetReconciliationList), h.ListReconciliationItems)
	g.POST("/provider-reconciliation/items/:id/status", adminauth.RequirePlatform(), adminauth.Require(adminauth.PermissionBetReconciliationResolve), h.UpdateReconciliationItemStatus)
}
