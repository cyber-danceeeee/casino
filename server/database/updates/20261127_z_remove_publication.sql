-- Phase 2: remove obsolete configuration draft/publication storage.

ALTER TABLE `game_site_announcements` DROP INDEX `idx_site_announcement_scope`,
 DROP COLUMN `draft_payload`,
 DROP COLUMN `published_payload`,
 DROP COLUMN `publish_status`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `published_at`,
 DROP COLUMN `published_by`;

ALTER TABLE `game_popup_campaigns` DROP INDEX `idx_popup_campaign_scope_status`,
 DROP COLUMN `popup_component_id_draft`,
 DROP COLUMN `draft_payload`,
 DROP COLUMN `published_payload`,
 DROP COLUMN `publish_status`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `published_at`,
 DROP COLUMN `published_by`;

ALTER TABLE `game_site_banners` DROP INDEX `idx_site_banner_scope_sort`,
 DROP INDEX `idx_site_banner_desktop_asset_draft`,
 DROP INDEX `idx_site_banner_mobile_asset_draft`,
 DROP COLUMN `desktop_asset_id_draft`,
 DROP COLUMN `mobile_asset_id_draft`,
 DROP COLUMN `draft_payload`,
 DROP COLUMN `published_payload`,
 DROP COLUMN `publish_status`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `published_at`,
 DROP COLUMN `published_by`;

ALTER TABLE `game_h5_navigation_items` DROP INDEX `idx_h5_navigation_publish`,
 DROP COLUMN `icon_asset_id_draft`,
 DROP COLUMN `lottie_asset_id_draft`,
 DROP COLUMN `sound_asset_id_draft`,
 DROP COLUMN `draft_payload`,
 DROP COLUMN `published_payload`,
 DROP COLUMN `publish_status`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `published_at`,
 DROP COLUMN `published_by`;

ALTER TABLE `game_site_config` DROP INDEX `idx_clean_21963770ee41b0cf`,
 DROP COLUMN `site_name_draft`,
 DROP COLUMN `country_code_draft`,
 DROP COLUMN `currency_code_draft`,
 DROP COLUMN `default_language_draft`,
 DROP COLUMN `timezone_draft`,
 DROP COLUMN `brand_config_draft`,
 DROP COLUMN `theme_config_draft`,
 DROP COLUMN `feature_config_draft`,
 DROP COLUMN `maintenance_config_draft`,
 DROP COLUMN `seo_config_draft`,
 DROP COLUMN `config_status`,
 DROP COLUMN `config_published_at`,
 DROP COLUMN `config_published_by`,
 DROP COLUMN `config_draft_updated_at`,
 DROP COLUMN `config_draft_updated_by`;

ALTER TABLE `game_games` DROP INDEX `idx_game_config_publish`,
 DROP COLUMN `category_ids_draft`,
 DROP COLUMN `config_status`,
 DROP COLUMN `config_published_at`,
 DROP COLUMN `config_published_by`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `draft_updated_by`,
 DROP COLUMN `maintenance_message_draft`,
 DROP COLUMN `platform_enabled_draft`,
 DROP COLUMN `is_hot_draft`,
 DROP COLUMN `is_featured_draft`,
 DROP COLUMN `is_new_draft`,
 DROP COLUMN `status_draft`,
 DROP COLUMN `sort_draft`;

ALTER TABLE `game_lobby_sections` DROP INDEX `idx_game_lobby_publish`,
 DROP COLUMN `draft_payload`,
 DROP COLUMN `publish_status`,
 DROP COLUMN `published_at`,
 DROP COLUMN `published_by`,
 DROP COLUMN `draft_updated_at`,
 DROP COLUMN `draft_updated_by`;

DROP TABLE game_config_releases;

CREATE INDEX idx_site_announcement_live ON game_site_announcements(site_id,enabled,is_pinned,priority,sort_order);
CREATE INDEX idx_popup_campaign_live ON game_popup_campaigns(site_id,enabled,deleted_at);
CREATE INDEX idx_site_banner_live ON game_site_banners(site_id,enabled,sort_order,id);
CREATE INDEX idx_game_config_live ON game_games(site_id,platform_enabled,sort);
DELETE rp FROM game_admin_role_permissions rp JOIN game_admin_permissions p ON p.id=rp.permission_id WHERE p.code LIKE 'release:%' OR p.code IN('system:site:config:preview','system:site:config:publish');
DELETE FROM game_admin_permissions WHERE code LIKE 'release:%' OR code IN('system:site:config:preview','system:site:config:publish');
