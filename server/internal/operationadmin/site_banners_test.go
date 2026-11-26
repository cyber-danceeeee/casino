package operationadmin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/tenancy"
)

func TestSiteBannerDraftIsolationAndPublicVisibility(t *testing.T) {
	db := newSiteBannerTestDB(t)
	svc := New(db)
	assetA := assetLibraryRecord{TenantID: 10, SiteID: 20, AssetScope: AssetScopeTenant, AssetType: "IMAGE", UsageType: "BANNER", ObjectKey: "tenant/10/banner-a.png", Status: AssetStatusActive}
	assetB := assetLibraryRecord{TenantID: 11, SiteID: 21, AssetScope: AssetScopeTenant, AssetType: "IMAGE", UsageType: "BANNER", ObjectKey: "tenant/11/banner-b.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&assetA).Error)
	require.NoError(t, db.Create(&assetB).Error)
	actorA := Actor{ID: 7, Username: "tenant-a", Scope: tenancy.Scope{TenantID: 10, SiteID: 20}}
	id, err := svc.SaveSiteBanner(0, SiteBannerInput{Name: "Hero", Title: "Welcome", DesktopAssetID: assetA.ID,
		ActionType: "NONE", VisibilityScope: "ALL", Enabled: true}, actorA)
	require.NoError(t, err)

	rows, err := svc.PublicSiteBanners(actorA.Scope, "en", "MY", false, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "draft banner must not be public")

	_, err = svc.GetSiteBanner(id, tenancy.Scope{TenantID: 11, SiteID: 21})
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "BANNER_NOT_FOUND", domain.Code)

	err = db.Model(&siteBannerRecord{}).Where("id=?", id).Updates(map[string]any{
		"title": "Welcome", "desktop_asset_id": assetA.ID, "action_type": "NONE", "visibility_scope": "ALL",
		"language_code": "ALL", "country_code": "ALL", "enabled": true, "publish_status": BannerPublishPublished, "config_version": 1,
	}).Error
	require.NoError(t, err)
	rows, err = svc.PublicSiteBanners(actorA.Scope, "en", "MY", false, time.Now())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Welcome", rows[0].Title)

	_, err = svc.SaveSiteBanner(id, SiteBannerInput{Name: "Hero Draft", Title: "Draft Title", DesktopAssetID: assetA.ID,
		ActionType: "NONE", VisibilityScope: "ALL", Enabled: true}, actorA)
	require.NoError(t, err)
	rows, err = svc.PublicSiteBanners(actorA.Scope, "en", "MY", false, time.Now())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Welcome", rows[0].Title, "published version must remain live while a draft exists")

	err = svc.DeleteAsset(assetA.ID, "attempt deleting referenced banner asset", actorA)
	require.Error(t, err)
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "ASSET_IN_USE", domain.Code)

	_, err = svc.SaveSiteBanner(id, SiteBannerInput{Name: "Attack", Title: "Attack", DesktopAssetID: assetB.ID,
		ActionType: "NONE", VisibilityScope: "ALL", Enabled: true}, actorA)
	require.Error(t, err, "tenant A must not use tenant B assets")
}

func TestPublicSiteBannersFiltersScheduleAndKeepsStableSort(t *testing.T) {
	db := newSiteBannerTestDB(t)
	svc := New(db)
	asset := assetLibraryRecord{TenantID: 10, SiteID: 20, AssetScope: AssetScopeTenant, AssetType: "IMAGE", UsageType: "BANNER", ObjectKey: "tenant/10/banner.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	nowValue := time.Now().UTC().Truncate(time.Second)
	startFuture := nowValue.Add(time.Hour)
	endPast := nowValue.Add(-time.Hour)
	for _, row := range []siteBannerRecord{
		{TenantID: 10, SiteID: 20, Name: "second", Title: "Second", DesktopAssetID: asset.ID, ActionType: "NONE", SortOrder: 5, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", Enabled: true, PublishStatus: BannerPublishPublished, ConfigVersion: 1},
		{TenantID: 10, SiteID: 20, Name: "first", Title: "First", DesktopAssetID: asset.ID, ActionType: "NONE", SortOrder: 5, LanguageCode: "EN", CountryCode: "MY", VisibilityScope: "LOGGED_OUT", Enabled: true, PublishStatus: BannerPublishPublished, ConfigVersion: 1},
		{TenantID: 10, SiteID: 20, Name: "future", Title: "Future", DesktopAssetID: asset.ID, SortOrder: 1, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", StartAt: &startFuture, Enabled: true, PublishStatus: BannerPublishPublished, ConfigVersion: 1},
		{TenantID: 10, SiteID: 20, Name: "expired", Title: "Expired", DesktopAssetID: asset.ID, SortOrder: 1, LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", EndAt: &endPast, Enabled: true, PublishStatus: BannerPublishPublished, ConfigVersion: 1},
	} {
		require.NoError(t, db.Create(&row).Error)
	}
	rows, err := svc.PublicSiteBanners(tenancy.Scope{TenantID: 10, SiteID: 20}, "en", "MY", false, nowValue)
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
	require.NoError(t, db.Exec(`CREATE TABLE game_saas_sites (
		id INTEGER PRIMARY KEY, tenant_id INTEGER, config_status TEXT, config_draft_updated_at DATETIME,
		country_code TEXT, currency_code TEXT, default_language TEXT,
		updated_by INTEGER, updated_at DATETIME, deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_saas_sites (id,tenant_id,config_status,country_code,currency_code,default_language) VALUES (20,10,'PUBLISHED','MY','MYR','en'),(21,11,'PUBLISHED','MY','MYR','en')`).Error)
	require.NoError(t, db.AutoMigrate(&model.GameAdminOperationLogs{}))
	return db
}
