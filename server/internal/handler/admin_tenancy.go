package handler

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/tenancy"
)

func adminTenancyScope(c *gin.Context) tenancy.Scope {
	actor, _ := adminauth.FromContext(c)
	return tenancy.Scope{TenantID: actor.TenantID, SiteID: actor.SiteID, Platform: actor.Platform}
}
