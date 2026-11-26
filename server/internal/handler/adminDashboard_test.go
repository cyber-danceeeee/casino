package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"casino/internal/content"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDashboardCacheKeyIncludesH5Scope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := dashboardTestContext("/api/v1/admin/dashboard/overview?currency=PKR&limit=5")

	h5A := dashboardCacheKey(ctx, content.Scope{SiteID: 3})
	h5B := dashboardCacheKey(ctx, content.Scope{SiteID: 5})
	siteB := dashboardCacheKey(ctx, content.Scope{SiteID: 4})
	platform := dashboardCacheKey(ctx, content.Scope{Platform: true})

	require.NotEqual(t, h5A, h5B)
	require.NotEqual(t, h5A, siteB)
	require.NotEqual(t, h5A, platform)
	require.Contains(t, h5A, "|site=3|platform=false")
	require.Contains(t, platform, "|site=0|platform=true")
}

func dashboardTestContext(target string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}
