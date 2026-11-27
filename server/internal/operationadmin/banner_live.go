package operationadmin

import (
	"casino/internal/content"
	"time"

	"gorm.io/gorm"
)

func bannerValues(configuration SiteBannerConfiguration, actorID uint64) (map[string]any, error) {
	values, err := content.ConfigurationValues(configuration, map[string]string{"Name": "name", "Title": "title", "Subtitle": "subtitle", "DesktopAssetID": "desktop_asset_id", "MobileAssetID": "mobile_asset_id", "ActionType": "action_type", "ActionTarget": "action_target", "SortOrder": "sort_order", "LanguageCode": "language_code", "CountryCode": "country_code", "CurrencyCode": "currency_code", "VisibilityScope": "visibility_scope", "StartAt": "start_at", "EndAt": "end_at", "Enabled": "enabled"})
	if err != nil {
		return nil, err
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["updated_by"] = actorID
	values["updated_at"] = time.Now()

	return values, nil
}
