DROP PROCEDURE IF EXISTS tenant_operation_defaults_drop_index;
DROP PROCEDURE IF EXISTS tenant_operation_defaults_add_index;
DELIMITER $$
CREATE PROCEDURE tenant_operation_defaults_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE tenant_operation_defaults_statement FROM @statement_value;
    EXECUTE tenant_operation_defaults_statement;
    DEALLOCATE PREPARE tenant_operation_defaults_statement;
  END IF;
END$$
CREATE PROCEDURE tenant_operation_defaults_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD ',definition_value);
    PREPARE tenant_operation_defaults_statement FROM @statement_value;
    EXECUTE tenant_operation_defaults_statement;
    DEALLOCATE PREPARE tenant_operation_defaults_statement;
  END IF;
END$$
DELIMITER ;

CALL tenant_operation_defaults_drop_index('game_operation_music','uk_game_operation_music_code');
CALL tenant_operation_defaults_add_index('game_operation_music','idx_game_operation_music_scope_code','INDEX `idx_game_operation_music_scope_code` (`tenant_id`,`site_id`,`code`,`deleted_at`)');

UPDATE game_h5_navigation_items
SET platform_source_id=0, updated_at=NOW(3)
WHERE tenant_id<>0 AND platform_source_id<>0 AND deleted_at IS NULL;

INSERT INTO game_site_banners
(tenant_id,site_id,name,title,subtitle,desktop_asset_id,mobile_asset_id,desktop_asset_id_draft,mobile_asset_id_draft,
 action_type,action_target,sort_order,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,
 draft_payload,published_payload,publish_status,config_version,created_by,updated_by,draft_updated_at,published_at,published_by,created_at,updated_at)
SELECT s.tenant_id,s.id,
       COALESCE(NULLIF(b.banner_code,''),NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,CONCAT('banner-',b.id)),
       COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,b.banner_code),
       '',0,0,0,0,
       CASE UPPER(b.target_type)
         WHEN 'EXTERNAL_URL' THEN 'OPEN_EXTERNAL_URL'
         WHEN 'INTERNAL_ROUTE' THEN 'OPEN_INTERNAL_PAGE'
         WHEN 'GAME' THEN 'OPEN_GAME'
         WHEN 'PROMOTION' THEN 'OPEN_PROMOTION'
         ELSE 'NONE'
       END,
       CASE
         WHEN UPPER(b.target_type) IN ('EXTERNAL_URL','INTERNAL_ROUTE','GAME','PROMOTION') THEN COALESCE(b.target_value,'')
         ELSE ''
       END,
       b.sort_order,
       COALESCE(NULLIF(s.default_language,''),'ALL'),
       COALESCE(NULLIF(s.country_code,''),'ALL'),
       COALESCE(NULLIF(s.currency_code,''),'ALL'),
       'ALL',b.start_at,b.end_at,IF(b.enabled=1 OR b.status=1,1,0),
       JSON_OBJECT(
         'name',COALESCE(NULLIF(b.banner_code,''),NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,CONCAT('banner-',b.id)),
         'title',COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,b.banner_code),
         'subtitle','','desktopAssetID',0,'mobileAssetID',0,
         'desktopImageUrl',COALESCE(b.image_url,''),'mobileImageUrl',COALESCE(NULLIF(b.mobile_image_url,''),b.image_url,''),
         'actionType',CASE UPPER(b.target_type)
           WHEN 'EXTERNAL_URL' THEN 'OPEN_EXTERNAL_URL'
           WHEN 'INTERNAL_ROUTE' THEN 'OPEN_INTERNAL_PAGE'
           WHEN 'GAME' THEN 'OPEN_GAME'
           WHEN 'PROMOTION' THEN 'OPEN_PROMOTION'
           ELSE 'NONE'
         END,
         'actionTarget',CASE WHEN UPPER(b.target_type) IN ('EXTERNAL_URL','INTERNAL_ROUTE','GAME','PROMOTION') THEN COALESCE(b.target_value,'') ELSE '' END,
         'sortOrder',b.sort_order,'languageCode',COALESCE(NULLIF(s.default_language,''),'ALL'),
         'countryCode',COALESCE(NULLIF(s.country_code,''),'ALL'),'currencyCode',COALESCE(NULLIF(s.currency_code,''),'ALL'),
         'visibilityScope','ALL','startAt',b.start_at,'endAt',b.end_at,'enabled',IF(b.enabled=1 OR b.status=1,TRUE,FALSE)
       ),
       JSON_OBJECT(
         'name',COALESCE(NULLIF(b.banner_code,''),NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,CONCAT('banner-',b.id)),
         'title',COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,b.banner_code),
         'subtitle','','desktopAssetID',0,'mobileAssetID',0,
         'desktopImageUrl',COALESCE(b.image_url,''),'mobileImageUrl',COALESCE(NULLIF(b.mobile_image_url,''),b.image_url,''),
         'actionType',CASE UPPER(b.target_type)
           WHEN 'EXTERNAL_URL' THEN 'OPEN_EXTERNAL_URL'
           WHEN 'INTERNAL_ROUTE' THEN 'OPEN_INTERNAL_PAGE'
           WHEN 'GAME' THEN 'OPEN_GAME'
           WHEN 'PROMOTION' THEN 'OPEN_PROMOTION'
           ELSE 'NONE'
         END,
         'actionTarget',CASE WHEN UPPER(b.target_type) IN ('EXTERNAL_URL','INTERNAL_ROUTE','GAME','PROMOTION') THEN COALESCE(b.target_value,'') ELSE '' END,
         'sortOrder',b.sort_order,'languageCode',COALESCE(NULLIF(s.default_language,''),'ALL'),
         'countryCode',COALESCE(NULLIF(s.country_code,''),'ALL'),'currencyCode',COALESCE(NULLIF(s.currency_code,''),'ALL'),
         'visibilityScope','ALL','startAt',b.start_at,'endAt',b.end_at,'enabled',IF(b.enabled=1 OR b.status=1,TRUE,FALSE)
       ),
       'PUBLISHED',1,0,0,NOW(3),NOW(3),0,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_banners b ON b.tenant_id=0 AND b.site_id=0 AND b.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_site_banners sb
    WHERE sb.tenant_id=s.tenant_id AND sb.site_id=s.id
      AND sb.name=COALESCE(NULLIF(b.banner_code,''),NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.title,CONCAT('banner-',b.id))
      AND sb.deleted_at IS NULL
  );

INSERT INTO game_social_links
(tenant_id,site_id,platform_code,label,icon_url,share_url,share_title,share_text,share_description,sort,status,created_at,updated_at)
SELECT s.tenant_id,s.id,p.platform_code,p.label,p.icon_url,p.share_url,p.share_title,p.share_text,p.share_description,p.sort,p.status,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_social_links p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_social_links t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.platform_code=p.platform_code AND t.deleted_at IS NULL
  );

INSERT INTO game_operation_music
(tenant_id,site_id,code,name_zh,name_en,audio_url,scene,loop_play,volume_percent,sort_order,enabled,start_at,end_at,created_at,updated_at)
SELECT s.tenant_id,s.id,p.code,p.name_zh,p.name_en,p.audio_url,p.scene,p.loop_play,p.volume_percent,p.sort_order,p.enabled,p.start_at,p.end_at,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_operation_music p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_operation_music t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.code=p.code AND t.deleted_at IS NULL
  );

INSERT INTO game_vip_levels
(tenant_id,site_id,level_no,code,name,name_zh,name_en,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,
 required_experience,required_deposit,required_valid_bet,withdrawal_limit,cashback_rate,birthday_reward,sort_order,status,enabled,created_at,updated_at)
SELECT s.tenant_id,s.id,p.level_no,p.code,p.name,p.name_zh,p.name_en,p.icon_url,p.level_up_reward,p.weekly_reward,p.monthly_reward,p.daily_withdraw_count,
       p.required_experience,p.required_deposit,p.required_valid_bet,p.withdrawal_limit,p.cashback_rate,p.birthday_reward,p.sort_order,p.status,p.enabled,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_vip_levels p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_vip_levels t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.level_no=p.level_no AND t.deleted_at IS NULL
  );

INSERT INTO game_vip_benefits
(tenant_id,site_id,vip_level_id,benefit_code,name_zh,name_en,benefit_value,value_unit,enabled,sort_order,created_at,updated_at)
SELECT tv.tenant_id,tv.site_id,tv.id,b.benefit_code,b.name_zh,b.name_en,b.benefit_value,b.value_unit,b.enabled,b.sort_order,NOW(3),NOW(3)
FROM game_vip_benefits b
JOIN game_vip_levels pv ON pv.id=b.vip_level_id AND pv.tenant_id=0 AND pv.site_id=0 AND pv.deleted_at IS NULL
JOIN game_vip_levels tv ON tv.tenant_id<>0 AND tv.level_no=pv.level_no AND tv.deleted_at IS NULL
WHERE b.tenant_id=0 AND b.site_id=0 AND b.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_vip_benefits tb
    WHERE tb.tenant_id=tv.tenant_id AND tb.site_id=tv.site_id AND tb.vip_level_id=tv.id AND tb.benefit_code=b.benefit_code AND tb.deleted_at IS NULL
  );

INSERT INTO game_deposit_amount_options
(tenant_id,site_id,option_code,currency_code,amount,bonus_reward,cash_reward,sort,status,created_at,updated_at)
SELECT s.tenant_id,s.id,p.option_code,COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR'),p.amount,p.bonus_reward,p.cash_reward,p.sort,p.status,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_deposit_amount_options p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_deposit_amount_options t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.currency_code=COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR')
      AND t.option_code=p.option_code AND t.deleted_at IS NULL
  );

INSERT INTO game_deposit_vouchers
(tenant_id,site_id,voucher_code,title,title_zh,title_en,description_zh,description_en,voucher_type,icon_url,
 min_amount,max_amount,max_reward,bonus_reward,cash_reward,sort,status,start_at,end_at,new_users_only,first_deposit_only,
 daily_limit,weekly_limit,per_user_limit,countries,currencies,method_codes,created_at,updated_at)
SELECT s.tenant_id,s.id,p.voucher_code,p.title,p.title_zh,p.title_en,p.description_zh,p.description_en,p.voucher_type,p.icon_url,
       p.min_amount,p.max_amount,p.max_reward,p.bonus_reward,p.cash_reward,p.sort,p.status,p.start_at,p.end_at,p.new_users_only,p.first_deposit_only,
       p.daily_limit,p.weekly_limit,p.per_user_limit,p.countries,COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR'),p.method_codes,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_deposit_vouchers p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_deposit_vouchers t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.voucher_code=p.voucher_code AND t.deleted_at IS NULL
  );

INSERT INTO game_deposit_voucher_rewards
(tenant_id,site_id,voucher_id,amount_option_id,amount,bonus_reward,cash_reward,reward_amount,sort,status,created_at,updated_at)
SELECT tv.tenant_id,tv.site_id,tv.id,COALESCE(ta.id,0),r.amount,r.bonus_reward,r.cash_reward,r.reward_amount,r.sort,r.status,NOW(3),NOW(3)
FROM game_deposit_voucher_rewards r
JOIN game_deposit_vouchers pv ON pv.id=r.voucher_id AND pv.tenant_id=0 AND pv.site_id=0 AND pv.deleted_at IS NULL
JOIN game_deposit_vouchers tv ON tv.tenant_id<>0 AND tv.voucher_code=pv.voucher_code AND tv.deleted_at IS NULL
LEFT JOIN game_deposit_amount_options ta ON ta.tenant_id=tv.tenant_id AND ta.site_id=tv.site_id AND ta.amount=r.amount AND ta.deleted_at IS NULL
WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_deposit_voucher_rewards tr
    WHERE tr.tenant_id=tv.tenant_id AND tr.site_id=tv.site_id AND tr.voucher_id=tv.id AND tr.amount=r.amount AND tr.deleted_at IS NULL
  );

INSERT INTO game_activity_events
(tenant_id,site_id,activity_code,slug,kind,title,name_zh,name_en,subtitle,description_zh,description_en,promotion_type,
 trigger_type,claim_mode,stacking_group,priority,countries,currencies,languages,reward_text,max_reward_text,note,cta_text,
 action_url,hero_url,card_image_url,status,status_code,total_budget,per_user_limit,total_claim_limit,claim_count,reward_amount,
 sort,start_at,end_at,created_at,updated_at)
SELECT s.tenant_id,s.id,p.activity_code,p.slug,p.kind,p.title,p.name_zh,p.name_en,p.subtitle,p.description_zh,p.description_en,p.promotion_type,
       p.trigger_type,p.claim_mode,p.stacking_group,p.priority,COALESCE(NULLIF(UPPER(s.country_code),''),p.countries),
       COALESCE(NULLIF(UPPER(s.currency_code),''),p.currencies),COALESCE(NULLIF(s.default_language,''),p.languages),
       p.reward_text,p.max_reward_text,p.note,p.cta_text,p.action_url,p.hero_url,p.card_image_url,p.status,p.status_code,p.total_budget,
       p.per_user_limit,p.total_claim_limit,0,0,p.sort,p.start_at,p.end_at,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_activity_events p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_activity_events t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.activity_code=p.activity_code AND t.deleted_at IS NULL
  );

INSERT INTO game_promotion_participation_rules
(tenant_id,site_id,promotion_id,rule_type,comparison_operator,numeric_value,text_value,enabled,sort_order,created_at,updated_at)
SELECT ta.tenant_id,ta.site_id,ta.id,r.rule_type,r.comparison_operator,r.numeric_value,r.text_value,r.enabled,r.sort_order,NOW(3),NOW(3)
FROM game_promotion_participation_rules r
JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
JOIN game_activity_events ta ON ta.tenant_id<>0 AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_promotion_participation_rules tr
    WHERE tr.tenant_id=ta.tenant_id AND tr.site_id=ta.site_id AND tr.promotion_id=ta.id
      AND tr.rule_type=r.rule_type AND tr.sort_order=r.sort_order AND tr.deleted_at IS NULL
  );

INSERT INTO game_promotion_reward_rules
(tenant_id,site_id,promotion_id,reward_type,calculation_type,reward_amount,percentage_rate,currency,wagering_multiplier,
 minimum_amount,maximum_reward,expiration_days,enabled,sort_order,created_at,updated_at)
SELECT ta.tenant_id,ta.site_id,ta.id,r.reward_type,r.calculation_type,r.reward_amount,r.percentage_rate,COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR'),
       r.wagering_multiplier,r.minimum_amount,r.maximum_reward,r.expiration_days,r.enabled,r.sort_order,NOW(3),NOW(3)
FROM game_promotion_reward_rules r
JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
JOIN game_activity_events ta ON ta.tenant_id<>0 AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
JOIN game_saas_sites s ON s.id=ta.site_id AND s.tenant_id=ta.tenant_id AND s.deleted_at IS NULL
WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_promotion_reward_rules tr
    WHERE tr.tenant_id=ta.tenant_id AND tr.site_id=ta.site_id AND tr.promotion_id=ta.id
      AND tr.reward_type=r.reward_type AND tr.sort_order=r.sort_order AND tr.deleted_at IS NULL
  );

INSERT INTO game_promotion_eligible_games
(tenant_id,site_id,promotion_id,game_id,created_at,updated_at)
SELECT ta.tenant_id,ta.site_id,ta.id,r.game_id,NOW(3),NOW(3)
FROM game_promotion_eligible_games r
JOIN game_activity_events pa ON pa.id=r.promotion_id AND pa.tenant_id=0 AND pa.site_id=0 AND pa.deleted_at IS NULL
JOIN game_activity_events ta ON ta.tenant_id<>0 AND ta.activity_code=pa.activity_code AND ta.deleted_at IS NULL
WHERE r.tenant_id=0 AND r.site_id=0 AND r.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_promotion_eligible_games tr
    WHERE tr.tenant_id=ta.tenant_id AND tr.site_id=ta.site_id AND tr.promotion_id=ta.id AND tr.game_id=r.game_id AND tr.deleted_at IS NULL
  );

DROP PROCEDURE IF EXISTS tenant_operation_defaults_drop_index;
DROP PROCEDURE IF EXISTS tenant_operation_defaults_add_index;
