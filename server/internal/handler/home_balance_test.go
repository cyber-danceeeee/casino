package handler

import (
	"casino/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWalletTotalDoesNotDoubleCountWithdrawableFunds(t *testing.T) {
	wallet := model.GameUserWallets{Currency: "IDR", CashBalance: h5DecimalPtr(decimal.RequireFromString("100.10")), BonusBalance: h5DecimalPtr(decimal.RequireFromString("20.20")), FrozenBalance: h5DecimalPtr(decimal.RequireFromString("5.30")), WithdrawableBalance: h5DecimalPtr(decimal.RequireFromString("80.00"))}
	require.Equal(t, "125.60", h5WalletPayload(wallet)["totalBalance"])
	require.Equal(t, "0.00", h5WalletPayload(model.GameUserWallets{})["totalBalance"])
	wallet.CashBalance = h5DecimalPtr(decimal.RequireFromString("90071992547409.91"))
	wallet.BonusBalance = h5DecimalPtr(decimal.RequireFromString("0.02"))
	wallet.FrozenBalance = h5DecimalPtr(decimal.RequireFromString("0.03"))
	require.Equal(t, "90071992547409.96", h5WalletPayload(wallet)["totalBalance"])
}

func TestHomeWalletAuthenticationReadsOnlyCurrentMembersBalance(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, _ := db.DB()
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.Table("game_site_config").AutoMigrate(&h5BootstrapSiteRow{}))
	require.NoError(t, db.Table("game_country_configs").AutoMigrate(&h5BootstrapCountryRow{}))
	require.NoError(t, db.Table("game_site_config").Create(&h5BootstrapSiteRow{ID: 1, CountryCode: "ID", CurrencyCode: "IDR"}).Error)
	require.NoError(t, db.Table("game_country_configs").Create(&h5BootstrapCountryRow{CountryCode: "ID", CurrencyCode: "IDR", Status: true}).Error)
	for _, sql := range []string{
		`CREATE TABLE game_users(id INTEGER PRIMARY KEY,site_id INTEGER,status INTEGER,login_enabled INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_users VALUES(1,1,1,1,NULL),(2,1,1,1,NULL)`,
		`CREATE TABLE game_user_sessions(site_id INTEGER,user_id INTEGER,token_hash TEXT,expired_at DATETIME,revoked_at DATETIME,deleted_at DATETIME)`,
		`CREATE TABLE game_user_wallets(id INTEGER PRIMARY KEY,user_id INTEGER,site_id INTEGER,currency TEXT,cash_balance REAL,bonus_balance REAL,frozen_balance REAL,withdrawable_balance REAL,deleted_at DATETIME)`,
		`INSERT INTO game_user_wallets VALUES(1,1,1,'IDR',100,20,5,80,NULL),(2,2,1,'IDR',900,0,0,0,NULL)`,
		`CREATE TABLE game_register_reward_claims(id INTEGER PRIMARY KEY,user_id INTEGER,site_id INTEGER,status INTEGER,currency TEXT,updated_at DATETIME,deleted_at DATETIME)`,
	} {
		require.NoError(t, db.Exec(sql).Error)
	}
	for userID, token := range map[int]string{1: "user_member_one", 2: "user_member_two"} {
		require.NoError(t, db.Exec("INSERT INTO game_user_sessions(site_id,user_id,token_hash,expired_at) VALUES(1,?,?,?)", userID, h5UserSessionHash(token), time.Now().Add(time.Hour)).Error)
	}
	for _, tc := range []struct {
		token string
		id    uint64
		total string
	}{
		{"user_member_one", 1, "125.00"}, {"user_member_two", 2, "900.00"}, {"", 0, ""}, {"user_invalid", 0, ""},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/api/v1/home", nil)
		if tc.token != "" {
			c.Request.Header.Set("Authorization", "Bearer "+tc.token)
		}
		user, wallet, ok := h5AuthenticatedUser(db, c)
		require.Equal(t, tc.id > 0, ok)
		require.Equal(t, tc.id, user.ID)
		if ok {
			require.Equal(t, tc.total, h5WalletPayload(wallet)["totalBalance"])
		} else {
			require.Zero(t, wallet.ID)
		}
	}
}
