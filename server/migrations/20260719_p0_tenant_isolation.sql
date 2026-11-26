DROP PROCEDURE IF EXISTS p0_tenant_add_column;
DROP PROCEDURE IF EXISTS p0_tenant_add_index;
DROP PROCEDURE IF EXISTS p0_tenant_add_unique_index;
DROP PROCEDURE IF EXISTS p0_tenant_drop_index;
DELIMITER $$
CREATE PROCEDURE p0_tenant_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE p0_tenant_statement FROM @statement_value; EXECUTE p0_tenant_statement; DEALLOCATE PREPARE p0_tenant_statement;
  END IF;
END$$
CREATE PROCEDURE p0_tenant_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE p0_tenant_statement FROM @statement_value; EXECUTE p0_tenant_statement; DEALLOCATE PREPARE p0_tenant_statement;
  END IF;
END$$
CREATE PROCEDURE p0_tenant_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE INDEX `',index_name_value,'` ',definition_value);
    PREPARE p0_tenant_statement FROM @statement_value; EXECUTE p0_tenant_statement; DEALLOCATE PREPARE p0_tenant_statement;
  END IF;
END$$
CREATE PROCEDURE p0_tenant_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE p0_tenant_statement FROM @statement_value; EXECUTE p0_tenant_statement; DEALLOCATE PREPARE p0_tenant_statement;
  END IF;
END$$
DELIMITER ;

CALL p0_tenant_add_column('game_commission_plans','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_commission_plans','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agents','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agents','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_channels','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_channels','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_user_agent_attributions','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_user_agent_attributions','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_commission_records','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_commission_records','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_settlements','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_settlements','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_settlement_items','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_settlement_items','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_wallets','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_wallets','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_wallet_transactions','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_wallet_transactions','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_hierarchy_history','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_hierarchy_history','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_user_agent_attribution_history','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_user_agent_attribution_history','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_agent_commission_rules','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_agent_commission_rules','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');

CALL p0_tenant_add_column('game_risk_rules','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_rules','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_risk_reviews','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_reviews','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_risk_review_histories','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_review_histories','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_risk_blacklists','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_blacklists','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_risk_blacklist_user_relations','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_blacklist_user_relations','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_risk_blacklist_histories','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_risk_blacklist_histories','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');

CALL p0_tenant_add_column('game_daily_financial_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_financial_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_game_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_game_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_provider_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_provider_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_user_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_user_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_channel_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_channel_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_agent_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_agent_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_daily_promotion_summaries','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_daily_promotion_summaries','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');
CALL p0_tenant_add_column('game_report_export_tasks','tenant_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `id`');
CALL p0_tenant_add_column('game_report_export_tasks','site_id','BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER `tenant_id`');

CALL p0_tenant_add_index('game_commission_plans','idx_commission_plan_tenant','(`tenant_id`,`site_id`,`enabled`,`deleted_at`)');
CALL p0_tenant_add_index('game_agents','idx_agents_tenant','(`tenant_id`,`site_id`,`status`,`deleted_at`)');
CALL p0_tenant_add_index('game_agent_channels','idx_agent_channels_tenant','(`tenant_id`,`site_id`,`enabled`,`deleted_at`)');
CALL p0_tenant_add_index('game_user_agent_attributions','idx_agent_attr_tenant','(`tenant_id`,`site_id`,`user_id`,`agent_id`,`active`,`deleted_at`)');
CALL p0_tenant_add_index('game_user_agent_attribution_history','idx_agent_attr_history_tenant','(`tenant_id`,`site_id`,`user_id`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_agent_commission_rules','idx_agent_commission_rules_tenant','(`tenant_id`,`site_id`,`level_no`,`deleted_at`)');
CALL p0_tenant_add_index('game_commission_records','idx_commission_records_tenant','(`tenant_id`,`site_id`,`agent_id`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_agent_settlements','idx_agent_settlements_tenant','(`tenant_id`,`site_id`,`agent_id`,`status`,`deleted_at`)');
CALL p0_tenant_add_index('game_agent_wallets','idx_agent_wallets_tenant','(`tenant_id`,`site_id`,`agent_id`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_agent_wallet_transactions','idx_agent_wallet_tx_tenant','(`tenant_id`,`site_id`,`agent_id`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_report_export_tasks','idx_report_export_tenant','(`tenant_id`,`site_id`,`status`,`created_at`,`deleted_at`)');

CALL p0_tenant_add_index('game_risk_rules','idx_risk_rules_tenant','(`tenant_id`,`site_id`,`enabled`,`deleted_at`)');
CALL p0_tenant_add_index('game_risk_reviews','idx_risk_reviews_tenant','(`tenant_id`,`site_id`,`status`,`risk_level`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_risk_review_histories','idx_risk_review_history_tenant','(`tenant_id`,`site_id`,`review_id`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_risk_blacklists','idx_risk_blacklists_tenant','(`tenant_id`,`site_id`,`blacklist_type`,`status`,`deleted_at`)');
CALL p0_tenant_add_index('game_risk_blacklist_user_relations','idx_risk_blacklist_rel_tenant','(`tenant_id`,`site_id`,`user_id`,`created_at`,`deleted_at`)');
CALL p0_tenant_add_index('game_risk_blacklist_histories','idx_risk_blacklist_history_tenant','(`tenant_id`,`site_id`,`blacklist_id`,`created_at`,`deleted_at`)');

CALL p0_tenant_add_index('game_daily_financial_summaries','idx_daily_financial_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_game_summaries','idx_daily_game_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_provider_summaries','idx_daily_provider_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_user_summaries','idx_daily_user_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_channel_summaries','idx_daily_channel_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_agent_summaries','idx_daily_agent_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');
CALL p0_tenant_add_index('game_daily_promotion_summaries','idx_daily_promotion_tenant','(`tenant_id`,`site_id`,`business_date`,`currency`,`deleted_at`)');

CALL p0_tenant_drop_index('game_daily_financial_summaries','uk_daily_financial_dimension');
CALL p0_tenant_drop_index('game_daily_game_summaries','uk_daily_game_dimension');
CALL p0_tenant_drop_index('game_daily_provider_summaries','uk_daily_provider_dimension');
CALL p0_tenant_drop_index('game_daily_user_summaries','uk_daily_user_dimension');
CALL p0_tenant_drop_index('game_daily_channel_summaries','uk_daily_channel_dimension');
CALL p0_tenant_drop_index('game_daily_agent_summaries','uk_daily_agent_dimension');
CALL p0_tenant_drop_index('game_daily_promotion_summaries','uk_daily_promotion_dimension');

CALL p0_tenant_add_unique_index('game_daily_financial_summaries','uk_daily_financial_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`country`,`currency`,`payment_channel`,`user_channel`)');
CALL p0_tenant_add_unique_index('game_daily_game_summaries','uk_daily_game_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`game_id`,`country`,`currency`,`device`)');
CALL p0_tenant_add_unique_index('game_daily_provider_summaries','uk_daily_provider_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`provider_code`,`country`,`currency`,`game_type`)');
CALL p0_tenant_add_unique_index('game_daily_user_summaries','uk_daily_user_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`user_id`,`currency`)');
CALL p0_tenant_add_unique_index('game_daily_channel_summaries','uk_daily_channel_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`channel_id`,`currency`)');
CALL p0_tenant_add_unique_index('game_daily_agent_summaries','uk_daily_agent_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`agent_id`,`currency`)');
CALL p0_tenant_add_unique_index('game_daily_promotion_summaries','uk_daily_promotion_tenant_dimension','(`tenant_id`,`site_id`,`business_date`,`timezone`,`promotion_id`,`country`,`currency`)');

DROP PROCEDURE IF EXISTS p0_tenant_add_column;
DROP PROCEDURE IF EXISTS p0_tenant_add_index;
DROP PROCEDURE IF EXISTS p0_tenant_add_unique_index;
DROP PROCEDURE IF EXISTS p0_tenant_drop_index;
