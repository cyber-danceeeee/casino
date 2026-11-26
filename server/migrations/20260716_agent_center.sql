CREATE TABLE IF NOT EXISTS game_commission_plans (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  model VARCHAR(32) NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  cpa_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  revenue_share_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
  negative_carryover TINYINT(1) NOT NULL DEFAULT 0,
  minimum_settlement_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  settlement_cycle VARCHAR(32) NOT NULL DEFAULT 'MONTHLY',
  deduction_rules JSON NULL,
  tier_rules JSON NULL,
  eligible_products JSON NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_commission_plans_code (code),
  KEY idx_game_commission_plans_model_enabled (model, enabled), KEY idx_game_commission_plans_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  agent_code VARCHAR(32) NOT NULL,
  name VARCHAR(128) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  parent_agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  level INT UNSIGNED NOT NULL DEFAULT 1,
  country VARCHAR(8) NOT NULL DEFAULT 'PK',
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
  commission_plan_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  activated_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agents_code (agent_code),
  KEY idx_game_agents_parent (parent_agent_id), KEY idx_game_agents_user (user_id),
  KEY idx_game_agents_status_country (status, country), KEY idx_game_agents_plan (commission_plan_id),
  KEY idx_game_agents_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL, name_zh VARCHAR(128) NOT NULL, name_en VARCHAR(128) NOT NULL,
  agent_id BIGINT UNSIGNED NOT NULL, country VARCHAR(8) NOT NULL DEFAULT 'PK',
  platform VARCHAR(24) NOT NULL DEFAULT 'H5', registration_url VARCHAR(500) NOT NULL DEFAULT '',
  tracking_parameters JSON NULL, secret_hash VARCHAR(128) NOT NULL DEFAULT '', enabled TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agent_channels_code (code),
  KEY idx_game_agent_channels_agent (agent_id), KEY idx_game_agent_channels_country_enabled (country, enabled),
  KEY idx_game_agent_channels_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_user_agent_attributions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, user_id BIGINT UNSIGNED NOT NULL,
  agent_id BIGINT UNSIGNED NOT NULL, channel_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  attribution_source VARCHAR(32) NOT NULL, active TINYINT(1) NOT NULL DEFAULT 1, attributed_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_user_agent_active (user_id, active),
  KEY idx_game_user_agent_agent (agent_id, attributed_at), KEY idx_game_user_agent_channel (channel_id, attributed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_user_agent_attribution_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, user_id BIGINT UNSIGNED NOT NULL,
  before_agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0, after_agent_id BIGINT UNSIGNED NOT NULL,
  before_channel_id BIGINT UNSIGNED NOT NULL DEFAULT 0, after_channel_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  attribution_source VARCHAR(32) NOT NULL, operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '', reason VARCHAR(500) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), KEY idx_game_attribution_history_user (user_id, created_at), KEY idx_game_attribution_history_agent (after_agent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_commission_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, record_no VARCHAR(48) NOT NULL,
  agent_id BIGINT UNSIGNED NOT NULL, user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  attribution_agent_id BIGINT UNSIGNED NOT NULL, commission_plan_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  commission_type VARCHAR(32) NOT NULL, calculation_base DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  commission_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000, commission_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR', business_date DATE NOT NULL, status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  source_type VARCHAR(32) NOT NULL DEFAULT '', source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reversal_of_id BIGINT UNSIGNED NOT NULL DEFAULT 0, confirmed_at DATETIME(3) NULL, settlement_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_commission_records_no (record_no),
  UNIQUE KEY uk_game_commission_source (agent_id, commission_type, source_type, source_id),
  KEY idx_game_commission_agent_date (agent_id, business_date), KEY idx_game_commission_user (user_id),
  KEY idx_game_commission_status (status, created_at), KEY idx_game_commission_settlement (settlement_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_settlements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, settlement_no VARCHAR(48) NOT NULL,
  agent_id BIGINT UNSIGNED NOT NULL, period_start DATE NOT NULL, period_end DATE NOT NULL,
  gross_commission DECIMAL(18,2) NOT NULL DEFAULT 0.00, adjustments DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  final_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  status VARCHAR(32) NOT NULL DEFAULT 'DRAFT', creator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  creator_username VARCHAR(64) NOT NULL DEFAULT '', reviewer_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reviewer_username VARCHAR(64) NOT NULL DEFAULT '', review_reason VARCHAR(500) NOT NULL DEFAULT '',
  idempotency_key VARCHAR(128) NOT NULL, submitted_at DATETIME(3) NULL, reviewed_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agent_settlement_no (settlement_no),
  UNIQUE KEY uk_game_agent_settlement_period (agent_id, period_start, period_end, currency),
  UNIQUE KEY uk_game_agent_settlement_idempotency (idempotency_key), KEY idx_game_agent_settlement_status (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_settlement_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, settlement_id BIGINT UNSIGNED NOT NULL,
  commission_record_id BIGINT UNSIGNED NOT NULL, amount DECIMAL(18,2) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agent_settlement_item (commission_record_id), KEY idx_game_agent_settlement_items (settlement_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_wallets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, wallet_no VARCHAR(48) NOT NULL, agent_id BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR', available_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  frozen_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00, settled_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  withdrawn_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00, status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE', version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agent_wallet_no (wallet_no), UNIQUE KEY uk_game_agent_wallet_agent_currency (agent_id, currency),
  KEY idx_game_agent_wallet_status (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_wallet_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, transaction_no VARCHAR(48) NOT NULL, wallet_id BIGINT UNSIGNED NOT NULL,
  agent_id BIGINT UNSIGNED NOT NULL, transaction_type VARCHAR(32) NOT NULL, amount DECIMAL(18,2) NOT NULL,
  balance_before DECIMAL(18,2) NOT NULL, balance_after DECIMAL(18,2) NOT NULL, currency VARCHAR(8) NOT NULL,
  reference_type VARCHAR(32) NOT NULL DEFAULT '', reference_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(128) NOT NULL, operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '', reason VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_game_agent_wallet_tx_no (transaction_no), UNIQUE KEY uk_game_agent_wallet_tx_idempotency (idempotency_key),
  KEY idx_game_agent_wallet_tx_wallet (wallet_id, created_at), KEY idx_game_agent_wallet_tx_agent (agent_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS game_agent_hierarchy_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, agent_id BIGINT UNSIGNED NOT NULL,
  before_parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0, after_parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  before_level INT UNSIGNED NOT NULL, after_level INT UNSIGNED NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0, operator_username VARCHAR(64) NOT NULL DEFAULT '', reason VARCHAR(500) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id), KEY idx_game_agent_hierarchy_agent (agent_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_permissions (code,name,created_at,updated_at) VALUES
('agent:list','List agents',NOW(3),NOW(3)),('agent:view','View agent',NOW(3),NOW(3)),('agent:create','Create agent',NOW(3),NOW(3)),('agent:update','Update agent',NOW(3),NOW(3)),('agent:status','Update agent status',NOW(3),NOW(3)),
('agent:hierarchy:view','View agent hierarchy',NOW(3),NOW(3)),('agent:hierarchy:update','Update agent hierarchy',NOW(3),NOW(3)),('agent:plan:assign','Assign commission plan',NOW(3),NOW(3)),
('agent:channel:list','List agent channels',NOW(3),NOW(3)),('agent:channel:create','Create agent channel',NOW(3),NOW(3)),('agent:channel:update','Update agent channel',NOW(3),NOW(3)),('agent:channel:status','Update agent channel status',NOW(3),NOW(3)),
('agent:commission-plan:list','List commission plans',NOW(3),NOW(3)),('agent:commission-plan:create','Create commission plan',NOW(3),NOW(3)),('agent:commission-plan:update','Update commission plan',NOW(3),NOW(3)),('agent:commission-plan:delete','Delete commission plan',NOW(3),NOW(3)),
('agent:commission-record:list','List commission records',NOW(3),NOW(3)),('agent:commission-record:view','View commission record',NOW(3),NOW(3)),('agent:commission-record:adjust','Adjust commission record',NOW(3),NOW(3)),
('agent:settlement:list','List agent settlements',NOW(3),NOW(3)),('agent:settlement:view','View agent settlement',NOW(3),NOW(3)),('agent:settlement:generate','Generate agent settlement',NOW(3),NOW(3)),('agent:settlement:submit','Submit agent settlement',NOW(3),NOW(3)),('agent:settlement:approve','Approve agent settlement',NOW(3),NOW(3)),('agent:settlement:reject','Reject agent settlement',NOW(3),NOW(3)),('agent:settlement:pay','Pay agent settlement',NOW(3),NOW(3)),('agent:settlement:cancel','Cancel agent settlement',NOW(3),NOW(3)),('agent:settlement:export','Export agent settlement',NOW(3),NOW(3)),
('agent:wallet:list','List agent wallets',NOW(3),NOW(3)),('agent:wallet:view','View agent wallet',NOW(3),NOW(3)),('agent:wallet:freeze','Freeze agent wallet',NOW(3),NOW(3)),('agent:wallet:adjust','Adjust agent wallet',NOW(3),NOW(3)),('agent:operation-log:view','View agent operation logs',NOW(3),NOW(3)),('agent:attribution:update','Update user agent attribution',NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_permissions p ON p.code LIKE 'agent:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL;
