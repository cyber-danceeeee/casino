DROP PROCEDURE IF EXISTS system_add_column;
DROP PROCEDURE IF EXISTS system_add_index;
DELIMITER $$
CREATE PROCEDURE system_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE system_statement FROM @statement_value; EXECUTE system_statement; DEALLOCATE PREPARE system_statement;
  END IF;
END$$
CREATE PROCEDURE system_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE system_statement FROM @statement_value; EXECUTE system_statement; DEALLOCATE PREPARE system_statement;
  END IF;
END$$
DELIMITER ;

CALL system_add_column('game_admin_accounts','password_hash','VARCHAR(255) NOT NULL DEFAULT '''' AFTER `display_name`');
CALL system_add_column('game_admin_accounts','department_id','BIGINT UNSIGNED NULL AFTER `password_hash`');
CALL system_add_column('game_admin_accounts','mobile','VARCHAR(32) NOT NULL DEFAULT '''' AFTER `department_id`');
CALL system_add_column('game_admin_accounts','email','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `mobile`');
CALL system_add_column('game_admin_accounts','mfa_enabled','TINYINT(1) NOT NULL DEFAULT 0 AFTER `email`');
CALL system_add_column('game_admin_accounts','mfa_secret_hash','VARCHAR(255) NOT NULL DEFAULT '''' AFTER `mfa_enabled`');
CALL system_add_column('game_admin_accounts','system_protected','TINYINT(1) NOT NULL DEFAULT 0 AFTER `mfa_secret_hash`');
CALL system_add_column('game_admin_accounts','last_login_ip','VARCHAR(64) NOT NULL DEFAULT '''' AFTER `system_protected`');
CALL system_add_column('game_admin_accounts','last_login_at','DATETIME(3) NULL AFTER `last_login_ip`');
CALL system_add_column('game_admin_accounts','failed_login_count','INT UNSIGNED NOT NULL DEFAULT 0 AFTER `last_login_at`');
CALL system_add_column('game_admin_accounts','last_password_changed_at','DATETIME(3) NULL AFTER `failed_login_count`');
CALL system_add_column('game_admin_accounts','version','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `last_password_changed_at`');
CALL system_add_index('game_admin_accounts','idx_admin_department','(`department_id`)');
CALL system_add_index('game_admin_accounts','idx_admin_status_created','(`status`,`created_at`)');
CALL system_add_index('game_admin_accounts','idx_admin_email','(`email`)');
CALL system_add_index('game_admin_accounts','idx_admin_mobile','(`mobile`)');

CALL system_add_column('game_admin_roles','name_zh','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `name`');
CALL system_add_column('game_admin_roles','name_en','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `name_zh`');
CALL system_add_column('game_admin_roles','description_zh','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `name_en`');
CALL system_add_column('game_admin_roles','description_en','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `description_zh`');
CALL system_add_column('game_admin_roles','data_scope','VARCHAR(32) NOT NULL DEFAULT ''ALL'' AFTER `description_en`');
CALL system_add_column('game_admin_roles','system_protected','TINYINT(1) NOT NULL DEFAULT 0 AFTER `status`');
CALL system_add_column('game_admin_roles','version','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `system_protected`');
CALL system_add_index('game_admin_roles','idx_admin_role_status','(`status`)');

UPDATE game_admin_roles SET name_zh=IF(name_zh='',name,name_zh), name_en=IF(name_en='',name,name_en),
  system_protected=IF(code='super-admin',1,system_protected);

CALL system_add_column('game_admin_permissions','name_zh','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `name`');
CALL system_add_column('game_admin_permissions','name_en','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `name_zh`');
CALL system_add_column('game_admin_permissions','module','VARCHAR(64) NOT NULL DEFAULT '''' AFTER `name_en`');
CALL system_add_column('game_admin_permissions','permission_type','VARCHAR(16) NOT NULL DEFAULT ''API'' AFTER `module`');
CALL system_add_column('game_admin_permissions','api_path','VARCHAR(255) NOT NULL DEFAULT '''' AFTER `permission_type`');
CALL system_add_column('game_admin_permissions','enabled','TINYINT(1) NOT NULL DEFAULT 1 AFTER `api_path`');
CALL system_add_column('game_admin_permissions','system_protected','TINYINT(1) NOT NULL DEFAULT 0 AFTER `enabled`');
CALL system_add_index('game_admin_permissions','idx_admin_permission_module','(`module`,`permission_type`,`enabled`)');

UPDATE game_admin_permissions SET name_zh=IF(name_zh='',name,name_zh), name_en=IF(name_en='',name,name_en),
  module=IF(module='',SUBSTRING_INDEX(code,':',1),module), system_protected=1;

CALL system_add_column('game_admin_sessions','ip_address','VARCHAR(64) NOT NULL DEFAULT '''' AFTER `token_hash`');
CALL system_add_column('game_admin_sessions','user_agent','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `ip_address`');

CALL system_add_column('game_admin_operation_logs','module','VARCHAR(64) NOT NULL DEFAULT '''' AFTER `permission`');
CALL system_add_column('game_admin_operation_logs','result','VARCHAR(16) NOT NULL DEFAULT ''SUCCESS'' AFTER `action`');
CALL system_add_column('game_admin_operation_logs','request_info','JSON NULL AFTER `correlation_id`');
CALL system_add_column('game_admin_operation_logs','error_info','JSON NULL AFTER `request_info`');
CALL system_add_index('game_admin_operation_logs','idx_admin_audit_module_time','(`module`,`created_at`)');
CALL system_add_index('game_admin_operation_logs','idx_admin_audit_result_time','(`result`,`created_at`)');

CALL system_add_column('game_system_configs','name_zh','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `config_key`');
CALL system_add_column('game_system_configs','name_en','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `name_zh`');
CALL system_add_column('game_system_configs','category','VARCHAR(64) NOT NULL DEFAULT ''BASIC'' AFTER `name_en`');
CALL system_add_column('game_system_configs','value_type','VARCHAR(16) NOT NULL DEFAULT ''STRING'' AFTER `config_value`');
CALL system_add_column('game_system_configs','description_zh','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `description`');
CALL system_add_column('game_system_configs','description_en','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `description_zh`');
CALL system_add_column('game_system_configs','editable','TINYINT(1) NOT NULL DEFAULT 1 AFTER `description_en`');
CALL system_add_column('game_system_configs','sensitive','TINYINT(1) NOT NULL DEFAULT 0 AFTER `editable`');
CALL system_add_column('game_system_configs','restart_required','TINYINT(1) NOT NULL DEFAULT 0 AFTER `sensitive`');
CALL system_add_column('game_system_configs','allowed_values','JSON NULL AFTER `restart_required`');
CALL system_add_index('game_system_configs','idx_system_config_category_status','(`category`,`status`)');

DROP PROCEDURE system_add_column;
DROP PROCEDURE system_add_index;

CREATE TABLE IF NOT EXISTS game_admin_departments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  parent_id BIGINT UNSIGNED NULL,
  code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  leader_id BIGINT UNSIGNED NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  system_protected TINYINT(1) NOT NULL DEFAULT 0,
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_department_code (code),
  KEY idx_admin_department_parent (parent_id, sort_order),
  KEY idx_admin_department_enabled (enabled, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_admin_menus (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  parent_id BIGINT UNSIGNED NULL,
  menu_type VARCHAR(16) NOT NULL,
  code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  route_path VARCHAR(255) NOT NULL DEFAULT '',
  component_path VARCHAR(255) NOT NULL DEFAULT '',
  icon VARCHAR(64) NOT NULL DEFAULT '',
  permission_code VARCHAR(128) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  visible TINYINT(1) NOT NULL DEFAULT 1,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  cacheable TINYINT(1) NOT NULL DEFAULT 0,
  external_link VARCHAR(500) NOT NULL DEFAULT '',
  system_protected TINYINT(1) NOT NULL DEFAULT 0,
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_menu_code (code),
  UNIQUE KEY uk_admin_menu_route (route_path, deleted_at),
  KEY idx_admin_menu_parent (parent_id, sort_order),
  KEY idx_admin_menu_permission (permission_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_admin_role_menus (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  role_id BIGINT UNSIGNED NOT NULL,
  menu_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_menu (role_id, menu_id),
  KEY idx_admin_role_menu_menu (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_admin_dictionaries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  dictionary_code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  description_zh VARCHAR(500) NOT NULL DEFAULT '',
  description_en VARCHAR(500) NOT NULL DEFAULT '',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  system_protected TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_dictionary_code (dictionary_code),
  KEY idx_admin_dictionary_enabled (enabled, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_admin_dictionary_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  dictionary_id BIGINT UNSIGNED NOT NULL,
  item_code VARCHAR(64) NOT NULL,
  label_zh VARCHAR(128) NOT NULL,
  label_en VARCHAR(128) NOT NULL,
  item_value VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  metadata JSON NULL,
  system_protected TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_dictionary_item (dictionary_id, item_code),
  KEY idx_admin_dictionary_item_sort (dictionary_id, enabled, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_admin_login_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id BIGINT UNSIGNED NULL,
  username VARCHAR(64) NOT NULL DEFAULT '',
  success TINYINT(1) NOT NULL,
  failure_reason VARCHAR(255) NOT NULL DEFAULT '',
  ip_address VARCHAR(64) NOT NULL DEFAULT '',
  country VARCHAR(64) NOT NULL DEFAULT '',
  device VARCHAR(128) NOT NULL DEFAULT '',
  operating_system VARCHAR(128) NOT NULL DEFAULT '',
  browser VARCHAR(128) NOT NULL DEFAULT '',
  user_agent VARCHAR(500) NOT NULL DEFAULT '',
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_admin_login_admin_time (admin_id, created_at),
  KEY idx_admin_login_username_time (username, created_at),
  KEY idx_admin_login_ip_time (ip_address, created_at),
  KEY idx_admin_login_success_time (success, created_at),
  KEY idx_admin_login_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_system_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  task_type VARCHAR(32) NOT NULL,
  handler_code VARCHAR(128) NOT NULL,
  cron_expression VARCHAR(64) NOT NULL,
  payload JSON NULL,
  timeout_seconds INT UNSIGNED NOT NULL DEFAULT 300,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  running TINYINT(1) NOT NULL DEFAULT 0,
  system_protected TINYINT(1) NOT NULL DEFAULT 0,
  last_run_at DATETIME(3) NULL,
  last_run_result VARCHAR(16) NOT NULL DEFAULT '',
  next_run_at DATETIME(3) NULL,
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_task_code (task_code),
  KEY idx_system_task_due (enabled, running, next_run_at),
  KEY idx_system_task_handler (handler_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_system_task_executions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  execution_no VARCHAR(64) NOT NULL,
  task_id BIGINT UNSIGNED NOT NULL,
  task_code VARCHAR(64) NOT NULL,
  triggered_by VARCHAR(64) NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(16) NOT NULL,
  started_at DATETIME(3) NOT NULL,
  completed_at DATETIME(3) NULL,
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  error_message VARCHAR(1000) NOT NULL DEFAULT '',
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_system_task_execution_no (execution_no),
  KEY idx_system_task_execution_task_time (task_id, started_at),
  KEY idx_system_task_execution_status_time (status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_departments (code,name_zh,name_en,sort_order,enabled,system_protected)
VALUES ('ROOT','总部','Head Office',1,1,1);

UPDATE game_admin_accounts a JOIN game_admin_departments d ON d.code='ROOT'
SET a.department_id=COALESCE(a.department_id,d.id), a.system_protected=IF(a.id=1,1,a.system_protected);

INSERT IGNORE INTO game_system_configs
  (config_key,name_zh,name_en,category,config_value,value_type,description,description_zh,description_en,editable,`sensitive`,restart_required,status)
VALUES
  ('basic.platform_name','平台名称','Platform Name','BASIC','Casino Admin','STRING','Platform display name','平台显示名称','Platform display name',1,0,0,1),
  ('security.session_timeout_minutes','会话超时分钟数','Session Timeout Minutes','SECURITY','720','NUMBER','Admin session timeout','管理员会话超时时间','Administrator session timeout',1,0,0,1),
  ('security.password_min_length','密码最小长度','Minimum Password Length','SECURITY','10','NUMBER','Minimum administrator password length','管理员密码最小长度','Minimum administrator password length',1,0,0,1),
  ('user.default_currency','默认用户币种','Default User Currency','USER','PKR','ENUM','Default currency','默认用户币种','Default user currency',1,0,0,1),
  ('game.default_provider','默认游戏供应商','Default Game Provider','GAME','softaggregator','STRING','Default provider','默认游戏供应商','Default game provider',1,0,0,1),
  ('file.storage_driver','文件存储驱动','File Storage Driver','FILE_STORAGE','local','ENUM','Storage driver','文件存储驱动','File storage driver',1,0,1,1);

UPDATE game_system_configs SET allowed_values=JSON_ARRAY('PKR','USD','EUR') WHERE config_key='user.default_currency';
UPDATE game_system_configs SET allowed_values=JSON_ARRAY('local','s3','oss','r2') WHERE config_key='file.storage_driver';

INSERT IGNORE INTO game_admin_dictionaries (dictionary_code,name_zh,name_en,enabled,system_protected) VALUES
  ('COUNTRY','国家','Country',1,1),('CURRENCY','币种','Currency',1,1),('LANGUAGE','语言','Language',1,1),
  ('GAME_TYPE','游戏类型','Game Type',1,1),('RISK_LEVEL','风险等级','Risk Level',1,1),('USER_STATUS','用户状态','User Status',1,1);

INSERT IGNORE INTO game_admin_dictionary_items (dictionary_id,item_code,label_zh,label_en,item_value,sort_order,enabled,system_protected)
SELECT id,'PK','巴基斯坦','Pakistan','PK',1,1,1 FROM game_admin_dictionaries WHERE dictionary_code='COUNTRY';
INSERT IGNORE INTO game_admin_dictionary_items (dictionary_id,item_code,label_zh,label_en,item_value,sort_order,enabled,system_protected)
SELECT id,'PKR','巴基斯坦卢比','Pakistani Rupee','PKR',1,1,1 FROM game_admin_dictionaries WHERE dictionary_code='CURRENCY';
INSERT IGNORE INTO game_admin_dictionary_items (dictionary_id,item_code,label_zh,label_en,item_value,sort_order,enabled,system_protected)
SELECT id,'zh-CN','简体中文','Simplified Chinese','zh-CN',1,1,1 FROM game_admin_dictionaries WHERE dictionary_code='LANGUAGE';
INSERT IGNORE INTO game_admin_dictionary_items (dictionary_id,item_code,label_zh,label_en,item_value,sort_order,enabled,system_protected)
SELECT id,'en-US','英语','English','en-US',2,1,1 FROM game_admin_dictionaries WHERE dictionary_code='LANGUAGE';

INSERT IGNORE INTO game_system_tasks
  (task_code,name_zh,name_en,task_type,handler_code,cron_expression,payload,timeout_seconds,enabled,system_protected,next_run_at)
VALUES
  ('REPORT_DAILY_AGGREGATION','报表每日汇总','Daily Report Aggregation','REPORT','report.daily_aggregation','10 0 * * *',JSON_OBJECT('timezone','+08:00'),600,1,1,DATE_ADD(NOW(3),INTERVAL 1 DAY)),
  ('ADMIN_SESSION_CLEANUP','管理员会话清理','Administrator Session Cleanup','CLEANUP','system.session_cleanup','0 * * * *',JSON_OBJECT(),120,1,1,DATE_ADD(NOW(3),INTERVAL 1 HOUR)),
  ('EXPIRED_BONUS_PROCESSING','过期奖励处理','Expired Bonus Processing','BONUS','operation.expired_bonus','15 * * * *',JSON_OBJECT(),300,1,1,DATE_ADD(NOW(3),INTERVAL 1 HOUR));

INSERT IGNORE INTO game_admin_roles
  (code,name,name_zh,name_en,description_zh,description_en,data_scope,status,system_protected)
VALUES
  ('super-admin','Super Administrator','超级管理员','Super Administrator','系统受保护的超级管理员角色','System-protected super administrator role','ALL',1,1);

INSERT IGNORE INTO game_admin_accounts
  (username,display_name,password_hash,department_id,status,system_protected)
SELECT 'system-admin','System Administrator','',d.id,1,1
FROM game_admin_departments d WHERE d.code='ROOT' AND d.deleted_at IS NULL LIMIT 1;

INSERT IGNORE INTO game_admin_account_roles (admin_id,role_id,created_at,updated_at)
SELECT a.id,r.id,NOW(3),NOW(3) FROM game_admin_accounts a JOIN game_admin_roles r ON r.code='super-admin'
WHERE a.username='system-admin' AND a.deleted_at IS NULL AND r.deleted_at IS NULL;

INSERT IGNORE INTO game_admin_permissions (code,name,name_zh,name_en,module,permission_type,enabled,system_protected) VALUES
  ('system:admin:list','List administrators','管理员列表','List administrators','system','API',1,1),
  ('system:admin:view','View administrator','查看管理员','View administrator','system','API',1,1),
  ('system:admin:create','Create administrator','创建管理员','Create administrator','system','API',1,1),
  ('system:admin:update','Update administrator','更新管理员','Update administrator','system','API',1,1),
  ('system:admin:status','Update administrator status','管理员状态','Update administrator status','system','API',1,1),
  ('system:admin:reset-password','Reset administrator password','重置管理员密码','Reset administrator password','system','ACTION',1,1),
  ('system:admin:reset-mfa','Reset administrator MFA','重置管理员 MFA','Reset administrator MFA','system','ACTION',1,1),
  ('system:admin:force-logout','Force administrator logout','强制管理员退出','Force administrator logout','system','ACTION',1,1),
  ('system:admin:assign-role','Assign administrator roles','分配管理员角色','Assign administrator roles','system','ACTION',1,1),
  ('system:role:list','List roles','角色列表','List roles','system','API',1,1),
  ('system:role:create','Create role','创建角色','Create role','system','API',1,1),
  ('system:role:update','Update role','更新角色','Update role','system','API',1,1),
  ('system:role:delete','Delete role','删除角色','Delete role','system','API',1,1),
  ('system:role:assign-permission','Assign role permissions','分配角色权限','Assign role permissions','system','ACTION',1,1),
  ('system:role:assign-menu','Assign role menus','分配角色菜单','Assign role menus','system','ACTION',1,1),
  ('system:permission:list','List permissions','权限列表','List permissions','system','API',1,1),
  ('system:permission:create','Create permission','创建权限','Create permission','system','API',1,1),
  ('system:permission:update','Update permission','更新权限','Update permission','system','API',1,1),
  ('system:permission:delete','Delete permission','删除权限','Delete permission','system','API',1,1),
  ('system:department:list','List departments','部门列表','List departments','system','API',1,1),
  ('system:department:create','Create department','创建部门','Create department','system','API',1,1),
  ('system:department:update','Update department','更新部门','Update department','system','API',1,1),
  ('system:department:delete','Delete department','删除部门','Delete department','system','API',1,1),
  ('system:menu:list','List menus','菜单列表','List menus','system','API',1,1),
  ('system:menu:create','Create menu','创建菜单','Create menu','system','API',1,1),
  ('system:menu:update','Update menu','更新菜单','Update menu','system','API',1,1),
  ('system:menu:delete','Delete menu','删除菜单','Delete menu','system','API',1,1),
  ('system:config:list','List system configuration','系统配置列表','List system configuration','system','API',1,1),
  ('system:config:update','Update system configuration','更新系统配置','Update system configuration','system','API',1,1),
  ('system:config:secret-update','Update secret configuration','更新密钥配置','Update secret configuration','system','ACTION',1,1),
  ('system:dictionary:list','List dictionaries','字典列表','List dictionaries','system','API',1,1),
  ('system:dictionary:create','Create dictionary','创建字典','Create dictionary','system','API',1,1),
  ('system:dictionary:update','Update dictionary','更新字典','Update dictionary','system','API',1,1),
  ('system:dictionary:delete','Delete dictionary','删除字典','Delete dictionary','system','API',1,1),
  ('system:audit-log:view','View audit logs','查看审计日志','View audit logs','system','API',1,1),
  ('system:login-log:view','View administrator login logs','查看登录日志','View administrator login logs','system','API',1,1),
  ('system:task:list','List scheduled tasks','定时任务列表','List scheduled tasks','system','API',1,1),
  ('system:task:create','Create scheduled task','创建定时任务','Create scheduled task','system','API',1,1),
  ('system:task:update','Update scheduled task','更新定时任务','Update scheduled task','system','API',1,1),
  ('system:task:delete','Delete scheduled task','删除定时任务','Delete scheduled task','system','API',1,1),
  ('system:task:run','Run scheduled task','运行定时任务','Run scheduled task','system','ACTION',1,1),
  ('system:monitoring:view','View system monitoring','查看系统监控','View system monitoring','system','API',1,1);

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_permissions p ON p.code LIKE 'system:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

INSERT IGNORE INTO game_admin_menus
  (parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected)
VALUES
  (NULL,'DIRECTORY','system','系统管理','System Management','','','Setting','',900,1,1,1);

SET @system_menu_id=(SELECT id FROM game_admin_menus WHERE code='system' AND deleted_at IS NULL LIMIT 1);
INSERT IGNORE INTO game_admin_menus
  (parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected)
VALUES
  (@system_menu_id,'MENU','system-admins','管理员管理','Administrator Management','/system/admins','system/SystemManagementPage','User','system:admin:list',10,1,1,1),
  (@system_menu_id,'MENU','system-roles','角色管理','Role Management','/system/roles','system/SystemManagementPage','Avatar','system:role:list',20,1,1,1),
  (@system_menu_id,'MENU','system-permissions','权限管理','Permission Management','/system/permissions','system/SystemManagementPage','Key','system:permission:list',30,1,1,1),
  (@system_menu_id,'MENU','system-departments','部门管理','Department Management','/system/departments','system/SystemManagementPage','OfficeBuilding','system:department:list',40,1,1,1),
  (@system_menu_id,'MENU','system-menus','菜单管理','Menu Management','/system/menus','system/SystemManagementPage','Menu','system:menu:list',50,1,1,1),
  (@system_menu_id,'MENU','system-configs','系统配置','System Configuration','/system/config','system/SystemManagementPage','SetUp','system:config:list',60,1,1,1),
  (@system_menu_id,'MENU','system-dictionaries','字典管理','Dictionary Management','/system/dictionaries','system/SystemManagementPage','Collection','system:dictionary:list',70,1,1,1),
  (@system_menu_id,'MENU','system-audit-logs','审计日志','Audit Logs','/system/audit-logs','system/SystemManagementPage','DocumentChecked','system:audit-log:view',80,1,1,1),
  (@system_menu_id,'MENU','system-login-logs','登录日志','Login Logs','/system/login-logs','system/SystemManagementPage','Clock','system:login-log:view',90,1,1,1),
  (@system_menu_id,'MENU','system-tasks','定时任务','Scheduled Tasks','/system/tasks','system/SystemManagementPage','Timer','system:task:list',100,1,1,1),
  (@system_menu_id,'MENU','system-monitoring','系统监控','System Monitoring','/system/monitor','system/SystemManagementPage','Monitor','system:monitoring:view',110,1,1,1);

INSERT IGNORE INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at)
SELECT r.id,m.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_menus m ON m.code='system' OR m.code LIKE 'system-%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL AND m.deleted_at IS NULL;
