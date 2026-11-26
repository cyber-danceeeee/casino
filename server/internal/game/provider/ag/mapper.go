package ag

import (
	"crypto/sha1" //nolint:gosec // Used only for stable non-secret catalog slugs.
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"casino/internal/game/domain"
)

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func mapGame(item gameListItem) domain.Game {
	raw, _ := json.Marshal(item)
	providerGameID := strings.TrimSpace(item.GameCode)
	if providerGameID == "" {
		providerGameID = strings.TrimSpace(item.OfficialGameCode)
	}
	return domain.Game{
		ProviderCode:      Code,
		ProviderGameID:    providerGameID,
		Name:              strings.TrimSpace(item.GameName),
		Slug:              stableSlug(providerGameID, item.GameName),
		Category:          strings.TrimSpace(item.ProviderCategoryCode),
		GameType:          normalizeGameType(item.GameType),
		SupportsDemo:      false,
		SupportsFreeSpins: false,
		IsEnabled:         item.Status == 1,
		RawMetadata:       raw,
	}
}

func deduplicateGames(games []domain.Game) []domain.Game {
	result := make([]domain.Game, 0, len(games))
	indexByID := make(map[string]int, len(games))
	for _, game := range games {
		providerGameID := strings.TrimSpace(game.ProviderGameID)
		if providerGameID == "" {
			continue
		}
		index, exists := indexByID[providerGameID]
		if !exists {
			indexByID[providerGameID] = len(result)
			result = append(result, game)
			continue
		}
		if promotionalCategory(result[index].Category) && !promotionalCategory(game.Category) {
			result[index] = game
		}
	}
	return result
}

func promotionalCategory(category string) bool {
	category = strings.ToUpper(strings.TrimSpace(category))
	return strings.HasPrefix(category, "HOT_") || strings.HasPrefix(category, "NEW_") || category == "HOT" || category == "NEW"
}

func normalizeGameType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "slot", "slots", "slot_game", "egame":
		return "slot"
	case "live", "live-casino", "casino":
		return "live"
	case "sports", "sport":
		return "sports"
	case "fish", "fishing":
		return "fish"
	case "table", "table-games", "card", "poker":
		return "table"
	case "crash", "instant":
		return "crash"
	case "bingo":
		return "bingo"
	default:
		if strings.TrimSpace(value) == "" {
			return "slot"
		}
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func stableSlug(values ...string) string {
	joined := strings.ToLower(strings.Join(values, "-"))
	slug := strings.Trim(slugPattern.ReplaceAllString(joined, "-"), "-")
	if slug == "" {
		slug = "game"
	}
	sum := sha1.Sum([]byte(joined)) //nolint:gosec
	suffix := hex.EncodeToString(sum[:])[:8]
	if len(slug) > 80 {
		slug = strings.Trim(slug[:80], "-")
	}
	return slug + "-" + suffix
}
