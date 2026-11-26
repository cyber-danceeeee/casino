package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestExpensiveAdminRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expensiveAdminRate.Lock()
	expensiveAdminRate.buckets = make(map[string]adminRateBucket)
	expensiveAdminRate.Unlock()
	for attempt := 1; attempt <= 3; attempt++ {
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = httptest.NewRequest("GET", "/api/v1/admin/support/players", nil)
		context.Request.RemoteAddr = "192.0.2.1:1234"
		allowed := allowExpensiveAdminRequest(context, "test", 2, time.Minute)
		if attempt <= 2 && !allowed {
			t.Fatalf("request %d was limited too early", attempt)
		}
		if attempt == 3 && (allowed || response.Body.Len() == 0) {
			t.Fatal("request above the limit was not rejected with an API response")
		}
	}
}
