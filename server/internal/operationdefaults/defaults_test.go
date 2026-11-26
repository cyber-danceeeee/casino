package operationdefaults

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

func TestVisibilityOverrideIsIsolatedByH5AndSite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:operation-defaults?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE game_social_links (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, status INTEGER NOT NULL, deleted_at DATETIME)").Error)
	require.NoError(t, db.AutoMigrate(&model.GameContentVisibility{}))
	require.NoError(t, db.Exec("INSERT INTO game_social_links (id, site_id, status) VALUES (7, 0, 1)").Error)

	scopeA := content.Scope{SiteID: 101}
	scopeB := content.Scope{SiteID: 201}
	require.NoError(t, SetVisibility(db, scopeA, ResourceSocialLink, 7, false, 99))

	a, err := VisibilityMap(db, scopeA, ResourceSocialLink, []uint64{7})
	require.NoError(t, err)
	b, err := VisibilityMap(db, scopeB, ResourceSocialLink, []uint64{7})
	require.NoError(t, err)
	require.False(t, Visible(a, 7, true))
	require.True(t, Visible(b, 7, true))
	require.False(t, Visible(b, 7, false), "a h5 cannot enable a platform-disabled default")
}

func TestVisibilityOverrideRejectsNonPlatformSource(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:operation-default-invalid?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE game_social_links (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, status INTEGER NOT NULL, deleted_at DATETIME)").Error)
	require.NoError(t, db.AutoMigrate(&model.GameContentVisibility{}))
	require.NoError(t, db.Exec("INSERT INTO game_social_links (id, site_id, status) VALUES (8, 101, 1)").Error)

	err = SetVisibility(db, content.Scope{SiteID: 101}, ResourceSocialLink, 8, false, 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
