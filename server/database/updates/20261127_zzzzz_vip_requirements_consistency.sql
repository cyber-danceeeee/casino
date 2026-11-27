-- Use the same upgrade thresholds in the admin catalog and H5 requirement table.
UPDATE game_vip_levels l JOIN game_vip_requirements r ON r.vip_level_id=l.id AND r.site_id=l.site_id
 AND r.requirement_type='level_up' AND r.deleted_at IS NULL
 SET l.required_deposit=r.deposit_amount,l.required_valid_bet=r.valid_bet_amount,l.updated_at=NOW(3)
 WHERE l.deleted_at IS NULL AND l.site_id IN(0,1)
 AND (l.required_deposit<>r.deposit_amount OR l.required_valid_bet<>r.valid_bet_amount);
