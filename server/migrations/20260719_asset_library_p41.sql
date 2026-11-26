CREATE TABLE IF NOT EXISTS game_asset_library (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  asset_scope VARCHAR(16) NOT NULL DEFAULT 'TENANT',
  asset_type VARCHAR(16) NOT NULL,
  usage_type VARCHAR(32) NOT NULL DEFAULT 'GENERAL',
  display_name VARCHAR(160) NOT NULL DEFAULT '',
  original_name VARCHAR(255) NOT NULL DEFAULT '',
  object_key VARCHAR(500) NOT NULL,
  public_url VARCHAR(1000) NOT NULL DEFAULT '',
  mime_type VARCHAR(128) NOT NULL DEFAULT '',
  file_ext VARCHAR(16) NOT NULL DEFAULT '',
  file_size BIGINT UNSIGNED NOT NULL DEFAULT 0,
  checksum_sha256 CHAR(64) NOT NULL DEFAULT '',
  width INT UNSIGNED NOT NULL DEFAULT 0,
  height INT UNSIGNED NOT NULL DEFAULT 0,
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reference_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  is_public TINYINT(1) NOT NULL DEFAULT 0,
  tags VARCHAR(500) NOT NULL DEFAULT '',
  alt_text VARCHAR(255) NOT NULL DEFAULT '',
  metadata_json JSON NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_by_name VARCHAR(64) NOT NULL DEFAULT '',
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_asset_scope_status (tenant_id, site_id, asset_scope, status, created_at),
  KEY idx_asset_type_usage (asset_type, usage_type, status),
  KEY idx_asset_checksum (checksum_sha256),
  KEY idx_asset_object_key (object_key(191)),
  KEY idx_asset_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_permissions
(code, name, name_zh, name_en, module, permission_type, enabled, system_protected, created_at, updated_at)
VALUES
('operation:asset:list', 'List Assets', '素材列表', 'List Assets', 'operation', 'API', 1, 1, NOW(3), NOW(3)),
('operation:asset:upload', 'Upload Operation Asset', '上传运营资源', 'Upload Operation Asset', 'operation', 'API', 1, 1, NOW(3), NOW(3)),
('operation:asset:update', 'Update Asset', '更新素材', 'Update Asset', 'operation', 'API', 1, 1, NOW(3), NOW(3)),
('operation:asset:delete', 'Delete Asset', '删除素材', 'Delete Asset', 'operation', 'API', 1, 1, NOW(3), NOW(3));

UPDATE game_admin_permissions
SET name_zh='上传运营资源', name_en='Upload Operation Asset', module='operation', permission_type='API', enabled=1, system_protected=1, updated_at=NOW(3)
WHERE code='operation:asset:upload' AND deleted_at IS NULL;

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
SELECT r.id, p.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('operation:asset:list','operation:asset:upload','operation:asset:update','operation:asset:delete')
WHERE r.deleted_at IS NULL
  AND (r.code='super-admin' OR r.code='tenant-owner');

UPDATE game_admin_role_permissions rp
JOIN game_admin_roles r ON r.id=rp.role_id
JOIN game_admin_permissions p ON p.id=rp.permission_id
SET rp.deleted_at=NULL, rp.updated_at=NOW(3)
WHERE r.deleted_at IS NULL
  AND (r.code='super-admin' OR r.code='tenant-owner')
  AND p.code IN ('operation:asset:list','operation:asset:upload','operation:asset:update','operation:asset:delete');

INSERT IGNORE INTO game_admin_menus
(parent_id, menu_type, code, name_zh, name_en, route_path, component_path, icon, permission_code, sort_order, visible, enabled, system_protected, created_at, updated_at)
VALUES
(NULL, 'DIRECTORY', 'operations', '运营中心', 'Operations Center', '', '', 'Message', '', 500, 1, 1, 1, NOW(3), NOW(3));

SET @operations_menu_id=(SELECT id FROM game_admin_menus WHERE code='operations' AND deleted_at IS NULL LIMIT 1);

INSERT IGNORE INTO game_admin_menus
(parent_id, menu_type, code, name_zh, name_en, route_path, component_path, icon, permission_code, sort_order, visible, enabled, system_protected, created_at, updated_at)
VALUES
(@operations_menu_id, 'MENU', 'asset-library', '素材中心', 'Asset Library', '/operations/assets', 'operations/AssetLibraryPage', 'FolderOpened', 'operation:asset:list', 5, 1, 1, 1, NOW(3), NOW(3));

UPDATE game_admin_menus
SET parent_id=@operations_menu_id,
    name_zh='素材中心',
    name_en='Asset Library',
    route_path='/operations/assets',
    component_path='operations/AssetLibraryPage',
    icon='FolderOpened',
    permission_code='operation:asset:list',
    visible=1,
    enabled=1,
    deleted_at=NULL,
    updated_at=NOW(3)
WHERE code='asset-library';

INSERT IGNORE INTO game_admin_role_menus (role_id, menu_id, created_at, updated_at, deleted_at)
SELECT r.id, m.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code IN ('operations','asset-library')
WHERE r.deleted_at IS NULL
  AND (r.code='super-admin' OR r.code='tenant-owner');

UPDATE game_admin_role_menus rm
JOIN game_admin_roles r ON r.id=rm.role_id
JOIN game_admin_menus m ON m.id=rm.menu_id
SET rm.deleted_at=NULL, rm.updated_at=NOW(3)
WHERE r.deleted_at IS NULL
  AND (r.code='super-admin' OR r.code='tenant-owner')
  AND m.code IN ('operations','asset-library');
