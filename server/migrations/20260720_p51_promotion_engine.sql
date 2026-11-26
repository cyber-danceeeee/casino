ALTER TABLE game_activity_events
  ADD COLUMN trigger_type varchar(48) NOT NULL DEFAULT 'MANUAL_CLAIM' AFTER promotion_type,
  ADD COLUMN claim_mode varchar(24) NOT NULL DEFAULT 'MANUAL' AFTER trigger_type,
  ADD COLUMN stacking_group varchar(64) NOT NULL DEFAULT '' AFTER claim_mode,
  ADD COLUMN priority int NOT NULL DEFAULT 0 AFTER stacking_group,
  ADD COLUMN languages varchar(128) NOT NULL DEFAULT 'zh-CN,en-US' AFTER currencies,
  ADD KEY idx_promotion_trigger_scope (tenant_id, site_id, trigger_type, status_code, start_at, end_at);

ALTER TABLE game_promotion_reward_rules
  ADD COLUMN calculation_type varchar(24) NOT NULL DEFAULT 'FIXED' AFTER reward_type,
  ADD COLUMN percentage_rate decimal(8,4) NOT NULL DEFAULT 0.0000 AFTER reward_amount,
  ADD COLUMN expiration_days int unsigned NOT NULL DEFAULT 0 AFTER maximum_reward;

-- Existing CMS promotions require an explicit operator review before automatic rewards are enabled.
UPDATE game_activity_events
SET trigger_type = CASE
    WHEN promotion_type IN ('FIRST_DEPOSIT_BONUS', 'DEPOSIT_BONUS') THEN 'DEPOSIT_SUCCEEDED'
    WHEN promotion_type = 'REGISTRATION_BONUS' THEN 'REGISTRATION_COMPLETED'
    WHEN promotion_type = 'CASHBACK' THEN 'GAME_BET_SETTLED'
    ELSE 'MANUAL_CLAIM'
  END,
  claim_mode = 'MANUAL'
WHERE trigger_type = 'MANUAL_CLAIM' AND claim_mode = 'MANUAL';

CREATE TABLE IF NOT EXISTS game_promotion_claims (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  tenant_id bigint unsigned NOT NULL,
  site_id bigint unsigned NOT NULL,
  promotion_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  trigger_type varchar(48) NOT NULL,
  trigger_event_id varchar(64) NOT NULL,
  trigger_order_no varchar(128) NOT NULL DEFAULT '',
  idempotency_key varchar(191) NOT NULL,
  currency varchar(8) NOT NULL,
  base_amount decimal(18,2) NOT NULL DEFAULT 0.00,
  reward_amount decimal(18,2) NOT NULL DEFAULT 0.00,
  status varchar(32) NOT NULL DEFAULT 'PROCESSING',
  failure_code varchar(64) NOT NULL DEFAULT '',
  failure_reason varchar(500) NOT NULL DEFAULT '',
  trace_id varchar(64) NOT NULL DEFAULT '',
  correlation_id varchar(64) NOT NULL DEFAULT '',
  processed_at datetime(3) DEFAULT NULL,
  created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_promotion_claim_idempotency (idempotency_key),
  KEY idx_promotion_claim_scope (tenant_id, site_id, promotion_id, status, created_at),
  KEY idx_promotion_claim_user (tenant_id, site_id, user_id, promotion_id, status)
);

CREATE TABLE IF NOT EXISTS game_promotion_claim_rewards (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  tenant_id bigint unsigned NOT NULL,
  site_id bigint unsigned NOT NULL,
  claim_id bigint unsigned NOT NULL,
  promotion_id bigint unsigned NOT NULL,
  reward_rule_id bigint unsigned NOT NULL,
  reward_type varchar(32) NOT NULL,
  calculation_type varchar(24) NOT NULL,
  currency varchar(8) NOT NULL,
  base_amount decimal(18,2) NOT NULL DEFAULT 0.00,
  reward_amount decimal(18,2) NOT NULL DEFAULT 0.00,
  wagering_multiplier decimal(8,2) NOT NULL DEFAULT 0.00,
  wallet_transaction_id bigint unsigned NOT NULL DEFAULT 0,
  bonus_record_id bigint unsigned NOT NULL DEFAULT 0,
  status varchar(32) NOT NULL DEFAULT 'COMPLETED',
  created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_promotion_claim_reward (claim_id, reward_rule_id),
  KEY idx_promotion_reward_scope (tenant_id, site_id, promotion_id, created_at),
  KEY idx_promotion_reward_wallet (wallet_transaction_id)
);
