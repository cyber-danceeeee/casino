package operationadmin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

func TestSiteBannerSaveEditDeleteImmediatelyVisible(t *testing.T) {
	db := newSiteBannerTestDB(t)
	svc := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", ObjectKey: "banner.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	actor := Actor{ID: 7, Scope: content.Scope{SiteID: 20}}
	in := SiteBannerInput{Name: "Hero", Title: "Welcome", DesktopAssetID: asset.ID, ActionType: "NONE", VisibilityScope: "ALL", Enabled: true}
	id, err := svc.SaveSiteBanner(0, in, actor)
	require.NoError(t, err)
	read := func() []PublicBanner {
		rows, err := svc.PublicSiteBanners(actor.Scope, "en", "MY", false, time.Now())
		require.NoError(t, err)
		return rows
	}
	require.Len(t, read(), 1)
	in.Title = "Updated"
	_, err = svc.SaveSiteBanner(id, in, actor)
	require.NoError(t, err)
	require.Equal(t, "Updated", read()[0].Title)
	_, err = svc.GetSiteBanner(id, content.Scope{SiteID: 21})
	require.Error(t, err)
	err = svc.DeleteAsset(asset.ID, "referenced", actor)
	require.Error(t, err)
	in.DesktopAssetID = 999
	_, err = svc.SaveSiteBanner(id, in, actor)
	require.Error(t, err)
	require.Equal(t, "Updated", read()[0].Title)
	require.NoError(t, svc.DeleteSiteBanner(id, actor))
	require.Empty(t, read())
}

func TestPublicSiteBannersFiltersScheduleAndKeepsStableSort(t *testing.T) {
	db := newSiteBannerTestDB(t)
	svc := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BANNER", ObjectKey: "h5/10/banner.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	nowValue := time.Now().UTC().Truncate(time.Second)
	startFuture := nowValue.Add(time.Hour)
	endPast := nowValue.Add(-time.Hour)
	for _, row := range []siteBannerRecord{
		{SiteID: 20, Name: "second", Title: "Second", DesktopAssetID: asset.ID, ActionType: "NONE", SortOrder: 5, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", Enabled: true, ConfigVersion: 1},
		{SiteID: 20, Name: "first", Title: "First", DesktopAssetID: asset.ID, ActionType: "NONE", SortOrder: 5, LanguageCode: "EN", CountryCode: "MY", VisibilityScope: "LOGGED_OUT", Enabled: true, ConfigVersion: 1},
		{SiteID: 20, Name: "future", Title: "Future", DesktopAssetID: asset.ID, SortOrder: 1, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", StartAt: &startFuture, Enabled: true, ConfigVersion: 1},
		{SiteID: 20, Name: "expired", Title: "Expired", DesktopAssetID: asset.ID, SortOrder: 1, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", EndAt: &endPast, Enabled: true, ConfigVersion: 1},
	} {
		require.NoError(t, db.Create(&row).Error)
	}
	rows, err := svc.PublicSiteBanners(content.Scope{SiteID: 20}, "en", "MY", false, nowValue)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "Second", rows[0].Title)
	require.Equal(t, "First", rows[1].Title)
}

func newSiteBannerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&assetLibraryRecord{}, &siteBannerRecord{}))
	require.NoError(t, db.Exec("CREATE TABLE game_site_config (id INTEGER PRIMARY KEY, config_version INTEGER DEFAULT 0, country_code TEXT, currency_code TEXT, default_language TEXT, updated_by INTEGER, updated_at DATETIME, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_site_config (id, country_code, currency_code, default_language) VALUES (20, 'MY', 'MYR', 'en'), (21, 'MY', 'MYR', 'en')").Error)
	require.NoError(t, db.AutoMigrate(&model.GameAdminOperationLogs{}))
	return db
}
