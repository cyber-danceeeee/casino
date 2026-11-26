INSERT IGNORE INTO game_admin_permissions (code, name, created_at, updated_at)
VALUES ('game:asset:upload', 'Upload game asset', NOW(3), NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
SELECT r.id, p.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code = 'game:asset:upload' AND p.deleted_at IS NULL
WHERE r.code = 'super-admin' AND r.deleted_at IS NULL;
