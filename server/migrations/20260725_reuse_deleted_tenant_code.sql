-- A soft-deleted tenant must not reserve its tenant code forever. Keep one
-- active tenant per code while preserving deleted tenant history. A generated
-- column is used because MySQL allows multiple NULL values in a unique index.
SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_tenants'
      AND index_name = 'uk_saas_tenant_code'
);
SET @sql := IF(@idx > 0,
    'ALTER TABLE game_saas_tenants DROP INDEX uk_saas_tenant_code',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_tenants'
      AND index_name = 'uk_saas_tenant_code_deleted_at'
);
SET @sql := IF(@idx = 0,
    'SELECT 1',
    'ALTER TABLE game_saas_tenants DROP INDEX uk_saas_tenant_code_deleted_at'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists := (
    SELECT COUNT(1)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_tenants'
      AND column_name = 'tenant_code_active'
);
SET @sql := IF(@column_exists = 0,
    'ALTER TABLE game_saas_tenants ADD COLUMN tenant_code_active VARCHAR(64) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `tenant_code`, NULL)) STORED',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx := (
    SELECT COUNT(1)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'game_saas_tenants'
      AND index_name = 'uk_saas_tenant_active_code'
);
SET @sql := IF(@idx = 0,
    'ALTER TABLE game_saas_tenants ADD UNIQUE KEY uk_saas_tenant_active_code (tenant_code_active)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
