CREATE TABLE IF NOT EXISTS `game_site_announcements` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `site_id` BIGINT UNSIGNED NOT NULL,
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `summary` VARCHAR(500) NOT NULL DEFAULT '',
  `content` MEDIUMTEXT NULL,
  `announcement_type` VARCHAR(32) NOT NULL DEFAULT 'GENERAL',
  `priority` INT NOT NULL DEFAULT 0,
  `sort_order` INT NOT NULL DEFAULT 0,
  `is_pinned` TINYINT(1) NOT NULL DEFAULT 0,
  `language_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `country_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `currency_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `visibility_scope` VARCHAR(24) NOT NULL DEFAULT 'ALL',
  `start_at` DATETIME(3) NULL,
  `end_at` DATETIME(3) NULL,
  `enabled` TINYINT(1) NOT NULL DEFAULT 0,
  `draft_payload` JSON NULL,
  `published_payload` JSON NULL,
  `publish_status` VARCHAR(24) NOT NULL DEFAULT 'DRAFT',
  `config_version` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `draft_updated_at` DATETIME(3) NULL,
  `published_at` DATETIME(3) NULL,
  `published_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_site_announcement_scope` (`tenant_id`,`site_id`,`publish_status`,`enabled`,`is_pinned`,`priority`,`sort_order`),
  KEY `idx_site_announcement_schedule` (`tenant_id`,`site_id`,`start_at`,`end_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `game_site_popups` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `site_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(128) NOT NULL DEFAULT '',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `content` TEXT NULL,
  `desktop_asset_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mobile_asset_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `desktop_asset_id_draft` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mobile_asset_id_draft` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `button_text` VARCHAR(64) NOT NULL DEFAULT '',
  `action_type` VARCHAR(32) NOT NULL DEFAULT 'NONE',
  `action_target` VARCHAR(500) NOT NULL DEFAULT '',
  `display_frequency` VARCHAR(32) NOT NULL DEFAULT 'EVERY_VISIT',
  `display_limit` INT UNSIGNED NOT NULL DEFAULT 0,
  `cooldown_minutes` INT UNSIGNED NOT NULL DEFAULT 0,
  `language_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `country_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `currency_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `visibility_scope` VARCHAR(24) NOT NULL DEFAULT 'ALL',
  `start_at` DATETIME(3) NULL,
  `end_at` DATETIME(3) NULL,
  `sort_order` INT NOT NULL DEFAULT 0,
  `enabled` TINYINT(1) NOT NULL DEFAULT 0,
  `draft_payload` JSON NULL,
  `published_payload` JSON NULL,
  `publish_status` VARCHAR(24) NOT NULL DEFAULT 'DRAFT',
  `config_version` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `draft_updated_at` DATETIME(3) NULL,
  `published_at` DATETIME(3) NULL,
  `published_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_site_popup_scope` (`tenant_id`,`site_id`,`publish_status`,`enabled`,`sort_order`,`id`),
  KEY `idx_site_popup_schedule` (`tenant_id`,`site_id`,`start_at`,`end_at`),
  KEY `idx_site_popup_desktop_asset` (`desktop_asset_id`),
  KEY `idx_site_popup_mobile_asset` (`mobile_asset_id`),
  KEY `idx_site_popup_desktop_asset_draft` (`desktop_asset_id_draft`),
  KEY `idx_site_popup_mobile_asset_draft` (`mobile_asset_id_draft`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `game_popup_impressions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `site_id` BIGINT UNSIGNED NOT NULL,
  `popup_id` BIGINT UNSIGNED NOT NULL,
  `player_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `session_id` VARCHAR(128) NOT NULL DEFAULT '',
  `device_id` VARCHAR(128) NOT NULL DEFAULT '',
  `event_id` VARCHAR(128) NOT NULL,
  `event_type` VARCHAR(24) NOT NULL,
  `business_date` DATE NOT NULL,
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '',
  `user_agent` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_popup_event` (`tenant_id`,`site_id`,`event_id`),
  KEY `idx_popup_player_frequency` (`tenant_id`,`site_id`,`popup_id`,`player_id`,`event_type`,`business_date`),
  KEY `idx_popup_device_frequency` (`tenant_id`,`site_id`,`popup_id`,`device_id`,`event_type`,`business_date`),
  KEY `idx_popup_session_frequency` (`tenant_id`,`site_id`,`popup_id`,`session_id`,`event_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `game_admin_permissions`
(`code`,`name`,`name_zh`,`name_en`,`module`,`permission_type`,`enabled`,`system_protected`,`created_at`,`updated_at`)
VALUES
('operations:announcement:view','View site announcements','查看站点公告','View site announcements','operations','API',1,1,NOW(3),NOW(3)),
('operations:announcement:edit','Edit site announcements','编辑站点公告','Edit site announcements','operations','API',1,1,NOW(3),NOW(3)),
('operations:announcement:delete','Delete site announcements','删除站点公告','Delete site announcements','operations','API',1,1,NOW(3),NOW(3)),
('operations:popup:view','View site popups','查看站点弹窗','View site popups','operations','API',1,1,NOW(3),NOW(3)),
('operations:popup:edit','Edit site popups','编辑站点弹窗','Edit site popups','operations','API',1,1,NOW(3),NOW(3)),
('operations:popup:delete','Delete site popups','删除站点弹窗','Delete site popups','operations','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO `game_admin_role_permissions` (`role_id`,`permission_id`,`created_at`,`updated_at`,`deleted_at`)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'operations:announcement:view','operations:announcement:edit','operations:announcement:delete',
  'operations:popup:view','operations:popup:edit','operations:popup:delete'
)
WHERE r.code IN ('super-admin','tenant-owner') AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

INSERT IGNORE INTO `game_admin_role_menus` (`role_id`,`menu_id`,`created_at`,`updated_at`,`deleted_at`)
SELECT r.id,m.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code IN ('operations','announcements','popups')
WHERE r.code IN ('super-admin','tenant-owner') AND r.deleted_at IS NULL AND m.deleted_at IS NULL;
