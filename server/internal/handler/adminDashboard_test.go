package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"casino/internal/tenancy"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDashboardCacheKeyIncludesTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := dashboardTestContext("/api/v1/admin/dashboard/overview?currency=PKR&limit=5")

	tenantA := dashboardCacheKey(ctx, tenancy.Scope{TenantID: 7, SiteID: 3})
	tenantB := dashboardCacheKey(ctx, tenancy.Scope{TenantID: 8, SiteID: 3})
	siteB := dashboardCacheKey(ctx, tenancy.Scope{TenantID: 7, SiteID: 4})
	platform := dashboardCacheKey(ctx, tenancy.Scope{Platform: true})

	require.NotEqual(t, tenantA, tenantB)
	require.NotEqual(t, tenantA, siteB)
	require.NotEqual(t, tenantA, platform)
	require.Contains(t, tenantA, "|tenant=7|site=3|platform=false")
	require.Contains(t, platform, "|tenant=0|site=0|platform=true")
}

func dashboardTestContext(target string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}
