DROP PROCEDURE IF EXISTS event_consumer_add_column;
DELIMITER $$
CREATE PROCEDURE event_consumer_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE event_consumer_statement FROM @statement_value; EXECUTE event_consumer_statement; DEALLOCATE PREPARE event_consumer_statement;
  END IF;
END$$
DELIMITER ;

CALL event_consumer_add_column('event_outbox','target_consumers','JSON NULL AFTER `causation_id`');

CREATE TABLE IF NOT EXISTS audit_event_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  event_version INT NOT NULL DEFAULT 1,
  consumer_name VARCHAR(128) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  aggregate_type VARCHAR(64) NOT NULL DEFAULT '',
  aggregate_id VARCHAR(128) NOT NULL DEFAULT '',
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  transaction_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  business_type VARCHAR(64) NOT NULL DEFAULT '',
  currency VARCHAR(16) NOT NULL DEFAULT '',
  amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  trace_id VARCHAR(128) NOT NULL DEFAULT '',
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  causation_id VARCHAR(128) NOT NULL DEFAULT '',
  payload_snapshot JSON NULL,
  occurred_at DATETIME(3) NOT NULL,
  consumed_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_audit_event_consumer (event_id, consumer_name),
  KEY idx_audit_event_scope_time (tenant_id, site_id, occurred_at, id),
  KEY idx_audit_event_trace (trace_id, correlation_id, causation_id),
  KEY idx_audit_event_transaction (transaction_id, currency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS dashboard_event_facts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  consumer_name VARCHAR(128) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  business_date DATE NOT NULL,
  metric_type VARCHAR(64) NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT '',
  amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  trace_id VARCHAR(128) NOT NULL DEFAULT '',
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  causation_id VARCHAR(128) NOT NULL DEFAULT '',
  occurred_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_dashboard_fact_event_metric (event_id, consumer_name, metric_type),
  KEY idx_dashboard_fact_scope_date (tenant_id, site_id, business_date, metric_type, currency),
  KEY idx_dashboard_fact_trace (trace_id, correlation_id, causation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP PROCEDURE IF EXISTS event_consumer_add_column;
