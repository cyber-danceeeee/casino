-- Existing installations imported a fixed ten-digit default for Indonesia.
-- Zero enables variable national-number lengths within the E.164 limit.
-- Keep the deployed baseline immutable and preserve explicit non-default rules.
UPDATE game_country_configs
SET phone_length=0, updated_at=NOW(3)
WHERE country_code='ID' AND phone_length=10 AND deleted_at IS NULL;
