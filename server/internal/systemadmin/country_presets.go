package systemadmin

import (
	"gorm.io/gorm"
	"strings"
)

func countryNotInUse(db *gorm.DB, code string) error {
	var count int64
	if err := db.Table("game_site_config").Where("deleted_at IS NULL AND (country_code=? OR country_code_draft=?)", code, code).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return NewError("COUNTRY_IN_USE", "this country is selected in H5 settings; choose another country there before disabling, deleting or changing its currency")
	}
	return nil
}

// Defaults for the deployment markets supported out of the box. Other countries
// can still be configured explicitly; these values never select the live market.
func CountryPresets() []CountryInput {
	return []CountryInput{
		{CountryCode: "ID", ISOCountryCode: "ID", CountryNameZh: "印度尼西亚", CountryNameEn: "Indonesia", CurrencyCode: "IDR", CurrencyNameZh: "印度尼西亚盾", CurrencyNameEn: "Indonesian Rupiah", CurrencySymbol: "Rp", CurrencyDecimalPlaces: 0, Timezone: "Asia/Jakarta", DefaultLanguage: "id", SupportedLanguages: "id,en", PhoneCountryCode: "+62", PhoneLength: 0},
		{CountryCode: "IN", ISOCountryCode: "IN", CountryNameZh: "印度", CountryNameEn: "India", CurrencyCode: "INR", CurrencyNameZh: "印度卢比", CurrencyNameEn: "Indian Rupee", CurrencySymbol: "₹", CurrencyDecimalPlaces: 2, Timezone: "Asia/Kolkata", DefaultLanguage: "en", SupportedLanguages: "en,hi", PhoneCountryCode: "+91", PhoneLength: 10},
		{CountryCode: "PK", ISOCountryCode: "PK", CountryNameZh: "巴基斯坦", CountryNameEn: "Pakistan", CurrencyCode: "PKR", CurrencyNameZh: "巴基斯坦卢比", CurrencyNameEn: "Pakistani Rupee", CurrencySymbol: "₨", CurrencyDecimalPlaces: 2, Timezone: "Asia/Karachi", DefaultLanguage: "en", SupportedLanguages: "en,ur", PhoneCountryCode: "+92", PhoneLength: 10},
	}
}

func countryInputDefaults(input CountryInput) CountryInput {
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	for _, preset := range CountryPresets() {
		if preset.CountryCode != input.CountryCode {
			continue
		}
		// A minimal request selects the preset; a complete request retains the
		// operator's explicit language, precision, timezone and phone settings.
		minimal := strings.TrimSpace(input.CurrencyCode) == ""
		if input.CountryNameZh == "" {
			input.CountryNameZh = preset.CountryNameZh
		}
		if input.CountryNameEn == "" {
			input.CountryNameEn = preset.CountryNameEn
		}
		if input.CurrencyCode == "" {
			input.CurrencyCode = preset.CurrencyCode
		}
		if input.CurrencyNameZh == "" {
			input.CurrencyNameZh = preset.CurrencyNameZh
		}
		if input.CurrencyNameEn == "" {
			input.CurrencyNameEn = preset.CurrencyNameEn
		}
		if input.CurrencySymbol == "" {
			input.CurrencySymbol = preset.CurrencySymbol
		}
		if input.Timezone == "" {
			input.Timezone = preset.Timezone
		}
		if input.DefaultLanguage == "" {
			input.DefaultLanguage = preset.DefaultLanguage
		}
		if input.SupportedLanguages == "" {
			input.SupportedLanguages = preset.SupportedLanguages
		}
		if input.PhoneCountryCode == "" {
			input.PhoneCountryCode = preset.PhoneCountryCode
		}
		if minimal {
			input.PhoneLength, input.CurrencyDecimalPlaces = preset.PhoneLength, preset.CurrencyDecimalPlaces
		}
		break
	}
	if input.ISOCountryCode == "" {
		input.ISOCountryCode = input.CountryCode
	}
	input.CurrencyCode = strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}
	if input.DefaultLanguage == "" {
		input.DefaultLanguage = "en"
	}
	if input.SupportedLanguages == "" {
		input.SupportedLanguages = input.DefaultLanguage
	}
	return input
}
