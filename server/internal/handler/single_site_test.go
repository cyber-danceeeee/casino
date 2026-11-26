package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestH5ConfigurationIsFixed(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "http://example.test/api/v1/h5/bootstrap?siteID=999", nil)
	if scope := h5ContentScope(c); scope.SiteID != 1 {
		t.Fatalf("unexpected H5 config: %+v", scope)
	}
}
