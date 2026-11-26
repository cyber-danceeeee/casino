UPDATE game_user_wallets w
JOIN game_saas_sites s ON s.id=w.site_id AND s.tenant_id=w.tenant_id AND s.deleted_at IS NULL
JOIN game_saas_tenants t ON t.id=w.tenant_id AND t.deleted_at IS NULL
LEFT JOIN game_country_configs c ON c.country_code=COALESCE(NULLIF(s.country_code,''),NULLIF(t.country_code,'')) AND c.status=1 AND c.deleted_at IS NULL
SET w.currency=COALESCE(NULLIF(UPPER(s.currency_code),''),NULLIF(UPPER(t.default_currency),''),NULLIF(UPPER(c.currency_code),''),w.currency),
    w.updated_at=NOW(3)
WHERE w.deleted_at IS NULL
  AND UPPER(w.currency) IN ('','PKR')
  AND COALESCE(NULLIF(UPPER(s.currency_code),''),NULLIF(UPPER(t.default_currency),''),NULLIF(UPPER(c.currency_code),''),'') <> ''
  AND UPPER(w.currency) <> COALESCE(NULLIF(UPPER(s.currency_code),''),NULLIF(UPPER(t.default_currency),''),NULLIF(UPPER(c.currency_code),''));

UPDATE game_register_reward_claims r
JOIN game_user_wallets w ON w.user_id=r.user_id AND w.tenant_id=r.tenant_id AND w.site_id=r.site_id AND w.deleted_at IS NULL
SET r.currency=w.currency,
    r.updated_at=NOW(3)
WHERE r.deleted_at IS NULL
  AND r.status=1
  AND UPPER(r.currency) IN ('','PKR')
  AND UPPER(r.currency) COLLATE utf8mb4_unicode_ci <> UPPER(w.currency) COLLATE utf8mb4_unicode_ci;
