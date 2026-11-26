START TRANSACTION;

UPDATE game_admin_role_permissions rp
INNER JOIN game_admin_roles r ON r.id = rp.role_id
INNER JOIN game_admin_permissions p ON p.id = rp.permission_id
SET rp.deleted_at = NULL,
    rp.updated_at = NOW(3)
WHERE r.code = 'super-admin'
  AND r.status = 1
  AND r.system_protected = 1
  AND r.deleted_at IS NULL
  AND p.enabled = 1
  AND p.deleted_at IS NULL
  AND rp.deleted_at IS NOT NULL;

INSERT INTO game_admin_role_permissions (
  role_id,
  permission_id,
  created_at,
  updated_at,
  deleted_at
)
SELECT
  r.id,
  p.id,
  NOW(3),
  NOW(3),
  NULL
FROM game_admin_roles r
CROSS JOIN game_admin_permissions p
LEFT JOIN game_admin_role_permissions rp
  ON rp.role_id = r.id
 AND rp.permission_id = p.id
WHERE r.code = 'super-admin'
  AND r.status = 1
  AND r.system_protected = 1
  AND r.deleted_at IS NULL
  AND p.enabled = 1
  AND p.deleted_at IS NULL
  AND rp.id IS NULL;

COMMIT;
