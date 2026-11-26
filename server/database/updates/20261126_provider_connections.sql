CREATE TABLE game_provider_connections (
 provider_code VARCHAR(64) NOT NULL PRIMARY KEY,
 adapter_type VARCHAR(32) NOT NULL,
 name VARCHAR(128) NOT NULL,
 api_base_url VARCHAR(500) NOT NULL,
 environment VARCHAR(24) NOT NULL DEFAULT 'sandbox',
 enabled BOOLEAN NOT NULL DEFAULT FALSE,
 supported_countries VARCHAR(500) NOT NULL DEFAULT '',
 supported_currencies VARCHAR(500) NOT NULL DEFAULT '',
 credentials_ciphertext LONGTEXT NOT NULL,
 callback_ip_allowlist TEXT NOT NULL,
 version BIGINT UNSIGNED NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- Fresh baselines create most menus during application initialization. Ensure
-- this page also exists before startup and preserve existing menu customization.
INSERT IGNORE INTO game_admin_menus(parent_id,menu_type,code,name_zh,name_en,route_path,icon,permission_code,sort_order,visible,enabled,system_protected)
 VALUES(NULL,'DIRECTORY','games','游戏中心','Game Center','','Platform','',40,1,1,1);
INSERT IGNORE INTO game_admin_menus(parent_id,menu_type,code,name_zh,name_en,route_path,icon,permission_code,sort_order,visible,enabled,system_protected)
 VALUES((SELECT id FROM (SELECT id FROM game_admin_menus WHERE code='games' AND deleted_at IS NULL LIMIT 1) parent_menu),'MENU','provider-connections','供应商配置','Supplier Connections','/games/providers','Connection','game:provider:list',5,1,1,1);
INSERT IGNORE INTO game_admin_role_menus(role_id,menu_id)
 SELECT r.id,m.id FROM game_admin_roles r CROSS JOIN game_admin_menus m
 WHERE r.code='super-admin' AND r.deleted_at IS NULL AND m.code='provider-connections' AND m.deleted_at IS NULL;
INSERT IGNORE INTO game_admin_role_menus(role_id,menu_id)
 SELECT DISTINCT rm.role_id,m.id FROM game_admin_role_menus rm
 JOIN game_admin_menus existing_menu ON existing_menu.id=rm.menu_id AND (existing_menu.code IN ('game-sync','provider-management') OR existing_menu.route_path='/games/sync')
 CROSS JOIN game_admin_menus m WHERE m.code='provider-connections' AND rm.deleted_at IS NULL;
