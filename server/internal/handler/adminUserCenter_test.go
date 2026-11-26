package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdminUserStatusValues(t *testing.T) {
	tests := []struct {
		input string
		want  int
		ok    bool
	}{{"ACTIVE", 1, true}, {"DISABLED", 2, true}, {"FROZEN", 3, true}, {"LOCKED", 4, true}, {"CLOSED", 5, true}, {"invalid", 0, false}}
	for _, test := range tests {
		got, ok := adminUserStatusValue(test.input)
		assert.Equal(t, test.want, got)
		assert.Equal(t, test.ok, ok)
	}
}

func TestAdminUserStatusTransitions(t *testing.T) {
	assert.True(t, adminValidUserStatusTransition(1, 2))
	assert.True(t, adminValidUserStatusTransition(2, 1))
	assert.True(t, adminValidUserStatusTransition(1, 3))
	assert.True(t, adminValidUserStatusTransition(3, 1))
	assert.False(t, adminValidUserStatusTransition(1, 1))
	assert.False(t, adminValidUserStatusTransition(5, 1))
	assert.False(t, adminValidUserStatusTransition(0, 1))
}

func TestAdminRiskEnums(t *testing.T) {
	for _, value := range []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"} {
		assert.True(t, adminValidRiskLevel(value))
	}
	assert.False(t, adminValidRiskLevel("UNKNOWN"))
	for _, value := range []string{"OBSERVED", "UNDER_REVIEW", "RESTRICTED", "CLEARED", "BLOCKED"} {
		assert.True(t, adminValidRiskStatus(value))
	}
	assert.False(t, adminValidRiskStatus("UNKNOWN"))
}

func TestAdminSensitiveMasks(t *testing.T) {
	assert.Equal(t, "30******88", adminMaskPhone("3088888888"))
	assert.Equal(t, "a***@example.com", adminMaskEmail("admin@example.com"))
	assert.Equal(t, "192.168.*.*", adminMaskIP("192.168.1.20"))
	assert.Equal(t, "abcdef...uvwxyz", adminMaskHash("abcdef000000uvwxyz"))
}

func TestAdminDateParsing(t *testing.T) {
	start, ok := adminParseDate("2026-07-15", false)
	assert.True(t, ok)
	assert.Equal(t, 0, start.Hour())
	end, ok := adminParseDate("2026-07-15", true)
	assert.True(t, ok)
	assert.Equal(t, 23, end.Hour())
	assert.Equal(t, 59, end.Minute())
	_, ok = adminParseDate("not-a-date", false)
	assert.False(t, ok)
}

func TestAdminListDTOExcludesCredentials(t *testing.T) {
	payload, err := json.Marshal(adminUserRow{ID: 1, UserNo: "U1"})
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "password")
	assert.NotContains(t, string(payload), "token")
}

func TestAdminStableDefaultOrdering(t *testing.T) {
	assert.Equal(t, "ACTIVE", adminUserStatusText(1))
	assert.Equal(t, "DISABLED", adminUserStatusText(2))
	assert.Equal(t, "FROZEN", adminUserStatusText(3))
	assert.Equal(t, "LOCKED", adminUserStatusText(4))
	assert.Equal(t, "CLOSED", adminUserStatusText(5))
}

func TestAdminFormatTime(t *testing.T) {
	value := time.Date(2026, 7, 15, 8, 9, 10, 0, time.Local)
	assert.Equal(t, "2026-07-15 08:09:10", adminFormatTime(value))
	assert.Equal(t, "", adminFormatTime(time.Time{}))
}

func TestAdminRestrictionBooleanUpdate(t *testing.T) {
	updates := map[string]any{}
	value := false
	applyRestrictionUpdate(updates, "login_enabled", &value)
	assert.Equal(t, false, updates["login_enabled"])
	applyRestrictionUpdate(updates, "game_access_enabled", nil)
	_, exists := updates["game_access_enabled"]
	assert.False(t, exists)
}
