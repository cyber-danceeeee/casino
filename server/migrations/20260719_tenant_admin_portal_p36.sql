DROP PROCEDURE IF EXISTS tenant_admin_portal_add_column;
DELIMITER $$
CREATE PROCEDURE tenant_admin_portal_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = table_name_value AND COLUMN_NAME = column_name_value
  ) THEN
    SET @statement_value = CONCAT('ALTER TABLE `', table_name_value, '` ADD COLUMN `', column_name_value, '` ', definition_value);
    PREPARE stmt FROM @statement_value;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$
DELIMITER ;

CALL tenant_admin_portal_add_column('game_saas_sites','is_initialized','TINYINT(1) NOT NULL DEFAULT 0 AFTER `config_draft_updated_by`');

UPDATE game_saas_sites
SET is_initialized = IF(config_version > 1, 1, is_initialized)
WHERE deleted_at IS NULL;

DROP PROCEDURE IF EXISTS tenant_admin_portal_add_column;
