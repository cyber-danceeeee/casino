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

func TestH5NavigationSaveEditSortDeleteImmediatelyVisible(t *testing.T) {
	db := newH5NavigationTestDB(t)
	svc := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", ObjectKey: "icon.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	actor := Actor{ID: 7, Scope: content.Scope{SiteID: 20}}
	in := H5NavigationInput{IconAssetID: asset.ID, LabelZh: "充值", LabelEn: "Deposit", DisplayPosition: "HOME_QUICK", BadgeType: "NONE", VisibilityScope: "ALL", Enabled: true, ActionType: "OPEN_DEPOSIT"}
	id, err := svc.SaveH5Navigation(0, in, actor)
	require.NoError(t, err)
	read := func() []PublicH5NavigationItem {
		rows, err := svc.H5Navigation(actor.Scope, "en", "ID", false, "HOME_QUICK", time.Now())
		require.NoError(t, err)
		return rows
	}
	require.Len(t, read(), 1)
	require.Equal(t, "Deposit", read()[0].Label)
	in.LabelEn = "Updated"
	_, err = svc.SaveH5Navigation(id, in, actor)
	require.NoError(t, err)
	require.Equal(t, "Updated", read()[0].Label)
	_, err = svc.SaveH5Navigation(id, in, Actor{Scope: content.Scope{SiteID: 21}})
	require.Error(t, err)
	_, err = svc.SaveH5Navigation(id, in, Actor{Scope: content.Scope{Platform: true}})
	require.Error(t, err)
	require.NoError(t, svc.DeleteH5Navigation(id, actor))
	require.Empty(t, read())
}

func TestH5NavigationPublicFiltersAndSortsConfigurations(t *testing.T) {
	db := newH5NavigationTestDB(t)
	service := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BUTTON_ICON", ObjectKey: "h5/10/icon.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	actor := Actor{ID: 7, Username: "h5-a", Scope: content.Scope{SiteID: 20}}
	first, err := service.SaveH5Navigation(0, H5NavigationInput{
		LabelZh: "活动", LabelEn: "Offers", IconAssetID: asset.ID, DisplayPosition: "BOTTOM_NAV", SortOrder: 20,
		BadgeType: "NEW", VisibilityScope: "GUEST", CountryCodes: []string{"MY"}, LanguageCodes: []string{"en-US"},
		Enabled: true, ActionType: "OPEN_INTERNAL_PAGE", ActionTarget: "/offers",
	}, actor)
	require.NoError(t, err)
	second, err := service.SaveH5Navigation(0, H5NavigationInput{
		LabelZh: "客服", LabelEn: "Support", IconAssetID: asset.ID, DisplayPosition: "BOTTOM_NAV", SortOrder: 10,
		BadgeType: "NONE", VisibilityScope: "ALL", Enabled: true, ActionType: "OPEN_SUPPORT",
	}, actor)
	require.NoError(t, err)
	futureStart := time.Now().Add(time.Hour)
	_, err = service.SaveH5Navigation(0, H5NavigationInput{
		LabelZh: "稍后", LabelEn: "Later", IconAssetID: asset.ID, DisplayPosition: "BOTTOM_NAV", SortOrder: 30,
		BadgeType: "NONE", VisibilityScope: "ALL", StartAt: &futureStart, Enabled: true,
		ActionType: "OPEN_INTERNAL_PAGE", ActionTarget: "/later",
	}, actor)
	require.NoError(t, err)
	require.NoError(t, service.SortH5Navigation(H5NavigationSortInput{DisplayPosition: "BOTTOM_NAV", IDs: []uint64{first, second}}, actor))
	guest, err := service.H5Navigation(actor.Scope, "en-GB", "MY", false, "BOTTOM_NAV", time.Now())
	require.NoError(t, err)
	require.Len(t, guest, 2)
	require.Equal(t, "Offers", guest[0].Label)
	require.Equal(t, "Support", guest[1].Label)

	loggedIn, err := service.H5Navigation(actor.Scope, "en-US", "MY", true, "BOTTOM_NAV", time.Now())
	require.NoError(t, err)
	require.Len(t, loggedIn, 1)
	require.Equal(t, "Support", loggedIn[0].Label)

	wrongLanguage, err := service.H5Navigation(actor.Scope, "zh-CN", "MY", false, "BOTTOM_NAV", time.Now())
	require.NoError(t, err)
	require.Len(t, wrongLanguage, 2)
}

func TestH5NavigationSaveRejectsInactiveAsset(t *testing.T) {
	db := newH5NavigationTestDB(t)
	svc := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetType: "IMAGE", AssetScope: AssetScopeSite, ObjectKey: "icon.png", Status: "DISABLED"}
	require.NoError(t, db.Create(&asset).Error)
	_, err := svc.SaveH5Navigation(0, H5NavigationInput{LabelZh: "提现", LabelEn: "Withdraw", IconAssetID: asset.ID, DisplayPosition: "HOME_QUICK", BadgeType: "NONE", VisibilityScope: "ALL", Enabled: true, ActionType: "OPEN_WITHDRAW"}, Actor{ID: 7, Scope: content.Scope{SiteID: 20}})
	require.Error(t, err)
}

func newH5NavigationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&assetLibraryRecord{}, &model.GameH5NavigationItems{}, &model.GameAdminOperationLogs{}))
	require.NoError(t, db.Exec("CREATE TABLE game_site_config (id INTEGER PRIMARY KEY, config_version INTEGER DEFAULT 0, country_code TEXT, currency_code TEXT, default_language TEXT, updated_by INTEGER, updated_at DATETIME, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_site_config (id, country_code, currency_code, default_language) VALUES (20, 'MY', 'MYR', 'en-US'), (21, 'MY', 'MYR', 'en-US')").Error)
	return db
}

func TestBundledBottomNavigationSavesWithoutUploadedAssets(t *testing.T) {
	db := newH5NavigationTestDB(t)
	service := New(db)
	actor := Actor{ID: 7, Scope: content.Scope{SiteID: 1}}
	input := H5NavigationInput{MenuCode: "vip", LabelZh: "VIP", LabelEn: "VIP", DisplayPosition: "BOTTOM_NAV", BadgeType: "NONE", VisibilityScope: "ALL", Enabled: true, ActionType: "OPEN_INTERNAL_PAGE", ActionTarget: "/vip"}
	_, err := service.SaveH5Navigation(0, input, actor)
	require.NoError(t, err)
	items, err := service.H5Navigation(actor.Scope, "en", "ID", false, "BOTTOM_NAV", time.Now())
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "VIP", items[0].Label)
	input.MenuCode = "custom-entry"
	_, err = service.SaveH5Navigation(0, input, actor)
	require.Error(t, err)
}
