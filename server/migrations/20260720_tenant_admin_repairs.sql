-- Tenant admin production repairs: H5 entry RBAC and initial tenant game catalog.

INSERT IGNORE INTO game_admin_permissions
  (code,name,name_zh,name_en,module,permission_type,api_path,enabled,system_protected,created_at,updated_at,deleted_at)
VALUES
  ('operation:h5-navigation:list','List H5 navigation','查看 H5 菜单','View H5 navigation','operations','API','/api/v1/admin/h5-navigation',1,1,NOW(3),NOW(3),NULL),
  ('operation:h5-navigation:create','Create H5 navigation','新增 H5 菜单','Create H5 navigation','operations','API','/api/v1/admin/h5-navigation',1,1,NOW(3),NOW(3),NULL),
  ('operation:h5-navigation:update','Update H5 navigation','编辑 H5 菜单','Update H5 navigation','operations','API','/api/v1/admin/h5-navigation/:id',1,1,NOW(3),NOW(3),NULL),
  ('operation:h5-navigation:delete','Delete H5 navigation','删除 H5 菜单','Delete H5 navigation','operations','API','/api/v1/admin/h5-navigation/:id',1,1,NOW(3),NOW(3),NULL);

INSERT IGNORE INTO game_admin_role_permissions
  (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'operation:h5-navigation:list',
  'operation:h5-navigation:create',
  'operation:h5-navigation:update',
  'operation:h5-navigation:delete'
)
WHERE r.code IN ('super-admin','tenant-owner')
  AND r.deleted_at IS NULL
  AND p.deleted_at IS NULL;

-- Older provisioning created selected tenant providers as disabled and did not
-- materialize their local tenant game catalog. Only repair untouched providers.
UPDATE game_game_providers tp
SET tp.status=1,tp.updated_at=NOW(3)
WHERE tp.tenant_id>0
  AND tp.status=3
  AND tp.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_games existing_game
    WHERE existing_game.tenant_id=tp.tenant_id
      AND existing_game.site_id=tp.site_id
      AND existing_game.provider_id=tp.id
      AND existing_game.deleted_at IS NULL
  );

INSERT IGNORE INTO game_games (
  tenant_id,site_id,game_code,provider_id,provider_game_code,name,name_i18n,slug,
  category,game_type,rtp,volatility,cover_url,local_thumbnail_url,background_url,
  launch_image_url,maintenance_message,maintenance_message_draft,supports_demo,
  supports_free_spins,provider_enabled,platform_enabled,platform_enabled_draft,
  is_hot,is_hot_draft,is_featured,is_featured_draft,is_new,is_new_draft,
  supported_countries,supported_currencies,supported_devices,raw_metadata,
  metadata_hash,last_synced_at,badge,player_count,hot_score,launch_url,status,
  status_draft,sort,sort_draft,created_at,updated_at,deleted_at,
  platform_category_id,platform_category_code,category_ids_draft,config_status,
  config_version,config_published_at,config_published_by,draft_updated_at,draft_updated_by
)
SELECT
  tp.tenant_id,tp.site_id,g.game_code,tp.id,g.provider_game_code,g.name,COALESCE(g.name_i18n, JSON_OBJECT()),g.slug,
  g.category,g.game_type,g.rtp,g.volatility,g.cover_url,g.local_thumbnail_url,g.background_url,
  g.launch_image_url,g.maintenance_message,NULL,g.supports_demo,g.supports_free_spins,
  IF(tp.status=1,g.provider_enabled,0),g.platform_enabled,NULL,
  g.is_hot,NULL,g.is_featured,NULL,g.is_new,NULL,
  g.supported_countries,g.supported_currencies,g.supported_devices,COALESCE(g.raw_metadata, JSON_OBJECT()),
  g.metadata_hash,g.last_synced_at,g.badge,g.player_count,g.hot_score,g.launch_url,g.status,
  NULL,g.sort,NULL,NOW(3),NOW(3),NULL,
  g.platform_category_id,g.platform_category_code,JSON_ARRAY(),'PUBLISHED',
  1,NOW(3),0,NULL,0
FROM game_game_providers tp
JOIN game_game_providers pp
  ON pp.tenant_id=0
 AND pp.site_id=0
 AND pp.provider_code=tp.provider_code
 AND pp.deleted_at IS NULL
JOIN game_games g
  ON g.tenant_id=0
 AND g.site_id=0
 AND g.provider_id=pp.id
 AND g.deleted_at IS NULL
WHERE tp.tenant_id>0
  AND tp.deleted_at IS NULL;
