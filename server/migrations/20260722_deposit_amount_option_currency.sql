DROP PROCEDURE IF EXISTS deposit_amount_option_currency_add_column;
DROP PROCEDURE IF EXISTS deposit_amount_option_currency_add_index;
DELIMITER $$
CREATE PROCEDURE deposit_amount_option_currency_add_column(IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='game_deposit_amount_options' AND column_name=column_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_deposit_amount_options ADD COLUMN ',definition_value);
    PREPARE option_currency_statement FROM @statement_value;
    EXECUTE option_currency_statement;
    DEALLOCATE PREPARE option_currency_statement;
  END IF;
END$$
CREATE PROCEDURE deposit_amount_option_currency_add_index(IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='game_deposit_amount_options' AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE game_deposit_amount_options ADD ',definition_value);
    PREPARE option_currency_statement FROM @statement_value;
    EXECUTE option_currency_statement;
    DEALLOCATE PREPARE option_currency_statement;
  END IF;
END$$
DELIMITER ;

CALL deposit_amount_option_currency_add_column('currency_code','`currency_code` VARCHAR(8) NOT NULL DEFAULT ''PKR'' AFTER `option_code`');

UPDATE game_deposit_amount_options o
LEFT JOIN game_saas_sites s ON s.id=o.site_id AND s.tenant_id=o.tenant_id AND s.deleted_at IS NULL
LEFT JOIN game_saas_tenants t ON t.id=o.tenant_id AND t.deleted_at IS NULL
LEFT JOIN game_country_configs c ON c.country_code=COALESCE(NULLIF(s.country_code,''),NULLIF(t.country_code,'')) AND c.status=1 AND c.deleted_at IS NULL
SET o.currency_code=COALESCE(NULLIF(UPPER(s.currency_code),''),NULLIF(UPPER(t.default_currency),''),NULLIF(UPPER(c.currency_code),''),'PKR')
WHERE o.tenant_id<>0
  AND (o.currency_code='' OR o.currency_code='PKR');

CALL deposit_amount_option_currency_add_index('idx_deposit_amount_option_scope_currency','INDEX `idx_deposit_amount_option_scope_currency` (`tenant_id`,`site_id`,`currency_code`,`status`,`deleted_at`)');

DROP PROCEDURE IF EXISTS deposit_amount_option_currency_add_column;
DROP PROCEDURE IF EXISTS deposit_amount_option_currency_add_index;
