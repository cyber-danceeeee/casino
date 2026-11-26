package handler

import (
	"strings"
	"testing"

	"casino/internal/content"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestH5CatalogFallbackUsesOnlyImportedRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	defer raw.Close()
	require.NoError(t, db.Exec("CREATE TABLE game_games(id INTEGER PRIMARY KEY,site_id INTEGER)").Error)
	require.NoError(t, db.Exec("INSERT INTO game_games VALUES(1,0),(2,1)").Error)
	var ids []uint64
	require.NoError(t, h5ApplyReadableScope(db.Table("game_games"), "game_games", h5PlatformCatalogScope()).Pluck("id", &ids).Error)
	require.Equal(t, []uint64{1}, ids)
}

func TestH5CatalogMaintenanceBindsDeploymentCountryAndCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	defer raw.Close()
	var rows []struct{ ID uint64 }
	query := h5ApplyGameMaintenance(db.Table("game_games"), content.Scope{SiteID: 1}, "PK", "PKR").Find(&rows)
	require.NoError(t, query.Error)
	require.Equal(t, strings.Count(query.Statement.SQL.String(), "?"), len(query.Statement.Vars))
	require.Equal(t, "PK", query.Statement.Vars[3])
	require.Equal(t, "PKR", query.Statement.Vars[4])
	require.Equal(t, "PK", query.Statement.Vars[8])
	require.Equal(t, "PKR", query.Statement.Vars[9])
}
