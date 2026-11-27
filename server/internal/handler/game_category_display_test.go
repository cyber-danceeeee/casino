package handler

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"casino/internal/config"
	"casino/internal/content"
	"casino/internal/game/connections"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var categoryDisplayDriver sync.Once

func TestPublicCategoryUsesOnlyCuratedMembershipAndDisplayOrder(t *testing.T) {
	categoryDisplayDriver.Do(func() {
		sql.Register("sqlite3_category_display", &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.RegisterFunc("FIND_IN_SET", func(value, list string) int {
				for i, entry := range strings.Split(list, ",") {
					if value == entry {
						return i + 1
					}
				}
				return 0
			}, true)
		}})
	})
	db, err := gorm.Open(sqlite.Dialector{DriverName: "sqlite3_category_display", DSN: ":memory:"}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, _ := db.DB()
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	var old *config.Config
	func() { defer func() { _ = recover() }(); old = config.Get() }()
	cfg := &config.Config{}
	cfg.Game.AllInTop = config.AllInTop{Enabled: true, ClientID: "fixture-client", ClientSecret: "fixture-secret"}
	config.Set(cfg)
	t.Cleanup(func() { config.Set(old) })
	require.NoError(t, db.AutoMigrate(&connections.Record{}))
	for _, query := range []string{
		`CREATE TABLE game_site_config(id INTEGER PRIMARY KEY,country_code TEXT,currency_code TEXT,deleted_at DATETIME)`,
		`INSERT INTO game_site_config VALUES(1,'ID','IDR',NULL)`,
		`CREATE TABLE game_game_providers(id INTEGER PRIMARY KEY,site_id INTEGER,provider_code TEXT,name TEXT,environment TEXT,status INTEGER,supported_countries TEXT,supported_currencies TEXT,deleted_at DATETIME)`,
		`INSERT INTO game_game_providers VALUES(1,0,'allintop','AllInTop','sandbox',1,'ID','IDR',NULL)`,
		`CREATE TABLE game_games(id INTEGER PRIMARY KEY,site_id INTEGER,provider_id INTEGER,status INTEGER,provider_enabled INTEGER,platform_enabled INTEGER,supported_countries TEXT,supported_currencies TEXT,platform_category_id INTEGER,platform_category_code TEXT,deleted_at DATETIME)`,
		`INSERT INTO game_games VALUES(1,0,1,1,1,1,'','',1,'slots',NULL),(2,0,1,1,1,1,'','',1,'slots',NULL),(3,0,1,1,1,1,'','',1,'slots',NULL)`,
		`CREATE TABLE game_game_categories(id INTEGER PRIMARY KEY,site_id INTEGER,category_key TEXT,status INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_game_categories VALUES(1,0,'slots',1,NULL)`,
		`CREATE TABLE game_game_category_relations(id INTEGER PRIMARY KEY,site_id INTEGER,game_id INTEGER,category_id INTEGER,sort INTEGER,deleted_at DATETIME)`,
		`INSERT INTO game_game_category_relations VALUES(1,0,1,1,20,NULL),(2,0,2,1,1,NULL)`,
		`CREATE TABLE game_provider_maintenance(provider_code TEXT,site_id INTEGER,enabled INTEGER,hide_from_lobby INTEGER,deleted_at DATETIME,start_at DATETIME,end_at DATETIME,affected_countries TEXT,affected_currencies TEXT)`,
		`CREATE TABLE game_game_maintenance(game_id INTEGER,site_id INTEGER,enabled INTEGER,hide_from_lobby INTEGER,deleted_at DATETIME,start_at DATETIME,end_at DATETIME,affected_countries TEXT,affected_currencies TEXT)`,
	} {
		require.NoError(t, db.Exec(query).Error)
	}
	require.NoError(t, db.Exec("UPDATE game_games SET status=2 WHERE id=2").Error)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/games?category=slots", nil)
	scope := content.Scope{SiteID: 1}
	q, _ := h5BuildPublicGameQuery(c, db, scope, h5GameListFilters{Category: "slots"})
	var ids []uint64
	require.NoError(t, q.Order("display_relation.sort ASC,game_games.id ASC").Pluck("game_games.id", &ids).Error)
	require.Equal(t, []uint64{2, 1}, ids)
	require.NoError(t, db.Exec("UPDATE game_game_category_relations SET deleted_at=CURRENT_TIMESTAMP WHERE game_id=1").Error)
	q, _ = h5BuildPublicGameQuery(c, db, scope, h5GameListFilters{Category: "slots"})
	ids = nil
	require.NoError(t, q.Pluck("game_games.id", &ids).Error)
	require.Equal(t, []uint64{2}, ids)
	require.NoError(t, db.Exec("UPDATE game_game_categories SET status=0 WHERE id=1").Error)
	q, _ = h5BuildPublicGameQuery(c, db, scope, h5GameListFilters{Category: "slots"})
	var count int64
	require.NoError(t, q.Count(&count).Error)
	require.Zero(t, count)
}
