-- Site codes are platform identifiers, but a soft-deleted site must not
-- reserve its code forever. Keep one active site per code and preserve the
-- deleted site's original value in history.
SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_sites'
      AND index_name = 'uk_saas_site_code'
);
SET @sql := IF(@idx > 0,
    'ALTER TABLE game_saas_sites DROP INDEX uk_saas_site_code',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
    SELECT COUNT(1)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_sites'
      AND column_name = 'site_code_active'
);
SET @sql := IF(@column_exists = 0,
    'ALTER TABLE game_saas_sites ADD COLUMN site_code_active VARCHAR(64) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `site_code`, NULL)) STORED',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_sites'
      AND index_name = 'uk_saas_site_active_code'
);
SET @sql := IF(@idx = 0,
    'ALTER TABLE game_saas_sites ADD UNIQUE KEY uk_saas_site_active_code (site_code_active)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
