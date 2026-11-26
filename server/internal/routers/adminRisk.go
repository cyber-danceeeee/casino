package routers

import (
	"casino/internal/adminauth"
	"casino/internal/handler"
	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminRiskRouter(group, handler.NewAdminRiskHandler()) })
}
func adminRiskRouter(group *gin.RouterGroup, h *handler.AdminRiskHandler) {
	g := group.Group("/admin/risk")
	g.Use(adminauth.Middleware())
	g.GET("/dashboard", adminauth.Require(adminauth.PermissionRiskDashboardView), h.Dashboard)
	g.GET("/rules", adminauth.Require(adminauth.PermissionRiskRuleList), h.ListRules)
	g.GET("/rules/:id", adminauth.Require(adminauth.PermissionRiskRuleList), h.GetRule)
	g.POST("/rules", adminauth.Require(adminauth.PermissionRiskRuleCreate), h.CreateRule)
	g.PUT("/rules/:id", adminauth.Require(adminauth.PermissionRiskRuleUpdate), h.UpdateRule)
	g.POST("/rules/:id/copy", adminauth.Require(adminauth.PermissionRiskRuleCreate), h.CopyRule)
	g.POST("/rules/:id/test", adminauth.Require(adminauth.PermissionRiskRuleTest), h.TestRule)
	g.PATCH("/rules/:id/status", adminauth.Require(adminauth.PermissionRiskRuleStatus), h.RuleStatus)
	g.DELETE("/rules/:id", adminauth.Require(adminauth.PermissionRiskRuleDelete), h.DeleteRule)
	g.GET("/events", adminauth.Require(adminauth.PermissionRiskEventList), h.ListEvents)
	g.GET("/events/:id", adminauth.Require(adminauth.PermissionRiskEventView), h.GetEvent)
	g.POST("/events", adminauth.Require(adminauth.PermissionRiskEventReview), h.CreateEvent)
	g.POST("/events/:id/assign", adminauth.Require(adminauth.PermissionRiskEventAssign), h.AssignEvent)
	g.POST("/events/:id/review", adminauth.Require(adminauth.PermissionRiskEventReview), h.EventAction("UNDER_REVIEW", adminauth.PermissionRiskEventReview))
	g.POST("/events/:id/confirm", adminauth.Require(adminauth.PermissionRiskEventReview), h.EventAction("CONFIRMED", adminauth.PermissionRiskEventReview))
	g.POST("/events/:id/clear", adminauth.Require(adminauth.PermissionRiskEventClear), h.EventAction("CLEARED", adminauth.PermissionRiskEventClear))
	g.POST("/events/:id/restrict", adminauth.Require(adminauth.PermissionRiskEventRestrict), h.EventAction("RESTRICTED", adminauth.PermissionRiskEventRestrict))
	g.POST("/events/:id/block", adminauth.Require(adminauth.PermissionRiskEventBlock), h.EventAction("BLOCKED", adminauth.PermissionRiskEventBlock))
	g.POST("/events/:id/notes", adminauth.Require(adminauth.PermissionRiskEventNote), h.EventNote)
	g.GET("/reviews", adminauth.Require(adminauth.PermissionRiskReviewList), h.ListReviews)
	g.GET("/reviews/:id", adminauth.Require(adminauth.PermissionRiskReviewView), h.GetReview)
	g.POST("/reviews/:id/claim", adminauth.Require(adminauth.PermissionRiskReviewClaim), h.ClaimReview)
	g.POST("/reviews/:id/assign", adminauth.Require(adminauth.PermissionRiskReviewAssign), h.AssignReview)
	g.POST("/reviews/:id/decision", adminauth.RequireAny(adminauth.PermissionRiskReviewDecision, adminauth.PermissionRiskReviewEscalate), h.DecideReview)
	g.POST("/reviews/:id/notes", adminauth.Require(adminauth.PermissionRiskReviewDecision), h.ReviewNote)
	g.GET("/associations", adminauth.Require(adminauth.PermissionRiskAssociationList), h.ListAssociations)
	g.GET("/associations/:userId", adminauth.Require(adminauth.PermissionRiskAssociationView), h.GetAssociations)
	g.GET("/blacklists", adminauth.Require(adminauth.PermissionRiskBlacklistList), h.ListBlacklists)
	g.GET("/blacklists/:id", adminauth.Require(adminauth.PermissionRiskBlacklistList), h.GetBlacklist)
	g.POST("/blacklists", adminauth.Require(adminauth.PermissionRiskBlacklistCreate), h.CreateBlacklist)
	g.PUT("/blacklists/:id", adminauth.Require(adminauth.PermissionRiskBlacklistUpdate), h.UpdateBlacklist)
	g.DELETE("/blacklists/:id", adminauth.Require(adminauth.PermissionRiskBlacklistRemove), h.RemoveBlacklist)
	g.GET("/withdrawals", adminauth.Require(adminauth.PermissionRiskWithdrawalList), h.ListWithdrawals)
	g.GET("/withdrawals/:id", adminauth.Require(adminauth.PermissionRiskWithdrawalView), h.GetWithdrawal)
	g.POST("/withdrawals/:id/decision", adminauth.RequireAny(adminauth.PermissionRiskWithdrawalApprove, adminauth.PermissionRiskWithdrawalReject, adminauth.PermissionRiskWithdrawalRestrict), h.DecideWithdrawal)
	g.GET("/operation-logs", adminauth.Require(adminauth.PermissionRiskOperationLogView), h.ListOperationLogs)
	g.GET("/operation-logs/:id", adminauth.Require(adminauth.PermissionRiskOperationLogView), h.GetOperationLog)
}
