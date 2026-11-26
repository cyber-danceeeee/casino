package softaggregator

import "strings"

// SoftAggregator accepts the platform ISO currency without a country prefix.
type currencyMapper struct{}

func (currencyMapper) ToProviderCurrency(_, currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func (currencyMapper) ToPlatformCurrency(providerCurrency string) string {
	return strings.ToUpper(strings.TrimSpace(providerCurrency))
}
