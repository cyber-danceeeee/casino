DROP PROCEDURE IF EXISTS game_user_phone_drop_index;
DROP PROCEDURE IF EXISTS game_user_phone_add_unique_index;

DELIMITER $$
CREATE PROCEDURE game_user_phone_drop_index(IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='game_users'
      AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_users DROP INDEX `',index_name_value,'`');
    PREPARE game_user_phone_statement FROM @statement_value;
    EXECUTE game_user_phone_statement;
    DEALLOCATE PREPARE game_user_phone_statement;
  END IF;
END$$

CREATE PROCEDURE game_user_phone_add_unique_index(IN index_name_value VARCHAR(64), IN columns_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='game_users'
      AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT(
      'ALTER TABLE game_users ADD UNIQUE INDEX `',index_name_value,'` ',columns_value
    );
    PREPARE game_user_phone_statement FROM @statement_value;
    EXECUTE game_user_phone_statement;
    DEALLOCATE PREPARE game_user_phone_statement;
  END IF;
END$$
DELIMITER ;

CALL game_user_phone_drop_index('uk_game_users_phone');
CALL game_user_phone_add_unique_index(
  'uk_game_users_scope_phone',
  '(`tenant_id`,`site_id`,`phone_country_code`,`phone`)'
);

DROP PROCEDURE IF EXISTS game_user_phone_drop_index;
DROP PROCEDURE IF EXISTS game_user_phone_add_unique_index;
