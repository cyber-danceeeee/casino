package service

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/model"
)

func normalizeCurrencyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func configuredScopeCurrency(ctx context.Context, db *gorm.DB, scope content.Scope) string {
	var row struct {
		SiteCurrency string `gorm:"column:site_currency"`

		CountryCurrency string `gorm:"column:country_currency"`
	}
	query := db.WithContext(ctx).Table("game_site_config s").
		Select("s.currency_code site_currency,c.currency_code country_currency").
		Joins("LEFT JOIN game_country_configs c ON c.country_code=s.country_code AND c.status=1 AND c.deleted_at IS NULL").
		Where("s.id = ? AND s.deleted_at IS NULL", scope.SiteID)
	if err := query.Take(&row).Error; err != nil {
		return ""
	}
	for _, value := range []string{row.SiteCurrency, row.CountryCurrency} {
		if currency := normalizeCurrencyCode(value); currency != "" {
			return currency
		}
	}
	return ""
}

func reconcileWalletCurrency(ctx context.Context, db *gorm.DB, scope content.Scope, wallet *model.GameUserWallets, expected string) (bool, error) {
	expected = normalizeCurrencyCode(expected)
	if expected == "" {
		return false, nil
	}
	if strings.EqualFold(wallet.Currency, expected) {
		wallet.Currency = expected
		return true, nil
	}
	configured := configuredScopeCurrency(ctx, db, scope)
	if configured == "" || !strings.EqualFold(configured, expected) {
		return false, nil
	}
	if err := content.ApplySite(db.WithContext(ctx).Model(&model.GameUserWallets{}), "", scope).
		Where("id = ? AND user_id = ?", wallet.ID, wallet.UserID).
		Update("currency", expected).Error; err != nil {
		return false, err
	}
	wallet.Currency = expected
	return true, nil
}
