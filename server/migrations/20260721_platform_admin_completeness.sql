DELIMITER $$

DROP PROCEDURE IF EXISTS pac_add_column$$
CREATE PROCEDURE pac_add_column(IN table_name_value VARCHAR(128), IN column_name_value VARCHAR(128), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE prepared_statement FROM @statement_value;
    EXECUTE prepared_statement;
    DEALLOCATE PREPARE prepared_statement;
  END IF;
END$$

DELIMITER ;

CREATE TABLE IF NOT EXISTS game_saas_tenant_packages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  package_code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL DEFAULT '',
  name_en VARCHAR(128) NOT NULL DEFAULT '',
  site_limit INT NOT NULL DEFAULT 1,
  game_limit INT NOT NULL DEFAULT 0,
  provider_limit INT NOT NULL DEFAULT 0,
  storage_gb INT NOT NULL DEFAULT 0,
  player_limit BIGINT NOT NULL DEFAULT 0,
  status VARCHAR(24) NOT NULL DEFAULT 'AVAILABLE',
  sort_order INT NOT NULL DEFAULT 0,
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_package_code (package_code),
  KEY idx_tenant_package_status_sort (status,sort_order,deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO game_saas_tenant_packages
(package_code,name_zh,name_en,site_limit,game_limit,provider_limit,storage_gb,player_limit,status,sort_order,created_at,updated_at)
VALUES
('BASIC','基础版','Basic',1,250,3,10,10000,'AVAILABLE',10,NOW(3),NOW(3)),
('PROFESSIONAL','专业版','Professional',3,1000,10,100,100000,'AVAILABLE',20,NOW(3),NOW(3)),
('ENTERPRISE','企业版','Enterprise',-1,-1,-1,-1,-1,'CONTACT',30,NOW(3),NOW(3)),
('STANDARD','标准版','Standard',1,500,5,25,50000,'LEGACY',40,NOW(3),NOW(3))
ON DUPLICATE KEY UPDATE deleted_at=NULL,updated_at=NOW(3);

CALL pac_add_column('game_saas_sites','android_apk_object_key','VARCHAR(1000) NOT NULL DEFAULT '''' AFTER `android_package_name`');
CALL pac_add_column('game_saas_sites','android_apk_name','VARCHAR(255) NOT NULL DEFAULT '''' AFTER `android_apk_object_key`');
CALL pac_add_column('game_saas_sites','android_apk_size','BIGINT NOT NULL DEFAULT 0 AFTER `android_apk_name`');
CALL pac_add_column('game_saas_sites','android_apk_uploaded_at','DATETIME(3) NULL AFTER `android_apk_size`');

INSERT INTO game_system_configs
(tenant_id,site_id,config_key,name_zh,name_en,category,config_value,value_type,description,description_zh,description_en,editable,`sensitive`,restart_required,allowed_values,status,created_at,updated_at)
VALUES
 (0,0,'basic.platform_logo_url','平台 Logo','Platform Logo','PLATFORM','','STRING','Platform admin logo','平台与租户后台左上角 Logo','Logo displayed in platform and tenant admin headers',1,0,0,NULL,1,NOW(3),NOW(3)),
 (0,0,'basic.platform_contact','平台联系方式','Platform Contact','PLATFORM','support@soranopro.com','STRING','Platform contact','后台页脚展示的平台联系方式','Platform contact shown in the admin footer',1,0,0,NULL,1,NOW(3),NOW(3)),
 (0,0,'finance.platform_usdt_address','平台 USDT 地址','Platform USDT Address','PLATFORM','','STRING','Platform USDT address','租户充值时展示的平台 USDT 收款地址','Platform USDT address shown during tenant deposits',1,0,0,NULL,1,NOW(3),NOW(3))
ON DUPLICATE KEY UPDATE name_zh=VALUES(name_zh),name_en=VALUES(name_en),category='PLATFORM',description_zh=VALUES(description_zh),description_en=VALUES(description_en),editable=1,deleted_at=NULL,updated_at=NOW(3);

UPDATE game_system_configs
SET category='PLATFORM',name_zh='平台名称',name_en='Platform Name',description_zh='平台与租户后台左上角名称',description_en='Name displayed in platform and tenant admin headers',updated_at=NOW(3)
WHERE config_key='basic.platform_name' AND deleted_at IS NULL;

UPDATE game_provider_sync_logs
SET status='FAILED',completed_at=NOW(3),duration_ms=TIMESTAMPDIFF(MICROSECOND,started_at,NOW(3))/1000,
    error_message='provider synchronization timed out and was recovered',updated_at=NOW(3)
WHERE status IN ('PENDING','RUNNING') AND started_at<DATE_SUB(NOW(3),INTERVAL 35 MINUTE) AND deleted_at IS NULL;

DROP PROCEDURE IF EXISTS pac_add_column;
