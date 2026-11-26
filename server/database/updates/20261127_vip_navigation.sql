-- Add standalone VIP navigation without replacing operator customization.
INSERT IGNORE INTO game_h5_navigation_items
 (site_id,menu_code,label_zh,label_en,display_position,action_type,action_target,page_route,
  sort_order,badge_type,visibility_scope,country_codes,language_codes,enabled,publish_status,
  config_version,published_at,created_at,updated_at)
 SELECT 1,'vip','VIP','VIP','BOTTOM_NAV','OPEN_INTERNAL_PAGE','/vip','/vip',
  30,'NONE','ALL','[]','[]',1,'PUBLISHED',1,NOW(3),NOW(3),NOW(3)
 FROM game_site_config WHERE id=1 AND deleted_at IS NULL;
