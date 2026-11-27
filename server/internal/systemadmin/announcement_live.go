package systemadmin

import (
	"casino/internal/content"

	"gorm.io/gorm"
)

func announcementValues(configuration AnnouncementConfiguration, actorID uint64) (map[string]any, error) {
	values, err := content.ConfigurationValues(configuration, map[string]string{"Title": "title", "Summary": "summary", "Content": "content", "AnnouncementType": "announcement_type", "Priority": "priority", "SortOrder": "sort_order", "IsPinned": "is_pinned", "LanguageCode": "language_code", "CountryCode": "country_code", "CurrencyCode": "currency_code", "VisibilityScope": "visibility_scope", "StartAt": "start_at", "EndAt": "end_at", "Enabled": "enabled"})
	if err != nil {
		return nil, err
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["updated_by"] = actorID
	values["updated_at"] = now()

	return values, nil
}
