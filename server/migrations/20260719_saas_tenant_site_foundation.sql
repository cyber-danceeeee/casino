DROP PROCEDURE IF EXISTS saas_add_column;
DROP PROCEDURE IF EXISTS saas_add_index;
DROP PROCEDURE IF EXISTS saas_add_unique_index;
DROP PROCEDURE IF EXISTS saas_drop_index;
DELIMITER $$
CREATE PROCEDURE saas_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE saas_statement FROM @statement_value; EXECUTE saas_statement; DEALLOCATE PREPARE saas_statement;
  END IF;
END$$
CREATE PROCEDURE saas_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE saas_statement FROM @statement_value; EXECUTE saas_statement; DEALLOCATE PREPARE saas_statement;
  END IF;
END$$
CREATE PROCEDURE saas_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE KEY `',index_name_value,'` ',definition_value);
    PREPARE saas_statement FROM @statement_value; EXECUTE saas_statement; DEALLOCATE PREPARE saas_statement;
  END IF;
END$$
CREATE PROCEDURE saas_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE saas_statement FROM @statement_value; EXECUTE saas_statement; DEALLOCATE PREPARE saas_statement;
  END IF;
END$$
DELIMITER ;

CREATE TABLE IF NOT EXISTS game_saas_tenants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_code VARCHAR(64) NOT NULL,
  tenant_name VARCHAR(128) NOT NULL,
  company_name VARCHAR(190) NOT NULL DEFAULT '',
  contact_name VARCHAR(128) NOT NULL DEFAULT '',
  contact_email VARCHAR(190) NOT NULL DEFAULT '',
  contact_phone VARCHAR(32) NOT NULL DEFAULT '',
  plan_code VARCHAR(64) NOT NULL DEFAULT 'STANDARD',
  default_currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  default_language VARCHAR(16) NOT NULL DEFAULT 'en',
  timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Karachi',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  remark VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_saas_tenant_code (tenant_code),
  KEY idx_saas_tenant_status (enabled, status, deleted_at),
  KEY idx_saas_tenant_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_saas_sites (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_code VARCHAR(64) NOT NULL,
  site_name VARCHAR(128) NOT NULL,
  primary_domain VARCHAR(190) NOT NULL DEFAULT '',
  admin_domain VARCHAR(190) NOT NULL DEFAULT '',
  h5_domain VARCHAR(190) NOT NULL DEFAULT '',
  android_package_name VARCHAR(190) NOT NULL DEFAULT '',
  ios_bundle_id VARCHAR(190) NOT NULL DEFAULT '',
  country_code VARCHAR(8) NOT NULL DEFAULT '',
  currency_code VARCHAR(8) NOT NULL DEFAULT '',
  default_language VARCHAR(16) NOT NULL DEFAULT 'en',
  timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Karachi',
  brand_config JSON NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL DEFAULT 'BUILDING',
  remark VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_saas_site_code (site_code),
  KEY idx_saas_site_tenant (tenant_id, enabled, deleted_at),
  KEY idx_saas_site_domains (primary_domain, admin_domain, h5_domain),
  KEY idx_saas_site_status (enabled, status, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CALL saas_add_column('game_admin_accounts','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL saas_add_column('game_admin_accounts','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL saas_add_column('game_admin_roles','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL saas_add_column('game_admin_departments','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL saas_add_column('game_admin_sessions','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `admin_id`');
CALL saas_add_column('game_admin_sessions','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL saas_add_column('game_admin_login_logs','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `admin_id`');
CALL saas_add_column('game_admin_login_logs','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL saas_add_column('game_admin_operation_logs','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `operator_id`');
CALL saas_add_column('game_admin_operation_logs','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');

CALL saas_add_index('game_admin_accounts','idx_admin_tenant_site','(`tenant_id`,`site_id`,`status`,`deleted_at`)');
CALL saas_add_index('game_admin_roles','idx_admin_role_tenant','(`tenant_id`,`status`,`deleted_at`)');
CALL saas_add_index('game_admin_departments','idx_admin_department_tenant','(`tenant_id`,`enabled`,`deleted_at`)');
CALL saas_add_index('game_admin_sessions','idx_admin_session_tenant','(`tenant_id`,`site_id`,`expired_at`)');
CALL saas_add_index('game_admin_login_logs','idx_admin_login_tenant_time','(`tenant_id`,`site_id`,`created_at`)');
CALL saas_add_index('game_admin_operation_logs','idx_admin_audit_tenant_time','(`tenant_id`,`site_id`,`created_at`)');

CALL saas_drop_index('game_admin_accounts','uk_game_admin_accounts_username');
CALL saas_drop_index('game_admin_roles','uk_game_admin_roles_code');
CALL saas_drop_index('game_admin_departments','uk_admin_department_code');
CALL saas_add_unique_index('game_admin_accounts','uk_admin_account_tenant_username','(`tenant_id`, `username`)');
CALL saas_add_unique_index('game_admin_roles','uk_admin_role_tenant_code','(`tenant_id`, `code`)');
CALL saas_add_unique_index('game_admin_departments','uk_admin_department_tenant_code','(`tenant_id`, `code`)');

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:tenant:list','List SaaS tenants','查看租户','List SaaS tenants','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:view','View SaaS tenant','查看租户详情','View SaaS tenant','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:create','Create SaaS tenant','创建租户','Create SaaS tenant','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:update','Update SaaS tenant','更新租户','Update SaaS tenant','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:status','Update SaaS tenant status','变更租户状态','Update SaaS tenant status','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:delete','Delete SaaS tenant','删除租户','Delete SaaS tenant','system','API',1,1,NOW(3),NOW(3)),
('system:site:list','List SaaS sites','查看站点','List SaaS sites','system','API',1,1,NOW(3),NOW(3)),
('system:site:view','View SaaS site','查看站点详情','View SaaS site','system','API',1,1,NOW(3),NOW(3)),
('system:site:create','Create SaaS site','创建站点','Create SaaS site','system','API',1,1,NOW(3),NOW(3)),
('system:site:update','Update SaaS site','更新站点','Update SaaS site','system','API',1,1,NOW(3),NOW(3)),
('system:site:status','Update SaaS site status','变更站点状态','Update SaaS site status','system','API',1,1,NOW(3),NOW(3)),
('system:site:delete','Delete SaaS site','删除站点','Delete SaaS site','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'system:tenant:list','system:tenant:view','system:tenant:create','system:tenant:update','system:tenant:status','system:tenant:delete',
  'system:site:list','system:site:view','system:site:create','system:site:update','system:site:status','system:site:delete'
)
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

DROP PROCEDURE saas_add_column;
DROP PROCEDURE saas_add_index;
DROP PROCEDURE saas_add_unique_index;
DROP PROCEDURE saas_drop_index;
