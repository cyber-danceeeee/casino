package systemadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type configurationSnapshot struct {
	Site          map[string]any   `json:"site"`
	Games         []map[string]any `json:"games"`
	Banners       []map[string]any `json:"banners"`
	Announcements []map[string]any `json:"announcements"`
	Popups        []map[string]any `json:"popups"`
	Navigation    []map[string]any `json:"navigation"`
}

func captureConfigurationSnapshot(tx *gorm.DB, siteID uint64) (configurationSnapshot, error) {
	var site SiteRow
	if err := tx.Table("game_site_config s").Select(siteSelect()).Where("s.id = ? AND s.deleted_at IS NULL", siteID).Take(&site).Error; err != nil {
		return configurationSnapshot{}, err
	}
	snapshot := configurationSnapshot{Site: siteConfiguration(site)}
	tables := []struct {
		name    string
		columns string
		target  *[]map[string]any
	}{{"game_games", "id,platform_enabled,is_hot,is_featured,is_new,sort,status,maintenance_message,platform_category_id,platform_category_code", &snapshot.Games}, {"game_site_banners", "id,name,title,subtitle,desktop_asset_id,mobile_asset_id,action_type,action_target,sort_order,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,configuration_json", &snapshot.Banners}, {"game_site_announcements", "id,title,summary,content,announcement_type,priority,sort_order,is_pinned,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,configuration_json", &snapshot.Announcements}, {"game_popup_campaigns", "id,popup_component_id,campaign_name,action_type,action_target,trigger_type,trigger_value,frequency_type,display_limit,cooldown_minutes,visibility_scope,start_at,end_at,priority,enabled,configuration_json", &snapshot.Popups}, {"game_h5_navigation_items", "id,menu_code,label_zh,label_en,labels_json,icon_asset_id,lottie_asset_id,sound_asset_id,display_position,action_type,action_target,page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,country_codes,language_codes,start_at,end_at,enabled,configuration_json", &snapshot.Navigation}}
	for _, table := range tables {
		if err := tx.Table(table.name).Select(table.columns).Where("site_id = ? AND deleted_at IS NULL", siteID).Order("id ASC").Find(table.target).Error; err != nil {
			return configurationSnapshot{}, err
		}
	}
	return snapshot, nil
}

func restoreConfigurationSnapshot(tx *gorm.DB, siteID uint64, version uint64, actorID uint64, snapshot configurationSnapshot) error {
	site := snapshot.Site
	var country struct {
		ID           uint64
		CurrencyCode string
	}
	if err := tx.Table("game_country_configs").Select("id,currency_code").Where("country_code=? AND deleted_at IS NULL", site["countryCode"]).Take(&country).Error; err != nil {
		return NewError("RESTORE_MARKET_INVALID", "backup country is not configured")
	}
	if country.CurrencyCode != stringFromAny(site["currencyCode"]) {
		return NewError("RESTORE_MARKET_INVALID", "backup currency must match its country")
	}
	if _, err := time.LoadLocation(stringFromAny(site["timezone"])); err != nil {
		return NewError("INVALID_TIMEZONE", "backup timezone is invalid")
	}
	if err := activateCountry(tx, country.ID, Actor{ID: actorID}); err != nil {
		return err
	}
	updates := map[string]any{"site_name": site["siteName"], "country_code": site["countryCode"], "currency_code": site["currencyCode"], "default_language": site["defaultLanguage"], "timezone": site["timezone"], "brand_config": marshal(site["brandConfig"]), "theme_config": marshal(site["themeConfig"]), "feature_config": marshal(site["featureConfig"]), "maintenance_config": marshal(site["maintenanceConfig"]), "seo_config": marshal(site["seoConfig"]), "config_version": version, "updated_by": actorID, "updated_at": now()}

	if err := tx.Table("game_site_config").Where("id = ? AND deleted_at IS NULL", siteID).Updates(updates).Error; err != nil {
		return err
	}
	tables := []struct {
		name    string
		rows    []map[string]any
		disable map[string]any
		allowed map[string]bool
	}{{"game_games", snapshot.Games, map[string]any{"platform_enabled": 0, "status": 3}, configurationColumns("id,platform_enabled,is_hot,is_featured,is_new,sort,status,maintenance_message,platform_category_id,platform_category_code")}, {"game_site_banners", snapshot.Banners, map[string]any{"enabled": 0}, configurationColumns("id,name,title,subtitle,desktop_asset_id,mobile_asset_id,action_type,action_target,sort_order,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,configuration_json")}, {"game_site_announcements", snapshot.Announcements, map[string]any{"enabled": 0}, configurationColumns("id,title,summary,content,announcement_type,priority,sort_order,is_pinned,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,configuration_json")}, {"game_popup_campaigns", snapshot.Popups, map[string]any{"enabled": 0}, configurationColumns("id,popup_component_id,campaign_name,action_type,action_target,trigger_type,trigger_value,frequency_type,display_limit,cooldown_minutes,visibility_scope,start_at,end_at,priority,enabled,configuration_json")}, {"game_h5_navigation_items", snapshot.Navigation, map[string]any{"enabled": 0}, configurationColumns("id,menu_code,label_zh,label_en,labels_json,icon_asset_id,lottie_asset_id,sound_asset_id,display_position,action_type,action_target,page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,country_codes,language_codes,start_at,end_at,enabled,configuration_json")}}
	for _, table := range tables {
		if err := tx.Table(table.name).Where("site_id = ? AND deleted_at IS NULL", siteID).Updates(table.disable).Error; err != nil {
			return err
		}
		for _, raw := range table.rows {
			id := uint64FromJSON(raw["id"])
			values := map[string]any{}
			for key, value := range raw {
				if key != "id" && table.allowed[key] {
					values[key] = value
				}
			}
			values["config_version"] = version
			values["updated_at"] = now()
			values["deleted_at"] = nil
			if err := tx.Table(table.name).Where("id = ? AND site_id = ?", id, siteID).Updates(values).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func configurationColumns(value string) map[string]bool {
	result := map[string]bool{}
	for _, column := range strings.Split(value, ",") {
		result[column] = true
	}
	return result
}

func uint64FromJSON(value any) uint64 {
	switch v := value.(type) {
	case float64:
		return uint64(v)
	case uint64:
		return v
	case int64:
		return uint64(v)
	case json.Number:
		result, _ := v.Int64()
		return uint64(result)
	default:
		return parseID(fmt.Sprint(value))
	}
}

func splitCSV(value string) []string {
	result := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
