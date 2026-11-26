package ag

import "strings"

type currencyMapper struct{}

func (currencyMapper) ToProviderCurrency(country string, currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" || strings.Contains(currency, "_") {
		return currency
	}
	country = strings.ToUpper(strings.TrimSpace(country))
	if country == "" {
		return currency
	}
	return country + "_" + currency
}

func (currencyMapper) ToPlatformCurrency(providerCurrency string) string {
	value := strings.ToUpper(strings.TrimSpace(providerCurrency))
	if value == "" {
		return ""
	}
	parts := strings.Split(value, "_")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return value
}
