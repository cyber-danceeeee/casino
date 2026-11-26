package systemadmin

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPublishBannerDraftsValidatesAssetsAndIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	for _, statement := range []string{
		`CREATE TABLE game_asset_library (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, asset_scope TEXT, asset_type TEXT, status TEXT, is_public INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE game_site_banners (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, name TEXT, title TEXT, subtitle TEXT,
			desktop_asset_id INTEGER, mobile_asset_id INTEGER, desktop_asset_id_draft INTEGER, mobile_asset_id_draft INTEGER,
			action_type TEXT, action_target TEXT, sort_order INTEGER, language_code TEXT, country_code TEXT, currency_code TEXT,
			visibility_scope TEXT, start_at DATETIME, end_at DATETIME, enabled INTEGER, draft_payload TEXT, published_payload TEXT,
			publish_status TEXT, config_version INTEGER, published_at DATETIME, published_by INTEGER, updated_by INTEGER, updated_at DATETIME, deleted_at DATETIME)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, db.Exec(`INSERT INTO game_asset_library VALUES (1,10,20,'TENANT','IMAGE','ACTIVE',0,NULL)`).Error)
	payload, err := json.Marshal(siteBannerPublishPayload{Name: "Hero", Title: "Welcome", DesktopAssetID: 1, ActionType: "NONE",
		SortOrder: 2, LanguageCode: "ALL", CountryCode: "ALL", CurrencyCode: "ALL", VisibilityScope: "ALL", Enabled: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`INSERT INTO game_site_banners (id,tenant_id,site_id,desktop_asset_id_draft,draft_payload,publish_status) VALUES (1,10,20,1,?,'DRAFT')`, string(payload)).Error)

	count, err := publishBannerDrafts(db, 10, 20, 7, 3)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = publishBannerDrafts(db, 10, 20, 7, 3)
	require.NoError(t, err)
	require.Zero(t, count)
	var row struct {
		PublishStatus string
		ConfigVersion uint64
		Enabled       bool
	}
	require.NoError(t, db.Table("game_site_banners").Where("id=?", 1).Take(&row).Error)
	require.Equal(t, "PUBLISHED", row.PublishStatus)
	require.Equal(t, uint64(3), row.ConfigVersion)
	require.True(t, row.Enabled)

	legacyPayload, err := json.Marshal(siteBannerPublishPayload{Name: "Legacy", Title: "Legacy", DesktopImageURL: "/static/h5/images/banner-default.webp", MobileImageURL: "/static/h5/images/banner-default.webp", ActionType: "NONE", VisibilityScope: "ALL", Enabled: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`INSERT INTO game_site_banners (id,tenant_id,site_id,draft_payload,publish_status) VALUES (3,10,20,?,'DRAFT')`, string(legacyPayload)).Error)
	count, err = publishBannerDrafts(db, 10, 20, 7, 4)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	badPayload, _ := json.Marshal(siteBannerPublishPayload{Name: "Bad", Title: "Bad", DesktopAssetID: 99, ActionType: "NONE", VisibilityScope: "ALL", Enabled: true})
	require.NoError(t, db.Exec(`INSERT INTO game_site_banners (id,tenant_id,site_id,desktop_asset_id_draft,draft_payload,publish_status) VALUES (2,10,20,99,?,'DRAFT')`, string(badPayload)).Error)
	_, err = publishBannerDrafts(db, 10, 20, 7, uint64(time.Now().Unix()))
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "BANNER_ASSET_INVALID", domain.Code)
}
