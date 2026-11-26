-- Every tenant site owns its category configuration. Platform categories are
-- copied once at site creation; this repairs sites created before that rule.
INSERT IGNORE INTO game_game_categories
    (tenant_id, site_id, category_key, label, name_zh, name_en, short_label,
     icon_url, banner_url, sort, status, created_at, updated_at)
SELECT
    s.tenant_id, s.id, c.category_key, c.label, c.name_zh, c.name_en, c.short_label,
    c.icon_url, c.banner_url, c.sort, c.status, NOW(3), NOW(3)
FROM game_saas_sites s
JOIN game_game_categories c
  ON c.tenant_id = 0
 AND c.site_id = 0
 AND c.deleted_at IS NULL
LEFT JOIN game_game_categories existing
  ON existing.tenant_id = s.tenant_id
 AND existing.site_id = s.id
 AND existing.category_key = c.category_key
WHERE s.deleted_at IS NULL
  AND existing.id IS NULL;
