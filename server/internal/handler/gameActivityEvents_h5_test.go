package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
)

func TestH5FindActivePromotionEnforcesScopeAndSchedule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE game_activity_events (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, slug TEXT NOT NULL, status INTEGER NOT NULL, status_code TEXT NOT NULL, start_at DATETIME, end_at DATETIME, deleted_at DATETIME)").Error)
	now := time.Now()
	require.NoError(t, db.Exec("INSERT INTO game_activity_events (id, site_id, slug, status, status_code, start_at, end_at) VALUES (1, 3, 'active', 1, 'ACTIVE', ?, ?), (2, 3, 'expired', 1, 'ACTIVE', ?, ?), (3, 3, 'draft', 1, 'DRAFT', ?, ?), (4, 4, 'other-h5', 1, 'ACTIVE', ?, ?)", now.Add(-time.Hour), now.Add(time.Hour),
		now.Add(-2*time.Hour), now.Add(-time.Hour),
		now.Add(-time.Hour), now.Add(time.Hour),
		now.Add(-time.Hour), now.Add(time.Hour),
	).Error)

	active, err := h5FindActivePromotion(db, content.Scope{SiteID: 3}, "active", now)
	require.NoError(t, err)
	require.Equal(t, uint64(1), active.ID)
	_, err = h5FindActivePromotion(db, content.Scope{SiteID: 4}, "active", now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = h5FindActivePromotion(db, content.Scope{SiteID: 3}, "expired", now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = h5FindActivePromotion(db, content.Scope{SiteID: 3}, "draft", now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
