CREATE TABLE IF NOT EXISTS game_tenant_maintenance (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  scope_type VARCHAR(16) NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 0,
  start_at DATETIME(3) NULL,
  end_at DATETIME(3) NULL,
  reason VARCHAR(500) NOT NULL DEFAULT '',
  title VARCHAR(160) NOT NULL DEFAULT '',
  message_zh VARCHAR(1000) NOT NULL DEFAULT '',
  message_en VARCHAR(1000) NOT NULL DEFAULT '',
  estimated_recovery_at DATETIME(3) NULL,
  support_url VARCHAR(500) NOT NULL DEFAULT '',
  support_email VARCHAR(190) NOT NULL DEFAULT '',
  support_chat VARCHAR(500) NOT NULL DEFAULT '',
  allowlisted_ips TEXT NULL,
  allow_platform_bypass TINYINT(1) NOT NULL DEFAULT 1,
  allow_tenant_bypass TINYINT(1) NOT NULL DEFAULT 1,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_maintenance_scope (tenant_id,site_id,scope_type),
  KEY idx_tenant_maintenance_active (tenant_id,site_id,enabled,start_at,end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_provider_maintenance (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  scope_type VARCHAR(16) NOT NULL DEFAULT 'GLOBAL',
  provider_code VARCHAR(64) NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 0,
  start_at DATETIME(3) NULL,
  end_at DATETIME(3) NULL,
  reason VARCHAR(500) NOT NULL DEFAULT '',
  message_zh VARCHAR(1000) NOT NULL DEFAULT '',
  message_en VARCHAR(1000) NOT NULL DEFAULT '',
  affected_countries VARCHAR(500) NOT NULL DEFAULT '',
  affected_currencies VARCHAR(500) NOT NULL DEFAULT '',
  allow_existing_sessions TINYINT(1) NOT NULL DEFAULT 1,
  hide_from_lobby TINYINT(1) NOT NULL DEFAULT 1,
  block_new_launches TINYINT(1) NOT NULL DEFAULT 1,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_provider_maintenance_scope (tenant_id,site_id,provider_code),
  KEY idx_provider_maintenance_active (provider_code,enabled,start_at,end_at),
  KEY idx_provider_maintenance_tenant (tenant_id,site_id,enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_game_maintenance (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL,
  game_id BIGINT UNSIGNED NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 0,
  start_at DATETIME(3) NULL,
  end_at DATETIME(3) NULL,
  reason VARCHAR(500) NOT NULL DEFAULT '',
  message_zh VARCHAR(1000) NOT NULL DEFAULT '',
  message_en VARCHAR(1000) NOT NULL DEFAULT '',
  affected_countries VARCHAR(500) NOT NULL DEFAULT '',
  affected_currencies VARCHAR(500) NOT NULL DEFAULT '',
  allow_existing_sessions TINYINT(1) NOT NULL DEFAULT 1,
  hide_from_lobby TINYINT(1) NOT NULL DEFAULT 1,
  block_new_launches TINYINT(1) NOT NULL DEFAULT 1,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_maintenance_scope (tenant_id,site_id,game_id),
  KEY idx_game_maintenance_active (tenant_id,site_id,enabled,start_at,end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_system_health_checks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  component_code VARCHAR(64) NOT NULL,
  component_type VARCHAR(32) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'UNKNOWN',
  latency_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  error_summary VARCHAR(500) NOT NULL DEFAULT '',
  consecutive_failures INT UNSIGNED NOT NULL DEFAULT 0,
  last_checked_at DATETIME(3) NOT NULL,
  last_recovered_at DATETIME(3) NULL,
  details JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_health_component (component_code),
  KEY idx_system_health_status (status,last_checked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_system_alerts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  fingerprint CHAR(64) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  severity VARCHAR(16) NOT NULL,
  alert_type VARCHAR(64) NOT NULL,
  source VARCHAR(128) NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'OPEN',
  occurrence_count BIGINT UNSIGNED NOT NULL DEFAULT 1,
  summary VARCHAR(500) NOT NULL,
  details JSON NULL,
  first_triggered_at DATETIME(3) NOT NULL,
  last_triggered_at DATETIME(3) NOT NULL,
  acknowledged_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  acknowledged_at DATETIME(3) NULL,
  resolved_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  resolved_at DATETIME(3) NULL,
  internal_note VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_alert_fingerprint (fingerprint),
  KEY idx_system_alert_status (status,severity,last_triggered_at),
  KEY idx_system_alert_scope (tenant_id,site_id,status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_system_retry_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  source_type VARCHAR(48) NOT NULL,
  source_id BIGINT UNSIGNED NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(191) NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  attempt_count INT UNSIGNED NOT NULL DEFAULT 0,
  max_attempts INT UNSIGNED NOT NULL DEFAULT 5,
  next_attempt_at DATETIME(3) NULL,
  last_error VARCHAR(500) NOT NULL DEFAULT '',
  locked_by VARCHAR(128) NOT NULL DEFAULT '',
  locked_at DATETIME(3) NULL,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_retry_idempotency (idempotency_key),
  KEY idx_system_retry_queue (status,next_attempt_at,created_at),
  KEY idx_system_retry_scope (tenant_id,site_id,status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_support_cases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  case_no VARCHAR(40) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  player_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  category VARCHAR(24) NOT NULL,
  priority VARCHAR(16) NOT NULL DEFAULT 'NORMAL',
  status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
  title VARCHAR(190) NOT NULL,
  description TEXT NULL,
  assigned_admin_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_by BIGINT UNSIGNED NOT NULL,
  resolution TEXT NULL,
  related_order_ids TEXT NULL,
  related_transaction_ids TEXT NULL,
  idempotency_key VARCHAR(191) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  closed_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_support_case_no (case_no),
  UNIQUE KEY uk_support_case_idempotency (created_by,idempotency_key),
  KEY idx_support_case_scope (tenant_id,site_id,status,updated_at),
  KEY idx_support_case_player (tenant_id,site_id,player_id,created_at),
  KEY idx_support_case_assignee (assigned_admin_id,status,updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_support_case_notes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  case_id BIGINT UNSIGNED NOT NULL,
  author_id BIGINT UNSIGNED NOT NULL,
  author_username VARCHAR(128) NOT NULL DEFAULT '',
  note TEXT NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_support_case_note_scope (tenant_id,site_id,case_id,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_operational_notes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  target_type VARCHAR(48) NOT NULL,
  target_id VARCHAR(128) NOT NULL,
  note TEXT NOT NULL,
  author_id BIGINT UNSIGNED NOT NULL,
  author_username VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_operational_note_target (tenant_id,site_id,target_type,target_id,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_config_releases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL,
  environment VARCHAR(24) NOT NULL DEFAULT 'PRODUCTION',
  version BIGINT UNSIGNED NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'DRAFT',
  modules VARCHAR(1000) NOT NULL DEFAULT '',
  snapshot JSON NULL,
  diff_snapshot JSON NULL,
  validation_result JSON NULL,
  rollback_source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(191) NOT NULL,
  failure_reason VARCHAR(500) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL,
  published_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  published_at DATETIME(3) NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_config_release_version (tenant_id,site_id,environment,version),
  UNIQUE KEY uk_config_release_idempotency (idempotency_key),
  KEY idx_config_release_scope (tenant_id,site_id,status,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_backup_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  backup_no VARCHAR(48) NOT NULL,
  scope_type VARCHAR(48) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  storage_location VARCHAR(1000) NOT NULL DEFAULT '',
  artifact_key VARCHAR(500) NOT NULL DEFAULT '',
  size_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0,
  checksum_sha256 CHAR(64) NOT NULL DEFAULT '',
  encrypted TINYINT(1) NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(191) NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  started_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL,
  expires_at DATETIME(3) NULL,
  failure_reason VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_backup_task_no (backup_no),
  UNIQUE KEY uk_backup_idempotency (created_by,idempotency_key),
  KEY idx_backup_status (status,created_at),
  KEY idx_backup_scope (tenant_id,site_id,status,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_support_export_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_no VARCHAR(48) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  export_type VARCHAR(32) NOT NULL,
  filters JSON NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  artifact_key VARCHAR(500) NOT NULL DEFAULT '',
  row_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(191) NOT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  failure_reason VARCHAR(500) NOT NULL DEFAULT '',
  completed_at DATETIME(3) NULL,
  expires_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_support_export_no (task_no),
  UNIQUE KEY uk_support_export_idempotency (created_by,idempotency_key),
  KEY idx_support_export_scope (tenant_id,site_id,status,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('support:player:read','Read player support data','玩家支持查询','Read Player Support Data','support','API',1,1,NOW(3),NOW(3)),
('support:timeline:read','Read player timeline','玩家时间线','Read Player Timeline','support','API',1,1,NOW(3),NOW(3)),
('support:case:read','Read support cases','查看支持工单','Read Support Cases','support','API',1,1,NOW(3),NOW(3)),
('support:case:write','Manage support cases','管理支持工单','Manage Support Cases','support','API',1,1,NOW(3),NOW(3)),
('support:export','Export support records','导出支持记录','Export Support Records','support','API',1,1,NOW(3),NOW(3)),
('monitor:payment:read','Read payment monitor','支付监控','Read Payment Monitor','monitor','API',1,1,NOW(3),NOW(3)),
('monitor:payment:review','Mark payment for review','标记支付人工复核','Mark Payment for Review','monitor','API',1,1,NOW(3),NOW(3)),
('monitor:provider:read','Read provider monitor','Provider监控','Read Provider Monitor','monitor','API',1,1,NOW(3),NOW(3)),
('operations:note:write','Write operational notes','添加内部运营备注','Write Operational Notes','operations','API',1,1,NOW(3),NOW(3)),
('maintenance:tenant:read','Read tenant maintenance','查看租户维护','Read Tenant Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('maintenance:tenant:write','Manage tenant maintenance','管理租户维护','Manage Tenant Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('maintenance:provider:read','Read provider maintenance','查看Provider维护','Read Provider Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('maintenance:provider:write','Manage provider maintenance','管理Provider维护','Manage Provider Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('maintenance:game:read','Read game maintenance','查看游戏维护','Read Game Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('maintenance:game:write','Manage game maintenance','管理游戏维护','Manage Game Maintenance','maintenance','API',1,1,NOW(3),NOW(3)),
('health:system:read','Read system health','系统健康','Read System Health','health','API',1,1,NOW(3),NOW(3)),
('jobs:read','Read jobs and retries','查看任务与重试','Read Jobs and Retries','jobs','API',1,1,NOW(3),NOW(3)),
('jobs:retry','Retry failed work','重试失败任务','Retry Failed Work','jobs','API',1,1,NOW(3),NOW(3)),
('jobs:execute','Execute supported jobs','手动执行任务','Execute Supported Jobs','jobs','API',1,1,NOW(3),NOW(3)),
('alerts:read','Read operational alerts','查看告警','Read Operational Alerts','alerts','API',1,1,NOW(3),NOW(3)),
('alerts:acknowledge','Acknowledge alerts','确认告警','Acknowledge Alerts','alerts','API',1,1,NOW(3),NOW(3)),
('alerts:resolve','Resolve alerts','解决告警','Resolve Alerts','alerts','API',1,1,NOW(3),NOW(3)),
('backup:read','Read backup tasks','查看备份','Read Backup Tasks','backup','API',1,1,NOW(3),NOW(3)),
('backup:create','Create backup tasks','创建备份','Create Backup Tasks','backup','API',1,1,NOW(3),NOW(3)),
('backup:restore','Restore verified backup','恢复备份','Restore Verified Backup','backup','API',1,1,NOW(3),NOW(3)),
('backup:delete','Delete backup metadata','删除备份','Delete Backup Metadata','backup','API',1,1,NOW(3),NOW(3)),
('release:read','Read configuration releases','查看发布记录','Read Configuration Releases','release','API',1,1,NOW(3),NOW(3)),
('release:publish','Publish configuration releases','发布配置','Publish Configuration Releases','release','API',1,1,NOW(3),NOW(3)),
('release:rollback','Rollback configuration releases','回滚配置','Rollback Configuration Releases','release','API',1,1,NOW(3),NOW(3));

INSERT INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.module IN ('support','monitor','operations','maintenance','health','jobs','alerts','backup','release') AND p.deleted_at IS NULL
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

INSERT IGNORE INTO game_admin_menus
(parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected,created_at,updated_at)
VALUES
(NULL,'DIRECTORY','support-center','支持中心','Support Center','','','Headset','',850,1,1,1,NOW(3),NOW(3)),
(NULL,'DIRECTORY','ops-stability','运维中心','Operations','','','Monitor','',875,1,1,1,NOW(3),NOW(3));

SET @support_menu_id=(SELECT id FROM game_admin_menus WHERE code='support-center' AND deleted_at IS NULL LIMIT 1);
SET @ops_stability_menu_id=(SELECT id FROM game_admin_menus WHERE code='ops-stability' AND deleted_at IS NULL LIMIT 1);

INSERT IGNORE INTO game_admin_menus
(parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected,created_at,updated_at)
VALUES
(@support_menu_id,'MENU','support-player-search','玩家查询','Player Search','/support/players','','Search','support:player:read',10,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-player-timeline','玩家时间线','Player Timeline','/support/timeline','','Clock','support:timeline:read',20,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-order-search','订单查询','Order Search','/support/orders','','Tickets','support:player:read',30,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-payment-diagnostics','支付诊断','Payment Diagnostics','/support/payment-diagnostics','','CreditCard','monitor:payment:read',40,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-game-diagnostics','游戏诊断','Game Diagnostics','/support/game-diagnostics','','Monitor','monitor:provider:read',50,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-cases','支持工单','Support Cases','/support/cases','','ChatLineSquare','support:case:read',60,1,1,1,NOW(3),NOW(3)),
(@support_menu_id,'MENU','support-exports','导出记录','Export Records','/support/exports','','Download','support:export',70,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','payment-monitor','支付监控','Payment Monitor','/ops/payments','','CreditCard','monitor:payment:read',10,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','provider-monitor','Provider监控','Provider Monitor','/ops/providers','','Odometer','monitor:provider:read',20,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','maintenance-center','维护管理','Maintenance','/ops/maintenance','','Warning','maintenance:tenant:read',30,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','release-center','发布中心','Release Center','/ops/releases','','UploadFilled','release:read',40,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','system-health','系统健康','System Health','/ops/health','','FirstAidKit','health:system:read',50,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','jobs-retries','任务与重试','Jobs & Retries','/ops/jobs','','Timer','jobs:read',60,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','alert-center','告警中心','Alert Center','/ops/alerts','','Bell','alerts:read',70,1,1,1,NOW(3),NOW(3)),
(@ops_stability_menu_id,'MENU','backup-center','备份中心','Backup Center','/ops/backups','','FolderOpened','backup:read',80,1,1,1,NOW(3),NOW(3));

INSERT INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at,deleted_at)
SELECT r.id,m.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r JOIN game_admin_menus m ON m.code IN (
  'support-center','support-player-search','support-player-timeline','support-order-search','support-payment-diagnostics','support-game-diagnostics','support-cases','support-exports',
  'ops-stability','payment-monitor','provider-monitor','maintenance-center','release-center','system-health','jobs-retries','alert-center','backup-center'
) AND m.deleted_at IS NULL
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

INSERT INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at,deleted_at)
SELECT r.id,m.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r JOIN game_admin_menus m ON m.code IN (
  'support-center','support-player-search','support-player-timeline','support-order-search','support-payment-diagnostics','support-game-diagnostics','support-cases','support-exports',
  'ops-stability','payment-monitor','provider-monitor','maintenance-center','release-center'
) AND m.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

INSERT INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'support:player:read','support:timeline:read','support:case:read','support:case:write','support:export',
  'monitor:payment:read','monitor:payment:review','monitor:provider:read','operations:note:write',
  'maintenance:tenant:read','maintenance:tenant:write','maintenance:provider:read','maintenance:provider:write','maintenance:game:read','maintenance:game:write',
  'release:read','release:publish','release:rollback'
) AND p.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);
