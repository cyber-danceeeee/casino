package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"casino/internal/gameadmin"
)

func TestGameBatchInputBindingAcceptsAddCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest("PATCH", "/api/v1/admin/games/batch", strings.NewReader(`{
		"ids":[15107],
		"addCategoryIds":[999999999],
		"reason":"batch category binding test"
	}`))
	request.Header.Set("Content-Type", "application/json")
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request
	var input gameadmin.GameBatchInput
	if err := context.ShouldBindJSON(&input); err != nil {
		t.Fatalf("valid add-category batch payload was rejected: %v", err)
	}
	if len(input.IDs) != 1 || len(*input.AddCategoryIDs) != 1 || (*input.AddCategoryIDs)[0] != 999999999 {
		t.Fatalf("unexpected bound input: %+v", input)
	}
}
