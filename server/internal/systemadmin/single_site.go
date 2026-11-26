package systemadmin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"casino/internal/content"
	"casino/internal/singlesite"

	"gorm.io/gorm"
)

// EnsureSingleSite adopts a sole existing site or creates one on an empty
// database. Never merge players, wallets or independent brands implicitly.
func (s *Service) EnsureSiteConfig() error {
	var count int64
	if err := s.db.Table("game_site_config").Where("id=1 AND deleted_at IS NULL").Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("H5 configuration is missing; run the database installer")
	}
	return nil
}

func singleSiteURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", NewError("INVALID_H5_URL", "H5 address must be an HTTP or HTTPS site root")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func (s *Service) SaveSingleSiteSettings(input SiteInput, actor Actor) error {
	scope, ok := singlesite.Scope()
	if !ok || actor.SiteID != scope.SiteID {
		return NewError("SITE_NOT_READY", "H5 site is not initialized")
	}
	address, err := singleSiteURL(input.H5Domain)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := New(tx).UpdateSite(scope.SiteID, input, actor); err != nil {
			return err
		}
		return tx.Table("game_site_config").Where("id=?", scope.SiteID).Updates(map[string]any{"h5_domain": address, "primary_domain": address}).Error
	})
}

func (s *Service) validateSingleSiteMarket(input SiteInput, scope content.Scope) error {
	defaults, err := s.provisionCountryDefaults(s.db, strings.ToUpper(strings.TrimSpace(input.CountryCode)))
	if err != nil {
		return err
	}
	if strings.ToUpper(strings.TrimSpace(input.CurrencyCode)) != defaults.CurrencyCode {
		return NewError("INVALID_CURRENCY", "currency must match the configured country")
	}
	var before struct{ CurrencyCode string }
	if err := s.db.Table("game_site_config").Select("currency_code").Where("id=?", scope.SiteID).Take(&before).Error; err != nil {
		return err
	}
	if before.CurrencyCode != strings.ToUpper(strings.TrimSpace(input.CurrencyCode)) {
		var wallets int64
		if err := s.db.Table("game_user_wallets").Where("site_id = ?", scope.SiteID).Count(&wallets).Error; err != nil {
			return err
		}
		if wallets > 0 {
			return NewError("CURRENCY_CHANGE_REQUIRES_MIGRATION", "existing player wallets must be migrated before changing the site currency")
		}
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return NewError("INVALID_TIMEZONE", "select a valid IANA timezone")
	}
	return nil
}
