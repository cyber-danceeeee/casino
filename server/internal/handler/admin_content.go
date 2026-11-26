package handler

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/content"
	"casino/internal/singlesite"
)

func adminContentScope(c *gin.Context) content.Scope {
	actor, _ := adminauth.FromContext(c)
	if singlesite.CatalogResource(c.Request.URL.Path) {
		return content.Scope{Platform: true}
	}
	return content.Scope{SiteID: actor.SiteID, Platform: actor.Platform}
}
