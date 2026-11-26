-- Built-in H5 popup components + tenant campaign configuration.

CREATE TABLE IF NOT EXISTS `game_popup_components` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `component_code` varchar(64) NOT NULL,
  `name_zh` varchar(128) NOT NULL DEFAULT '',
  `name_en` varchar(128) NOT NULL DEFAULT '',
  `description_zh` varchar(500) NOT NULL DEFAULT '',
  `description_en` varchar(500) NOT NULL DEFAULT '',
  `category` varchar(32) NOT NULL DEFAULT 'GENERAL',
  `preview_url` varchar(1000) NOT NULL DEFAULT '',
  `preview_video_url` varchar(1000) NOT NULL DEFAULT '',
  `h5_min_version` varchar(32) NOT NULL DEFAULT '1.0.0',
  `default_action_type` varchar(32) NOT NULL DEFAULT 'CLOSE_ONLY',
  `allowed_action_types` json NOT NULL,
  `status` tinyint NOT NULL DEFAULT 1,
  `sort_order` int NOT NULL DEFAULT 0,
  `created_by` bigint unsigned NOT NULL DEFAULT 0,
  `updated_by` bigint unsigned NOT NULL DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_popup_component_code` (`component_code`),
  KEY `idx_popup_component_status_sort` (`status`,`sort_order`,`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `game_tenant_popup_campaigns` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned NOT NULL,
  `site_id` bigint unsigned NOT NULL,
  `popup_component_id` bigint unsigned NOT NULL,
  `popup_component_id_draft` bigint unsigned NOT NULL DEFAULT 0,
  `campaign_name` varchar(128) NOT NULL,
  `action_type` varchar(32) NOT NULL DEFAULT 'CLOSE_ONLY',
  `action_target` varchar(1000) NOT NULL DEFAULT '',
  `trigger_type` varchar(32) NOT NULL DEFAULT 'HOME_ENTER',
  `trigger_value` varchar(255) NOT NULL DEFAULT '',
  `frequency_type` varchar(32) NOT NULL DEFAULT 'ONCE_PER_SESSION',
  `display_limit` int NOT NULL DEFAULT 0,
  `cooldown_minutes` int NOT NULL DEFAULT 0,
  `visibility_scope` varchar(16) NOT NULL DEFAULT 'ALL',
  `start_at` datetime(3) DEFAULT NULL,
  `end_at` datetime(3) DEFAULT NULL,
  `priority` int NOT NULL DEFAULT 0,
  `enabled` tinyint(1) NOT NULL DEFAULT 0,
  `draft_payload` json NOT NULL,
  `published_payload` json NOT NULL,
  `publish_status` varchar(24) NOT NULL DEFAULT 'DRAFT',
  `config_version` bigint unsigned NOT NULL DEFAULT 0,
  `created_by` bigint unsigned NOT NULL DEFAULT 0,
  `updated_by` bigint unsigned NOT NULL DEFAULT 0,
  `draft_updated_at` datetime(3) DEFAULT NULL,
  `published_at` datetime(3) DEFAULT NULL,
  `published_by` bigint unsigned NOT NULL DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_popup_campaign_scope_status` (`tenant_id`,`site_id`,`publish_status`,`enabled`,`deleted_at`),
  KEY `idx_popup_campaign_component` (`popup_component_id`,`deleted_at`),
  KEY `idx_popup_campaign_schedule` (`tenant_id`,`site_id`,`start_at`,`end_at`),
  KEY `idx_popup_campaign_priority` (`tenant_id`,`site_id`,`priority`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `game_popup_campaign_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned NOT NULL,
  `site_id` bigint unsigned NOT NULL,
  `campaign_id` bigint unsigned NOT NULL,
  `player_id` bigint unsigned NOT NULL DEFAULT 0,
  `session_id` varchar(128) NOT NULL DEFAULT '',
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `event_id` varchar(128) NOT NULL,
  `event_type` varchar(16) NOT NULL,
  `business_date` varchar(10) NOT NULL,
  `ip_address` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_popup_campaign_event` (`tenant_id`,`site_id`,`event_id`),
  KEY `idx_popup_campaign_event_viewer` (`tenant_id`,`site_id`,`campaign_id`,`player_id`,`business_date`),
  KEY `idx_popup_campaign_event_session` (`tenant_id`,`site_id`,`campaign_id`,`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO game_popup_components
  (component_code,name_zh,name_en,description_zh,description_en,category,h5_min_version,default_action_type,allowed_action_types,status,sort_order,created_at,updated_at)
VALUES
  ('DEPOSIT_POPUP','充值活动弹窗','Deposit Popup','金币与首充引导动画','Animated deposit and first-deposit prompt','DEPOSIT','1.0.0','OPEN_DEPOSIT',JSON_ARRAY('OPEN_DEPOSIT','OPEN_PROMOTION','CLOSE_ONLY'),1,10,NOW(3),NOW(3)),
  ('WELCOME_BONUS_POPUP','欢迎奖励弹窗','Welcome Bonus Popup','欢迎礼物与奖励动画','Welcome gift and reward animation','PROMOTION','1.0.0','OPEN_PROMOTION',JSON_ARRAY('OPEN_PROMOTION','OPEN_COUPON','CLOSE_ONLY'),1,20,NOW(3),NOW(3)),
  ('CASHBACK_POPUP','返水活动弹窗','Cashback Popup','返水金币与奖励动画','Cashback coin and reward animation','PROMOTION','1.0.0','OPEN_PROMOTION',JSON_ARRAY('OPEN_PROMOTION','OPEN_INTERNAL_PAGE','CLOSE_ONLY'),1,30,NOW(3),NOW(3)),
  ('VIP_UPGRADE_POPUP','VIP 升级弹窗','VIP Upgrade Popup','VIP 徽章升级动画','Animated VIP upgrade badge','VIP','1.0.0','OPEN_VIP',JSON_ARRAY('OPEN_VIP','OPEN_PROMOTION','CLOSE_ONLY'),1,40,NOW(3),NOW(3)),
  ('APP_DOWNLOAD_POPUP','App 下载弹窗','App Download Popup','App 下载与安装引导动画','Animated app install prompt','DOWNLOAD','1.0.0','OPEN_DOWNLOAD',JSON_ARRAY('OPEN_DOWNLOAD','OPEN_EXTERNAL_URL','CLOSE_ONLY'),1,50,NOW(3),NOW(3))
ON DUPLICATE KEY UPDATE
  name_zh=VALUES(name_zh),name_en=VALUES(name_en),description_zh=VALUES(description_zh),description_en=VALUES(description_en),
  category=VALUES(category),h5_min_version=VALUES(h5_min_version),default_action_type=VALUES(default_action_type),
  allowed_action_types=VALUES(allowed_action_types),status=VALUES(status),sort_order=VALUES(sort_order),deleted_at=NULL,updated_at=NOW(3);

INSERT IGNORE INTO game_admin_permissions
  (code,name,name_zh,name_en,module,permission_type,api_path,enabled,system_protected,created_at,updated_at,deleted_at)
VALUES
  ('system:popup-component:list','List popup components','查看弹窗组件','List Popup Components','system','API','/api/v1/admin/system/popup-components',1,1,NOW(3),NOW(3),NULL),
  ('system:popup-component:create','Create popup component','登记弹窗组件','Create Popup Component','system','API','/api/v1/admin/system/popup-components',1,1,NOW(3),NOW(3),NULL),
  ('system:popup-component:update','Update popup component','编辑弹窗组件','Update Popup Component','system','API','/api/v1/admin/system/popup-components/:id',1,1,NOW(3),NOW(3),NULL),
  ('system:popup-component:delete','Delete popup component','删除弹窗组件','Delete Popup Component','system','API','/api/v1/admin/system/popup-components/:id',1,1,NOW(3),NOW(3),NULL);

INSERT INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code LIKE 'system:popup-component:%' AND p.deleted_at IS NULL
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

INSERT IGNORE INTO game_admin_menus
  (parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected,created_at,updated_at)
VALUES
  (NULL,'MENU','popup-components','弹窗组件','Popup Components','/platform/resources/popup-components','platform/PopupComponentsPage','Files','system:popup-component:list',15,1,1,1,NOW(3),NOW(3));

UPDATE game_admin_menus
SET name_zh='弹窗投放',name_en='Popup Campaigns',component_path='operations/PopupManagementPage',
    permission_code='operations:popup:view',visible=1,enabled=1,deleted_at=NULL,updated_at=NOW(3)
WHERE code='popups';

INSERT INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at,deleted_at)
SELECT r.id,m.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code='popup-components' AND m.deleted_at IS NULL
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

UPDATE game_admin_permissions
SET enabled=0,updated_at=NOW(3)
WHERE code IN ('operation:popup:list','operation:popup:create','operation:popup:update','operation:popup:delete');

DELETE FROM game_tenant_operation_overrides WHERE resource_type='POPUP';
DROP TABLE IF EXISTS game_popup_impressions;
DROP TABLE IF EXISTS game_site_popups;
DROP TABLE IF EXISTS game_home_popup_dialogs;
