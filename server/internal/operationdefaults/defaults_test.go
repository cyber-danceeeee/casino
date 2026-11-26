package operationdefaults

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/tenancy"
)

func TestVisibilityOverrideIsIsolatedByTenantAndSite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:operation-defaults?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE game_social_links (
		id integer primary key, tenant_id integer not null, site_id integer not null,
		status integer not null, deleted_at datetime
	)`).Error)
	require.NoError(t, db.AutoMigrate(&model.GameTenantOperationOverrides{}))
	require.NoError(t, db.Exec("INSERT INTO game_social_links(id,tenant_id,site_id,status) VALUES (7,0,0,1)").Error)

	scopeA := tenancy.Scope{TenantID: 10, SiteID: 101}
	scopeB := tenancy.Scope{TenantID: 20, SiteID: 201}
	require.NoError(t, SetVisibility(db, scopeA, ResourceSocialLink, 7, false, 99))

	a, err := VisibilityMap(db, scopeA, ResourceSocialLink, []uint64{7})
	require.NoError(t, err)
	b, err := VisibilityMap(db, scopeB, ResourceSocialLink, []uint64{7})
	require.NoError(t, err)
	require.False(t, Visible(a, 7, true))
	require.True(t, Visible(b, 7, true))
	require.False(t, Visible(b, 7, false), "a tenant cannot enable a platform-disabled default")
}

func TestVisibilityOverrideRejectsNonPlatformSource(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:operation-default-invalid?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE game_social_links (
		id integer primary key, tenant_id integer not null, site_id integer not null,
		status integer not null, deleted_at datetime
	)`).Error)
	require.NoError(t, db.AutoMigrate(&model.GameTenantOperationOverrides{}))
	require.NoError(t, db.Exec("INSERT INTO game_social_links(id,tenant_id,site_id,status) VALUES (8,10,101,1)").Error)

	err = SetVisibility(db, tenancy.Scope{TenantID: 10, SiteID: 101}, ResourceSocialLink, 8, false, 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
