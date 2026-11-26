DROP PROCEDURE IF EXISTS bonus_audit_add_index;
DELIMITER $$
CREATE PROCEDURE bonus_audit_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema=DATABASE() AND table_name=table_name_value
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE bonus_audit_statement FROM @statement_value;
    EXECUTE bonus_audit_statement;
    DEALLOCATE PREPARE bonus_audit_statement;
  END IF;
END$$
DELIMITER ;

CALL bonus_audit_add_index(
  'game_bet_records',
  'idx_game_bet_records_audit_progress',
  '(`tenant_id`,`site_id`,`user_id`,`currency`,`bet_status`,`rollback_status`,`bet_at`)'
);

DROP PROCEDURE IF EXISTS bonus_audit_add_index;
