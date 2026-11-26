-- Current bonus services scope these models by tenant/site; the old SQL chain
-- omitted those columns. Backfill through the owning player/bonus record.
ALTER TABLE game_bonus_records
  ADD COLUMN tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id,
  ADD COLUMN site_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER tenant_id,
  ADD KEY idx_bonus_records_scope (tenant_id, site_id, user_id);

UPDATE game_bonus_records b
JOIN game_users u ON u.id=b.user_id
SET b.tenant_id=u.tenant_id, b.site_id=u.site_id;

ALTER TABLE game_bonus_status_histories
  ADD COLUMN tenant_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER id,
  ADD COLUMN site_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER tenant_id,
  ADD KEY idx_bonus_histories_scope (tenant_id, site_id, bonus_record_id);

UPDATE game_bonus_status_histories h
JOIN game_bonus_records b ON b.id=h.bonus_record_id
SET h.tenant_id=b.tenant_id, h.site_id=b.site_id;
