package market

import (
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestGameRuleMatchingUsesExactCountryAndCurrency(t *testing.T) {
	m := Market{Country: "PK", Currency: "PKR"}
	require.False(t, m.ProviderAllowed("", ""))
	require.True(t, m.ProviderAllowed("pk,ca", "PKR,CAD"))
	require.False(t, m.ProviderAllowed("PKX", "PKR"))
	require.False(t, m.ProviderAllowed("PK", "CAD"))
	require.True(t, m.GameAllowed("", ""))
	require.True(t, m.GameAllowed(`["PK","CA"]`, `["PKR"]`))
	require.False(t, m.GameAllowed("CA", "PKR"))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, _ := db.DB()
	defer raw.Close()
	require.NoError(t, db.Exec(`CREATE TABLE games (id INTEGER PRIMARY KEY,supported_countries TEXT,supported_currencies TEXT)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO games VALUES(1,'',''),(2,'PK','PKR'),(3,'PKX','PKR'),(4,'PK','CAD'),(5,'["PK","CA"]','["PKR"]')`).Error)
	var ids []int
	require.NoError(t, ApplyGames(db.Table("games g"), "g", m).Order("id").Pluck("id", &ids).Error)
	require.Equal(t, []int{1, 2, 5}, ids)
}
