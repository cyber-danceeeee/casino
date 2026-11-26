DROP PROCEDURE IF EXISTS game_config_publish_add_column;
DROP PROCEDURE IF EXISTS game_config_publish_add_index;
DELIMITER $$
CREATE PROCEDURE game_config_publish_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE game_config_publish_statement FROM @statement_value; EXECUTE game_config_publish_statement; DEALLOCATE PREPARE game_config_publish_statement;
  END IF;
END$$
CREATE PROCEDURE game_config_publish_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE game_config_publish_statement FROM @statement_value; EXECUTE game_config_publish_statement; DEALLOCATE PREPARE game_config_publish_statement;
  END IF;
END$$
DELIMITER ;

CALL game_config_publish_add_column('game_games','platform_enabled_draft','TINYINT(1) NULL AFTER `platform_enabled`');
CALL game_config_publish_add_column('game_games','is_hot_draft','TINYINT(1) NULL AFTER `is_hot`');
CALL game_config_publish_add_column('game_games','is_featured_draft','TINYINT(1) NULL AFTER `is_featured`');
CALL game_config_publish_add_column('game_games','is_new_draft','TINYINT(1) NULL AFTER `is_new`');
CALL game_config_publish_add_column('game_games','sort_draft','INT NULL AFTER `sort`');
CALL game_config_publish_add_column('game_games','status_draft','TINYINT(4) NULL AFTER `status`');
CALL game_config_publish_add_column('game_games','maintenance_message_draft','VARCHAR(500) NULL AFTER `maintenance_message`');
CALL game_config_publish_add_column('game_games','category_ids_draft','JSON NULL AFTER `platform_category_code`');
CALL game_config_publish_add_column('game_games','config_status','VARCHAR(24) NOT NULL DEFAULT ''PUBLISHED'' AFTER `category_ids_draft`');
CALL game_config_publish_add_column('game_games','config_version','BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `config_status`');
CALL game_config_publish_add_column('game_games','config_published_at','DATETIME(3) NULL AFTER `config_version`');
CALL game_config_publish_add_column('game_games','config_published_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `config_published_at`');
CALL game_config_publish_add_column('game_games','draft_updated_at','DATETIME(3) NULL AFTER `config_published_by`');
CALL game_config_publish_add_column('game_games','draft_updated_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `draft_updated_at`');
CALL game_config_publish_add_index('game_games','idx_game_config_publish','(`tenant_id`,`site_id`,`config_status`,`platform_enabled`,`sort`)');

UPDATE game_games
SET platform_enabled_draft = COALESCE(platform_enabled_draft, platform_enabled),
    is_hot_draft = COALESCE(is_hot_draft, is_hot),
    is_featured_draft = COALESCE(is_featured_draft, is_featured),
    is_new_draft = COALESCE(is_new_draft, is_new),
    sort_draft = COALESCE(sort_draft, sort),
    status_draft = COALESCE(status_draft, status),
    maintenance_message_draft = COALESCE(maintenance_message_draft, maintenance_message),
    category_ids_draft = COALESCE(category_ids_draft, (
      SELECT JSON_ARRAYAGG(JSON_OBJECT('categoryId', x.category_id, 'sort', x.sort_order))
      FROM (
        SELECT r.category_id, r.sort AS sort_order
        FROM game_game_category_relations r
        WHERE r.game_id = game_games.id AND r.deleted_at IS NULL
        ORDER BY r.sort ASC, r.id ASC
      ) x
    )),
    config_status = IF(config_status='', 'PUBLISHED', config_status),
    config_version = IF(config_version=0, 1, config_version),
    config_published_at = COALESCE(config_published_at, updated_at, created_at)
WHERE deleted_at IS NULL;

DROP PROCEDURE IF EXISTS game_config_publish_add_column;
DROP PROCEDURE IF EXISTS game_config_publish_add_index;
