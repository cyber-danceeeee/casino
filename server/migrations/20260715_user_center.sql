ALTER TABLE game_users
  ADD COLUMN username VARCHAR(64) NOT NULL DEFAULT '' AFTER user_no,
  ADD COLUMN email VARCHAR(190) NOT NULL DEFAULT '' AFTER username,
  ADD COLUMN country_code VARCHAR(8) NOT NULL DEFAULT 'PK' AFTER status,
  ADD COLUMN language VARCHAR(16) NOT NULL DEFAULT 'en' AFTER country_code,
  ADD COLUMN registration_channel VARCHAR(64) NOT NULL DEFAULT 'h5' AFTER language,
  ADD COLUMN last_login_ip VARCHAR(64) NOT NULL DEFAULT '' AFTER registered_ip,
  ADD COLUMN login_enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER last_login_ip,
  ADD COLUMN game_access_enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER login_enabled,
  ADD COLUMN deposit_enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER game_access_enabled,
  ADD COLUMN withdrawal_enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER deposit_enabled,
  ADD COLUMN risk_review_required TINYINT(1) NOT NULL DEFAULT 0 AFTER withdrawal_enabled,
  ADD COLUMN failed_login_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER risk_review_required,
  ADD COLUMN last_password_changed_at DATETIME(3) NULL AFTER failed_login_count,
  ADD KEY idx_game_users_username (username),
  ADD KEY idx_game_users_email (email),
  ADD KEY idx_game_users_status_country (status, country_code),
  ADD KEY idx_game_users_created_at (created_at),
  ADD KEY idx_game_users_last_login_at (last_login_at),
  ADD KEY idx_game_users_registered_ip (registered_ip),
  ADD KEY idx_game_users_last_login_ip (last_login_ip);

ALTER TABLE game_user_devices
  ADD COLUMN device_model VARCHAR(128) NOT NULL DEFAULT '' AFTER platform,
  ADD COLUMN os VARCHAR(128) NOT NULL DEFAULT '' AFTER device_model,
  ADD COLUMN browser VARCHAR(128) NOT NULL DEFAULT '' AFTER os,
  ADD COLUMN fingerprint_hash CHAR(64) NOT NULL DEFAULT '' AFTER app_version,
  ADD COLUMN first_ip VARCHAR(64) NOT NULL DEFAULT '' AFTER fingerprint_hash,
  ADD COLUMN last_ip VARCHAR(64) NOT NULL DEFAULT '' AFTER first_ip,
  ADD COLUMN trust_status TINYINT NOT NULL DEFAULT 1 AFTER status,
  ADD COLUMN risk_status TINYINT NOT NULL DEFAULT 1 AFTER trust_status,
  ADD COLUMN login_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER risk_status,
  ADD KEY idx_game_user_devices_device_id (device_id),
  ADD KEY idx_game_user_devices_fingerprint (fingerprint_hash),
  ADD KEY idx_game_user_devices_last_seen (last_seen_at),
  ADD KEY idx_game_user_devices_status_risk (status, risk_status);

CREATE TABLE game_user_login_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  login_identifier VARCHAR(190) NOT NULL DEFAULT '',
  success TINYINT(1) NOT NULL DEFAULT 0,
  failure_reason VARCHAR(128) NOT NULL DEFAULT '',
  ip_address VARCHAR(64) NOT NULL DEFAULT '',
  country VARCHAR(8) NOT NULL DEFAULT '',
  region VARCHAR(128) NOT NULL DEFAULT '',
  device_id VARCHAR(128) NOT NULL DEFAULT '',
  device_type VARCHAR(64) NOT NULL DEFAULT '',
  platform VARCHAR(32) NOT NULL DEFAULT '',
  os VARCHAR(128) NOT NULL DEFAULT '',
  browser VARCHAR(128) NOT NULL DEFAULT '',
  app_version VARCHAR(32) NOT NULL DEFAULT '',
  user_agent VARCHAR(500) NOT NULL DEFAULT '',
  occurred_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_login_records_user_time (user_id, occurred_at),
  KEY idx_game_login_records_ip_time (ip_address, occurred_at),
  KEY idx_game_login_records_device_time (device_id, occurred_at),
  KEY idx_game_login_records_success_time (success, occurred_at),
  KEY idx_game_login_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_admin_notes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  content VARCHAR(1000) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_admin_notes_user_time (user_id, created_at),
  KEY idx_game_user_admin_notes_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_risk_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  risk_level VARCHAR(16) NOT NULL DEFAULT 'LOW',
  risk_status VARCHAR(32) NOT NULL DEFAULT 'OBSERVED',
  trigger_count INT UNSIGNED NOT NULL DEFAULT 0,
  associated_accounts INT UNSIGNED NOT NULL DEFAULT 0,
  assigned_reviewer_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  assigned_reviewer VARCHAR(64) NOT NULL DEFAULT '',
  latest_reason VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_risk_profiles_user (user_id),
  KEY idx_game_user_risk_level_status (risk_level, risk_status),
  KEY idx_game_user_risk_updated (updated_at),
  KEY idx_game_user_risk_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_risk_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  risk_level VARCHAR(16) NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
  description VARCHAR(1000) NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_risk_events_user_time (user_id, created_at),
  KEY idx_game_user_risk_events_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_status_histories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  change_type VARCHAR(32) NOT NULL,
  before_value VARCHAR(64) NOT NULL DEFAULT '',
  after_value VARCHAR(64) NOT NULL DEFAULT '',
  reason VARCHAR(500) NOT NULL DEFAULT '',
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_status_histories_user_time (user_id, created_at),
  KEY idx_game_user_status_histories_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_tags (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  tag_type VARCHAR(16) NOT NULL DEFAULT 'INTERNAL',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_tags_name_type (name, tag_type),
  KEY idx_game_user_tags_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_user_tag_relations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  tag_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_tag_relations (user_id, tag_id),
  KEY idx_game_user_tag_relations_tag (tag_id),
  KEY idx_game_user_tag_relations_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE game_admin_operation_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  permission VARCHAR(128) NOT NULL,
  action VARCHAR(128) NOT NULL,
  target_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  target_device_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  before_state JSON NULL,
  after_state JSON NULL,
  reason VARCHAR(500) NOT NULL DEFAULT '',
  ip_address VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(500) NOT NULL DEFAULT '',
  correlation_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_game_admin_logs_operator_time (operator_id, created_at),
  KEY idx_game_admin_logs_target_user_time (target_user_id, created_at),
  KEY idx_game_admin_logs_target_device_time (target_device_id, created_at),
  KEY idx_game_admin_logs_correlation (correlation_id),
  KEY idx_game_admin_logs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
