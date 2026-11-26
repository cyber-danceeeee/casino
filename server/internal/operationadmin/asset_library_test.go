package operationadmin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
)

func TestAssetLibraryH5Scope(t *testing.T) {
	db := newAssetLibraryTestDB(t)
	svc := New(db)
	require.NoError(t, db.Create(&assetLibraryRecord{AssetScope: AssetScopePlatform, AssetType: "IMAGE", UsageType: "BANNER", DisplayName: "platform banner", ObjectKey: "admin/assets/platform/banner.png", IsPublic: true, Status: AssetStatusActive}).Error)
	require.NoError(t, db.Create(&assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "LOGO", DisplayName: "h5 logo", ObjectKey: "admin/assets/h5/10/logo.png", Status: AssetStatusActive}).Error)
	require.NoError(t, db.Create(&assetLibraryRecord{SiteID: 21, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "LOGO", DisplayName: "other h5", ObjectKey: "admin/assets/h5/11/logo.png", Status: AssetStatusActive}).Error)

	result, err := svc.ListAssets(ListQuery{Page: 1, PageSize: 20, Scope: content.Scope{SiteID: 20}})
	require.NoError(t, err)
	rows := result.List.([]assetLibraryRecord)
	require.Len(t, rows, 2)
	require.ElementsMatch(t, []string{"platform banner", "h5 logo"}, []string{rows[0].DisplayName, rows[1].DisplayName})

	_, err = svc.GetAsset(3, content.Scope{SiteID: 20})
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "ASSET_NOT_FOUND", domain.Code)
}

func TestAssetLibraryDeleteRejectsReferencedAsset(t *testing.T) {
	db := newAssetLibraryTestDB(t)
	svc := New(db)
	row := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BANNER", DisplayName: "home banner", ObjectKey: "admin/assets/h5/10/banner.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&row).Error)
	require.NoError(t, db.Exec("CREATE TABLE game_banners (id INTEGER PRIMARY KEY AUTOINCREMENT, site_id INTEGER NOT NULL DEFAULT 0, image_url TEXT NOT NULL DEFAULT '', mobile_image_url TEXT NOT NULL DEFAULT '', deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_banners (site_id, image_url) VALUES (20, ?)", row.ObjectKey).Error)

	err := svc.DeleteAsset(row.ID, "cleanup unused assets", Actor{ID: 7, Username: "ops", Scope: content.Scope{SiteID: 20}})
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "ASSET_IN_USE", domain.Code)
}

func newAssetLibraryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.AutoMigrate(&assetLibraryRecord{}))
	return db
}
