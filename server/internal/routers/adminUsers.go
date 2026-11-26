package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		adminUsersRouter(group, handler.NewAdminUsersHandler())
	})
}

func adminUsersRouter(group *gin.RouterGroup, h handler.AdminUsersHandler) {
	g := group.Group("/admin/users")
	g.Use(adminauth.Middleware())
	g.GET("/me", h.Me)
	g.GET("", adminauth.Require(adminauth.PermissionUserList), h.ListUsers)
	g.GET("/", adminauth.Require(adminauth.PermissionUserList), h.ListUsers)
	g.GET("/login-records", adminauth.Require(adminauth.PermissionLoginRecordView), h.ListLoginRecords)
	g.GET("/devices", adminauth.Require(adminauth.PermissionDeviceView), h.ListDevices)
	g.GET("/push-devices", adminauth.Require(adminauth.PermissionDeviceView), h.ListPushDevices)
	g.GET("/devices/:id", adminauth.Require(adminauth.PermissionDeviceView), h.GetDevice)
	g.GET("/devices/:id/users", adminauth.Require(adminauth.PermissionDeviceView), h.ListDeviceUsers)
	g.PATCH("/devices/:id/status", adminauth.Require(adminauth.PermissionDeviceManage), h.UpdateDeviceStatus)
	g.GET("/risk", adminauth.Require(adminauth.PermissionUserView), h.ListRiskUsers)
	g.GET("/risk/:id", adminauth.Require(adminauth.PermissionUserView), h.GetRiskUser)
	g.GET("/:id", adminauth.Require(adminauth.PermissionUserView), h.GetUser)
	g.GET("/:id/notes", adminauth.Require(adminauth.PermissionUserView), h.ListUserNotes)
	g.POST("/:id/notes", adminauth.Require(adminauth.PermissionUserNoteCreate), h.AddUserNote)
	g.PATCH("/:id/tags", adminauth.Require(adminauth.PermissionUserNoteCreate), h.UpdateUserTags)
	g.GET("/:id/operation-logs", adminauth.Require(adminauth.PermissionUserView), h.ListUserOperationLogs)
	g.PATCH("/:id/status", adminauth.Require(adminauth.PermissionUserStatusUpdate), h.UpdateUserStatus)
	g.PATCH("/:id/restrictions", adminauth.Require(adminauth.PermissionUserRestrictionUpdate), h.UpdateUserRestrictions)
	g.PATCH("/:id/risk-status", adminauth.Require(adminauth.PermissionUserRiskUpdate), h.UpdateUserRisk)
}
