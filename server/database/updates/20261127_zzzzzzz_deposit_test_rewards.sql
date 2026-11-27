-- Populate only the existing, unconfigured IDR voucher explicitly named "test".
-- Never replace operator-defined tiers or create payment channels.
START TRANSACTION;
CREATE TEMPORARY TABLE casino_test_voucher_targets (
 id BIGINT UNSIGNED PRIMARY KEY,
 site_id BIGINT UNSIGNED NOT NULL,
 before_state JSON NOT NULL
);
INSERT INTO casino_test_voucher_targets
SELECT v.id,v.site_id,JSON_OBJECT('minAmount',v.min_amount,'maxAmount',v.max_amount,'maxReward',v.max_reward,'rewardTiers',JSON_ARRAY())
FROM game_deposit_vouchers v
WHERE v.voucher_code='test' AND v.site_id=1 AND v.currencies='IDR' AND v.status=1 AND v.deleted_at IS NULL
 AND NOT EXISTS(SELECT 1 FROM game_deposit_voucher_rewards r WHERE r.voucher_id=v.id)
 AND EXISTS(SELECT 1 FROM game_country_configs WHERE country_code='ID' AND currency_code='IDR' AND status=1 AND deleted_at IS NULL)
 AND (SELECT COUNT(DISTINCT o.amount) FROM game_deposit_amount_options o WHERE o.site_id=v.site_id AND o.currency_code='IDR' AND o.status=1 AND o.deleted_at IS NULL AND o.amount IN(10000,20000,50000,100000,200000,500000,1000000))=7;

INSERT INTO game_deposit_voucher_rewards
 (site_id,voucher_id,amount_option_id,amount,bonus_reward,cash_reward,reward_amount,sort,status)
SELECT t.site_id,t.id,MIN(o.id),o.amount,o.amount*0.05,o.amount*0.05,o.amount*0.10,MIN(o.sort),1
FROM casino_test_voucher_targets t
JOIN game_deposit_amount_options o ON o.site_id=t.site_id AND o.currency_code='IDR' AND o.status=1 AND o.deleted_at IS NULL
 AND o.amount IN(10000,20000,50000,100000,200000,500000,1000000)
GROUP BY t.site_id,t.id,o.amount;

UPDATE game_deposit_vouchers v JOIN casino_test_voucher_targets t ON t.id=v.id
SET v.min_amount=10000,v.max_amount=1000000,v.max_reward=100000,v.updated_at=NOW(3);

INSERT INTO game_admin_operation_logs
 (site_id,operator_username,permission,action,target_type,target_id,before_state,after_state,currency,reason)
SELECT t.site_id,'deployment-maintenance','operation:voucher:tier:update','configure_tiers','deposit_voucher',t.id,t.before_state,
 JSON_OBJECT('minAmount',10000,'maxAmount',1000000,'maxReward',100000,'rewardTiers',
 (SELECT JSON_ARRAYAGG(JSON_OBJECT('amount',r.amount,'bonusReward',r.bonus_reward,'cashReward',r.cash_reward,'rewardAmount',r.reward_amount)) FROM game_deposit_voucher_rewards r WHERE r.voucher_id=t.id AND r.deleted_at IS NULL)),
 'IDR','Add requested test rewards: 5% bonus plus 5% cash; no payment channel enabled'
FROM casino_test_voucher_targets t;
DROP TEMPORARY TABLE casino_test_voucher_targets;
COMMIT;
