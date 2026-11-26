CREATE TABLE IF NOT EXISTS game_hourly_platform_summaries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  bucket_at DATETIME NOT NULL,
  timezone VARCHAR(40) NOT NULL,
  country VARCHAR(8) NOT NULL DEFAULT '',
  currency VARCHAR(8) NOT NULL,
  provider_code VARCHAR(64) NOT NULL DEFAULT '',
  category_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  channel_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  registered_users BIGINT UNSIGNED NOT NULL DEFAULT 0,
  active_users BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_deposit_users BIGINT UNSIGNED NOT NULL DEFAULT 0,
  deposit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
  withdrawal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
  total_bet DECIMAL(20,2) NOT NULL DEFAULT 0,
  total_win DECIMAL(20,2) NOT NULL DEFAULT 0,
  bonus_cost DECIMAL(20,2) NOT NULL DEFAULT 0,
  provider_cost DECIMAL(20,2) NOT NULL DEFAULT 0,
  payment_fee DECIMAL(20,2) NOT NULL DEFAULT 0,
  agent_commission DECIMAL(20,2) NOT NULL DEFAULT 0,
  other_cost DECIMAL(20,2) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_dashboard_hour_dimension (bucket_at,timezone,country,currency,provider_code,category_id,channel_id,agent_id),
  KEY idx_dashboard_hour_query (bucket_at,currency,country,provider_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS game_daily_risk_summaries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  business_date DATE NOT NULL,
  timezone VARCHAR(40) NOT NULL,
  country VARCHAR(8) NOT NULL DEFAULT '',
  risk_type VARCHAR(64) NOT NULL DEFAULT '',
  risk_level VARCHAR(16) NOT NULL DEFAULT '',
  review_status VARCHAR(32) NOT NULL DEFAULT '',
  event_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  affected_users BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_dashboard_risk_dimension (business_date,timezone,country,risk_type,risk_level,review_status),
  KEY idx_dashboard_risk_query (business_date,risk_level,review_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT IGNORE INTO game_admin_permissions (code,name,created_at,updated_at) VALUES
('dashboard:view','View BI dashboard',NOW(3),NOW(3)),
('dashboard:financial:view','View dashboard financial data',NOW(3),NOW(3)),
('dashboard:user:view','View dashboard user data',NOW(3),NOW(3)),
('dashboard:game:view','View dashboard game data',NOW(3),NOW(3)),
('dashboard:provider:view','View dashboard provider data',NOW(3),NOW(3)),
('dashboard:channel:view','View dashboard channel data',NOW(3),NOW(3)),
('dashboard:agent:view','View dashboard agent data',NOW(3),NOW(3)),
('dashboard:promotion:view','View dashboard promotion data',NOW(3),NOW(3)),
('dashboard:risk:view','View dashboard risk data',NOW(3),NOW(3)),
('dashboard:realtime:view','View dashboard realtime data',NOW(3),NOW(3)),
('dashboard:alert:view','View dashboard alerts',NOW(3),NOW(3)),
('dashboard:sensitive:view','View dashboard sensitive data',NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_permissions p ON p.code LIKE 'dashboard:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL;
