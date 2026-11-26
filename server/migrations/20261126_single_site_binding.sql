-- One immutable deployment binding. Existing business IDs and foreign keys are
-- preserved; initialization refuses ambiguous multi-site databases.
CREATE TABLE IF NOT EXISTS casino_site_binding (
  id TINYINT UNSIGNED NOT NULL PRIMARY KEY,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  CONSTRAINT chk_casino_single_site CHECK (id = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT IGNORE INTO casino_site_binding (id) VALUES (1);
