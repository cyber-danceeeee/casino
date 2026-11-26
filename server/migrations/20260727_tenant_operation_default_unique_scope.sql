-- Platform defaults are copied into tenant/site-owned rows. Their legacy
-- global unique keys must therefore be scoped to tenant and site.

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_social_links' AND index_name='uk_game_social_links_platform');
SET @sql := IF(@idx>0,'ALTER TABLE game_social_links DROP INDEX uk_game_social_links_platform','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_social_links' AND index_name='uk_game_social_links_scope_platform');
SET @sql := IF(@idx=0,'ALTER TABLE game_social_links ADD UNIQUE KEY uk_game_social_links_scope_platform (tenant_id,site_id,platform_code)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_vip_levels' AND index_name='uk_game_vip_levels_level_no');
SET @sql := IF(@idx>0,'ALTER TABLE game_vip_levels DROP INDEX uk_game_vip_levels_level_no','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_vip_levels' AND index_name='uk_game_vip_levels_scope_level_no');
SET @sql := IF(@idx=0,'ALTER TABLE game_vip_levels ADD UNIQUE KEY uk_game_vip_levels_scope_level_no (tenant_id,site_id,level_no)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_deposit_vouchers' AND index_name='uk_game_deposit_vouchers_code');
SET @sql := IF(@idx>0,'ALTER TABLE game_deposit_vouchers DROP INDEX uk_game_deposit_vouchers_code','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_deposit_vouchers' AND index_name='uk_game_deposit_vouchers_scope_code');
SET @sql := IF(@idx=0,'ALTER TABLE game_deposit_vouchers ADD UNIQUE KEY uk_game_deposit_vouchers_scope_code (tenant_id,site_id,voucher_code)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_deposit_amount_options' AND index_name='uk_game_deposit_amount_options_code');
SET @sql := IF(@idx>0,'ALTER TABLE game_deposit_amount_options DROP INDEX uk_game_deposit_amount_options_code','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_deposit_amount_options' AND index_name='uk_game_deposit_amount_options_scope_code');
SET @sql := IF(@idx=0,'ALTER TABLE game_deposit_amount_options ADD UNIQUE KEY uk_game_deposit_amount_options_scope_code (tenant_id,site_id,currency_code,option_code)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_activity_events' AND index_name='uk_game_activity_events_code');
SET @sql := IF(@idx>0,'ALTER TABLE game_activity_events DROP INDEX uk_game_activity_events_code','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_activity_events' AND index_name='uk_game_activity_events_scope_code');
SET @sql := IF(@idx=0,'ALTER TABLE game_activity_events ADD UNIQUE KEY uk_game_activity_events_scope_code (tenant_id,site_id,activity_code)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_activity_events' AND index_name='uk_game_activity_events_slug');
SET @sql := IF(@idx>0,'ALTER TABLE game_activity_events DROP INDEX uk_game_activity_events_slug','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @idx := (SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='game_activity_events' AND index_name='uk_game_activity_events_scope_slug');
SET @sql := IF(@idx=0,'ALTER TABLE game_activity_events ADD UNIQUE KEY uk_game_activity_events_scope_slug (tenant_id,site_id,slug)','SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
