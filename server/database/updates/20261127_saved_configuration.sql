-- Phase 1: adopt the last saved configuration, then let all saves take effect immediately.

-- Retain obsolete nullable columns until the new application has deployed successfully.

ALTER TABLE `game_site_announcements` ADD COLUMN configuration_json JSON NULL;

UPDATE `game_site_announcements` SET configuration_json=CASE
 WHEN publish_status='DRAFT' AND JSON_VALID(draft_payload) AND JSON_LENGTH(draft_payload)>0 THEN CAST(draft_payload AS JSON)
 WHEN JSON_VALID(published_payload) AND JSON_LENGTH(published_payload)>0 THEN CAST(published_payload AS JSON)
 ELSE JSON_OBJECT('title',`title`,'summary',`summary`,'content',`content`,'announcementType',`announcement_type`,'priority',`priority`,'sortOrder',`sort_order`,'isPinned',IF(`is_pinned`=1,CAST('true' AS JSON),CAST('false' AS JSON)),'languageCode',`language_code`,'countryCode',`country_code`,'currencyCode',`currency_code`,'visibilityScope',`visibility_scope`,'startAt',`start_at`,'endAt',`end_at`,'enabled',IF(`enabled`=1,CAST('true' AS JSON),CAST('false' AS JSON))) END;

UPDATE `game_site_announcements` SET `title`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.title') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.title')) ELSE `title` END,
 `summary`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.summary') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.summary')) ELSE `summary` END,
 `content`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.content') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.content')) ELSE `content` END,
 `announcement_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.announcementType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.announcementType')) ELSE `announcement_type` END,
 `priority`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.priority') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.priority')) AS SIGNED) ELSE `priority` END,
 `sort_order`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.sortOrder') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.sortOrder')) AS SIGNED) ELSE `sort_order` END,
 `is_pinned`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.isPinned') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.isPinned')) IN ('true','1'),1,0) ELSE `is_pinned` END,
 `language_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.languageCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.languageCode')) ELSE `language_code` END,
 `country_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.countryCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.countryCode')) ELSE `country_code` END,
 `currency_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.currencyCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.currencyCode')) ELSE `currency_code` END,
 `visibility_scope`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.visibilityScope') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.visibilityScope')) ELSE `visibility_scope` END,
 `start_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.startAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `start_at` END,
 `end_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.endAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `end_at` END,
 `enabled`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.enabled') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.enabled')) IN ('true','1'),1,0) ELSE `enabled` END,config_version=config_version+1,deleted_at=IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.deleted'))='true',NOW(3),deleted_at);

ALTER TABLE `game_popup_campaigns` ADD COLUMN configuration_json JSON NULL;

UPDATE `game_popup_campaigns` SET configuration_json=CASE
 WHEN publish_status='DRAFT' AND JSON_VALID(draft_payload) AND JSON_LENGTH(draft_payload)>0 THEN CAST(draft_payload AS JSON)
 WHEN JSON_VALID(published_payload) AND JSON_LENGTH(published_payload)>0 THEN CAST(published_payload AS JSON)
 ELSE JSON_OBJECT('popupComponentID',`popup_component_id`,'campaignName',`campaign_name`,'actionType',`action_type`,'actionTarget',`action_target`,'triggerType',`trigger_type`,'triggerValue',`trigger_value`,'frequencyType',`frequency_type`,'displayLimit',`display_limit`,'cooldownMinutes',`cooldown_minutes`,'visibilityScope',`visibility_scope`,'startAt',`start_at`,'endAt',`end_at`,'priority',`priority`,'enabled',IF(`enabled`=1,CAST('true' AS JSON),CAST('false' AS JSON))) END;

UPDATE `game_popup_campaigns` SET `popup_component_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.popupComponentID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.popupComponentID')) AS SIGNED) ELSE `popup_component_id` END,
 `campaign_name`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.campaignName') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.campaignName')) ELSE `campaign_name` END,
 `action_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionType')) ELSE `action_type` END,
 `action_target`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionTarget') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionTarget')) ELSE `action_target` END,
 `trigger_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.triggerType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.triggerType')) ELSE `trigger_type` END,
 `trigger_value`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.triggerValue') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.triggerValue')) ELSE `trigger_value` END,
 `frequency_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.frequencyType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.frequencyType')) ELSE `frequency_type` END,
 `display_limit`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.displayLimit') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.displayLimit')) AS SIGNED) ELSE `display_limit` END,
 `cooldown_minutes`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.cooldownMinutes') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.cooldownMinutes')) AS SIGNED) ELSE `cooldown_minutes` END,
 `visibility_scope`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.visibilityScope') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.visibilityScope')) ELSE `visibility_scope` END,
 `start_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.startAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `start_at` END,
 `end_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.endAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `end_at` END,
 `priority`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.priority') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.priority')) AS SIGNED) ELSE `priority` END,
 `enabled`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.enabled') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.enabled')) IN ('true','1'),1,0) ELSE `enabled` END,config_version=config_version+1,deleted_at=IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.deleted'))='true',NOW(3),deleted_at);

ALTER TABLE `game_popup_campaigns` MODIFY `draft_payload` JSON NULL;

ALTER TABLE `game_popup_campaigns` MODIFY `published_payload` JSON NULL;

ALTER TABLE `game_site_banners` ADD COLUMN configuration_json JSON NULL;

UPDATE `game_site_banners` SET configuration_json=CASE
 WHEN publish_status='DRAFT' AND JSON_VALID(draft_payload) AND JSON_LENGTH(draft_payload)>0 THEN CAST(draft_payload AS JSON)
 WHEN JSON_VALID(published_payload) AND JSON_LENGTH(published_payload)>0 THEN CAST(published_payload AS JSON)
 ELSE JSON_OBJECT('name',`name`,'title',`title`,'subtitle',`subtitle`,'desktopAssetID',`desktop_asset_id`,'mobileAssetID',`mobile_asset_id`,'actionType',`action_type`,'actionTarget',`action_target`,'sortOrder',`sort_order`,'languageCode',`language_code`,'countryCode',`country_code`,'currencyCode',`currency_code`,'visibilityScope',`visibility_scope`,'startAt',`start_at`,'endAt',`end_at`,'enabled',IF(`enabled`=1,CAST('true' AS JSON),CAST('false' AS JSON))) END;

UPDATE `game_site_banners` SET `name`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.name') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.name')) ELSE `name` END,
 `title`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.title') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.title')) ELSE `title` END,
 `subtitle`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.subtitle') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.subtitle')) ELSE `subtitle` END,
 `desktop_asset_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.desktopAssetID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.desktopAssetID')) AS SIGNED) ELSE `desktop_asset_id` END,
 `mobile_asset_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.mobileAssetID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.mobileAssetID')) AS SIGNED) ELSE `mobile_asset_id` END,
 `action_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionType')) ELSE `action_type` END,
 `action_target`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionTarget') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionTarget')) ELSE `action_target` END,
 `sort_order`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.sortOrder') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.sortOrder')) AS SIGNED) ELSE `sort_order` END,
 `language_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.languageCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.languageCode')) ELSE `language_code` END,
 `country_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.countryCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.countryCode')) ELSE `country_code` END,
 `currency_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.currencyCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.currencyCode')) ELSE `currency_code` END,
 `visibility_scope`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.visibilityScope') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.visibilityScope')) ELSE `visibility_scope` END,
 `start_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.startAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `start_at` END,
 `end_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.endAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `end_at` END,
 `enabled`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.enabled') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.enabled')) IN ('true','1'),1,0) ELSE `enabled` END,config_version=config_version+1,deleted_at=IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.deleted'))='true',NOW(3),deleted_at);

ALTER TABLE `game_h5_navigation_items` ADD COLUMN configuration_json JSON NULL;

UPDATE `game_h5_navigation_items` SET configuration_json=CASE
 WHEN publish_status='DRAFT' AND JSON_VALID(draft_payload) AND JSON_LENGTH(draft_payload)>0 THEN CAST(draft_payload AS JSON)
 WHEN JSON_VALID(published_payload) AND JSON_LENGTH(published_payload)>0 THEN CAST(published_payload AS JSON)
 ELSE JSON_OBJECT('menuCode',`menu_code`,'labels',IF(JSON_VALID(`labels_json`),CAST(`labels_json` AS JSON),JSON_OBJECT()),'iconAssetID',`icon_asset_id`,'lottieAssetID',`lottie_asset_id`,'soundAssetID',`sound_asset_id`,'displayPosition',`display_position`,'actionType',`action_type`,'actionTarget',`action_target`,'providerCode',`provider_code`,'gameID',`game_id`,'providerGameCode',`provider_game_code`,'gameName',`game_name`,'sortOrder',`sort_order`,'badgeType',`badge_type`,'visibilityScope',`visibility_scope`,'countryCodes',IF(JSON_VALID(`country_codes`),CAST(`country_codes` AS JSON),JSON_ARRAY()),'languageCodes',IF(JSON_VALID(`language_codes`),CAST(`language_codes` AS JSON),JSON_ARRAY()),'startAt',`start_at`,'endAt',`end_at`,'enabled',IF(`enabled`=1,CAST('true' AS JSON),CAST('false' AS JSON))) END;

UPDATE `game_h5_navigation_items` SET `menu_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.menuCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.menuCode')) ELSE `menu_code` END,
 `labels_json`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.labels') THEN JSON_EXTRACT(configuration_json,'$.labels') ELSE `labels_json` END,
 `icon_asset_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.iconAssetID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.iconAssetID')) AS SIGNED) ELSE `icon_asset_id` END,
 `lottie_asset_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.lottieAssetID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.lottieAssetID')) AS SIGNED) ELSE `lottie_asset_id` END,
 `sound_asset_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.soundAssetID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.soundAssetID')) AS SIGNED) ELSE `sound_asset_id` END,
 `display_position`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.displayPosition') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.displayPosition')) ELSE `display_position` END,
 `action_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionType')) ELSE `action_type` END,
 `action_target`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.actionTarget') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.actionTarget')) ELSE `action_target` END,
 `provider_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.providerCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.providerCode')) ELSE `provider_code` END,
 `game_id`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.gameID') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.gameID')) AS SIGNED) ELSE `game_id` END,
 `provider_game_code`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.providerGameCode') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.providerGameCode')) ELSE `provider_game_code` END,
 `game_name`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.gameName') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.gameName')) ELSE `game_name` END,
 `sort_order`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.sortOrder') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.sortOrder')) AS SIGNED) ELSE `sort_order` END,
 `badge_type`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.badgeType') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.badgeType')) ELSE `badge_type` END,
 `visibility_scope`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.visibilityScope') THEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.visibilityScope')) ELSE `visibility_scope` END,
 `country_codes`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.countryCodes') THEN JSON_EXTRACT(configuration_json,'$.countryCodes') ELSE `country_codes` END,
 `language_codes`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.languageCodes') THEN JSON_EXTRACT(configuration_json,'$.languageCodes') ELSE `language_codes` END,
 `start_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.startAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.startAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `start_at` END,
 `end_at`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.endAt') THEN CASE WHEN JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt'))='null' THEN NULL ELSE STR_TO_DATE(REPLACE(LEFT(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.endAt')),19),'T',' '),'%Y-%m-%d %H:%i:%s') END ELSE `end_at` END,
 `enabled`=CASE WHEN JSON_CONTAINS_PATH(configuration_json,'one','$.enabled') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.enabled')) IN ('true','1'),1,0) ELSE `enabled` END,config_version=config_version+1,deleted_at=IF(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.deleted'))='true',NOW(3),deleted_at);

UPDATE game_site_config s SET site_name=COALESCE(NULLIF(site_name_draft,''),site_name),
 country_code=COALESCE(NULLIF(country_code_draft,''),country_code),
 currency_code=COALESCE(NULLIF(currency_code_draft,''),currency_code),
 default_language=COALESCE(NULLIF(default_language_draft,''),default_language),
 timezone=COALESCE(NULLIF(timezone_draft,''),timezone),
 brand_config=COALESCE(brand_config_draft,brand_config),
 theme_config=COALESCE(theme_config_draft,theme_config),
 feature_config=COALESCE(feature_config_draft,feature_config),
 maintenance_config=COALESCE(maintenance_config_draft,maintenance_config),
 seo_config=COALESCE(seo_config_draft,seo_config),config_version=config_version+1 WHERE config_status='DRAFT' AND (COALESCE(NULLIF(currency_code_draft,''),currency_code)=currency_code OR NOT EXISTS(SELECT 1 FROM game_user_wallets w WHERE w.site_id=s.id));

UPDATE game_games SET platform_enabled=COALESCE(platform_enabled_draft,platform_enabled),is_hot=COALESCE(is_hot_draft,is_hot),is_featured=COALESCE(is_featured_draft,is_featured),is_new=COALESCE(is_new_draft,is_new),sort=COALESCE(sort_draft,sort),maintenance_message=COALESCE(maintenance_message_draft,maintenance_message),status=COALESCE(status_draft,status),config_version=config_version+1 WHERE config_status='DRAFT';

UPDATE game_games SET status=3 WHERE provider_enabled=0 OR platform_enabled=0;

DELETE r FROM game_game_category_relations r JOIN game_games g ON g.id=r.game_id AND g.site_id=r.site_id WHERE g.config_status='DRAFT' AND g.category_ids_draft IS NOT NULL;

INSERT INTO game_game_category_relations(site_id,game_id,category_id,sort,created_at,updated_at)
 SELECT g.site_id,g.id,j.category_id,j.sort_order,NOW(3),NOW(3) FROM game_games g
 JOIN JSON_TABLE(COALESCE(g.category_ids_draft,JSON_ARRAY()),'$[*]' COLUMNS(category_id BIGINT PATH '$.categoryId',sort_order INT PATH '$.sort')) j
 JOIN game_game_categories c ON c.id=j.category_id AND c.site_id=g.site_id AND c.deleted_at IS NULL
 WHERE g.config_status='DRAFT' AND g.category_ids_draft IS NOT NULL;

UPDATE game_lobby_sections SET section_code=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.sectionCode') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.sectionCode')) ELSE section_code END,section_type=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.sectionType') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.sectionType')) ELSE section_type END,title_zh=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.titleZh') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.titleZh')) ELSE title_zh END,title_en=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.titleEn') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.titleEn')) ELSE title_en END,display_style=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.displayStyle') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.displayStyle')) ELSE display_style END,source_type=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.sourceType') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.sourceType')) ELSE source_type END,source_ref=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.sourceRef') THEN JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.sourceRef')) ELSE source_ref END,max_items=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.maxItems') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.maxItems')) AS SIGNED) ELSE max_items END,sort_order=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.sortOrder') THEN CAST(JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.sortOrder')) AS SIGNED) ELSE sort_order END,enabled=CASE WHEN JSON_CONTAINS_PATH(draft_payload,'one','$.enabled') THEN IF(JSON_UNQUOTE(JSON_EXTRACT(draft_payload,'$.enabled')) IN('true','1'),1,0) ELSE enabled END,config_json=COALESCE(JSON_EXTRACT(draft_payload,'$.config'),config_json),config_version=config_version+1 WHERE publish_status='DRAFT' AND draft_payload IS NOT NULL;

UPDATE game_lobby_sections SET supported_countries='',supported_currencies='';

UPDATE game_h5_navigation_items SET country_codes='[]',language_codes='[]',configuration_json=JSON_SET(configuration_json,'$.countryCodes',JSON_ARRAY(),'$.languageCodes',JSON_ARRAY());

UPDATE game_country_configs SET status=0,is_default=0;

UPDATE game_country_configs c JOIN game_site_config s ON s.id=1 AND c.country_code=s.country_code AND c.deleted_at IS NULL SET c.status=1,c.is_default=1;

ALTER TABLE game_country_configs ADD COLUMN active_slot TINYINT GENERATED ALWAYS AS (CASE WHEN status=1 AND deleted_at IS NULL THEN 1 ELSE NULL END) STORED, ADD UNIQUE KEY uk_single_active_country(active_slot);

DELETE rm FROM game_admin_role_menus rm JOIN game_admin_menus m ON m.id=rm.menu_id WHERE m.code IN('release-center','brand-preview','brand-publish','music');

DELETE FROM game_admin_menus WHERE code IN('release-center','brand-preview','brand-publish','music');

UPDATE game_notifications SET status='READY' WHERE status='DRAFT';
ALTER TABLE game_notifications ALTER COLUMN status SET DEFAULT 'READY';
UPDATE game_agent_settlements SET status='READY' WHERE status='DRAFT';
ALTER TABLE game_agent_settlements ALTER COLUMN status SET DEFAULT 'READY';
UPDATE game_activity_events SET status_code='PAUSED',status=2 WHERE status_code='DRAFT';
ALTER TABLE game_activity_events ALTER COLUMN status_code SET DEFAULT 'PAUSED';
UPDATE game_cms_pages SET status=3 WHERE status=2;
