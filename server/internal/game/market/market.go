// Package market defines the published deployment market shared by catalog and launch.
package market

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrUnavailable = errors.New("published H5 country and currency are not configured")

type Market struct {
	Country  string
	Currency string
}

func Load(db *gorm.DB) (Market, error) {
	var row struct {
		CountryCode  string
		CurrencyCode string
	}
	if err := db.Table("game_site_config").Select("country_code,currency_code").Where("id=1 AND deleted_at IS NULL").Take(&row).Error; err != nil {
		return Market{}, err
	}
	m := Market{strings.ToUpper(strings.TrimSpace(row.CountryCode)), strings.ToUpper(strings.TrimSpace(row.CurrencyCode))}
	if !Letters(m.Country, 2) || !Letters(m.Currency, 3) {
		return Market{}, ErrUnavailable
	}
	return m, nil
}
func Letters(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
func Values(value string) []string {
	value = strings.TrimSpace(value)
	var parts []string
	if strings.HasPrefix(value, "[") {
		if json.Unmarshal([]byte(value), &parts) != nil {
			return []string{"INVALID"}
		}
	} else {
		parts = strings.Split(value, ",")
	}
	out := []string{}
	seen := map[string]bool{}
	for _, s := range parts {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s != "" && !seen[s] {
			out = append(out, s)
			seen[s] = true
		}
	}
	return out
}
func Allows(rule, value string, emptyAllowed bool) bool {
	vs := Values(rule)
	if len(vs) == 0 {
		return emptyAllowed
	}
	for _, v := range vs {
		if v == strings.ToUpper(strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}
func (m Market) ProviderAllowed(countries, currencies string) bool {
	return Allows(countries, m.Country, false) && Allows(currencies, m.Currency, false)
}
func (m Market) GameAllowed(countries, currencies string) bool {
	return Allows(countries, m.Country, true) && Allows(currencies, m.Currency, true)
}

// ApplyGames filters optional game restrictions. Provider restrictions are explicit
// and enforced separately using the configured, credential-ready connection list.
func ApplyGames(db *gorm.DB, alias string, m Market) *gorm.DB {
	for _, r := range []struct{ column, value string }{{"supported_countries", m.Country}, {"supported_currencies", m.Currency}} {
		col := r.column
		if alias != "" {
			col = alias + "." + col
		}
		norm := "UPPER(REPLACE(REPLACE(REPLACE(REPLACE(COALESCE(" + col + ",''),' ',''),'[',''),']',''),'\"',''))"
		list := "CONCAT(','," + norm + ",',')"
		if db.Dialector.Name() == "sqlite" {
			list = "(',' || " + norm + " || ',')"
		}
		db = db.Where("("+norm+"='' OR "+list+" LIKE ?)", "%,"+r.value+",%")
	}
	return db
}
