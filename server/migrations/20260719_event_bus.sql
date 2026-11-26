CREATE TABLE IF NOT EXISTS event_outbox (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  event_type VARCHAR(128) NOT NULL,
  aggregate_type VARCHAR(64) NOT NULL DEFAULT '',
  aggregate_id VARCHAR(128) NOT NULL DEFAULT '',
  source VARCHAR(64) NOT NULL DEFAULT '',
  idempotency_key VARCHAR(160) NOT NULL,
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  payload JSON NULL,
  metadata JSON NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
  attempts INT UNSIGNED NOT NULL DEFAULT 0,
  next_attempt_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  published_at DATETIME(3) NULL,
  last_error VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_event_outbox_event_id (event_id),
  UNIQUE KEY uk_event_outbox_idempotency (idempotency_key),
  KEY idx_event_outbox_pending (status, next_attempt_at, id),
  KEY idx_event_outbox_scope_time (tenant_id, site_id, created_at, id),
  KEY idx_event_outbox_type_time (event_type, created_at),
  KEY idx_event_outbox_aggregate (aggregate_type, aggregate_id),
  CONSTRAINT chk_event_outbox_status CHECK (status IN ('PENDING','DISPATCHING','PUBLISHED','FAILED'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS event_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  outbox_id BIGINT UNSIGNED NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  event_type VARCHAR(128) NOT NULL,
  consumer VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
  attempts INT UNSIGNED NOT NULL DEFAULT 0,
  last_error VARCHAR(1000) NOT NULL DEFAULT '',
  consumed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_event_delivery_consumer (event_id, consumer),
  KEY idx_event_delivery_outbox (outbox_id, status),
  KEY idx_event_delivery_consumer_status (consumer, status, updated_at),
  KEY idx_event_delivery_scope_time (tenant_id, site_id, updated_at),
  CONSTRAINT chk_event_delivery_status CHECK (status IN ('PENDING','SUCCESS','FAILED'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:event:list','List event outbox','查看事件中心','List event outbox','system','API',1,1,NOW(3),NOW(3)),
('system:event:view','View event detail','查看事件详情','View event detail','system','API',1,1,NOW(3),NOW(3)),
('system:event:dispatch','Dispatch pending events','分发事件','Dispatch pending events','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('system:event:list','system:event:view','system:event:dispatch')
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;
