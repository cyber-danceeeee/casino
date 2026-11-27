package systemadmin

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCountryDefaultsDistinguishIndonesiaFromIndia(t *testing.T) {
	id := countryInputDefaults(CountryInput{CountryCode: "id"})
	require.NoError(t, validateCountryInput(id, true))
	require.Equal(t, "ID", id.ISOCountryCode)
	require.Equal(t, "IDR", id.CurrencyCode)
	require.Equal(t, "+62", id.PhoneCountryCode)
	require.Zero(t, id.PhoneLength)
	in := countryInputDefaults(CountryInput{CountryCode: "IN"})
	require.NoError(t, validateCountryInput(in, true))
	require.Equal(t, "INR", in.CurrencyCode)
	require.Equal(t, "+91", in.PhoneCountryCode)
	id.ISOCountryCode = "IN"
	require.Error(t, validateCountryInput(id, true))
}

func TestSimplifiedCountrySavePreservesBrandingAndProtectsSelectedMarket(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	require.NoError(t, db.Table("game_country_configs").AutoMigrate(&countryConfigRecord{}))
	require.NoError(t, db.Exec("ALTER TABLE game_country_configs ADD COLUMN deleted_at DATETIME").Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_admin_operation_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT, operator_id INTEGER, site_id INTEGER,
        operator_username TEXT, permission TEXT, module TEXT, action TEXT, result TEXT,
        target_type TEXT, target_id INTEGER, before_state TEXT, after_state TEXT,
        reason TEXT, ip_address TEXT, user_agent TEXT, correlation_id TEXT,
        currency TEXT, created_at DATETIME, updated_at DATETIME
    )`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_site_config (id INTEGER PRIMARY KEY, country_code TEXT, currency_code TEXT,default_language TEXT,timezone TEXT,config_version INTEGER DEFAULT 0,updated_by INTEGER,updated_at DATETIME,deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_site_config(id,country_code,currency_code) VALUES(1,'ID','IDR')`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_user_wallets(site_id INTEGER)`).Error)
	s := New(db)
	id, err := s.CreateCountry(CountryInput{CountryCode: "ID"}, Actor{})
	require.NoError(t, err)
	row, err := s.countryByID(id)
	require.NoError(t, err)
	require.Equal(t, "IDR", row.CurrencyCode)
	require.Zero(t, row.PhoneLength)
	require.NoError(t, db.Table("game_country_configs").Where("id=?", id).Updates(map[string]any{"branding_config": `{"headerLogoUrl":"/static/existing.svg"}`, "theme_config": `{"primaryColor":"#112233"}`}).Error)
	input := CountryPresets()[0]
	input.Timezone = "Asia/Makassar"
	require.NoError(t, s.UpdateCountry(id, input, Actor{}))
	row, err = s.countryByID(id)
	require.NoError(t, err)
	require.Contains(t, row.BrandingConfig, "existing.svg")
	require.Contains(t, row.ThemeConfig, "#112233")
	require.Equal(t, "Asia/Makassar", row.Timezone)
	require.NoError(t, db.Exec("INSERT INTO game_user_wallets VALUES(1)").Error)
	require.Error(t, s.CountryStatus(id, false, "test", Actor{}))
	require.Error(t, s.DeleteCountry(id, "test", Actor{}))
	disabled := false
	input.Enabled = &disabled
	require.Error(t, s.UpdateCountry(id, input, Actor{}))
	input.Enabled = nil
	input.CurrencyCode = "INR"
	require.Error(t, s.UpdateCountry(id, input, Actor{}))
	row, err = s.countryByID(id)
	require.NoError(t, err)
	require.True(t, row.Enabled)
	require.Equal(t, "IDR", row.CurrencyCode)
}

func TestCountryActivationIsExclusiveAndAppliesH5RegistrationCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, _ := db.DB()
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.Table("game_country_configs").AutoMigrate(&countryConfigRecord{}))
	require.NoError(t, db.Exec(`ALTER TABLE game_country_configs ADD COLUMN deleted_at DATETIME`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX active_country_test ON game_country_configs(status) WHERE status=1 AND deleted_at IS NULL`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_site_config(id INTEGER PRIMARY KEY,country_code TEXT,currency_code TEXT,default_language TEXT,timezone TEXT,config_version INTEGER DEFAULT 0,updated_by INTEGER,updated_at DATETIME,deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO game_site_config(id,country_code,currency_code)VALUES(1,'ID','IDR')`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_user_wallets(site_id INTEGER)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE game_admin_operation_logs(id INTEGER PRIMARY KEY,operator_id INTEGER,site_id INTEGER,operator_username TEXT,permission TEXT,module TEXT,action TEXT,result TEXT,target_type TEXT,target_id INTEGER,before_state TEXT,after_state TEXT,reason TEXT,ip_address TEXT,user_agent TEXT,correlation_id TEXT,currency TEXT,created_at DATETIME,updated_at DATETIME)`).Error)
	svc := New(db)
	actor := Actor{ID: 7, Platform: true}
	id, err := svc.CreateCountry(CountryInput{CountryCode: "ID"}, actor)
	require.NoError(t, err)
	disabled := false
	india, err := svc.CreateCountry(CountryInput{CountryCode: "IN", Enabled: &disabled}, actor)
	require.NoError(t, err)
	require.NoError(t, svc.CountryStatus(india, true, "switch", actor))
	var enabled int64
	require.NoError(t, db.Table("game_country_configs").Where("status=1").Count(&enabled).Error)
	require.Equal(t, int64(1), enabled)
	app, err := svc.PublicAppConfig("ID", "en")
	require.NoError(t, err)
	require.Equal(t, "IN", app["country"].(map[string]any)["code"])
	require.Equal(t, "+91", app["country"].(map[string]any)["phoneCountryCode"])
	require.NoError(t, svc.CountryStatus(id, true, "switch back", actor))
	app, err = svc.PublicAppConfig("IN", "en")
	require.NoError(t, err)
	require.Equal(t, "ID", app["country"].(map[string]any)["code"])
	require.Equal(t, "+62", app["country"].(map[string]any)["phoneCountryCode"])
	require.Equal(t, "IDR", app["currency"].(map[string]any)["code"])
	require.NoError(t, db.Exec(`INSERT INTO game_user_wallets VALUES(1)`).Error)
	err = svc.CountryStatus(india, true, "must be rejected", actor)
	require.Equal(t, "CURRENCY_CHANGE_REQUIRES_MIGRATION", ErrorCode(err))
	current, err := svc.countryByID(id)
	require.NoError(t, err)
	require.True(t, current.Enabled)
	require.NoError(t, db.Table("game_country_configs").Where("status=1").Count(&enabled).Error)
	require.Equal(t, int64(1), enabled)
}
