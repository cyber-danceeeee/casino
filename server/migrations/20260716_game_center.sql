ALTER TABLE game_game_providers
    ADD COLUMN environment varchar(24) NOT NULL DEFAULT 'production' AFTER api_base_url,
    ADD COLUMN api_status varchar(24) NOT NULL DEFAULT 'UNKNOWN' AFTER environment,
    ADD COLUMN last_tested_at datetime(3) DEFAULT NULL AFTER api_status,
    ADD COLUMN last_sync_at datetime(3) DEFAULT NULL AFTER last_tested_at;

ALTER TABLE game_games
    ADD COLUMN local_thumbnail_url varchar(500) NOT NULL DEFAULT '' AFTER cover_url,
    ADD COLUMN provider_enabled tinyint(1) NOT NULL DEFAULT 1 AFTER supports_free_spins,
    ADD COLUMN platform_enabled tinyint(1) NOT NULL DEFAULT 0 AFTER provider_enabled,
    ADD COLUMN is_hot tinyint(1) NOT NULL DEFAULT 0 AFTER platform_enabled,
    ADD COLUMN is_featured tinyint(1) NOT NULL DEFAULT 0 AFTER is_hot,
    ADD COLUMN is_new tinyint(1) NOT NULL DEFAULT 0 AFTER is_featured,
    ADD COLUMN supported_countries varchar(500) NOT NULL DEFAULT '' AFTER is_new,
    ADD COLUMN supported_currencies varchar(500) NOT NULL DEFAULT '' AFTER supported_countries,
    ADD COLUMN supported_devices varchar(255) NOT NULL DEFAULT '' AFTER supported_currencies,
    ADD COLUMN metadata_hash char(64) NOT NULL DEFAULT '' AFTER raw_metadata;

UPDATE game_games
SET provider_enabled = (status <> 3), platform_enabled = (status = 1),
    is_hot = (hot_score > 0), is_featured = (badge = 'featured'), is_new = (badge = 'new');

ALTER TABLE game_game_categories
    ADD COLUMN name_zh varchar(64) NOT NULL DEFAULT '' AFTER label,
    ADD COLUMN name_en varchar(64) NOT NULL DEFAULT '' AFTER name_zh;

UPDATE game_game_categories SET name_zh=label, name_en=label WHERE name_zh='';

CREATE TABLE IF NOT EXISTS game_game_tags (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    code varchar(64) NOT NULL,
    name_zh varchar(64) NOT NULL,
    name_en varchar(64) NOT NULL,
    icon_url varchar(500) NOT NULL DEFAULT '',
    sort_order int NOT NULL DEFAULT 0,
    enabled tinyint(1) NOT NULL DEFAULT 1,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_game_tag_code (code), KEY idx_game_tag_status_sort (enabled, sort_order)
);

CREATE TABLE IF NOT EXISTS game_game_tag_relations (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    game_id bigint unsigned NOT NULL,
    tag_id bigint unsigned NOT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_game_tag_relation (game_id, tag_id), KEY idx_game_tag_relation_tag (tag_id, game_id)
);

CREATE TABLE IF NOT EXISTS game_game_collections (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    code varchar(64) NOT NULL,
    name_zh varchar(128) NOT NULL,
    name_en varchar(128) NOT NULL,
    description_zh text NULL,
    description_en text NULL,
    banner_url varchar(500) NOT NULL DEFAULT '',
    sort_order int NOT NULL DEFAULT 0,
    enabled tinyint(1) NOT NULL DEFAULT 1,
    start_at datetime(3) DEFAULT NULL,
    end_at datetime(3) DEFAULT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_game_collection_code (code), KEY idx_game_collection_display (enabled, start_at, end_at, sort_order)
);

CREATE TABLE IF NOT EXISTS game_game_collection_relations (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    collection_id bigint unsigned NOT NULL,
    game_id bigint unsigned NOT NULL,
    sort_order int NOT NULL DEFAULT 0,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_game_collection_relation (collection_id, game_id), KEY idx_game_collection_relation_game (game_id, collection_id)
);

CREATE TABLE IF NOT EXISTS game_provider_sync_logs (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    sync_no varchar(64) NOT NULL,
    provider_code varchar(64) NOT NULL,
    status varchar(24) NOT NULL,
    new_count int unsigned NOT NULL DEFAULT 0,
    updated_count int unsigned NOT NULL DEFAULT 0,
    unchanged_count int unsigned NOT NULL DEFAULT 0,
    disabled_count int unsigned NOT NULL DEFAULT 0,
    failed_count int unsigned NOT NULL DEFAULT 0,
    duration_ms bigint unsigned NOT NULL DEFAULT 0,
    triggered_by_id bigint unsigned NOT NULL DEFAULT 0,
    triggered_by varchar(64) NOT NULL DEFAULT '',
    error_message varchar(500) NOT NULL DEFAULT '',
    raw_provider_response longtext NULL,
    started_at datetime(3) NOT NULL,
    completed_at datetime(3) DEFAULT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_game_sync_no (sync_no), KEY idx_game_sync_provider_time (provider_code, started_at), KEY idx_game_sync_status_time (status, started_at)
);

CREATE TABLE IF NOT EXISTS game_provider_sync_log_items (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    sync_log_id bigint unsigned NOT NULL,
    provider_game_id varchar(128) NOT NULL DEFAULT '',
    game_id bigint unsigned NOT NULL DEFAULT 0,
    game_name varchar(128) NOT NULL DEFAULT '',
    change_type varchar(24) NOT NULL,
    before_state longtext NULL,
    after_state longtext NULL,
    error_message varchar(500) NOT NULL DEFAULT '',
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), KEY idx_game_sync_item_log (sync_log_id, change_type), KEY idx_game_sync_item_game (game_id)
);

ALTER TABLE game_game_launch_logs
    ADD COLUMN launch_no varchar(64) NOT NULL DEFAULT '' AFTER provider_id,
    ADD COLUMN launch_type varchar(16) NOT NULL DEFAULT 'REAL' AFTER launch_no,
    ADD COLUMN currency varchar(8) NOT NULL DEFAULT '' AFTER launch_type,
    ADD COLUMN language varchar(16) NOT NULL DEFAULT '' AFTER currency,
    ADD COLUMN country varchar(8) NOT NULL DEFAULT '' AFTER language,
    ADD COLUMN device varchar(32) NOT NULL DEFAULT '' AFTER country,
    ADD COLUMN provider_request longtext NULL AFTER device,
    ADD COLUMN provider_response longtext NULL AFTER provider_request,
    ADD COLUMN correlation_id varchar(128) NOT NULL DEFAULT '' AFTER provider_response;

UPDATE game_game_launch_logs SET launch_no=CONCAT('LCH-',id) WHERE launch_no='';

INSERT IGNORE INTO game_game_tags (code,name_zh,name_en,sort_order,enabled) VALUES
('hot','热门','Hot',10,1),('new','新品','New',20,1),('featured','推荐','Featured',30,1),
('exclusive','独家','Exclusive',40,1),('high-rtp','高返奖率','High RTP',50,1),('mexico-popular','墨西哥热门','Mexico Popular',60,1);

INSERT IGNORE INTO game_admin_permissions (code,name,created_at,updated_at)
VALUES
('game:provider:list','List game providers',NOW(3),NOW(3)),
('game:provider:view','View game provider',NOW(3),NOW(3)),
('game:provider:update','Update game provider',NOW(3),NOW(3)),
('game:provider:test','Test game provider',NOW(3),NOW(3)),
('game:provider:sync','Synchronize game provider',NOW(3),NOW(3)),
('game:list','List games',NOW(3),NOW(3)),('game:view','View game',NOW(3),NOW(3)),
('game:update','Update game',NOW(3),NOW(3)),('game:status:update','Update game status',NOW(3),NOW(3)),
('game:category:assign','Assign game categories',NOW(3),NOW(3)),('game:tag:assign','Assign game tags',NOW(3),NOW(3)),
('game:collection:assign','Assign game collections',NOW(3),NOW(3)),('game:demo:launch','Launch game demo',NOW(3),NOW(3)),
('game:category:list','List game categories',NOW(3),NOW(3)),('game:category:create','Create game category',NOW(3),NOW(3)),
('game:category:update','Update game category',NOW(3),NOW(3)),('game:category:delete','Delete game category',NOW(3),NOW(3)),
('game:tag:list','List game tags',NOW(3),NOW(3)),('game:tag:create','Create game tag',NOW(3),NOW(3)),
('game:tag:update','Update game tag',NOW(3),NOW(3)),('game:tag:delete','Delete game tag',NOW(3),NOW(3)),
('game:collection:list','List game collections',NOW(3),NOW(3)),('game:collection:create','Create game collection',NOW(3),NOW(3)),
('game:collection:update','Update game collection',NOW(3),NOW(3)),('game:collection:delete','Delete game collection',NOW(3),NOW(3)),
('game:sync-log:view','View game synchronization logs',NOW(3),NOW(3)),('game:launch-log:view','View game launch logs',NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_permissions p ON p.code LIKE 'game:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL;
