-- Separate category membership priority from the game's order inside a category.
ALTER TABLE game_game_category_relations ADD COLUMN category_priority INT NOT NULL DEFAULT 0;
UPDATE game_game_category_relations SET category_priority=sort;
UPDATE game_game_categories SET name_en=label WHERE TRIM(name_en)='' AND deleted_at IS NULL;

SET @new_hot_category := NOT EXISTS(SELECT 1 FROM game_game_categories WHERE category_key='hot' AND deleted_at IS NULL);
INSERT IGNORE INTO game_game_categories(site_id,category_key,label,name_zh,name_en,short_label,sort,status)
 SELECT 0,'hot','HOT','热门','HOT','HOT',0,1 WHERE @new_hot_category;

-- Import only missing memberships. A soft-deleted operator membership stays removed.
INSERT IGNORE INTO game_game_category_relations(site_id,game_id,category_id,sort,category_priority)
 SELECT g.site_id,g.id,c.id,g.sort,0 FROM game_games g
 JOIN game_game_categories c ON c.id=g.platform_category_id AND c.deleted_at IS NULL
 WHERE g.deleted_at IS NULL AND g.site_id IN(0,1);
INSERT IGNORE INTO game_game_category_relations(site_id,game_id,category_id,sort,category_priority)
 SELECT 0,g.id,c.id,0,1 FROM game_games g
 JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL AND p.status=1
 JOIN game_game_categories c ON c.category_key='hot' AND c.deleted_at IS NULL
 JOIN game_site_config s ON s.id=1 AND s.deleted_at IS NULL
 WHERE @new_hot_category AND g.deleted_at IS NULL AND g.status=1 AND g.provider_enabled=1 AND g.platform_enabled=1
 AND FIND_IN_SET(s.country_code,p.supported_countries)>0 AND FIND_IN_SET(s.currency_code,p.supported_currencies)>0
 ORDER BY g.hot_score DESC,g.sort,g.id LIMIT 12;

-- Backfill default navigation labels without changing existing custom labels.
UPDATE game_h5_navigation_items SET label_zh='首页',label_en='Home',labels_json=JSON_OBJECT('zh-CN','首页','en-US','Home'),
 configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',JSON_OBJECT('zh-CN','首页','en-US','Home')),
 config_version=config_version+1
 WHERE menu_code='home' AND display_position='BOTTOM_NAV' AND deleted_at IS NULL AND TRIM(label_zh)='' AND TRIM(label_en)=''
 AND COALESCE(JSON_LENGTH(JSON_EXTRACT(configuration_json,'$.labels')),0)=0;
UPDATE game_h5_navigation_items SET label_zh='优惠',label_en='Offers',labels_json=JSON_OBJECT('zh-CN','优惠','en-US','Offers'),
 configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',JSON_OBJECT('zh-CN','优惠','en-US','Offers')),
 config_version=config_version+1
 WHERE menu_code='offers' AND display_position='BOTTOM_NAV' AND deleted_at IS NULL AND TRIM(label_zh)='' AND TRIM(label_en)=''
 AND COALESCE(JSON_LENGTH(JSON_EXTRACT(configuration_json,'$.labels')),0)=0;
UPDATE game_h5_navigation_items SET label_zh='VIP',label_en='VIP',labels_json=JSON_OBJECT('zh-CN','VIP','en-US','VIP'),
 configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',JSON_OBJECT('zh-CN','VIP','en-US','VIP')),
 config_version=config_version+1
 WHERE menu_code='vip' AND display_position='BOTTOM_NAV' AND deleted_at IS NULL AND TRIM(label_zh)='' AND TRIM(label_en)=''
 AND COALESCE(JSON_LENGTH(JSON_EXTRACT(configuration_json,'$.labels')),0)=0;
UPDATE game_h5_navigation_items SET label_zh='分享',label_en='Share',labels_json=JSON_OBJECT('zh-CN','分享','en-US','Share'),
 configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',JSON_OBJECT('zh-CN','分享','en-US','Share')),
 config_version=config_version+1
 WHERE menu_code='share' AND display_position='BOTTOM_NAV' AND deleted_at IS NULL AND TRIM(label_zh)='' AND TRIM(label_en)=''
 AND COALESCE(JSON_LENGTH(JSON_EXTRACT(configuration_json,'$.labels')),0)=0;
UPDATE game_h5_navigation_items SET label_zh='我的',label_en='Profile',labels_json=JSON_OBJECT('zh-CN','我的','en-US','Profile'),
 configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',JSON_OBJECT('zh-CN','我的','en-US','Profile')),
 config_version=config_version+1
 WHERE menu_code='profile' AND display_position='BOTTOM_NAV' AND deleted_at IS NULL AND TRIM(label_zh)='' AND TRIM(label_en)=''
 AND COALESCE(JSON_LENGTH(JSON_EXTRACT(configuration_json,'$.labels')),0)=0;

-- The H5 homepage now uses the category catalog directly.
DELETE rm FROM game_admin_role_menus rm JOIN game_admin_menus m ON m.id=rm.menu_id WHERE m.code='lobby-sections';
DELETE FROM game_admin_menus WHERE code='lobby-sections';
DELETE rp FROM game_admin_role_permissions rp JOIN game_admin_permissions p ON p.id=rp.permission_id WHERE p.code LIKE 'lobby:%';
DELETE FROM game_admin_permissions WHERE code LIKE 'lobby:%';

-- User-approved reference values, interpreted in the deployment currency (IDR).
-- Seed only an empty catalog, preserving any pre-existing operator VIP policy.
SET @seed_reference_vip := NOT EXISTS(SELECT 1 FROM game_vip_levels WHERE site_id IN(0,1) AND deleted_at IS NULL);
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,1,'VIP 1','',2.00,2.00,2.00,0,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,2,'VIP 2','',5.00,5.00,5.00,0,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,3,'VIP 3','',10.00,10.00,10.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,4,'VIP 4','',30.00,20.00,30.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,5,'VIP 5','',50.00,25.00,50.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,6,'VIP 6','',100.00,35.00,100.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,7,'VIP 7','',300.00,60.00,300.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,8,'VIP 8','',500.00,100.00,500.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,9,'VIP 9','',1000.00,200.00,800.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,10,'VIP 10','',3000.00,350.00,1200.00,1,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,11,'VIP 11','',5000.00,600.00,3000.00,2,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,12,'VIP 12','',10000.00,1200.00,6000.00,2,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,13,'VIP 13','',30000.00,1800.00,12000.00,2,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,14,'VIP 14','',50000.00,2500.00,40000.00,3,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_levels(site_id,level_no,name,icon_url,level_up_reward,weekly_reward,monthly_reward,daily_withdraw_count,status) SELECT 1,15,'VIP 15','',100000.00,3000.00,80000.00,5,1 WHERE @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',1,0,100,500 FROM game_vip_levels WHERE site_id=1 AND level_no=1 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',1,30,100,300 FROM game_vip_levels WHERE site_id=1 AND level_no=1 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',2,0,500,5000 FROM game_vip_levels WHERE site_id=1 AND level_no=2 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',2,30,200,600 FROM game_vip_levels WHERE site_id=1 AND level_no=2 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',3,0,1000,10000 FROM game_vip_levels WHERE site_id=1 AND level_no=3 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',3,30,300,900 FROM game_vip_levels WHERE site_id=1 AND level_no=3 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',4,0,3000,30000 FROM game_vip_levels WHERE site_id=1 AND level_no=4 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',4,30,800,2400 FROM game_vip_levels WHERE site_id=1 AND level_no=4 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',5,0,5000,50000 FROM game_vip_levels WHERE site_id=1 AND level_no=5 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',5,30,1500,4500 FROM game_vip_levels WHERE site_id=1 AND level_no=5 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',6,0,10000,100000 FROM game_vip_levels WHERE site_id=1 AND level_no=6 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',6,45,5000,15000 FROM game_vip_levels WHERE site_id=1 AND level_no=6 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',7,0,30000,300000 FROM game_vip_levels WHERE site_id=1 AND level_no=7 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',7,45,10000,30000 FROM game_vip_levels WHERE site_id=1 AND level_no=7 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',8,0,50000,500000 FROM game_vip_levels WHERE site_id=1 AND level_no=8 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',8,45,20000,60000 FROM game_vip_levels WHERE site_id=1 AND level_no=8 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',9,0,100000,1000000 FROM game_vip_levels WHERE site_id=1 AND level_no=9 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',9,45,36000,108000 FROM game_vip_levels WHERE site_id=1 AND level_no=9 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',10,0,300000,3000000 FROM game_vip_levels WHERE site_id=1 AND level_no=10 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',10,45,60000,180000 FROM game_vip_levels WHERE site_id=1 AND level_no=10 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',11,0,500000,5000000 FROM game_vip_levels WHERE site_id=1 AND level_no=11 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',11,60,120000,360000 FROM game_vip_levels WHERE site_id=1 AND level_no=11 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',12,0,1000000,10000000 FROM game_vip_levels WHERE site_id=1 AND level_no=12 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',12,60,240000,720000 FROM game_vip_levels WHERE site_id=1 AND level_no=12 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',13,0,3000000,30000000 FROM game_vip_levels WHERE site_id=1 AND level_no=13 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',13,60,400000,1200000 FROM game_vip_levels WHERE site_id=1 AND level_no=13 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',14,0,5000000,50000000 FROM game_vip_levels WHERE site_id=1 AND level_no=14 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',14,60,1000000,3000000 FROM game_vip_levels WHERE site_id=1 AND level_no=14 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'level_up',15,0,10000000,100000000 FROM game_vip_levels WHERE site_id=1 AND level_no=15 AND deleted_at IS NULL AND @seed_reference_vip;
INSERT INTO game_vip_requirements(site_id,vip_level_id,requirement_type,target_level_no,period_days,deposit_amount,valid_bet_amount) SELECT 1,id,'retain',15,60,1800000,5400000 FROM game_vip_levels WHERE site_id=1 AND level_no=15 AND deleted_at IS NULL AND @seed_reference_vip;
