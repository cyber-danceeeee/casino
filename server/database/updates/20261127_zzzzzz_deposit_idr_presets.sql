-- Add neutral IDR presets only when no IDR catalog has been configured.
-- Do not seed payment merchants or promotional rewards.
INSERT INTO game_deposit_amount_options
 (site_id,currency_code,option_code,amount,bonus_reward,cash_reward,sort,status)
SELECT 1,'IDR',CONCAT('idr-',preset.amount),preset.amount,0,0,preset.sort,1
FROM (
 SELECT 10000 amount,1 sort UNION ALL SELECT 20000,2 UNION ALL SELECT 50000,3
 UNION ALL SELECT 100000,4 UNION ALL SELECT 200000,5 UNION ALL SELECT 500000,6 UNION ALL SELECT 1000000,7
) preset
WHERE EXISTS(SELECT 1 FROM game_country_configs WHERE country_code='ID' AND currency_code='IDR' AND status=1 AND deleted_at IS NULL)
 AND NOT EXISTS(SELECT 1 FROM game_deposit_amount_options existing WHERE existing.site_id IN(0,1) AND existing.currency_code='IDR');
