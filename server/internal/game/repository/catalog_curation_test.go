package repository

import (
	"casino/internal/game/domain"
	"casino/internal/model"
	"context"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
	"testing"
)

func TestResyncPreservesOperatorCategoryIncludingRemovedMembership(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, _ := db.DB()
	t.Cleanup(func() { raw.Close() })
	stmt := &gorm.Statement{DB: db}
	require.NoError(t, stmt.Parse(&model.GameGames{}))
	cols := []string{}
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		kind := "TEXT"
		switch f.GORMDataType {
		case schema.Int, schema.Uint, schema.Bool:
			kind = "INTEGER"
		case schema.Time:
			kind = "DATETIME"
		}
		if f.DBName == "id" {
			kind = "INTEGER PRIMARY KEY"
		}
		cols = append(cols, "`"+f.DBName+"` "+kind)
	}
	require.NoError(t, db.Exec("CREATE TABLE game_games("+strings.Join(cols, ",")+")").Error)
	provider := model.GameGameProviders{SiteID: 0, ProviderCode: "allintop"}
	provider.ID = 1
	row := model.GameGames{SiteID: 0, ProviderID: 1, ProviderGameCode: "v/game", GameCode: "game-one", Name: "Game", PlatformCategoryID: 99, PlatformCategoryCode: "curated", PlatformEnabled: true, ProviderEnabled: true, Status: 1}
	require.NoError(t, db.Create(&row).Error)
	upstream := domain.Game{ProviderGameID: "v/game", ProviderCode: "allintop", Name: "Updated", Slug: "updated", Category: "fishing", GameType: "fish", IsEnabled: true, RawMetadata: []byte(`{}`)}
	for _, categoryID := range []uint64{99, 0} {
		require.NoError(t, db.Model(&row).Updates(map[string]any{"platform_category_id": categoryID}).Error)
		_, _, err = New(db).UpsertGame(context.Background(), &provider, upstream)
		require.NoError(t, err)
		var saved model.GameGames
		require.NoError(t, db.Take(&saved, row.ID).Error)
		require.Equal(t, categoryID, saved.PlatformCategoryID)
		require.Equal(t, "curated", saved.PlatformCategoryCode)
	}
}
