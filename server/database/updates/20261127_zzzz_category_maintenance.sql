-- Category membership controls display; maintenance is represented by status=2.
UPDATE game_games g SET g.platform_enabled=1,
 g.status=CASE WHEN g.status=2 THEN 2 WHEN g.provider_enabled=1 THEN 1 ELSE 3 END,
 g.config_version=g.config_version+1,g.updated_at=NOW(3)
 WHERE g.deleted_at IS NULL AND g.site_id IN(0,1) AND g.platform_enabled=0
 AND EXISTS(SELECT 1 FROM game_game_category_relations r JOIN game_game_categories c ON c.id=r.category_id AND c.deleted_at IS NULL
 WHERE r.game_id=g.id AND r.deleted_at IS NULL);
