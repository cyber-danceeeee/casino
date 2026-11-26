-- Scope game categories by tenant/site so tenant-created category keys do not collide globally.

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_game_categories'
      AND index_name = 'uk_game_game_categories_category_key'
);
SET @sql := IF(@idx > 0,
    'ALTER TABLE game_game_categories DROP INDEX uk_game_game_categories_category_key',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_game_categories'
      AND index_name = 'uk_game_categories_key'
);
SET @sql := IF(@idx > 0,
    'ALTER TABLE game_game_categories DROP INDEX uk_game_categories_key',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_game_categories'
      AND index_name = 'uk_game_game_categories_scope_key'
);
SET @sql := IF(@idx = 0,
    'ALTER TABLE game_game_categories ADD UNIQUE KEY uk_game_game_categories_scope_key (tenant_id, site_id, category_key)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
