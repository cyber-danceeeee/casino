-- This catalog mapping was previously created by seed-h5/EnsureCatalogSchema.
-- Define it explicitly so production does not require demo seeding or runtime DDL.
CREATE TABLE IF NOT EXISTS game_provider_category_mapping (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  provider_code VARCHAR(64) NOT NULL DEFAULT '',
  provider_category_code VARCHAR(128) NOT NULL DEFAULT '',
  platform_category_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_game_provider_category_mappings_provider_category (provider_code, provider_category_code),
  KEY idx_provider_category_mapping_scope (tenant_id, site_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
