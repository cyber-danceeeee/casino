package handler

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/singlesite"
	"casino/internal/tenancy"
)

func adminTenancyScope(c *gin.Context) tenancy.Scope {
	actor, _ := adminauth.FromContext(c)
	if singlesite.Active() && singlesite.CatalogResource(c.Request.URL.Path) {
		return tenancy.Scope{Platform: true}
	}
	return tenancy.Scope{TenantID: actor.TenantID, SiteID: actor.SiteID, Platform: actor.Platform}
}
