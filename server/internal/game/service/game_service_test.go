package service

import (
	"testing"

	"casino/internal/game/domain"
)

func TestUniqueProviderGamesNormalizesAndDeduplicatesIDs(t *testing.T) {
	games := uniqueProviderGames([]domain.Game{
		{ProviderGameID: "game-a", Name: "first"},
		{ProviderGameID: "GAME-A", Name: "duplicate"},
		{ProviderGameID: " game-b ", Name: "second"},
		{ProviderGameID: "", Name: "invalid"},
	})
	if len(games) != 2 {
		t.Fatalf("unique game count = %d, want 2", len(games))
	}
	if games[0].Name != "first" || games[1].ProviderGameID != "game-b" {
		t.Fatalf("unexpected normalized games: %#v", games)
	}
}
