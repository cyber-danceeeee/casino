DROP PROCEDURE IF EXISTS game_session_add_column;
DROP PROCEDURE IF EXISTS game_session_add_index;
DROP PROCEDURE IF EXISTS game_session_add_unique_index;
DELIMITER $$
CREATE PROCEDURE game_session_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE game_session_statement FROM @statement_value; EXECUTE game_session_statement; DEALLOCATE PREPARE game_session_statement;
  END IF;
END$$
CREATE PROCEDURE game_session_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE game_session_statement FROM @statement_value; EXECUTE game_session_statement; DEALLOCATE PREPARE game_session_statement;
  END IF;
END$$
CREATE PROCEDURE game_session_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE INDEX `',index_name_value,'` ',definition_value);
    PREPARE game_session_statement FROM @statement_value; EXECUTE game_session_statement; DEALLOCATE PREPARE game_session_statement;
  END IF;
END$$
DELIMITER ;

CALL game_session_add_column('game_game_launch_logs','session_id','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `launch_no`');
CALL game_session_add_column('game_game_launch_logs','country_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `country`');
CALL game_session_add_column('game_game_launch_logs','country_code_snapshot','VARCHAR(8) NOT NULL DEFAULT '''' AFTER `country_id`');
CALL game_session_add_column('game_game_launch_logs','currency_code_snapshot','VARCHAR(8) NOT NULL DEFAULT '''' AFTER `country_code_snapshot`');
CALL game_session_add_column('game_game_launch_logs','language_code_snapshot','VARCHAR(16) NOT NULL DEFAULT '''' AFTER `currency_code_snapshot`');
CALL game_session_add_column('game_game_launch_logs','session_status','VARCHAR(32) NOT NULL DEFAULT ''LAUNCH_REQUESTED'' AFTER `launch_status`');
CALL game_session_add_column('game_game_launch_logs','launch_time','DATETIME(3) NULL AFTER `session_status`');
CALL game_session_add_column('game_game_launch_logs','launch_success_time','DATETIME(3) NULL AFTER `launch_time`');
CALL game_session_add_column('game_game_launch_logs','exit_time','DATETIME(3) NULL AFTER `launch_success_time`');
CALL game_session_add_column('game_game_launch_logs','close_time','DATETIME(3) NULL AFTER `exit_time`');
CALL game_session_add_column('game_game_launch_logs','last_heartbeat_time','DATETIME(3) NULL AFTER `close_time`');
CALL game_session_add_column('game_game_launch_logs','duration_seconds','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `last_heartbeat_time`');
CALL game_session_add_column('game_game_launch_logs','exit_reason','VARCHAR(32) NOT NULL DEFAULT '''' AFTER `duration_seconds`');
CALL game_session_add_column('game_game_launch_logs','provider_launch_request','LONGTEXT NULL AFTER `provider_request`');
CALL game_session_add_column('game_game_launch_logs','provider_launch_response','LONGTEXT NULL AFTER `provider_response`');

UPDATE game_game_launch_logs
SET session_id=CASE
  WHEN session_id<>'' THEN session_id
  WHEN session_token<>'' THEN session_token
  ELSE CONCAT('legacy-',id)
END,
country_code_snapshot=IF(country_code_snapshot='',country,country_code_snapshot),
currency_code_snapshot=IF(currency_code_snapshot='',currency,currency_code_snapshot),
language_code_snapshot=IF(language_code_snapshot='',language,language_code_snapshot),
session_status=CASE
  WHEN session_status<>'' AND session_status<>'LAUNCH_REQUESTED' THEN session_status
  WHEN launch_status=2 THEN 'LAUNCH_SUCCESS'
  WHEN launch_status=3 THEN 'LAUNCH_FAILED'
  ELSE 'LAUNCH_REQUESTED'
END,
launch_time=IFNULL(launch_time,created_at),
launch_success_time=IF(launch_status=2 AND launch_success_time IS NULL,updated_at,launch_success_time),
close_time=IF(launch_status=3 AND close_time IS NULL,updated_at,close_time),
provider_launch_request=IF(provider_launch_request IS NULL,provider_request,provider_launch_request),
provider_launch_response=IF(provider_launch_response IS NULL,provider_response,provider_launch_response)
WHERE deleted_at IS NULL;

UPDATE game_game_launch_logs l
JOIN game_country_configs c ON c.country_code=l.country_code_snapshot AND c.deleted_at IS NULL
SET l.country_id=IF(l.country_id=0,c.id,l.country_id)
WHERE l.country_code_snapshot<>'' AND l.country_id=0;

CALL game_session_add_unique_index('game_game_launch_logs','uk_game_launch_session_id','(`session_id`)');
CALL game_session_add_index('game_game_launch_logs','idx_game_launch_user_created','(`user_id`,`created_at`)');
CALL game_session_add_index('game_game_launch_logs','idx_game_launch_provider_created','(`provider_id`,`created_at`)');
CALL game_session_add_index('game_game_launch_logs','idx_game_launch_game_created','(`game_id`,`created_at`)');
CALL game_session_add_index('game_game_launch_logs','idx_game_launch_status_created','(`session_status`,`created_at`)');
CALL game_session_add_index('game_game_launch_logs','idx_game_launch_country_created','(`country_id`,`created_at`)');

DROP PROCEDURE IF EXISTS game_session_add_column;
DROP PROCEDURE IF EXISTS game_session_add_index;
DROP PROCEDURE IF EXISTS game_session_add_unique_index;
