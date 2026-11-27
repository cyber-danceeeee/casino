-- Complete cleanup after the application no longer reads the lobby table.
DROP TABLE IF EXISTS game_lobby_sections;

-- Some legacy rows have names in columns but an empty authoritative labels map.
-- Keep existing translated labels and use legacy columns for missing languages.
UPDATE game_h5_navigation_items
 SET label_zh=COALESCE(NULLIF(TRIM(label_zh),''),CASE menu_code WHEN 'home' THEN '首页' WHEN 'offers' THEN '优惠' WHEN 'vip' THEN 'VIP' WHEN 'share' THEN '分享' ELSE '我的' END),
 label_en=COALESCE(NULLIF(TRIM(label_en),''),CASE menu_code WHEN 'home' THEN 'Home' WHEN 'offers' THEN 'Offers' WHEN 'vip' THEN 'VIP' WHEN 'share' THEN 'Share' ELSE 'Profile' END)
 WHERE menu_code IN('home','offers','vip','share','profile') AND display_position='BOTTOM_NAV' AND deleted_at IS NULL;
UPDATE game_h5_navigation_items
 SET configuration_json=JSON_SET(COALESCE(configuration_json,JSON_OBJECT()),'$.labels',
 JSON_MERGE_PATCH(COALESCE(JSON_EXTRACT(configuration_json,'$.labels'),JSON_OBJECT()),JSON_OBJECT(
 'zh-CN',COALESCE(NULLIF(TRIM(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.labels."zh-CN"'))),''),label_zh),
 'en-US',COALESCE(NULLIF(TRIM(JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.labels."en-US"'))),''),label_en)))),
 labels_json=JSON_EXTRACT(configuration_json,'$.labels'),
 label_zh=JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.labels."zh-CN"')),
 label_en=JSON_UNQUOTE(JSON_EXTRACT(configuration_json,'$.labels."en-US"')),
 config_version=config_version+1
 WHERE menu_code IN('home','offers','vip','share','profile') AND display_position='BOTTOM_NAV' AND deleted_at IS NULL;
