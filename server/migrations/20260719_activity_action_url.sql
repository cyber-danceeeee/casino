ALTER TABLE game_activity_events
    ADD COLUMN action_url varchar(500) NOT NULL DEFAULT '' AFTER cta_text;
