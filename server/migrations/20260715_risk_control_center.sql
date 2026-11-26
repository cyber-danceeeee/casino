ALTER TABLE game_user_risk_events
  ADD COLUMN event_no VARCHAR(64) NOT NULL DEFAULT '' AFTER id,
  ADD COLUMN rule_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER user_id,
  ADD COLUMN risk_type VARCHAR(64) NOT NULL DEFAULT 'MANUAL' AFTER event_type,
  ADD COLUMN trigger_value VARCHAR(500) NOT NULL DEFAULT '' AFTER risk_level,
  ADD COLUMN provider_code VARCHAR(64) NOT NULL DEFAULT '' AFTER trigger_value,
  ADD COLUMN game_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER provider_code,
  ADD COLUMN ip_address VARCHAR(64) NOT NULL DEFAULT '' AFTER game_id,
  ADD COLUMN device_id VARCHAR(128) NOT NULL DEFAULT '' AFTER ip_address,
  ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT 'PENDING' AFTER device_id,
  ADD COLUMN assigned_reviewer_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER status,
  ADD COLUMN assigned_reviewer VARCHAR(64) NOT NULL DEFAULT '' AFTER assigned_reviewer_id,
  ADD UNIQUE KEY uk_game_user_risk_events_no (event_no),
  ADD KEY idx_game_user_risk_events_status_level (status, risk_level, created_at),
  ADD KEY idx_game_user_risk_events_rule (rule_id, created_at),
  ADD KEY idx_game_user_risk_events_ip (ip_address, created_at),
  ADD KEY idx_game_user_risk_events_device (device_id, created_at);

CREATE TABLE game_risk_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name_zh VARCHAR(128) NOT NULL,
  name_en VARCHAR(128) NOT NULL,
  description_zh VARCHAR(1000) NOT NULL DEFAULT '',
  description_en VARCHAR(1000) NOT NULL DEFAULT '',
  rule_type VARCHAR(32) NOT NULL,
  risk_level VARCHAR(16) NOT NULL,
  priority INT NOT NULL DEFAULT 0,
  enabled TINYINT(1) NOT NULL DEFAULT 0,
  countries VARCHAR(500) NOT NULL DEFAULT '',
  currencies VARCHAR(128) NOT NULL DEFAULT '',
  channels VARCHAR(500) NOT NULL DEFAULT '',
  trigger_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_triggered_at DATETIME(3) NULL,
  start_at DATETIME(3) NULL,
  end_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_risk_rules_code (code),
  KEY idx_game_risk_rules_type_level (rule_type, risk_level),
  KEY idx_game_risk_rules_enabled_priority (enabled, priority),
  KEY idx_game_risk_rules_updated (updated_at),
  KEY idx_game_risk_rules_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_rule_conditions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  rule_id BIGINT UNSIGNED NOT NULL,
  field_name VARCHAR(64) NOT NULL,
  operator VARCHAR(32) NOT NULL,
  comparison_value VARCHAR(1000) NOT NULL DEFAULT '',
  window_seconds INT UNSIGNED NOT NULL DEFAULT 0,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_conditions_rule (rule_id, sort_order),
  KEY idx_game_risk_conditions_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_rule_actions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  rule_id BIGINT UNSIGNED NOT NULL,
  action_type VARCHAR(32) NOT NULL,
  parameter_value VARCHAR(500) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_actions_rule (rule_id, sort_order),
  KEY idx_game_risk_actions_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_rule_test_runs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  rule_id BIGINT UNSIGNED NOT NULL,
  input_facts JSON NOT NULL,
  result JSON NOT NULL,
  matched TINYINT(1) NOT NULL DEFAULT 0,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_test_runs_rule_time (rule_id, created_at),
  KEY idx_game_risk_test_runs_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_event_evidence (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id BIGINT UNSIGNED NOT NULL,
  evidence_type VARCHAR(64) NOT NULL,
  evidence_data JSON NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_evidence_event (event_id, created_at),
  KEY idx_game_risk_evidence_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_reviews (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  review_no VARCHAR(64) NOT NULL,
  business_type VARCHAR(32) NOT NULL,
  business_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  risk_level VARCHAR(16) NOT NULL,
  reason VARCHAR(500) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
  assigned_reviewer_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  assigned_reviewer VARCHAR(64) NOT NULL DEFAULT '',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  decision VARCHAR(32) NOT NULL DEFAULT '',
  internal_remark VARCHAR(1000) NOT NULL DEFAULT '',
  risk_tags VARCHAR(500) NOT NULL DEFAULT '',
  version INT UNSIGNED NOT NULL DEFAULT 1,
  completed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_risk_reviews_no (review_no),
  KEY idx_game_risk_reviews_status_level (status, risk_level, created_at),
  KEY idx_game_risk_reviews_business (business_type, business_id),
  KEY idx_game_risk_reviews_user (user_id, created_at),
  KEY idx_game_risk_reviews_reviewer (assigned_reviewer_id, status),
  KEY idx_game_risk_reviews_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_review_histories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  review_id BIGINT UNSIGNED NOT NULL,
  before_status VARCHAR(32) NOT NULL,
  after_status VARCHAR(32) NOT NULL,
  decision VARCHAR(32) NOT NULL DEFAULT '',
  reason VARCHAR(500) NOT NULL,
  internal_remark VARCHAR(1000) NOT NULL DEFAULT '',
  risk_tags VARCHAR(500) NOT NULL DEFAULT '',
  operator_id BIGINT UNSIGNED NOT NULL,
  operator_username VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_review_history_review (review_id, created_at),
  KEY idx_game_risk_review_history_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_internal_notes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  target_type VARCHAR(32) NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  content VARCHAR(2000) NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL,
  operator_username VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_notes_target (target_type, target_id, created_at),
  KEY idx_game_risk_notes_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_blacklists (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  blacklist_no VARCHAR(64) NOT NULL,
  blacklist_type VARCHAR(32) NOT NULL,
  value_hash CHAR(64) NOT NULL,
  value_masked VARCHAR(255) NOT NULL,
  value_last4 VARCHAR(8) NOT NULL DEFAULT '',
  reason VARCHAR(500) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'ENABLED',
  created_by BIGINT UNSIGNED NOT NULL,
  created_by_name VARCHAR(64) NOT NULL,
  expires_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_risk_blacklists_value (blacklist_type, value_hash),
  KEY idx_game_risk_blacklists_status_expiry (status, expires_at),
  KEY idx_game_risk_blacklists_created (created_at),
  KEY idx_game_risk_blacklists_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_blacklist_user_relations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  blacklist_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  relation_type VARCHAR(32) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_risk_blacklist_user (blacklist_id, user_id, relation_type),
  KEY idx_game_risk_blacklist_rel_user (user_id, created_at),
  KEY idx_game_risk_blacklist_rel_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_risk_blacklist_histories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  blacklist_id BIGINT UNSIGNED NOT NULL,
  before_status VARCHAR(16) NOT NULL,
  after_status VARCHAR(16) NOT NULL,
  reason VARCHAR(500) NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL,
  operator_username VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_risk_blacklist_history (blacklist_id, created_at),
  KEY idx_game_risk_blacklist_history_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO game_admin_permissions (code, name) VALUES
('risk:dashboard:view','View risk dashboard'),('risk:rule:list','List risk rules'),('risk:rule:create','Create risk rules'),('risk:rule:update','Update risk rules'),('risk:rule:delete','Delete risk rules'),('risk:rule:test','Test risk rules'),('risk:rule:status','Update risk rule status'),('risk:event:list','List risk events'),('risk:event:view','View risk events'),('risk:event:assign','Assign risk events'),('risk:event:review','Review risk events'),('risk:event:clear','Clear risk events'),('risk:event:restrict','Restrict users'),('risk:event:block','Block users'),('risk:event:note','Add risk notes'),('risk:review:list','List risk reviews'),('risk:review:view','View risk reviews'),('risk:review:claim','Claim risk reviews'),('risk:review:assign','Assign risk reviews'),('risk:review:decision','Decide risk reviews'),('risk:review:escalate','Escalate risk reviews'),('risk:association:list','List user associations'),('risk:association:view','View user associations'),('risk:blacklist:list','List blacklists'),('risk:blacklist:create','Create blacklists'),('risk:blacklist:update','Update blacklists'),('risk:blacklist:remove','Remove blacklists'),('risk:withdrawal:list','List risk withdrawals'),('risk:withdrawal:view','View risk withdrawals'),('risk:withdrawal:approve','Approve withdrawal risk review'),('risk:withdrawal:reject','Reject withdrawal risk review'),('risk:withdrawal:restrict','Restrict withdrawal'),('risk:operation-log:view','View risk operation logs');

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM game_admin_roles r CROSS JOIN game_admin_permissions p
WHERE r.code='SUPER_ADMIN' AND p.code LIKE 'risk:%';
