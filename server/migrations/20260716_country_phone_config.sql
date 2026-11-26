DROP PROCEDURE IF EXISTS country_config_add_column;

DELIMITER //
CREATE PROCEDURE country_config_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE statement_handle FROM @statement_value; EXECUTE statement_handle; DEALLOCATE PREPARE statement_handle;
  END IF;
END//
DELIMITER ;

CALL country_config_add_column('game_country_configs','phone_country_code','VARCHAR(8) NOT NULL DEFAULT ''+92'' AFTER `supported_languages`');
CALL country_config_add_column('game_country_configs','phone_length','TINYINT UNSIGNED NOT NULL DEFAULT 10 AFTER `phone_country_code`');

UPDATE game_country_configs
SET phone_country_code=CASE country_code
  WHEN 'PK' THEN '+92'
  WHEN 'US' THEN '+1'
  WHEN 'MX' THEN '+52'
  WHEN 'JP' THEN '+81'
  WHEN 'TH' THEN '+66'
  WHEN 'ID' THEN '+62'
  ELSE phone_country_code
END,
phone_length=CASE country_code
  WHEN 'TH' THEN 9
  ELSE IF(phone_length=0,10,phone_length)
END,
updated_at=NOW(3)
WHERE deleted_at IS NULL;

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected)
VALUES
('system:country:delete','Delete country configuration','删除国家配置','Delete country configuration','system','ACTION',1,1);

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3)
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code='system:country:delete'
WHERE r.code='super-admin' AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

DROP PROCEDURE country_config_add_column;
