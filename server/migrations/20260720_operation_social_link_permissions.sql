-- Social-link operations were routed through RBAC before their permission rows
-- existed, which made the tenant menu visible but every request return 403.

INSERT IGNORE INTO game_admin_permissions
  (code,name,name_zh,name_en,module,permission_type,api_path,enabled,system_protected,created_at,updated_at,deleted_at)
VALUES
  ('operation:social-link:list','List social links','查看客服与社媒','View social links','operations','API','/api/v1/admin/social-links',1,1,NOW(3),NOW(3),NULL),
  ('operation:social-link:create','Create social link','新增客服与社媒','Create social link','operations','API','/api/v1/admin/social-links',1,1,NOW(3),NOW(3),NULL),
  ('operation:social-link:update','Update social link','编辑客服与社媒','Update social link','operations','API','/api/v1/admin/social-links/:id',1,1,NOW(3),NOW(3),NULL),
  ('operation:social-link:delete','Delete social link','删除客服与社媒','Delete social link','operations','API','/api/v1/admin/social-links/:id',1,1,NOW(3),NOW(3),NULL);

INSERT IGNORE INTO game_admin_role_permissions
  (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'operation:social-link:list',
  'operation:social-link:create',
  'operation:social-link:update',
  'operation:social-link:delete'
)
WHERE r.code IN ('super-admin','tenant-owner')
  AND r.deleted_at IS NULL
  AND p.deleted_at IS NULL;
