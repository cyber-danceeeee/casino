SET @mfa_column_exists = (
  SELECT COUNT(*)
  FROM `information_schema`.`columns`
  WHERE `table_schema` = DATABASE()
    AND `table_name` = 'game_admin_accounts'
    AND `column_name` = 'mfa_secret_encrypted'
);
SET @mfa_column_sql = IF(
  @mfa_column_exists = 0,
  'ALTER TABLE `game_admin_accounts` ADD COLUMN `mfa_secret_encrypted` TEXT NULL AFTER `mfa_secret_hash`',
  'SELECT 1'
);
PREPARE mfa_column_statement FROM @mfa_column_sql;
EXECUTE mfa_column_statement;
DEALLOCATE PREPARE mfa_column_statement;

CREATE TABLE IF NOT EXISTS `game_admin_mfa_challenges` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `admin_id` bigint unsigned NOT NULL,
  `token_hash` char(64) NOT NULL,
  `purpose` varchar(16) NOT NULL,
  `secret_encrypted` text NULL,
  `attempts` int unsigned NOT NULL DEFAULT 0,
  `expires_at` datetime(3) NOT NULL,
  `consumed_at` datetime(3) NULL,
  `ip_address` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_mfa_challenge_token` (`token_hash`),
  KEY `idx_admin_mfa_challenge_account` (`admin_id`,`purpose`,`expires_at`),
  KEY `idx_admin_mfa_challenge_expiry` (`expires_at`,`consumed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
