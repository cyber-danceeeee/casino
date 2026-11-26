package adminauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestActorPermission(t *testing.T) {
	actor := Actor{Permissions: []string{PermissionUserView}}
	assert.True(t, actor.Can(PermissionUserView))
	assert.False(t, actor.Can(PermissionUserStatusUpdate))
	assert.True(t, Actor{Permissions: []string{"*"}}.Can(PermissionUserStatusUpdate))
}

func TestAllUserCenterPermissionsReturnsCopy(t *testing.T) {
	permissions := AllUserCenterPermissions()
	permissions[0] = "changed"
	assert.NotEqual(t, "changed", AllUserCenterPermissions()[0])
}

func TestRequireRejectsUnauthorizedActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", func(c *gin.Context) {
		c.Set(actorContextKey, Actor{ID: 7, Username: "readonly", Permissions: []string{PermissionUserView}})
	}, Require(PermissionUserStatusUpdate), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"code":403,"msg":"permission denied","data":{"errorCode":"PERMISSION_DENIED"}}`, recorder.Body.String())
}

func TestRequireAllowsAuthorizedActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", func(c *gin.Context) {
		c.Set(actorContextKey, Actor{ID: 8, Username: "operator", Permissions: []string{PermissionUserStatusUpdate}})
	}, Require(PermissionUserStatusUpdate), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"code":0}`, recorder.Body.String())
}

func TestRequirePlatformRejectsTenantActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/platform", func(c *gin.Context) {
		c.Set(actorContextKey, Actor{ID: 8, TenantID: 100, AdminType: "TENANT"})
	}, RequirePlatform(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/platform", nil)
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "PLATFORM_ACCESS_REQUIRED")
}
