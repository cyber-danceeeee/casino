package systemadmin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"casino/internal/singlesite"
	"casino/internal/tenancy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EnsureSingleSite adopts a sole existing site or creates one on an empty
// database. Never merge players, wallets or independent brands implicitly.
func (s *Service) EnsureSingleSite() error {
	var scope tenancy.Scope
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var binding struct {
			ID       uint64
			TenantID uint64
			SiteID   uint64
		}
		if err := tx.Table("casino_site_binding").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=1").Take(&binding).Error; err != nil {
			return fmt.Errorf("single-site migration is required: %w", err)
		}
		var sites, tenants int64
		if err := tx.Table("game_saas_sites").Where("deleted_at IS NULL").Count(&sites).Error; err != nil {
			return err
		}
		if err := tx.Table("game_saas_tenants").Where("deleted_at IS NULL").Count(&tenants).Error; err != nil {
			return err
		}
		if sites > 1 || tenants > 1 {
			return fmt.Errorf("single-site conversion requires exactly one existing site and owner; found %d sites and %d owners; preserve a database backup and select the site before upgrading", sites, tenants)
		}
		var unscoped int64
		if err := tx.Table("game_users").Where("tenant_id=0 OR site_id=0").Count(&unscoped).Error; err != nil {
			return err
		}
		if unscoped > 0 {
			return fmt.Errorf("%d legacy players have no site binding; migrate their related business data before upgrading", unscoped)
		}
		if sites == 1 {
			var site struct {
				ID       uint64
				TenantID uint64
			}
			if err := tx.Table("game_saas_sites").Select("id,tenant_id").Where("deleted_at IS NULL").Take(&site).Error; err != nil {
				return err
			}
			var ownerCount int64
			if err := tx.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", site.TenantID).Count(&ownerCount).Error; err != nil {
				return err
			}
			if ownerCount != 1 {
				return fmt.Errorf("existing site's owner record is missing")
			}
			scope = tenancy.Scope{TenantID: site.TenantID, SiteID: site.ID}
		} else {
			if binding.SiteID != 0 {
				return fmt.Errorf("bound H5 site was removed; restore it before starting")
			}
			defaults, err := s.provisionCountryDefaults(tx, "")
			if err != nil {
				return err
			}
			var ownerID uint64
			if tenants == 1 {
				if err := tx.Table("game_saas_tenants").Select("id").Where("deleted_at IS NULL").Take(&ownerID).Error; err != nil {
					return err
				}
			} else {
				input := TenantInput{TenantCode: "casino", TenantName: "Casino", CountryCode: defaults.CountryCode, DefaultCurrency: defaults.CurrencyCode, DefaultLanguage: defaults.DefaultLanguage, Timezone: defaults.Timezone, PlanCode: "STANDARD", PackageCode: "STANDARD"}
				if err := tx.Table("game_saas_tenants").Create(provisionTenantRow(input, 1)).Error; err != nil {
					return err
				}
				if err := tx.Table("game_saas_tenants").Select("id").Where("tenant_code=? AND deleted_at IS NULL", "casino").Take(&ownerID).Error; err != nil {
					return err
				}
			}
			input := SiteInput{TenantID: ownerID, SiteCode: "main", SiteName: "Casino", CountryCode: defaults.CountryCode, CurrencyCode: defaults.CurrencyCode, DefaultLanguage: defaults.DefaultLanguage, Timezone: defaults.Timezone,
				BrandConfig: map[string]any{"name": "Casino", "appName": "Casino"}, ThemeConfig: map[string]any{"primary": "#3d8aef", "button": "#2f69d1", "background": "#050b18", "text": "#ffffff"}, FeatureConfig: map[string]any{}, MaintenanceConfig: map[string]any{"enabled": false}, SEOConfig: map[string]any{"title": "Casino"}}
			row := siteInputRow(input, true, 1)
			row["site_code"] = "main"
			row["is_initialized"] = 1
			stripUnavailableSiteColumns(tx, row)
			if err := tx.Table("game_saas_sites").Create(row).Error; err != nil {
				return err
			}
			var siteID uint64
			if err := tx.Table("game_saas_sites").Select("id").Where("tenant_id=? AND site_code=? AND deleted_at IS NULL", ownerID, "main").Take(&siteID).Error; err != nil {
				return err
			}
			scope = tenancy.Scope{TenantID: ownerID, SiteID: siteID}
			if _, err := provisionLobbySections(tx, ownerID, siteID); err != nil {
				return err
			}
			if err := provisionTenantOperationDefaults(tx, ownerID, siteID, defaults.CurrencyCode, defaults.CountryCode, defaults.DefaultLanguage, 1); err != nil {
				return err
			}
			if _, err := provisionProviders(tx, ownerID, siteID, nil, -1); err != nil {
				return err
			}
			if _, err := provisionGames(tx, ownerID, siteID, nil, -1, 1); err != nil {
				return err
			}
		}
		if binding.SiteID != 0 && (binding.SiteID != scope.SiteID || binding.TenantID != scope.TenantID) {
			return fmt.Errorf("stored single-site binding does not match the existing site")
		}
		var duplicates int64
		if err := tx.Raw("SELECT COUNT(*) FROM (SELECT username FROM game_admin_accounts WHERE deleted_at IS NULL GROUP BY username HAVING COUNT(*)>1) duplicate_accounts").Scan(&duplicates).Error; err != nil {
			return err
		}
		if duplicates > 0 {
			return fmt.Errorf("administrator usernames must be unique before converting to a single login")
		}
		return tx.Table("casino_site_binding").Where("id=1").Updates(map[string]any{"tenant_id": scope.TenantID, "site_id": scope.SiteID}).Error
	})
	if err != nil {
		return err
	}
	singlesite.Bind(scope)
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
	if !ok || actor.TenantID != scope.TenantID || actor.SiteID != scope.SiteID {
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
		return tx.Table("game_saas_sites").Where("id=?", scope.SiteID).Updates(map[string]any{"h5_domain": address, "primary_domain": address}).Error
	})
}

func (s *Service) validateSingleSiteMarket(input SiteInput, scope tenancy.Scope) error {
	defaults, err := s.provisionCountryDefaults(s.db, strings.ToUpper(strings.TrimSpace(input.CountryCode)))
	if err != nil {
		return err
	}
	if strings.ToUpper(strings.TrimSpace(input.CurrencyCode)) != defaults.CurrencyCode {
		return NewError("INVALID_CURRENCY", "currency must match the configured country")
	}
	var before struct{ CurrencyCode string }
	if err := s.db.Table("game_saas_sites").Select("currency_code").Where("id=?", scope.SiteID).Take(&before).Error; err != nil {
		return err
	}
	if before.CurrencyCode != strings.ToUpper(strings.TrimSpace(input.CurrencyCode)) {
		var wallets int64
		if err := s.db.Table("game_user_wallets").Where("tenant_id=? AND site_id=?", scope.TenantID, scope.SiteID).Count(&wallets).Error; err != nil {
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
