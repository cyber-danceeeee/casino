DROP PROCEDURE IF EXISTS payment_channel_add_column;
DROP PROCEDURE IF EXISTS payment_channel_drop_index;
DROP PROCEDURE IF EXISTS payment_channel_add_index;

DELIMITER $$
CREATE PROCEDURE payment_channel_add_column(IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='game_payment_channels' AND column_name=column_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_payment_channels ADD COLUMN ',definition_value);
    PREPARE payment_channel_statement FROM @statement_value;
    EXECUTE payment_channel_statement;
    DEALLOCATE PREPARE payment_channel_statement;
  END IF;
END$$

CREATE PROCEDURE payment_channel_drop_index(IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='game_payment_channels' AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_payment_channels DROP INDEX `',index_name_value,'`');
    PREPARE payment_channel_statement FROM @statement_value;
    EXECUTE payment_channel_statement;
    DEALLOCATE PREPARE payment_channel_statement;
  END IF;
END$$

CREATE PROCEDURE payment_channel_add_index(IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='game_payment_channels' AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_payment_channels ADD ',definition_value);
    PREPARE payment_channel_statement FROM @statement_value;
    EXECUTE payment_channel_statement;
    DEALLOCATE PREPARE payment_channel_statement;
  END IF;
END$$
DELIMITER ;

CALL payment_channel_add_column('country_code','`country_code` VARCHAR(8) NOT NULL DEFAULT '''' AFTER `name`');
CALL payment_channel_add_column('currency_code','`currency_code` VARCHAR(8) NOT NULL DEFAULT '''' AFTER `country_code`');

UPDATE game_payment_channels c
LEFT JOIN game_saas_sites s ON s.id=c.site_id AND s.tenant_id=c.tenant_id AND s.deleted_at IS NULL
LEFT JOIN game_saas_tenants t ON t.id=c.tenant_id AND t.deleted_at IS NULL
SET c.country_code=COALESCE(NULLIF(UPPER(s.country_code),''),NULLIF(UPPER(t.country_code),''),'PK'),
    c.currency_code=COALESCE(NULLIF(UPPER(s.currency_code),''),NULLIF(UPPER(t.default_currency),''),'PKR')
WHERE c.country_code='' OR c.currency_code='';

CALL payment_channel_drop_index('uk_game_payment_channels_code');
CALL payment_channel_add_index(
  'uk_payment_channel_scope_region_code',
  'UNIQUE INDEX `uk_payment_channel_scope_region_code` (`tenant_id`,`site_id`,`country_code`,`currency_code`,`channel_code`)'
);
CALL payment_channel_add_index(
  'idx_payment_channel_region_status',
  'INDEX `idx_payment_channel_region_status` (`tenant_id`,`site_id`,`country_code`,`currency_code`,`status`,`sort`)'
);

DROP PROCEDURE IF EXISTS payment_channel_add_column;
DROP PROCEDURE IF EXISTS payment_channel_drop_index;
DROP PROCEDURE IF EXISTS payment_channel_add_index;
