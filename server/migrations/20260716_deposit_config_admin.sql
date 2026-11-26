ALTER TABLE game_deposit_vouchers
  ADD COLUMN title_zh VARCHAR(128) NOT NULL DEFAULT '' AFTER title,
  ADD COLUMN title_en VARCHAR(128) NOT NULL DEFAULT '' AFTER title_zh,
  ADD COLUMN description_zh TEXT NULL AFTER title_en,
  ADD COLUMN description_en TEXT NULL AFTER description_zh,
  ADD COLUMN voucher_type VARCHAR(32) NOT NULL DEFAULT 'DEPOSIT_BONUS' AFTER description_en,
  ADD COLUMN start_at DATETIME(3) NULL AFTER status,
  ADD COLUMN end_at DATETIME(3) NULL AFTER start_at,
  ADD COLUMN new_users_only TINYINT(1) NOT NULL DEFAULT 0 AFTER end_at,
  ADD COLUMN first_deposit_only TINYINT(1) NOT NULL DEFAULT 0 AFTER new_users_only,
  ADD COLUMN daily_limit INT UNSIGNED NOT NULL DEFAULT 0 AFTER first_deposit_only,
  ADD COLUMN weekly_limit INT UNSIGNED NOT NULL DEFAULT 0 AFTER daily_limit,
  ADD COLUMN per_user_limit INT UNSIGNED NOT NULL DEFAULT 0 AFTER weekly_limit,
  ADD COLUMN countries VARCHAR(500) NOT NULL DEFAULT '' AFTER per_user_limit,
  ADD COLUMN currencies VARCHAR(128) NOT NULL DEFAULT 'PKR' AFTER countries,
  ADD COLUMN method_codes VARCHAR(500) NOT NULL DEFAULT '' AFTER currencies,
  ADD KEY idx_game_deposit_vouchers_availability (status, start_at, end_at);

UPDATE game_deposit_vouchers
SET title_zh = title,
    title_en = title,
    voucher_type = CASE
      WHEN voucher_code LIKE '%first%' THEN 'FIRST_DEPOSIT'
      WHEN voucher_code LIKE '%weekly%' THEN 'WEEKLY_DEPOSIT'
      WHEN voucher_code LIKE '%daily%' THEN 'DAILY_DEPOSIT'
      ELSE 'DEPOSIT_BONUS'
    END,
    currencies = 'PKR'
WHERE title_zh = '';

ALTER TABLE game_deposit_voucher_rewards
  ADD COLUMN amount_option_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER voucher_id,
  ADD COLUMN reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER cash_reward,
  ADD KEY idx_game_deposit_voucher_rewards_option (amount_option_id);

UPDATE game_deposit_voucher_rewards
SET reward_amount = bonus_reward + cash_reward;

ALTER TABLE game_deposit_orders
  ADD COLUMN voucher_bonus_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER voucher_id,
  ADD COLUMN voucher_cash_reward DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER voucher_bonus_reward,
  ADD COLUMN voucher_reward_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER voucher_cash_reward;

INSERT IGNORE INTO game_admin_permissions
  (code, name, name_zh, name_en, module, permission_type, enabled, system_protected)
VALUES
  ('fund:deposit-method:list', 'Deposit Method List', '充值方式列表', 'Deposit Method List', 'fund', 'API', 1, 1),
  ('fund:deposit-method:create', 'Create Deposit Method', '创建充值方式', 'Create Deposit Method', 'fund', 'API', 1, 1),
  ('fund:deposit-method:update', 'Update Deposit Method', '更新充值方式', 'Update Deposit Method', 'fund', 'API', 1, 1),
  ('fund:deposit-method:delete', 'Delete Deposit Method', '删除充值方式', 'Delete Deposit Method', 'fund', 'API', 1, 1),
  ('fund:deposit-channel:list', 'Deposit Channel List', '充值渠道列表', 'Deposit Channel List', 'fund', 'API', 1, 1),
  ('fund:deposit-channel:create', 'Create Deposit Channel', '创建充值渠道', 'Create Deposit Channel', 'fund', 'API', 1, 1),
  ('fund:deposit-channel:update', 'Update Deposit Channel', '更新充值渠道', 'Update Deposit Channel', 'fund', 'API', 1, 1),
  ('fund:deposit-channel:delete', 'Delete Deposit Channel', '删除充值渠道', 'Delete Deposit Channel', 'fund', 'API', 1, 1),
  ('fund:deposit-option:list', 'Deposit Amount Option List', '充值金额档位列表', 'Deposit Amount Option List', 'fund', 'API', 1, 1),
  ('fund:deposit-option:create', 'Create Deposit Amount Option', '创建充值金额档位', 'Create Deposit Amount Option', 'fund', 'API', 1, 1),
  ('fund:deposit-option:update', 'Update Deposit Amount Option', '更新充值金额档位', 'Update Deposit Amount Option', 'fund', 'API', 1, 1),
  ('fund:deposit-option:delete', 'Delete Deposit Amount Option', '删除充值金额档位', 'Delete Deposit Amount Option', 'fund', 'API', 1, 1),
  ('operation:voucher:list', 'Voucher List', '优惠券列表', 'Voucher List', 'operation', 'API', 1, 1),
  ('operation:voucher:create', 'Create Voucher', '创建优惠券', 'Create Voucher', 'operation', 'API', 1, 1),
  ('operation:voucher:update', 'Update Voucher', '更新优惠券', 'Update Voucher', 'operation', 'API', 1, 1),
  ('operation:voucher:delete', 'Delete Voucher', '删除优惠券', 'Delete Voucher', 'operation', 'API', 1, 1),
  ('operation:voucher:tier:update', 'Configure Voucher Tiers', '配置优惠券奖励档位', 'Configure Voucher Tiers', 'operation', 'API', 1, 1),
  ('operation:voucher:usage:view', 'View Voucher Usage', '查看优惠券使用记录', 'View Voucher Usage', 'operation', 'API', 1, 1);

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, NOW(3), NOW(3)
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN (
  'fund:deposit-method:list','fund:deposit-method:create','fund:deposit-method:update','fund:deposit-method:delete',
  'fund:deposit-channel:list','fund:deposit-channel:create','fund:deposit-channel:update','fund:deposit-channel:delete',
  'fund:deposit-option:list','fund:deposit-option:create','fund:deposit-option:update','fund:deposit-option:delete',
  'operation:voucher:list','operation:voucher:create','operation:voucher:update','operation:voucher:delete',
  'operation:voucher:tier:update','operation:voucher:usage:view'
)
WHERE r.code = 'super-admin';
