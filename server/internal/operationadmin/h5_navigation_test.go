package operationadmin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

func TestH5NavigationDraftIsolationValidationAndPublishing(t *testing.T) {
	db := newH5NavigationTestDB(t)
	service := New(db)
	assetA := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BUTTON_ICON", ObjectKey: "h5/10/icon.png", Status: AssetStatusActive}
	assetB := assetLibraryRecord{SiteID: 21, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BUTTON_ICON", ObjectKey: "h5/11/icon.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&assetA).Error)
	require.NoError(t, db.Create(&assetB).Error)
	actorA := Actor{ID: 7, Username: "h5-a", Scope: content.Scope{SiteID: 20}}
	_, err := service.ListH5Navigation(ListQuery{Page: 1, PageSize: 20, Scope: content.Scope{Platform: true}})
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_SCOPE_REQUIRED", domain.Code)
	input := H5NavigationInput{
		LabelZh: "充值", LabelEn: "Deposit", IconAssetID: assetA.ID, DisplayPosition: "HOME_QUICK",
		SortOrder: 10, BadgeType: "HOT", VisibilityScope: "ALL", Enabled: true,
		ActionType: "OPEN_DEPOSIT", Reason: "create deposit entry",
	}
	id, err := service.SaveH5Navigation(0, input, actorA)
	require.NoError(t, err)
	require.NotZero(t, id)

	public, err := service.H5Navigation(actorA.Scope, "en-US", "MY", false, "HOME_QUICK", time.Now())
	require.NoError(t, err)
	require.Empty(t, public, "draft entry must not be public")

	_, err = service.GetH5Navigation(id, content.Scope{SiteID: 21})
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_NOT_FOUND", domain.Code)
	err = service.DeleteH5Navigation(id, Actor{ID: 1, Username: "platform", Scope: content.Scope{Platform: true}})
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_SCOPE_REQUIRED", domain.Code)
	_, err = service.SaveH5Navigation(id, input, Actor{ID: 8, Username: "h5-b", Scope: content.Scope{SiteID: 21}})
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_NOT_FOUND", domain.Code)

	err = service.DeleteAsset(assetA.ID, "attempt to delete referenced asset", actorA)
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "ASSET_IN_USE", domain.Code)

	input.IconAssetID = assetB.ID
	_, err = service.SaveH5Navigation(id, input, actorA)
	require.Error(t, err, "h5 A must not reference h5 B assets")
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_ASSET_INVALID", domain.Code)

	input.IconAssetID = assetA.ID
	input.ActionType = "RUN_SCRIPT"
	_, err = service.SaveH5Navigation(id, input, actorA)
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "INVALID_H5_NAVIGATION_ACTION", domain.Code)

	input.ActionType = "OPEN_EXTERNAL_URL"
	input.ActionTarget = "http://example.com"
	_, err = service.SaveH5Navigation(id, input, actorA)
	require.Error(t, err)
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "INVALID_H5_NAVIGATION_TARGET", domain.Code)

	count, err := PublishH5NavigationDrafts(db, 20, actorA.ID, 2)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = PublishH5NavigationDrafts(db, 20, actorA.ID, 2)
	require.NoError(t, err)
	require.Zero(t, count, "repeated publish must be idempotent")

	public, err = service.H5Navigation(actorA.Scope, "en-US", "MY", false, "HOME_QUICK", time.Now())
	require.NoError(t, err)
	require.Len(t, public, 1)
	require.Equal(t, "Deposit", public[0].Label)
	require.Equal(t, "OPEN_DEPOSIT", public[0].ActionType)

	input.LabelEn = "Deposit Draft"
	input.ActionType = "OPEN_DEPOSIT"
	input.ActionTarget = ""
	_, err = service.SaveH5Navigation(id, input, actorA)
	require.NoError(t, err)
	public, err = service.H5Navigation(actorA.Scope, "en-US", "MY", false, "HOME_QUICK", time.Now())
	require.NoError(t, err)
	require.Len(t, public, 1)
	require.Equal(t, "Deposit", public[0].Label, "published value must stay live while a new draft exists")
}

func TestH5NavigationPublicFiltersAndSortsDrafts(t *testing.T) {
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
	count, err := PublishH5NavigationDrafts(db, 20, actor.ID, 3)
	require.NoError(t, err)
	require.Equal(t, 3, count)

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
	require.Empty(t, wrongLanguage)
}

func TestPublishH5NavigationRejectsInactiveAsset(t *testing.T) {
	db := newH5NavigationTestDB(t)
	service := New(db)
	asset := assetLibraryRecord{SiteID: 20, AssetScope: AssetScopeSite, AssetType: "IMAGE", UsageType: "BUTTON_ICON", ObjectKey: "h5/10/icon.png", Status: AssetStatusActive}
	require.NoError(t, db.Create(&asset).Error)
	actor := Actor{ID: 7, Username: "h5-a", Scope: content.Scope{SiteID: 20}}
	_, err := service.SaveH5Navigation(0, H5NavigationInput{
		LabelZh: "提现", LabelEn: "Withdraw", IconAssetID: asset.ID, DisplayPosition: "HOME_QUICK",
		BadgeType: "NONE", VisibilityScope: "LOGGED_IN", Enabled: true, ActionType: "OPEN_WITHDRAW",
	}, actor)
	require.NoError(t, err)
	require.NoError(t, db.Model(&assetLibraryRecord{}).Where("id=?", asset.ID).Update("status", "DISABLED").Error)
	_, err = PublishH5NavigationDrafts(db, 20, actor.ID, 2)
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "H5_NAVIGATION_ASSET_INVALID", domain.Code)
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
	require.NoError(t, db.Exec("CREATE TABLE game_site_config (id INTEGER PRIMARY KEY, config_status TEXT, config_draft_updated_at DATETIME, country_code TEXT, currency_code TEXT, default_language TEXT, updated_by INTEGER, updated_at DATETIME, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_site_config (id, config_status, country_code, currency_code, default_language) VALUES (20, 'PUBLISHED', 'MY', 'MYR', 'en-US'), (21, 'PUBLISHED', 'MY', 'MYR', 'en-US')").Error)
	return db
}
