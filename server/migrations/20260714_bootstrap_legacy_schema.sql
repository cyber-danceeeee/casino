-- The original base schema omitted this historical table, but the July 15
-- operations migration still alters it. July 20 replaces and drops it.
-- Keep this compatibility table empty; current popup data uses popup campaigns.
CREATE TABLE IF NOT EXISTS game_home_popup_dialogs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(128) NOT NULL DEFAULT '',
  popup_type VARCHAR(32) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- The legacy seed command created this table dynamically before migrations.
-- Define its pre-July-16 shape so a clean deployment needs no seed command.
CREATE TABLE IF NOT EXISTS game_deposit_voucher_rewards (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  voucher_id BIGINT UNSIGNED NOT NULL,
  amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  bonus_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  cash_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  sort INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_voucher_amount (voucher_id, amount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE game_games
  ADD COLUMN platform_category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER category,
  ADD COLUMN platform_category_code VARCHAR(64) NOT NULL DEFAULT '' AFTER platform_category_id;
