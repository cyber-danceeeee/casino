DROP PROCEDURE IF EXISTS tenant_provision_drop_index;
DROP PROCEDURE IF EXISTS tenant_provision_add_unique_index;

DELIMITER $$
CREATE PROCEDURE tenant_provision_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE tenant_provision_statement FROM @statement_value;
    EXECUTE tenant_provision_statement;
    DEALLOCATE PREPARE tenant_provision_statement;
  END IF;
END$$

CREATE PROCEDURE tenant_provision_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema=DATABASE() AND table_name=table_name_value
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE INDEX `',index_name_value,'` ',definition_value);
    PREPARE tenant_provision_statement FROM @statement_value;
    EXECUTE tenant_provision_statement;
    DEALLOCATE PREPARE tenant_provision_statement;
  END IF;
END$$
DELIMITER ;

CALL tenant_provision_drop_index('game_game_providers','uk_game_providers_code');
CALL tenant_provision_add_unique_index('game_game_providers','uk_game_providers_scope_code','(`tenant_id`,`site_id`,`provider_code`)');

DROP PROCEDURE IF EXISTS tenant_provision_drop_index;
DROP PROCEDURE IF EXISTS tenant_provision_add_unique_index;
