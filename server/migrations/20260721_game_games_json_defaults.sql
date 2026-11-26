UPDATE game_games
SET name_i18n = JSON_OBJECT()
WHERE name_i18n IS NULL
  AND deleted_at IS NULL;

UPDATE game_games
SET raw_metadata = JSON_OBJECT()
WHERE raw_metadata IS NULL
  AND deleted_at IS NULL;

UPDATE game_games
SET category_ids_draft = JSON_ARRAY()
WHERE category_ids_draft IS NULL
  AND deleted_at IS NULL;
