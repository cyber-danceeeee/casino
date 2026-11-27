package systemadmin

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func siteConfiguration(row SiteRow) map[string]any {
	return map[string]any{
		"siteName":          row.SiteName,
		"countryCode":       row.CountryCode,
		"currencyCode":      row.CurrencyCode,
		"defaultLanguage":   row.DefaultLanguage,
		"timezone":          row.Timezone,
		"brandConfig":       parseJSONMap(row.BrandConfig),
		"themeConfig":       parseJSONMap(row.ThemeConfig),
		"featureConfig":     parseJSONMap(row.FeatureConfig),
		"maintenanceConfig": parseJSONMap(row.MaintenanceConfig),
		"seoConfig":         parseJSONMap(row.SEOConfig),
	}
}
func (s *Service) siteRowForActor(db *gorm.DB, id uint64, actor Actor, forUpdate bool) (SiteRow, error) {
	var row SiteRow
	query := db.Table("game_site_config s").
		Select(siteSelect()).
		Where("s.id=? AND s.deleted_at IS NULL", id)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if !actor.Platform {
		query = query
		if actor.SiteID > 0 {
			query = query.Where("s.id=?", actor.SiteID)
		}
	}
	if err := query.Take(&row).Error; err != nil {
		return row, rowNotFound("site")
	}
	return row, nil
}
func clearWhiteLabelConfigCache(tx *gorm.DB, siteID uint64) error {
	return nil
}
func siteConfigMeta(row SiteRow) map[string]any {
	return map[string]any{"configVersion": row.ConfigVersion, "updatedAt": row.UpdatedAt}
}
func touchSiteConfiguration(tx *gorm.DB, siteID uint64, actorID uint64) error {
	return tx.Table("game_site_config").Where("id=? AND deleted_at IS NULL", siteID).Updates(map[string]any{"config_version": gorm.Expr("config_version+1"), "updated_by": actorID, "updated_at": now()}).Error
}
