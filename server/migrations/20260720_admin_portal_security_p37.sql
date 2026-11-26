DROP PROCEDURE IF EXISTS admin_portal_add_column;
DROP PROCEDURE IF EXISTS admin_portal_add_index;
DELIMITER $$
CREATE PROCEDURE admin_portal_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE admin_portal_statement FROM @statement_value;
    EXECUTE admin_portal_statement;
    DEALLOCATE PREPARE admin_portal_statement;
  END IF;
END$$
CREATE PROCEDURE admin_portal_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value
  ) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE admin_portal_statement FROM @statement_value;
    EXECUTE admin_portal_statement;
    DEALLOCATE PREPARE admin_portal_statement;
  END IF;
END$$
DELIMITER ;

CALL admin_portal_add_column('game_admin_accounts','must_change_password','TINYINT(1) NOT NULL DEFAULT 0 AFTER `last_password_changed_at`');
CALL admin_portal_add_column('game_admin_accounts','preferred_language','VARCHAR(16) NOT NULL DEFAULT ''zh-CN'' AFTER `must_change_password`');
CALL admin_portal_add_column('game_admin_accounts','timezone','VARCHAR(64) NOT NULL DEFAULT ''Asia/Kuala_Lumpur'' AFTER `preferred_language`');
CALL admin_portal_add_column('game_admin_sessions','last_active_at','DATETIME(3) NULL AFTER `expired_at`');
CALL admin_portal_add_index('game_admin_sessions','idx_admin_session_account_active','(`admin_id`,`revoked_at`,`expired_at`)');

DROP PROCEDURE admin_portal_add_column;
DROP PROCEDURE admin_portal_add_index;

CREATE TABLE IF NOT EXISTS `game_admin_password_history` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_admin_password_history_account` (`admin_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `game_admin_mfa_recovery_codes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT UNSIGNED NOT NULL,
  `code_hash` CHAR(64) NOT NULL,
  `used_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_mfa_recovery_code` (`admin_id`,`code_hash`),
  KEY `idx_admin_mfa_recovery_active` (`admin_id`,`used_at`,`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
SELECT r.id, p.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.enabled=1 AND p.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
  AND (
    (p.code NOT LIKE 'system:%' AND p.code NOT IN (
      'game:provider:list','game:provider:view','game:provider:update','game:provider:test','game:provider:sync','game:sync-log:view',
      'bet:callback-log:list','bet:callback-log:view','bet:reconciliation:list','bet:reconciliation:view','bet:reconciliation:generate','bet:reconciliation:resolve',
      'report:provider:view'
    ))
    OR p.code IN ('system:site:list','system:site:view','system:site:update','system:site:config:preview','system:site:config:publish','system:domain:list','system:domain:view')
  )
ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at);

UPDATE game_admin_role_permissions rp
JOIN game_admin_roles r ON r.id=rp.role_id AND r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
JOIN game_admin_permissions p ON p.id=rp.permission_id AND p.deleted_at IS NULL
SET rp.deleted_at=NOW(3), rp.updated_at=NOW(3)
WHERE rp.deleted_at IS NULL
  AND (
    (p.code LIKE 'system:%' AND p.code NOT IN (
      'system:site:list','system:site:view','system:site:update','system:site:config:preview','system:site:config:publish','system:domain:list','system:domain:view'
    ))
    OR p.code IN (
      'game:provider:list','game:provider:view','game:provider:update','game:provider:test','game:provider:sync','game:sync-log:view',
      'bet:callback-log:list','bet:callback-log:view','bet:reconciliation:list','bet:reconciliation:view','bet:reconciliation:generate','bet:reconciliation:resolve',
      'report:provider:view'
    )
  );

INSERT INTO game_admin_role_menus (role_id, menu_id, created_at, updated_at, deleted_at)
SELECT r.id, m.id, NOW(3), NOW(3), NULL
FROM game_admin_roles r
JOIN game_admin_menus m ON m.deleted_at IS NULL
WHERE r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
  AND m.code IN (
    'dashboard','users','user-list','login-records','device-management','risk-users',
    'funds','wallet-accounts','wallet-ledgers','manual-adjustments','deposit-orders','withdrawal-orders','financial-reconciliation','deposit-methods','deposit-channels','deposit-amount-options',
    'games','game-list','game-categories','lobby-sections','game-launch-records',
    'bets','bet-records','rollback-records','abnormal-bets',
    'operations','asset-library','banners','announcements','activities','vip','push-messages','popups','music','social-links','h5-navigation','vouchers',
    'risk','risk-dashboard','risk-rules','risk-events','review-queue','user-associations','blacklist','withdrawal-risk','risk-operation-logs',
    'agents','agent-list','agent-hierarchy','channel-management','commission-plans','commission-records','agent-settlements','agent-wallets','agent-operation-logs',
    'reports','financial-report','game-report','user-report','channel-report','agent-report','promotion-report','export-tasks',
    'system','sites','brand-config','system-audit-logs','system-login-logs'
  )
ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at);

UPDATE game_admin_role_menus rm
JOIN game_admin_roles r ON r.id=rm.role_id AND r.code='tenant-owner' AND r.tenant_id>0 AND r.deleted_at IS NULL
JOIN game_admin_menus m ON m.id=rm.menu_id AND m.deleted_at IS NULL
SET rm.deleted_at=NOW(3), rm.updated_at=NOW(3)
WHERE rm.deleted_at IS NULL
  AND m.code IN (
    'create-brand','tenants','game-sync','provider-callback-logs','provider-reconciliation','provider-report',
    'admins','roles','permissions','departments','menus','system-config','country-config','scheduled-tasks','system-monitor'
  );
