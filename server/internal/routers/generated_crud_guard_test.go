package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"casino/internal/adminauth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeneratedCRUDGuardBlocksTenantAdminIDOR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		adminauth.StoreActor(c, adminauth.Actor{ID: 11, TenantID: 7, SiteID: 3, Permissions: []string{adminauth.PermissionBetView}})
		c.Next()
	})
	r.Use(generatedCRUDGuard("/api/v1"))
	r.GET("/api/v1/bet-records/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/bet-records/99", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "ADMIN_AUTHENTICATION_REQUIRED")
	require.NotContains(t, w.Body.String(), `"ok":true`)
}

func TestGeneratedCRUDGuardAllowsPlatformSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		adminauth.StoreActor(c, adminauth.Actor{ID: 1, Platform: true, Permissions: []string{"*"}})
		c.Next()
	})
	r.Use(generatedCRUDGuard("/api/v1"))
	r.GET("/api/v1/bet-records/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/bet-records/99", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"ok":true`)
}

func TestGeneratedCRUDGuardAllowsPublicH5GameList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(generatedCRUDGuard("/api/v1"))
	r.GET("/api/v1/games", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/games", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"ok":true`)
}

func TestActivityRoutesDoNotExposeGeneratedWriteCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	gameActivityEventsRouter(r.Group("/api/v1"), activityRoutesTestHandler{})

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	require.True(t, routes[http.MethodGet+" /api/v1/offers"])
	require.True(t, routes[http.MethodGet+" /api/v1/activities/:id"])
	require.True(t, routes[http.MethodPost+" /api/v1/activities/:slug/action"])
	require.False(t, routes[http.MethodPost+" /api/v1/activities/"])
	require.False(t, routes[http.MethodPut+" /api/v1/activities/:id"])
	require.False(t, routes[http.MethodDelete+" /api/v1/activities/:id"])
	require.False(t, routes[http.MethodPost+" /api/v1/activities/list"])
}

type activityRoutesTestHandler struct{}

func (activityRoutesTestHandler) Create(*gin.Context)        {}
func (activityRoutesTestHandler) DeleteByID(*gin.Context)    {}
func (activityRoutesTestHandler) UpdateByID(*gin.Context)    {}
func (activityRoutesTestHandler) GetByID(*gin.Context)       {}
func (activityRoutesTestHandler) List(*gin.Context)          {}
func (activityRoutesTestHandler) Offers(*gin.Context)        {}
func (activityRoutesTestHandler) GetByIDOrSlug(*gin.Context) {}
func (activityRoutesTestHandler) Action(*gin.Context)        {}
