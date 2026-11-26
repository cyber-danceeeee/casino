DELIMITER $$

DROP PROCEDURE IF EXISTS p38_add_column$$
CREATE PROCEDURE p38_add_column(IN table_name_value VARCHAR(128), IN column_name_value VARCHAR(128), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = table_name_value AND column_name = column_name_value
  ) THEN
    SET @statement_value = CONCAT('ALTER TABLE `', table_name_value, '` ADD COLUMN `', column_name_value, '` ', definition_value);
    PREPARE prepared_statement FROM @statement_value;
    EXECUTE prepared_statement;
    DEALLOCATE PREPARE prepared_statement;
  END IF;
END$$

DROP PROCEDURE IF EXISTS p38_add_index$$
CREATE PROCEDURE p38_add_index(IN table_name_value VARCHAR(128), IN index_name_value VARCHAR(128), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = table_name_value AND index_name = index_name_value
  ) THEN
    SET @statement_value = CONCAT('ALTER TABLE `', table_name_value, '` ADD INDEX `', index_name_value, '` ', definition_value);
    PREPARE prepared_statement FROM @statement_value;
    EXECUTE prepared_statement;
    DEALLOCATE PREPARE prepared_statement;
  END IF;
END$$

DELIMITER ;

CALL p38_add_column('game_saas_domains','verification_status','VARCHAR(24) NOT NULL DEFAULT ''PENDING'' AFTER `status`');
CALL p38_add_column('game_saas_domains','ssl_status','VARCHAR(24) NOT NULL DEFAULT ''PENDING'' AFTER `verification_status`');
CALL p38_add_column('game_saas_domains','ssl_expires_at','DATETIME(3) NULL AFTER `ssl_status`');
CALL p38_add_column('game_saas_domains','last_verified_at','DATETIME(3) NULL AFTER `ssl_expires_at`');
CALL p38_add_column('game_saas_domains','verification_error','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `last_verified_at`');
CALL p38_add_index('game_saas_domains','idx_saas_domain_verification','(`verification_status`,`ssl_status`,`ssl_expires_at`)');

CREATE TABLE IF NOT EXISTS game_saas_brand_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL,
  admin_id BIGINT UNSIGNED NOT NULL,
  admin_username VARCHAR(128) NOT NULL DEFAULT '',
  primary_domain VARCHAR(190) NOT NULL DEFAULT '',
  h5_url VARCHAR(500) NOT NULL DEFAULT '',
  admin_url VARCHAR(500) NOT NULL DEFAULT '',
  domain_status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  initial_config_version BIGINT UNSIGNED NOT NULL DEFAULT 1,
  status VARCHAR(24) NOT NULL DEFAULT 'DELIVERED',
  delivered_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  delivered_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_brand_delivery_tenant_site (tenant_id, site_id),
  KEY idx_brand_delivery_time (delivered_at, id),
  KEY idx_brand_delivery_status (domain_status, status, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_saas_brand_deliveries
(tenant_id,site_id,admin_id,admin_username,primary_domain,h5_url,admin_url,domain_status,initial_config_version,status,delivered_by,delivered_at,created_at,updated_at)
SELECT t.id,s.id,COALESCE(a.id,0),COALESCE(a.username,''),s.primary_domain,
       IF(s.h5_domain='','',CONCAT('https://',s.h5_domain)),
       IF(s.admin_domain='','',CONCAT('https://',s.admin_domain,'/tenant/login')),
       COALESCE(d.status,'PENDING'),GREATEST(s.config_version,1),'DELIVERED',t.created_by,t.created_at,t.created_at,t.updated_at
FROM game_saas_tenants t
JOIN game_saas_sites s ON s.tenant_id=t.id AND s.deleted_at IS NULL
LEFT JOIN game_admin_accounts a ON a.id=(SELECT MIN(a2.id) FROM game_admin_accounts a2 WHERE a2.tenant_id=t.id AND a2.deleted_at IS NULL)
LEFT JOIN game_saas_domains d ON d.id=(SELECT MIN(d2.id) FROM game_saas_domains d2 WHERE d2.tenant_id=t.id AND d2.site_id=s.id AND d2.deleted_at IS NULL)
WHERE t.deleted_at IS NULL;

SET @system_menu_id=(SELECT id FROM game_admin_menus WHERE code='system' AND deleted_at IS NULL LIMIT 1);

INSERT IGNORE INTO game_admin_menus
(parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected,created_at,updated_at)
VALUES
(@system_menu_id,'MENU','brand-deliveries','品牌交付记录','Brand Deliveries','/platform/tenants/deliveries','platform/BrandDeliveryPage','Tickets','system:tenant:list',25,1,1,1,NOW(3),NOW(3)),
(@system_menu_id,'MENU','tenant-domains','域名管理','Domain Management','/platform/tenants/domains','platform/DomainManagementPage','Link','system:domain:list',35,1,1,1,NOW(3),NOW(3)),
(@system_menu_id,'MENU','tenant-packages','租户套餐','Tenant Packages','/platform/tenants/packages','platform/TenantPackagesPage','Box','system:tenant:list',40,1,1,1,NOW(3),NOW(3)),
(@system_menu_id,'MENU','tenant-audit','租户审计','Tenant Audit','/platform/tenants/audit','platform/TenantAuditPage','DocumentChecked','system:audit-log:view',45,1,1,1,NOW(3),NOW(3));

INSERT INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at,deleted_at)
SELECT r.id,m.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code IN ('system','tenants','create-brand','brand-deliveries','sites','tenant-domains','tenant-packages','tenant-audit')
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND m.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=VALUES(updated_at);

DROP PROCEDURE IF EXISTS p38_add_column;
DROP PROCEDURE IF EXISTS p38_add_index;
