package systemadmin

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type provisionCountryDefaults struct {
	CountryCode     string
	CurrencyCode    string
	DefaultLanguage string
	Timezone        string
}

func (s *Service) provisionCountryDefaults(tx *gorm.DB, countryCode string) (provisionCountryDefaults, error) {
	var row struct {
		CountryCode     string `gorm:"column:country_code"`
		CurrencyCode    string `gorm:"column:currency_code"`
		DefaultLanguage string `gorm:"column:default_language"`
		Timezone        string `gorm:"column:timezone"`
	}
	query := tx.Table("game_country_configs").
		Select("country_code,currency_code,default_language,timezone").
		Where("deleted_at IS NULL AND status=1")
	if countryCode != "" {
		if err := query.Where("country_code=?", countryCode).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return provisionCountryDefaults{}, NewError("COUNTRY_NOT_FOUND", "country is not configured or enabled")
			}
			return provisionCountryDefaults{}, err
		}
		return validateProvisionCountryDefaults(row)
	}
	err := tx.Table("game_country_configs").
		Select("country_code,currency_code,default_language,timezone").
		Where("deleted_at IS NULL AND status=1").
		Order("is_default DESC,id ASC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return provisionCountryDefaults{}, NewError("COUNTRY_NOT_FOUND", "no enabled country configuration is available")
		}
		return provisionCountryDefaults{}, err
	}
	return validateProvisionCountryDefaults(row)
}

func validateProvisionCountryDefaults(row struct {
	CountryCode     string `gorm:"column:country_code"`
	CurrencyCode    string `gorm:"column:currency_code"`
	DefaultLanguage string `gorm:"column:default_language"`
	Timezone        string `gorm:"column:timezone"`
}) (provisionCountryDefaults, error) {
	defaults := provisionCountryDefaults{
		CountryCode: strings.ToUpper(strings.TrimSpace(row.CountryCode)), CurrencyCode: strings.ToUpper(strings.TrimSpace(row.CurrencyCode)),
		DefaultLanguage: strings.TrimSpace(row.DefaultLanguage), Timezone: strings.TrimSpace(row.Timezone),
	}
	if defaults.CountryCode == "" || defaults.CurrencyCode == "" || defaults.DefaultLanguage == "" || defaults.Timezone == "" {
		return provisionCountryDefaults{}, NewError("COUNTRY_CONFIG_INCOMPLETE", "country currency, language, and timezone must be configured")
	}
	return defaults, nil
}

// provisionGameCategories gives every site its own editable copy of the
// platform catalog. Game relations are created by provisionGames after the
// H5 providers and game rows have been provisioned.
