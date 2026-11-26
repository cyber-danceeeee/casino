DROP PROCEDURE IF EXISTS h5_entry_add_column;
DROP PROCEDURE IF EXISTS h5_entry_add_index;
DROP PROCEDURE IF EXISTS h5_entry_add_unique_index;
DROP PROCEDURE IF EXISTS h5_entry_drop_index;
DELIMITER $$
CREATE PROCEDURE h5_entry_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE h5_entry_statement FROM @statement_value;
    EXECUTE h5_entry_statement;
    DEALLOCATE PREPARE h5_entry_statement;
  END IF;
END$$
CREATE PROCEDURE h5_entry_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE h5_entry_statement FROM @statement_value;
    EXECUTE h5_entry_statement;
    DEALLOCATE PREPARE h5_entry_statement;
  END IF;
END$$
CREATE PROCEDURE h5_entry_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE KEY `',index_name_value,'` ',definition_value);
    PREPARE h5_entry_statement FROM @statement_value;
    EXECUTE h5_entry_statement;
    DEALLOCATE PREPARE h5_entry_statement;
  END IF;
END$$
CREATE PROCEDURE h5_entry_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE h5_entry_statement FROM @statement_value;
    EXECUTE h5_entry_statement;
    DEALLOCATE PREPARE h5_entry_statement;
  END IF;
END$$
DELIMITER ;

CALL h5_entry_add_column('game_h5_navigation_items','labels_json','TEXT NULL AFTER `label_en`');
CALL h5_entry_add_column('game_h5_navigation_items','icon_asset_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `labels_json`');
CALL h5_entry_add_column('game_h5_navigation_items','lottie_asset_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `icon_asset_id`');
CALL h5_entry_add_column('game_h5_navigation_items','sound_asset_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `lottie_asset_id`');
CALL h5_entry_add_column('game_h5_navigation_items','icon_asset_id_draft','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `sound_asset_id`');
CALL h5_entry_add_column('game_h5_navigation_items','lottie_asset_id_draft','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `icon_asset_id_draft`');
CALL h5_entry_add_column('game_h5_navigation_items','sound_asset_id_draft','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `lottie_asset_id_draft`');
CALL h5_entry_add_column('game_h5_navigation_items','display_position','VARCHAR(32) NOT NULL DEFAULT ''BOTTOM_NAV'' AFTER `sound_asset_id_draft`');
CALL h5_entry_add_column('game_h5_navigation_items','action_target','VARCHAR(500) NOT NULL DEFAULT '''' AFTER `action_type`');
CALL h5_entry_add_column('game_h5_navigation_items','badge_type','VARCHAR(16) NOT NULL DEFAULT ''NONE'' AFTER `sort_order`');
CALL h5_entry_add_column('game_h5_navigation_items','visibility_scope','VARCHAR(16) NOT NULL DEFAULT ''ALL'' AFTER `badge_type`');
CALL h5_entry_add_column('game_h5_navigation_items','country_codes','TEXT NULL AFTER `visibility_scope`');
CALL h5_entry_add_column('game_h5_navigation_items','language_codes','TEXT NULL AFTER `country_codes`');
CALL h5_entry_add_column('game_h5_navigation_items','start_at','DATETIME(3) NULL AFTER `language_codes`');
CALL h5_entry_add_column('game_h5_navigation_items','end_at','DATETIME(3) NULL AFTER `start_at`');
CALL h5_entry_add_column('game_h5_navigation_items','draft_payload','LONGTEXT NULL AFTER `enabled`');
CALL h5_entry_add_column('game_h5_navigation_items','published_payload','LONGTEXT NULL AFTER `draft_payload`');
CALL h5_entry_add_column('game_h5_navigation_items','publish_status','VARCHAR(16) NOT NULL DEFAULT ''DRAFT'' AFTER `published_payload`');
CALL h5_entry_add_column('game_h5_navigation_items','config_version','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `publish_status`');
CALL h5_entry_add_column('game_h5_navigation_items','created_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `config_version`');
CALL h5_entry_add_column('game_h5_navigation_items','updated_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `created_by`');
CALL h5_entry_add_column('game_h5_navigation_items','draft_updated_at','DATETIME(3) NULL AFTER `updated_by`');
CALL h5_entry_add_column('game_h5_navigation_items','published_at','DATETIME(3) NULL AFTER `draft_updated_at`');
CALL h5_entry_add_column('game_h5_navigation_items','published_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `published_at`');

ALTER TABLE game_h5_navigation_items
  MODIFY COLUMN action_type VARCHAR(32) NOT NULL DEFAULT 'OPEN_INTERNAL_PAGE';

CALL h5_entry_drop_index('game_h5_navigation_items','uk_h5_navigation_code');
CALL h5_entry_add_unique_index('game_h5_navigation_items','uk_h5_navigation_scope_code','(`tenant_id`,`site_id`,`menu_code`)');
CALL h5_entry_add_index('game_h5_navigation_items','idx_h5_navigation_public','(`tenant_id`,`site_id`,`display_position`,`enabled`,`sort_order`)');
CALL h5_entry_add_index('game_h5_navigation_items','idx_h5_navigation_publish','(`tenant_id`,`site_id`,`publish_status`,`updated_at`)');

UPDATE game_h5_navigation_items
SET action_target=CASE
      WHEN action_type='PAGE' THEN page_route
      WHEN action_type='GAME' THEN provider_game_code
      ELSE action_target
    END,
    action_type=CASE
      WHEN action_type='PAGE' THEN 'OPEN_INTERNAL_PAGE'
      WHEN action_type='GAME' THEN 'OPEN_GAME'
      ELSE action_type
    END,
    display_position=COALESCE(NULLIF(display_position,''),'BOTTOM_NAV'),
    badge_type=COALESCE(NULLIF(badge_type,''),'NONE'),
    visibility_scope=COALESCE(NULLIF(visibility_scope,''),'ALL'),
    country_codes=COALESCE(country_codes,'[]'),
    language_codes=COALESCE(language_codes,'[]'),
    publish_status='PUBLISHED',
    config_version=CASE WHEN config_version=0 THEN 1 ELSE config_version END,
    published_at=COALESCE(published_at,updated_at,created_at,NOW(3))
WHERE deleted_at IS NULL;

INSERT IGNORE INTO game_h5_navigation_items
(tenant_id,site_id,menu_code,label_zh,label_en,labels_json,display_position,action_type,action_target,
 page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,
 country_codes,language_codes,enabled,publish_status,config_version,created_at,updated_at,published_at)
SELECT s.tenant_id,s.id,g.menu_code,g.label_zh,g.label_en,g.labels_json,'BOTTOM_NAV',g.action_type,g.action_target,
       g.page_route,g.provider_code,g.game_id,g.provider_game_code,g.game_name,g.sort_order,'NONE','ALL',
       '[]','[]',g.enabled,'PUBLISHED',1,NOW(3),NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_h5_navigation_items g ON g.tenant_id=0 AND g.site_id=0 AND g.deleted_at IS NULL
WHERE s.deleted_at IS NULL;

DROP PROCEDURE h5_entry_add_column;
DROP PROCEDURE h5_entry_add_index;
DROP PROCEDURE h5_entry_add_unique_index;
DROP PROCEDURE h5_entry_drop_index;
