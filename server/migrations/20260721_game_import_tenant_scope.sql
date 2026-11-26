-- Allow tenants to import platform games into their own scoped game catalog.

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_games'
      AND index_name = 'uk_game_games_code'
);
SET @sql := IF(@idx > 0,
    'ALTER TABLE game_games DROP INDEX uk_game_games_code',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_games'
      AND index_name = 'uk_game_games_scope_code'
);
SET @sql := IF(@idx = 0,
    'ALTER TABLE game_games ADD UNIQUE KEY uk_game_games_scope_code (tenant_id, site_id, game_code)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
