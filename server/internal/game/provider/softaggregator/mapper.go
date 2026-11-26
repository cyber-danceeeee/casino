package softaggregator

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
	gameType := normalizeGameType(item.Type)
	providerGameID := strings.TrimSpace(item.IDHash)
	if providerGameID == "" {
		providerGameID = strings.TrimSpace(item.Name)
	}
	return domain.Game{
		ProviderCode:      Code,
		ProviderGameID:    providerGameID,
		Name:              strings.TrimSpace(item.Name),
		Slug:              stableSlug(providerGameID, item.Name),
		Category:          strings.TrimSpace(item.Category),
		GameType:          gameType,
		ThumbnailURL:      strings.TrimSpace(item.Image),
		SupportsDemo:      item.PlayForFunSupported,
		SupportsFreeSpins: item.FreeroundsSupported,
		IsEnabled:         true,
		RawMetadata:       raw,
	}
}

func normalizeGameType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "video-slots", "slots", "slot":
		return "slot"
	case "live", "live-casino", "casino":
		return "live"
	case "sports", "sport":
		return "sports"
	case "fish", "fishing":
		return "fish"
	case "table", "table-games", "card":
		return "table"
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
