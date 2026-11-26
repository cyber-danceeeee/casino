package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

func newSupportNoticeTestContext(scope content.Scope) *gin.Context {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodGet, "/support?language=en", nil)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	ctx.Set("h5_content_scope", scope)
	return ctx
}

func newSupportNoticeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE game_notices (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL DEFAULT 0, notice_type TEXT, title TEXT, title_zh TEXT, title_en TEXT, content_html TEXT, content_zh TEXT, content_en TEXT, status INTEGER, sort INTEGER, start_at DATETIME, end_at DATETIME, created_at DATETIME, deleted_at DATETIME)").Error)
	return db
}

func insertSupportNotice(t *testing.T, db *gorm.DB, id uint64, siteID uint64, title string, createdAt time.Time) {
	t.Helper()
	require.NoError(t, db.Exec("INSERT INTO game_notices (id, site_id, notice_type, title_en, content_en, status, sort, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", id, siteID, "GENERAL", title, title+" content", 1, int(id), createdAt).Error)
}

func TestH5SupportNoticesForUserOnlyIncludesNoticesAfterRegistration(t *testing.T) {
	db := newSupportNoticeTestDB(t)
	registeredAt := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	insertSupportNotice(t, db, 1, 1, "old notice", registeredAt.Add(-time.Hour))
	insertSupportNotice(t, db, 2, 1, "new notice", registeredAt.Add(time.Minute))

	user := model.GameUsers{}
	user.ID = 7
	user.CreatedAt = registeredAt

	items := h5SupportNotices(db, newSupportNoticeTestContext(content.Scope{SiteID: 1}), user, true)
	require.Len(t, items, 1)
	require.Equal(t, "notice:2", items[0].ID)
	require.Equal(t, "new notice", items[0].Title)
}

func TestH5SupportNoticesForGuestKeepsPublicNotices(t *testing.T) {
	db := newSupportNoticeTestDB(t)
	insertSupportNotice(t, db, 1, 1, "public notice", time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC))

	items := h5SupportNotices(db, newSupportNoticeTestContext(content.Scope{SiteID: 1}), model.GameUsers{}, false)
	require.Len(t, items, 1)
	require.Equal(t, "notice:1", items[0].ID)
}

func TestH5SupportNoticesAreH5Scoped(t *testing.T) {
	db := newSupportNoticeTestDB(t)
	nowValue := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	insertSupportNotice(t, db, 1, 1, "h5 a notice", nowValue)
	insertSupportNotice(t, db, 2, 202, "h5 b notice", nowValue)

	items := h5SupportNotices(db, newSupportNoticeTestContext(content.Scope{SiteID: 1}), model.GameUsers{}, false)
	require.Len(t, items, 1)
	require.Equal(t, "notice:1", items[0].ID)
}
