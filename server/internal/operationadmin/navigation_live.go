package operationadmin

import (
	"casino/internal/content"
	"time"

	"gorm.io/gorm"
)

func navigationValues(configuration H5NavigationConfiguration, actorID uint64) (map[string]any, error) {
	values, err := content.ConfigurationValues(configuration, map[string]string{"MenuCode": "menu_code", "Labels": "labels_json", "IconAssetID": "icon_asset_id", "LottieAssetID": "lottie_asset_id", "SoundAssetID": "sound_asset_id", "DisplayPosition": "display_position", "ActionType": "action_type", "ActionTarget": "action_target", "ProviderCode": "provider_code", "GameID": "game_id", "ProviderGameCode": "provider_game_code", "GameName": "game_name", "SortOrder": "sort_order", "BadgeType": "badge_type", "VisibilityScope": "visibility_scope", "CountryCodes": "country_codes", "LanguageCodes": "language_codes", "StartAt": "start_at", "EndAt": "end_at", "Enabled": "enabled"})
	if err != nil {
		return nil, err
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["updated_by"] = actorID
	values["updated_at"] = time.Now()
	values["page_route"] = ""
	if configuration.ActionType == "OPEN_INTERNAL_PAGE" {
		values["page_route"] = configuration.ActionTarget
	}
	values["label_zh"] = configuration.Labels["zh-CN"]
	values["label_en"] = configuration.Labels["en-US"]
	return values, nil
}
