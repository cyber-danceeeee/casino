CREATE TABLE IF NOT EXISTS `game_tenant_operation_overrides` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned NOT NULL,
  `site_id` bigint unsigned NOT NULL,
  `resource_type` varchar(32) NOT NULL,
  `resource_id` bigint unsigned NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `created_by` bigint unsigned NOT NULL DEFAULT 0,
  `updated_by` bigint unsigned NOT NULL DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_operation_override` (`tenant_id`,`site_id`,`resource_type`,`resource_id`),
  KEY `idx_tenant_operation_override_visible` (`tenant_id`,`site_id`,`resource_type`,`enabled`),
  KEY `idx_tenant_operation_override_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET @add_h5_navigation_platform_source_column = IF(
  (SELECT COUNT(*) FROM information_schema.columns
   WHERE table_schema=DATABASE() AND table_name='game_h5_navigation_items' AND column_name='platform_source_id')=0,
  'ALTER TABLE game_h5_navigation_items ADD COLUMN platform_source_id bigint unsigned NOT NULL DEFAULT 0 AFTER site_id',
  'SELECT 1'
);
PREPARE add_h5_navigation_platform_source_column_stmt FROM @add_h5_navigation_platform_source_column;
EXECUTE add_h5_navigation_platform_source_column_stmt;
DEALLOCATE PREPARE add_h5_navigation_platform_source_column_stmt;

SET @add_h5_navigation_platform_source_index = IF(
  (SELECT COUNT(*) FROM information_schema.statistics
   WHERE table_schema=DATABASE() AND table_name='game_h5_navigation_items' AND index_name='idx_h5_navigation_platform_source')=0,
  'ALTER TABLE game_h5_navigation_items ADD KEY idx_h5_navigation_platform_source (tenant_id,site_id,platform_source_id)',
  'SELECT 1'
);
PREPARE add_h5_navigation_platform_source_index_stmt FROM @add_h5_navigation_platform_source_index;
EXECUTE add_h5_navigation_platform_source_index_stmt;
DEALLOCATE PREPARE add_h5_navigation_platform_source_index_stmt;

UPDATE `game_h5_navigation_items` tenant_item
JOIN `game_h5_navigation_items` platform_item
  ON platform_item.tenant_id=0
 AND platform_item.site_id=0
 AND platform_item.menu_code=tenant_item.menu_code
 AND platform_item.deleted_at IS NULL
SET tenant_item.platform_source_id=platform_item.id
WHERE tenant_item.tenant_id<>0
  AND tenant_item.platform_source_id=0
  AND tenant_item.deleted_at IS NULL;
