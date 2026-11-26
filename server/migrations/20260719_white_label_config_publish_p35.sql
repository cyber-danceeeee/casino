DROP PROCEDURE IF EXISTS white_label_publish_add_column;
DROP PROCEDURE IF EXISTS white_label_publish_add_index;
DELIMITER $$
CREATE PROCEDURE white_label_publish_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE white_label_publish_statement FROM @statement_value; EXECUTE white_label_publish_statement; DEALLOCATE PREPARE white_label_publish_statement;
  END IF;
END$$
CREATE PROCEDURE white_label_publish_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE white_label_publish_statement FROM @statement_value; EXECUTE white_label_publish_statement; DEALLOCATE PREPARE white_label_publish_statement;
  END IF;
END$$
DELIMITER ;

CALL white_label_publish_add_column('game_saas_sites','site_name_draft','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `seo_config`');
CALL white_label_publish_add_column('game_saas_sites','country_code_draft','VARCHAR(8) NOT NULL DEFAULT '''' AFTER `site_name_draft`');
CALL white_label_publish_add_column('game_saas_sites','currency_code_draft','VARCHAR(16) NOT NULL DEFAULT '''' AFTER `country_code_draft`');
CALL white_label_publish_add_column('game_saas_sites','default_language_draft','VARCHAR(16) NOT NULL DEFAULT '''' AFTER `currency_code_draft`');
CALL white_label_publish_add_column('game_saas_sites','timezone_draft','VARCHAR(64) NOT NULL DEFAULT '''' AFTER `default_language_draft`');
CALL white_label_publish_add_column('game_saas_sites','brand_config_draft','JSON NULL AFTER `timezone_draft`');
CALL white_label_publish_add_column('game_saas_sites','theme_config_draft','JSON NULL AFTER `brand_config_draft`');
CALL white_label_publish_add_column('game_saas_sites','feature_config_draft','JSON NULL AFTER `theme_config_draft`');
CALL white_label_publish_add_column('game_saas_sites','maintenance_config_draft','JSON NULL AFTER `feature_config_draft`');
CALL white_label_publish_add_column('game_saas_sites','seo_config_draft','JSON NULL AFTER `maintenance_config_draft`');
CALL white_label_publish_add_column('game_saas_sites','config_status','VARCHAR(24) NOT NULL DEFAULT ''PUBLISHED'' AFTER `seo_config_draft`');
CALL white_label_publish_add_column('game_saas_sites','config_version','BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `config_status`');
CALL white_label_publish_add_column('game_saas_sites','config_published_at','DATETIME(3) NULL AFTER `config_version`');
CALL white_label_publish_add_column('game_saas_sites','config_published_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `config_published_at`');
CALL white_label_publish_add_column('game_saas_sites','config_draft_updated_at','DATETIME(3) NULL AFTER `config_published_by`');
CALL white_label_publish_add_column('game_saas_sites','config_draft_updated_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `config_draft_updated_at`');
CALL white_label_publish_add_index('game_saas_sites','idx_saas_site_config_version','(`tenant_id`,`id`,`config_version`,`config_status`)');

UPDATE game_saas_sites
SET site_name_draft = IF(site_name_draft='', site_name, site_name_draft),
    country_code_draft = IF(country_code_draft='', country_code, country_code_draft),
    currency_code_draft = IF(currency_code_draft='', currency_code, currency_code_draft),
    default_language_draft = IF(default_language_draft='', default_language, default_language_draft),
    timezone_draft = IF(timezone_draft='', timezone, timezone_draft),
    brand_config_draft = COALESCE(brand_config_draft, brand_config),
    theme_config_draft = COALESCE(theme_config_draft, theme_config),
    feature_config_draft = COALESCE(feature_config_draft, feature_config),
    maintenance_config_draft = COALESCE(maintenance_config_draft, maintenance_config),
    seo_config_draft = COALESCE(seo_config_draft, seo_config),
    config_status = IF(config_status='', 'PUBLISHED', config_status),
    config_version = IF(config_version=0, 1, config_version),
    config_published_at = COALESCE(config_published_at, updated_at, created_at)
WHERE deleted_at IS NULL;

CALL white_label_publish_add_column('game_lobby_sections','draft_payload','JSON NULL AFTER `config_json`');
CALL white_label_publish_add_column('game_lobby_sections','publish_status','VARCHAR(24) NOT NULL DEFAULT ''PUBLISHED'' AFTER `draft_payload`');
CALL white_label_publish_add_column('game_lobby_sections','config_version','BIGINT UNSIGNED NOT NULL DEFAULT 1 AFTER `publish_status`');
CALL white_label_publish_add_column('game_lobby_sections','published_at','DATETIME(3) NULL AFTER `config_version`');
CALL white_label_publish_add_column('game_lobby_sections','published_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `published_at`');
CALL white_label_publish_add_column('game_lobby_sections','draft_updated_at','DATETIME(3) NULL AFTER `published_by`');
CALL white_label_publish_add_column('game_lobby_sections','draft_updated_by','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `draft_updated_at`');
CALL white_label_publish_add_index('game_lobby_sections','idx_game_lobby_publish','(`tenant_id`,`site_id`,`publish_status`,`enabled`,`sort_order`)');

UPDATE game_lobby_sections
SET draft_payload = COALESCE(draft_payload, JSON_OBJECT(
      'sectionCode', section_code,
      'sectionType', section_type,
      'titleZh', title_zh,
      'titleEn', title_en,
      'displayStyle', display_style,
      'sourceType', source_type,
      'sourceRef', source_ref,
      'maxItems', max_items,
      'sortOrder', sort_order,
      'enabled', enabled,
      'supportedCountries', supported_countries,
      'supportedCurrencies', supported_currencies,
      'config', COALESCE(config_json, JSON_OBJECT())
    )),
    publish_status = IF(publish_status='', 'PUBLISHED', publish_status),
    config_version = IF(config_version=0, 1, config_version),
    published_at = COALESCE(published_at, updated_at, created_at)
WHERE deleted_at IS NULL;

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:site:config:preview','Preview site config','预览站点配置','Preview site config','system','API',1,1,NOW(3),NOW(3)),
('system:site:config:publish','Publish site config','发布站点配置','Publish site config','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('system:site:config:preview','system:site:config:publish')
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('system:site:config:preview','system:site:config:publish')
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

DROP PROCEDURE IF EXISTS white_label_publish_add_column;
DROP PROCEDURE IF EXISTS white_label_publish_add_index;
