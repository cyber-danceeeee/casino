package systemadmin

import (
	"casino/internal/content"

	"gorm.io/gorm"
)

func popupValues(configuration PopupConfiguration, actorID uint64) (map[string]any, error) {
	values, err := content.ConfigurationValues(configuration, map[string]string{"PopupComponentID": "popup_component_id", "CampaignName": "campaign_name", "ActionType": "action_type", "ActionTarget": "action_target", "TriggerType": "trigger_type", "TriggerValue": "trigger_value", "FrequencyType": "frequency_type", "DisplayLimit": "display_limit", "CooldownMinutes": "cooldown_minutes", "VisibilityScope": "visibility_scope", "StartAt": "start_at", "EndAt": "end_at", "Priority": "priority", "Enabled": "enabled"})
	if err != nil {
		return nil, err
	}
	values["config_version"] = gorm.Expr("config_version+1")
	values["updated_by"] = actorID
	values["updated_at"] = now()

	return values, nil
}
