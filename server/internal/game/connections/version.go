package connections

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"gorm.io/gorm"
)

// PolicyVersion invalidates cached H5 content after connection/rule/status edits.
// No credential values or ciphertext participate in this public cache key.
func PolicyVersion(db *gorm.DB) string {
	var accounts []struct {
		ProviderCode string
		Version      uint64
		Enabled      bool
	}
	if err := db.Table("game_provider_connections").Select("provider_code,version,enabled").Order("provider_code").Scan(&accounts).Error; err != nil {
		return "unavailable"
	}
	var catalog []struct {
		ProviderCode        string
		SiteID              uint64
		Status              int
		SupportedCountries  string
		SupportedCurrencies string
		LogoURL             string
	}
	if err := db.Table("game_game_providers").Select("provider_code,site_id,status,supported_countries,supported_currencies,logo_url").Where("site_id IN (0,1) AND deleted_at IS NULL").Order("provider_code,site_id,id").Scan(&catalog).Error; err != nil {
		return "unavailable"
	}
	raw, _ := json.Marshal([]any{accounts, catalog})
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:8])
}
