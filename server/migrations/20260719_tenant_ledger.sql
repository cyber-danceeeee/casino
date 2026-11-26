CREATE TABLE IF NOT EXISTS ledger_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_no VARCHAR(64) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  account_type VARCHAR(64) NOT NULL,
  normal_balance VARCHAR(8) NOT NULL DEFAULT 'DEBIT',
  balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  available_balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  frozen_balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_ledger_account_no (account_no),
  UNIQUE KEY uk_ledger_account_scope_type (tenant_id, site_id, currency, account_type),
  KEY idx_ledger_account_scope (tenant_id, site_id, currency, status),
  KEY idx_ledger_account_deleted (deleted_at),
  CONSTRAINT chk_ledger_account_balance_nonnegative CHECK (balance >= 0 AND available_balance >= 0 AND frozen_balance >= 0),
  CONSTRAINT chk_ledger_account_normal_balance CHECK (normal_balance IN ('DEBIT','CREDIT'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ledger_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  transaction_no VARCHAR(64) NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  business_type VARCHAR(64) NOT NULL DEFAULT '',
  business_order_no VARCHAR(128) NOT NULL DEFAULT '',
  idempotency_key VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'POSTED',
  debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  reversal_of_transaction_id BIGINT UNSIGNED NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_username VARCHAR(64) NOT NULL DEFAULT '',
  approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reason VARCHAR(1000) NOT NULL DEFAULT '',
  metadata JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_ledger_transaction_no (transaction_no),
  UNIQUE KEY uk_ledger_transaction_idempotency (idempotency_key),
  UNIQUE KEY uk_ledger_transaction_reversal_once (reversal_of_transaction_id),
  KEY idx_ledger_transaction_scope_time (tenant_id, site_id, created_at, id),
  KEY idx_ledger_transaction_business (business_type, business_order_no),
  KEY idx_ledger_transaction_status (status, created_at),
  CONSTRAINT chk_ledger_transaction_balanced CHECK (debit_amount = credit_amount),
  CONSTRAINT chk_ledger_transaction_currency CHECK (currency <> ''),
  CONSTRAINT chk_ledger_transaction_status CHECK (status IN ('POSTED','FAILED'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ledger_entries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  transaction_id BIGINT UNSIGNED NOT NULL,
  entry_no SMALLINT UNSIGNED NOT NULL,
  account_id BIGINT UNSIGNED NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  account_type VARCHAR(64) NOT NULL,
  direction VARCHAR(8) NOT NULL,
  amount DECIMAL(20,2) NOT NULL,
  balance_before DECIMAL(20,2) NOT NULL,
  balance_after DECIMAL(20,2) NOT NULL,
  memo VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_ledger_entry_transaction_no (transaction_id, entry_no),
  KEY idx_ledger_entry_account_time (account_id, created_at),
  KEY idx_ledger_entry_scope_time (tenant_id, site_id, created_at),
  CONSTRAINT chk_ledger_entry_direction CHECK (direction IN ('DEBIT','CREDIT')),
  CONSTRAINT chk_ledger_entry_amount_positive CHECK (amount > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ledger_balance_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id BIGINT UNSIGNED NOT NULL,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  account_type VARCHAR(64) NOT NULL,
  snapshot_date DATE NOT NULL,
  balance DECIMAL(20,2) NOT NULL,
  available_balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  frozen_balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  transaction_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_ledger_snapshot_account_transaction (account_id, transaction_id),
  KEY idx_ledger_snapshot_scope_date (tenant_id, site_id, snapshot_date, currency),
  KEY idx_ledger_snapshot_account_date (account_id, snapshot_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tenant_credit_limits (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant_id BIGINT UNSIGNED NOT NULL,
  site_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL,
  credit_limit DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  used_credit DECIMAL(20,2) NOT NULL DEFAULT 0.00,
  status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  last_transaction_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_credit_limit_scope (tenant_id, site_id, currency),
  KEY idx_tenant_credit_limit_status (tenant_id, site_id, status, deleted_at),
  CONSTRAINT chk_tenant_credit_limit_nonnegative CHECK (credit_limit >= 0 AND used_credit >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TRIGGER IF EXISTS trg_ledger_accounts_balance_guard;
DROP TRIGGER IF EXISTS trg_ledger_transactions_no_update;
DROP TRIGGER IF EXISTS trg_ledger_transactions_no_delete;
DROP TRIGGER IF EXISTS trg_ledger_entries_no_update;
DROP TRIGGER IF EXISTS trg_ledger_entries_no_delete;
DELIMITER $$
CREATE TRIGGER trg_ledger_accounts_balance_guard
BEFORE UPDATE ON ledger_accounts
FOR EACH ROW
BEGIN
  IF COALESCE(@ledger_posting, 0) <> 1
     AND (OLD.balance <> NEW.balance
       OR OLD.available_balance <> NEW.available_balance
       OR OLD.frozen_balance <> NEW.frozen_balance) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'ledger account balances must be updated through ledger service';
  END IF;
END$$
CREATE TRIGGER trg_ledger_transactions_no_update
BEFORE UPDATE ON ledger_transactions
FOR EACH ROW
BEGIN
  IF OLD.status = 'POSTED' THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'posted ledger transactions are immutable';
  END IF;
END$$
CREATE TRIGGER trg_ledger_transactions_no_delete
BEFORE DELETE ON ledger_transactions
FOR EACH ROW
BEGIN
  IF OLD.status = 'POSTED' THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'posted ledger transactions cannot be deleted';
  END IF;
END$$
CREATE TRIGGER trg_ledger_entries_no_update
BEFORE UPDATE ON ledger_entries
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'ledger entries are immutable';
END$$
CREATE TRIGGER trg_ledger_entries_no_delete
BEFORE DELETE ON ledger_entries
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'ledger entries cannot be deleted';
END$$
DELIMITER ;

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:tenant:ledger:view','View tenant ledger','查看租户账本','View tenant ledger','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:ledger:deposit','Post tenant deposit','租户保证金入账','Post tenant deposit','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:ledger:credit','Set tenant credit limit','设置租户授信','Set tenant credit limit','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:ledger:freeze','Freeze tenant funds','冻结/解冻租户资金','Freeze tenant funds','system','API',1,1,NOW(3),NOW(3)),
('system:tenant:ledger:reverse','Reverse tenant ledger transaction','租户账本冲正','Reverse tenant ledger transaction','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'system:tenant:ledger:view',
  'system:tenant:ledger:deposit',
  'system:tenant:ledger:credit',
  'system:tenant:ledger:freeze',
  'system:tenant:ledger:reverse'
)
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;
