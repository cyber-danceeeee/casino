UPDATE game_admin_role_permissions rp
JOIN game_admin_permissions p ON p.id = rp.permission_id
SET rp.deleted_at = NOW(3), rp.updated_at = NOW(3)
WHERE p.code LIKE 'system:dictionary:%' AND rp.deleted_at IS NULL;

UPDATE game_admin_permissions
SET deleted_at = NOW(3), enabled = 0, updated_at = NOW(3)
WHERE code LIKE 'system:dictionary:%' AND deleted_at IS NULL;

UPDATE game_admin_role_menus rm
JOIN game_admin_menus m ON m.id = rm.menu_id
SET rm.deleted_at = NOW(3), rm.updated_at = NOW(3)
WHERE (m.code = 'system-dictionaries' OR m.route_path = '/system/dictionaries')
  AND rm.deleted_at IS NULL;

UPDATE game_admin_menus
SET deleted_at = NOW(3), visible = 0, enabled = 0, updated_at = NOW(3)
WHERE (code = 'system-dictionaries' OR route_path = '/system/dictionaries')
  AND deleted_at IS NULL;

DROP TABLE IF EXISTS game_admin_dictionary_items;
DROP TABLE IF EXISTS game_admin_dictionaries;
