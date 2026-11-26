DROP PROCEDURE IF EXISTS white_label_add_column;
DROP PROCEDURE IF EXISTS white_label_add_index;
DELIMITER $$
CREATE PROCEDURE white_label_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE white_label_statement FROM @statement_value; EXECUTE white_label_statement; DEALLOCATE PREPARE white_label_statement;
  END IF;
END$$
CREATE PROCEDURE white_label_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE white_label_statement FROM @statement_value; EXECUTE white_label_statement; DEALLOCATE PREPARE white_label_statement;
  END IF;
END$$
DELIMITER ;

CALL white_label_add_column('game_saas_tenants','country_code','VARCHAR(8) NOT NULL DEFAULT '''' AFTER `plan_code`');
CALL white_label_add_column('game_saas_tenants','package_code','VARCHAR(64) NOT NULL DEFAULT ''STANDARD'' AFTER `country_code`');
CALL white_label_add_column('game_saas_tenants','expires_at','DATETIME(3) NULL AFTER `timezone`');
CALL white_label_add_column('game_saas_tenants','maintenance_enabled','TINYINT(1) NOT NULL DEFAULT 0 AFTER `status`');
CALL white_label_add_column('game_saas_tenants','maintenance_message','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `maintenance_enabled`');
CALL white_label_add_index('game_saas_tenants','idx_saas_tenant_country','(`country_code`,`enabled`,`deleted_at`)');
CALL white_label_add_index('game_saas_tenants','idx_saas_tenant_expiry','(`expires_at`,`enabled`,`deleted_at`)');

CALL white_label_add_column('game_saas_sites','theme_config','JSON NULL AFTER `brand_config`');
CALL white_label_add_column('game_saas_sites','feature_config','JSON NULL AFTER `theme_config`');
CALL white_label_add_column('game_saas_sites','maintenance_config','JSON NULL AFTER `feature_config`');
CALL white_label_add_column('game_saas_sites','seo_config','JSON NULL AFTER `maintenance_config`');

CREATE TABLE IF NOT EXISTS game_saas_domains (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  domain VARCHAR(190) NOT NULL,
  domain_key VARCHAR(255) NOT NULL,
  domain_type VARCHAR(24) NOT NULL DEFAULT 'H5',
  is_primary TINYINT(1) NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  remark VARCHAR(1000) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_saas_domain_key (domain_key),
  KEY idx_saas_domain_lookup (domain, enabled, deleted_at),
  KEY idx_saas_domain_scope (tenant_id, site_id, domain_type, enabled, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_saas_domains
(tenant_id, site_id, domain, domain_key, domain_type, is_primary, enabled, status, created_at, updated_at)
SELECT tenant_id, id, LOWER(TRIM(primary_domain)), LOWER(TRIM(primary_domain)), 'PRIMARY', 1, enabled, status, NOW(3), NOW(3)
FROM game_saas_sites
WHERE deleted_at IS NULL AND TRIM(primary_domain) <> '';

INSERT IGNORE INTO game_saas_domains
(tenant_id, site_id, domain, domain_key, domain_type, is_primary, enabled, status, created_at, updated_at)
SELECT tenant_id, id, LOWER(TRIM(h5_domain)), LOWER(TRIM(h5_domain)), 'H5', 1, enabled, status, NOW(3), NOW(3)
FROM game_saas_sites
WHERE deleted_at IS NULL AND TRIM(h5_domain) <> '';

INSERT IGNORE INTO game_saas_domains
(tenant_id, site_id, domain, domain_key, domain_type, is_primary, enabled, status, created_at, updated_at)
SELECT tenant_id, id, LOWER(TRIM(admin_domain)), LOWER(TRIM(admin_domain)), 'ADMIN', 0, enabled, status, NOW(3), NOW(3)
FROM game_saas_sites
WHERE deleted_at IS NULL AND TRIM(admin_domain) <> '';

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:domain:list','List SaaS domains','查看域名','List SaaS domains','system','API',1,1,NOW(3),NOW(3)),
('system:domain:view','View SaaS domain','查看域名详情','View SaaS domain','system','API',1,1,NOW(3),NOW(3)),
('system:domain:create','Create SaaS domain','创建域名','Create SaaS domain','system','API',1,1,NOW(3),NOW(3)),
('system:domain:update','Update SaaS domain','更新域名','Update SaaS domain','system','API',1,1,NOW(3),NOW(3)),
('system:domain:status','Update SaaS domain status','变更域名状态','Update SaaS domain status','system','API',1,1,NOW(3),NOW(3)),
('system:domain:delete','Delete SaaS domain','删除域名','Delete SaaS domain','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code LIKE 'system:domain:%'
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

DROP PROCEDURE IF EXISTS white_label_add_column;
DROP PROCEDURE IF EXISTS white_label_add_index;
