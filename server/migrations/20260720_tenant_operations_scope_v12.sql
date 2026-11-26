-- Tenant owners may operate only their own scoped retry records, dead letters, and alerts.
INSERT INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
SELECT r.id, p.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'jobs:read', 'jobs:retry',
  'alerts:read', 'alerts:acknowledge', 'alerts:resolve'
) AND p.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at);

INSERT INTO game_admin_role_menus (role_id, menu_id, created_at, updated_at, deleted_at)
SELECT r.id, m.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code IN ('ops-stability', 'jobs-retries', 'alert-center')
  AND m.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at);
