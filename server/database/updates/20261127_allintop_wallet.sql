-- Additive migration: older application images remain compatible for rollback.
ALTER TABLE game_provider_transactions
 ADD COLUMN provider_history_id VARCHAR(32) NOT NULL DEFAULT '',
 ADD INDEX idx_provider_history (provider_code,user_id,provider_history_id,status);
CREATE TABLE game_provider_wallet_histories (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 site_id BIGINT UNSIGNED NOT NULL DEFAULT 1,
 provider_code VARCHAR(64) NOT NULL,
 user_id BIGINT UNSIGNED NOT NULL,
 provider_history_id VARCHAR(32) NOT NULL,
 provider_round_id VARCHAR(128) NOT NULL,
 provider_game_id VARCHAR(128) NOT NULL,
 status VARCHAR(16) NOT NULL,
 bet_minor BIGINT NOT NULL DEFAULT 0,
 win_minor BIGINT NOT NULL DEFAULT 0,
 transaction_count INT UNSIGNED NOT NULL DEFAULT 0,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 UNIQUE KEY provider_history(provider_code,user_id,provider_history_id),
 INDEX provider_round(provider_code,user_id,provider_round_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
CREATE TABLE game_provider_wallet_receipts (
 provider_code VARCHAR(64) NOT NULL,
 transaction_code VARCHAR(128) NOT NULL,
 user_id BIGINT UNSIGNED NOT NULL,
 provider_history_id VARCHAR(32) NOT NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 PRIMARY KEY(provider_code,transaction_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
