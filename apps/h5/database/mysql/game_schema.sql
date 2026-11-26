CREATE DATABASE IF NOT EXISTS game_casino DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE game_casino;

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS game_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_no VARCHAR(32) NOT NULL COMMENT 'Public user id',
  phone_country_code VARCHAR(8) NOT NULL DEFAULT '+92',
  phone VARCHAR(32) NOT NULL,
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  avatar_url VARCHAR(500) NOT NULL DEFAULT '',
  invite_code VARCHAR(32) NOT NULL DEFAULT '',
  referrer_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  vip_level INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 disabled, 3 risk locked',
  registered_ip VARCHAR(64) NOT NULL DEFAULT '',
  last_login_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_users_user_no (user_no),
  UNIQUE KEY uk_game_users_phone (phone_country_code, phone),
  UNIQUE KEY uk_game_users_invite_code (invite_code),
  KEY idx_game_users_referrer (referrer_user_id),
  KEY idx_game_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Members';

CREATE TABLE IF NOT EXISTS game_user_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  token_hash VARCHAR(255) NOT NULL,
  device_id VARCHAR(128) NOT NULL DEFAULT '',
  client_ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(500) NOT NULL DEFAULT '',
  expired_at DATETIME(3) NULL DEFAULT NULL,
  revoked_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_sessions_token_hash (token_hash),
  KEY idx_game_user_sessions_user_id (user_id),
  KEY idx_game_user_sessions_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Login sessions';

CREATE TABLE IF NOT EXISTS game_sms_verification_codes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  phone_country_code VARCHAR(8) NOT NULL DEFAULT '+92',
  phone VARCHAR(32) NOT NULL,
  code_hash VARCHAR(255) NOT NULL,
  scene VARCHAR(32) NOT NULL COMMENT 'login, register, bind_phone, reset_password',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 unused, 2 used, 3 expired',
  expired_at DATETIME(3) NOT NULL,
  used_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_sms_codes_phone_scene (phone_country_code, phone, scene),
  KEY idx_game_sms_codes_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SMS verification codes';

CREATE TABLE IF NOT EXISTS game_user_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  language VARCHAR(16) NOT NULL DEFAULT 'en',
  music_enabled TINYINT NOT NULL DEFAULT 1,
  push_enabled TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_settings_user_id (user_id),
  KEY idx_game_user_settings_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member settings';

CREATE TABLE IF NOT EXISTS game_user_devices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  device_id VARCHAR(128) NOT NULL,
  platform VARCHAR(32) NOT NULL DEFAULT 'h5',
  app_version VARCHAR(32) NOT NULL DEFAULT '',
  client_ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(500) NOT NULL DEFAULT '',
  first_seen_at DATETIME(3) NULL DEFAULT NULL,
  last_seen_at DATETIME(3) NULL DEFAULT NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 blocked',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_devices_device (user_id, device_id),
  KEY idx_game_user_devices_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member devices';

CREATE TABLE IF NOT EXISTS game_user_wallets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  cash_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  bonus_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  withdrawable_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  frozen_balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_wallets_user_id (user_id),
  KEY idx_game_user_wallets_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member wallet balances';

CREATE TABLE IF NOT EXISTS game_withdraw_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  account_type VARCHAR(16) NOT NULL COMMENT 'bank, wallet',
  bank_name VARCHAR(128) NOT NULL DEFAULT '',
  wallet_type VARCHAR(32) NOT NULL DEFAULT '',
  account_name VARCHAR(128) NOT NULL DEFAULT '',
  account_no VARCHAR(128) NOT NULL DEFAULT '',
  phone_country_code VARCHAR(8) NOT NULL DEFAULT '+92',
  phone VARCHAR(32) NOT NULL DEFAULT '',
  is_default TINYINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 disabled, 3 reviewing',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_withdraw_accounts_user_id (user_id),
  KEY idx_game_withdraw_accounts_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Withdrawal bank and e-wallet accounts';

CREATE TABLE IF NOT EXISTS game_payment_methods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  method_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  method_type VARCHAR(32) NOT NULL COMMENT 'deposit, withdraw, both',
  icon_url VARCHAR(500) NOT NULL DEFAULT '',
  min_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  max_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_payment_methods_code (method_code),
  KEY idx_game_payment_methods_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment methods';

CREATE TABLE IF NOT EXISTS game_payment_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  method_id BIGINT UNSIGNED NOT NULL,
  channel_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  fee_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
  min_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  max_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_payment_channels_code (channel_code),
  KEY idx_game_payment_channels_method_id (method_id),
  KEY idx_game_payment_channels_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment channels';

CREATE TABLE IF NOT EXISTS game_deposit_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  paid_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  method_code VARCHAR(64) NOT NULL DEFAULT '',
  channel_code VARCHAR(64) NOT NULL DEFAULT '',
  voucher_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 succeeded, 3 failed, 4 cancelled',
  paid_at DATETIME(3) NULL DEFAULT NULL,
  expired_at DATETIME(3) NULL DEFAULT NULL,
  fail_reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_deposit_orders_order_no (order_no),
  KEY idx_game_deposit_orders_user_id (user_id),
  KEY idx_game_deposit_orders_status (status),
  KEY idx_game_deposit_orders_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Deposit orders';

CREATE TABLE IF NOT EXISTS game_withdraw_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  withdraw_account_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  fee_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  actual_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  currency VARCHAR(8) NOT NULL DEFAULT 'PKR',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 succeeded, 3 failed, 4 returned, 5 cancelled, 6 reviewing',
  audit_status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 passed, 3 rejected',
  review_remark VARCHAR(255) NOT NULL DEFAULT '',
  submitted_at DATETIME(3) NULL DEFAULT NULL,
  completed_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_withdraw_orders_order_no (order_no),
  KEY idx_game_withdraw_orders_user_id (user_id),
  KEY idx_game_withdraw_orders_status (status),
  KEY idx_game_withdraw_orders_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Withdraw orders';

CREATE TABLE IF NOT EXISTS game_wallet_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  txn_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  wallet_type VARCHAR(32) NOT NULL COMMENT 'cash, bonus, withdrawable, frozen',
  direction VARCHAR(16) NOT NULL COMMENT 'credit, debit',
  category VARCHAR(64) NOT NULL COMMENT 'deposit, withdrawal, bet, payout, cash_reward, commission',
  source_type VARCHAR(64) NOT NULL DEFAULT '',
  source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  order_no VARCHAR(64) NOT NULL DEFAULT '',
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  balance_before DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  balance_after DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_wallet_transactions_txn_no (txn_no),
  KEY idx_game_wallet_transactions_user_id (user_id),
  KEY idx_game_wallet_transactions_category (category),
  KEY idx_game_wallet_transactions_created_at (created_at),
  KEY idx_game_wallet_transactions_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Wallet statements';

CREATE TABLE IF NOT EXISTS game_cash_audit_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  source_type VARCHAR(64) NOT NULL,
  source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  requirement_multiplier DECIMAL(8,2) NOT NULL DEFAULT 1.00,
  required_valid_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  completed_valid_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 completed, 3 cancelled',
  started_at DATETIME(3) NULL DEFAULT NULL,
  completed_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_cash_audit_user_id (user_id),
  KEY idx_game_cash_audit_status (status),
  KEY idx_game_cash_audit_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Cash audit requirements';

CREATE TABLE IF NOT EXISTS game_bonus_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  source_type VARCHAR(64) NOT NULL,
  source_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  requirement_multiplier DECIMAL(8,2) NOT NULL DEFAULT 1.00,
  required_valid_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  completed_valid_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 locked, 2 unlocked, 3 expired, 4 cancelled',
  issued_at DATETIME(3) NULL DEFAULT NULL,
  unlocked_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_bonus_records_order_no (order_no),
  KEY idx_game_bonus_records_user_id (user_id),
  KEY idx_game_bonus_records_status (status),
  KEY idx_game_bonus_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Bonus center records';

CREATE TABLE IF NOT EXISTS game_game_categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  category_key VARCHAR(32) NOT NULL,
  label VARCHAR(64) NOT NULL,
  short_label VARCHAR(16) NOT NULL DEFAULT '',
  icon_url VARCHAR(500) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_categories_key (category_key),
  KEY idx_game_categories_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Game categories';

CREATE TABLE IF NOT EXISTS game_game_providers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  provider_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  logo_url VARCHAR(500) NOT NULL DEFAULT '',
  api_base_url VARCHAR(500) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 maintenance, 3 disabled',
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_providers_code (provider_code),
  KEY idx_game_providers_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Game providers';

CREATE TABLE IF NOT EXISTS game_games (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  game_code VARCHAR(64) NOT NULL,
  provider_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  provider_game_code VARCHAR(128) NOT NULL DEFAULT '',
  name VARCHAR(128) NOT NULL,
  game_type VARCHAR(32) NOT NULL COMMENT 'slot, chance, sports, live, fish, table',
  cover_url VARCHAR(500) NOT NULL DEFAULT '',
  badge VARCHAR(32) NOT NULL DEFAULT '',
  player_count INT UNSIGNED NOT NULL DEFAULT 0,
  hot_score INT NOT NULL DEFAULT 0,
  launch_url VARCHAR(500) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 online, 2 maintenance, 3 offline',
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_games_code (game_code),
  KEY idx_game_games_provider_id (provider_id),
  KEY idx_game_games_game_type (game_type),
  KEY idx_game_games_status (status),
  KEY idx_game_games_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Game list';

CREATE TABLE IF NOT EXISTS game_game_category_relations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  game_id BIGINT UNSIGNED NOT NULL,
  category_id BIGINT UNSIGNED NOT NULL,
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_category_relations (game_id, category_id),
  KEY idx_game_category_relations_category (category_id),
  KEY idx_game_category_relations_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Game and category relations';

CREATE TABLE IF NOT EXISTS game_user_favorite_games (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  game_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_favorite_games (user_id, game_id),
  KEY idx_game_user_favorite_games_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Favorite games';

CREATE TABLE IF NOT EXISTS game_game_launch_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  game_id BIGINT UNSIGNED NOT NULL,
  provider_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  session_token VARCHAR(255) NOT NULL DEFAULT '',
  launch_status TINYINT NOT NULL DEFAULT 1 COMMENT '1 loading, 2 entered, 3 failed',
  client_ip VARCHAR(64) NOT NULL DEFAULT '',
  error_message VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_launch_logs_user_id (user_id),
  KEY idx_game_launch_logs_game_id (game_id),
  KEY idx_game_launch_logs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Game launch logs';

CREATE TABLE IF NOT EXISTS game_bet_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  bet_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  game_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  provider_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  game_type VARCHAR(32) NOT NULL DEFAULT '',
  bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  valid_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  payout_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  profit_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 not settled, 2 settled, 3 cancelled',
  bet_at DATETIME(3) NOT NULL,
  settled_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_bet_records_bet_no (bet_no),
  KEY idx_game_bet_records_user_id (user_id),
  KEY idx_game_bet_records_game_type (game_type),
  KEY idx_game_bet_records_bet_at (bet_at),
  KEY idx_game_bet_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Bet records';

CREATE TABLE IF NOT EXISTS game_bet_reports (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  report_date DATE NOT NULL,
  game_type VARCHAR(32) NOT NULL DEFAULT '',
  bet_count INT UNSIGNED NOT NULL DEFAULT 0,
  bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  valid_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  payout_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  profit_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_bet_reports_user_date_type (user_id, report_date, game_type),
  KEY idx_game_bet_reports_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Bet report aggregates';

CREATE TABLE IF NOT EXISTS game_banners (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  banner_code VARCHAR(64) NOT NULL,
  position VARCHAR(64) NOT NULL DEFAULT 'home',
  title VARCHAR(128) NOT NULL,
  image_url VARCHAR(500) NOT NULL DEFAULT '',
  link_type VARCHAR(32) NOT NULL DEFAULT 'none' COMMENT 'none, route, activity, external',
  link_url VARCHAR(500) NOT NULL DEFAULT '',
  activity_slug VARCHAR(128) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  start_at DATETIME(3) NULL DEFAULT NULL,
  end_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_banners_code (banner_code),
  KEY idx_game_banners_position (position),
  KEY idx_game_banners_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Home banners and download bars';

CREATE TABLE IF NOT EXISTS game_home_popup_dialogs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  popup_code VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL DEFAULT '',
  popup_type VARCHAR(32) NOT NULL DEFAULT 'activity',
  image_url VARCHAR(500) NOT NULL DEFAULT '',
  action_text VARCHAR(64) NOT NULL DEFAULT '',
  action_link VARCHAR(500) NOT NULL DEFAULT '',
  config_text MEDIUMTEXT NULL,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  start_at DATETIME(3) NULL DEFAULT NULL,
  end_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_home_popup_dialogs_code (popup_code),
  KEY idx_game_home_popup_dialogs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Home popups and guide dialogs';

CREATE TABLE IF NOT EXISTS game_home_feature_cards (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  feature_code VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL,
  image_url VARCHAR(500) NOT NULL DEFAULT '',
  background_style VARCHAR(500) NOT NULL DEFAULT '',
  link_type VARCHAR(32) NOT NULL DEFAULT 'activity',
  link_url VARCHAR(500) NOT NULL DEFAULT '',
  countdown_end_at DATETIME(3) NULL DEFAULT NULL,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_home_feature_cards_code (feature_code),
  KEY idx_game_home_feature_cards_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Home feature cards';

CREATE TABLE IF NOT EXISTS game_rebate_jackpots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  period_type VARCHAR(32) NOT NULL DEFAULT 'weekly',
  period_start DATETIME(3) NOT NULL,
  period_end DATETIME(3) NOT NULL,
  jackpot_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  digit_text VARCHAR(32) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 settled, 3 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_rebate_jackpots_period (period_type, period_start, period_end),
  KEY idx_game_rebate_jackpots_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Home rebate jackpot display';

CREATE TABLE IF NOT EXISTS game_activity_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_code VARCHAR(64) NOT NULL,
  slug VARCHAR(128) NOT NULL,
  kind VARCHAR(32) NOT NULL COMMENT 'deposit, spin, chest, red_packet, checkin, card',
  title VARCHAR(128) NOT NULL,
  subtitle VARCHAR(255) NOT NULL DEFAULT '',
  reward_text VARCHAR(128) NOT NULL DEFAULT '',
  max_reward_text VARCHAR(128) NOT NULL DEFAULT '',
  note VARCHAR(255) NOT NULL DEFAULT '',
  cta_text VARCHAR(64) NOT NULL DEFAULT '',
  hero_url VARCHAR(500) NOT NULL DEFAULT '',
  card_image_url VARCHAR(500) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled, 3 ended',
  sort INT NOT NULL DEFAULT 0,
  start_at DATETIME(3) NULL DEFAULT NULL,
  end_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_activity_events_code (activity_code),
  UNIQUE KEY uk_game_activity_events_slug (slug),
  KEY idx_game_activity_events_kind (kind),
  KEY idx_game_activity_events_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Offer and activity list';

CREATE TABLE IF NOT EXISTS game_activity_assets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  asset_key VARCHAR(64) NOT NULL,
  asset_url VARCHAR(500) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_activity_assets_key (activity_id, asset_key),
  KEY idx_game_activity_assets_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Activity visual assets';

CREATE TABLE IF NOT EXISTS game_activity_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(128) NOT NULL DEFAULT '',
  content_html MEDIUMTEXT NULL,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_activity_rules_activity_id (activity_id),
  KEY idx_game_activity_rules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Activity rule texts';

CREATE TABLE IF NOT EXISTS game_activity_reward_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  label VARCHAR(128) NOT NULL DEFAULT '',
  reward_value VARCHAR(128) NOT NULL DEFAULT '',
  bonus_text VARCHAR(128) NOT NULL DEFAULT '',
  min_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  max_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_cash DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_bonus DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  requirement_multiplier DECIMAL(8,2) NOT NULL DEFAULT 1.00,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_activity_reward_rules_activity_id (activity_id),
  KEY idx_game_activity_reward_rules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Activity reward tiers';

CREATE TABLE IF NOT EXISTS game_user_activity_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  action_type VARCHAR(64) NOT NULL DEFAULT 'claim',
  order_no VARCHAR(64) NOT NULL DEFAULT '',
  reward_cash DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_bonus DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 claimed, 3 rejected, 4 expired',
  claimed_at DATETIME(3) NULL DEFAULT NULL,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_activity_records_user_id (user_id),
  KEY idx_game_user_activity_records_activity_id (activity_id),
  KEY idx_game_user_activity_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member activity claims';

CREATE TABLE IF NOT EXISTS game_user_checkins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  checkin_date DATE NOT NULL,
  day_no INT UNSIGNED NOT NULL DEFAULT 1,
  streak_count INT UNSIGNED NOT NULL DEFAULT 1,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_wallet VARCHAR(32) NOT NULL DEFAULT 'bonus',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 checked, 2 claimed, 3 expired',
  claimed_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_user_checkins_date (user_id, checkin_date),
  KEY idx_game_user_checkins_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Daily check-in records';

CREATE TABLE IF NOT EXISTS game_invitation_wheel_rounds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  round_no VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL DEFAULT '',
  start_at DATETIME(3) NOT NULL,
  end_at DATETIME(3) NOT NULL,
  target_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  current_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 ended, 3 settled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_invitation_wheel_rounds_no (round_no),
  KEY idx_game_invitation_wheel_rounds_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Invitation wheel rounds';

CREATE TABLE IF NOT EXISTS game_invitation_wheel_prizes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  prize_code VARCHAR(64) NOT NULL,
  label VARCHAR(64) NOT NULL,
  reward_type VARCHAR(32) NOT NULL COMMENT 'cash, bonus, spin, chest, empty',
  min_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  max_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  weight INT UNSIGNED NOT NULL DEFAULT 0,
  icon_url VARCHAR(500) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_invitation_wheel_prizes_code (prize_code),
  KEY idx_game_invitation_wheel_prizes_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Invitation wheel prizes';

CREATE TABLE IF NOT EXISTS game_invitation_wheel_spins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  spin_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  prize_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reward_type VARCHAR(32) NOT NULL DEFAULT '',
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 won, 3 failed',
  spun_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_invitation_wheel_spins_no (spin_no),
  KEY idx_game_invitation_wheel_spins_user_id (user_id),
  KEY idx_game_invitation_wheel_spins_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Invitation wheel spin records';

CREATE TABLE IF NOT EXISTS game_invitation_chest_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chest_code VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL,
  required_invites INT UNSIGNED NOT NULL DEFAULT 0,
  reward_cash DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_bonus DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  icon_url VARCHAR(500) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_invitation_chest_configs_code (chest_code),
  KEY idx_game_invitation_chest_configs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Invitation chest tiers';

CREATE TABLE IF NOT EXISTS game_user_invitation_chests (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  chest_config_id BIGINT UNSIGNED NOT NULL,
  progress_invites INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 locked, 2 available, 3 opened, 4 expired',
  opened_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_invitation_chests_user_id (user_id),
  KEY idx_game_user_invitation_chests_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member invitation chests';

CREATE TABLE IF NOT EXISTS game_red_packet_rounds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  round_no VARCHAR(64) NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  title VARCHAR(128) NOT NULL DEFAULT '',
  start_at DATETIME(3) NOT NULL,
  end_at DATETIME(3) NOT NULL,
  total_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  packet_count INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 active, 3 ended',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_red_packet_rounds_no (round_no),
  KEY idx_game_red_packet_rounds_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Red packet rain rounds';

CREATE TABLE IF NOT EXISTS game_red_packet_claims (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  claim_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 claimed, 3 failed',
  claimed_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_red_packet_claims_no (claim_no),
  KEY idx_game_red_packet_claims_user_id (user_id),
  KEY idx_game_red_packet_claims_round_id (round_id),
  KEY idx_game_red_packet_claims_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Red packet rain claim records';

CREATE TABLE IF NOT EXISTS game_welfare_cards (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  card_code VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL,
  period_type VARCHAR(32) NOT NULL DEFAULT 'day',
  price_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_wallet VARCHAR(32) NOT NULL DEFAULT 'bonus',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_welfare_cards_code (card_code),
  KEY idx_game_welfare_cards_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Welfare cards';

CREATE TABLE IF NOT EXISTS game_user_welfare_cards (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  welfare_card_id BIGINT UNSIGNED NOT NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 expired, 3 cancelled',
  activated_at DATETIME(3) NULL DEFAULT NULL,
  expired_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_welfare_cards_user_id (user_id),
  KEY idx_game_user_welfare_cards_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member welfare cards';

CREATE TABLE IF NOT EXISTS game_invites (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  inviter_user_id BIGINT UNSIGNED NOT NULL,
  invitee_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  invite_code VARCHAR(32) NOT NULL DEFAULT '',
  invitee_phone VARCHAR(32) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 invited, 2 registered, 3 valid, 4 invalid',
  registered_at DATETIME(3) NULL DEFAULT NULL,
  first_deposit_at DATETIME(3) NULL DEFAULT NULL,
  valid_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_invites_inviter (inviter_user_id),
  KEY idx_game_invites_invitee (invitee_user_id),
  KEY idx_game_invites_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Referral records';

CREATE TABLE IF NOT EXISTS game_agent_commission_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  level_no INT UNSIGNED NOT NULL DEFAULT 1,
  min_referrals INT UNSIGNED NOT NULL DEFAULT 0,
  min_valid_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  commission_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
  bonus_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_agent_commission_rules_level (level_no),
  KEY idx_game_agent_commission_rules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent commission rules';

CREATE TABLE IF NOT EXISTS game_agent_commission_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  period_start DATETIME(3) NOT NULL,
  period_end DATETIME(3) NOT NULL,
  referrals_count INT UNSIGNED NOT NULL DEFAULT 0,
  valid_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  commission_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  bonus_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 settled, 3 rejected',
  settled_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_agent_commission_records_user_id (user_id),
  KEY idx_game_agent_commission_records_period (period_start, period_end),
  KEY idx_game_agent_commission_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent commission settlements';

CREATE TABLE IF NOT EXISTS game_ranking_periods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ranking_type VARCHAR(64) NOT NULL DEFAULT 'commission',
  title VARCHAR(128) NOT NULL DEFAULT '',
  ranking_date DATE NULL DEFAULT NULL,
  period_start DATETIME(3) NOT NULL,
  period_end DATETIME(3) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 active, 2 settled, 3 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_ranking_periods_type_date (ranking_type, ranking_date),
  KEY idx_game_ranking_periods_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Ranking periods';

CREATE TABLE IF NOT EXISTS game_ranking_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ranking_period_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  rank_no INT UNSIGNED NOT NULL DEFAULT 0,
  display_name VARCHAR(64) NOT NULL DEFAULT '',
  avatar_url VARCHAR(500) NOT NULL DEFAULT '',
  total_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending, 2 rewarded, 3 invalid',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_ranking_records_period_rank (ranking_period_id, rank_no),
  KEY idx_game_ranking_records_user_id (user_id),
  KEY idx_game_ranking_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Ranking records';

CREATE TABLE IF NOT EXISTS game_vip_levels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  level_no INT UNSIGNED NOT NULL,
  name VARCHAR(64) NOT NULL DEFAULT '',
  icon_url VARCHAR(500) NOT NULL DEFAULT '',
  level_up_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  weekly_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  monthly_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  daily_withdraw_count INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_vip_levels_level_no (level_no),
  KEY idx_game_vip_levels_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VIP levels';

CREATE TABLE IF NOT EXISTS game_vip_requirements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  vip_level_id BIGINT UNSIGNED NOT NULL,
  requirement_type VARCHAR(32) NOT NULL COMMENT 'level_up, retain',
  target_level_no INT UNSIGNED NOT NULL DEFAULT 0,
  period_days INT UNSIGNED NOT NULL DEFAULT 0,
  deposit_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  valid_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_vip_requirements_level_id (vip_level_id),
  KEY idx_game_vip_requirements_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VIP level-up and retain requirements';

CREATE TABLE IF NOT EXISTS game_user_vip_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  old_level_no INT UNSIGNED NOT NULL DEFAULT 0,
  new_level_no INT UNSIGNED NOT NULL DEFAULT 0,
  reason VARCHAR(64) NOT NULL DEFAULT '',
  deposit_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  valid_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  changed_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_user_vip_logs_user_id (user_id),
  KEY idx_game_user_vip_logs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Member VIP change logs';

CREATE TABLE IF NOT EXISTS game_support_categories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  category_code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_support_categories_code (category_code),
  KEY idx_game_support_categories_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='FAQ categories';

CREATE TABLE IF NOT EXISTS game_support_questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  category_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  question VARCHAR(255) NOT NULL,
  answer_html MEDIUMTEXT NULL,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_support_questions_category_id (category_id),
  KEY idx_game_support_questions_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='FAQ questions and answers';

CREATE TABLE IF NOT EXISTS game_support_tickets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ticket_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  category_code VARCHAR(64) NOT NULL DEFAULT '',
  subject VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 open, 2 pending, 3 closed',
  last_message_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_support_tickets_no (ticket_no),
  KEY idx_game_support_tickets_user_id (user_id),
  KEY idx_game_support_tickets_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Support tickets';

CREATE TABLE IF NOT EXISTS game_support_messages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ticket_id BIGINT UNSIGNED NOT NULL,
  sender_type VARCHAR(32) NOT NULL COMMENT 'user, admin, system',
  sender_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  content_type VARCHAR(32) NOT NULL DEFAULT 'text',
  content MEDIUMTEXT NULL,
  read_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_support_messages_ticket_id (ticket_id),
  KEY idx_game_support_messages_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Support ticket messages';

CREATE TABLE IF NOT EXISTS game_notices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  notice_code VARCHAR(64) NOT NULL,
  notice_type VARCHAR(32) NOT NULL DEFAULT 'system',
  title VARCHAR(128) NOT NULL,
  content_html MEDIUMTEXT NULL,
  image_url VARCHAR(500) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  start_at DATETIME(3) NULL DEFAULT NULL,
  end_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_notices_code (notice_code),
  KEY idx_game_notices_type (notice_type),
  KEY idx_game_notices_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Notices and messages';

CREATE TABLE IF NOT EXISTS game_cms_pages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  page_code VARCHAR(64) NOT NULL,
  title VARCHAR(128) NOT NULL,
  content_html MEDIUMTEXT NULL,
  version_no INT UNSIGNED NOT NULL DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 published, 2 draft, 3 disabled',
  effective_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_cms_pages_code_version (page_code, version_no),
  KEY idx_game_cms_pages_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Static pages such as about and agreement';

CREATE TABLE IF NOT EXISTS game_app_versions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  platform VARCHAR(32) NOT NULL DEFAULT 'android',
  version_name VARCHAR(32) NOT NULL,
  version_code INT UNSIGNED NOT NULL DEFAULT 0,
  download_url VARCHAR(500) NOT NULL DEFAULT '',
  force_update TINYINT NOT NULL DEFAULT 0,
  release_note VARCHAR(500) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_app_versions_platform_code (platform, version_code),
  KEY idx_game_app_versions_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='App download versions';

CREATE TABLE IF NOT EXISTS game_app_download_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  device_id VARCHAR(128) NOT NULL DEFAULT '',
  app_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  reward_wallet VARCHAR(32) NOT NULL DEFAULT 'bonus',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 downloaded, 2 rewarded, 3 rejected',
  rewarded_at DATETIME(3) NULL DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_game_app_download_records_user_id (user_id),
  KEY idx_game_app_download_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='App download reward records';

CREATE TABLE IF NOT EXISTS game_system_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  config_key VARCHAR(128) NOT NULL,
  config_value MEDIUMTEXT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 enabled, 2 disabled',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_game_system_configs_key (config_key),
  KEY idx_game_system_configs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System configuration';
