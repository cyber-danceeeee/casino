ALTER TABLE game_game_providers
    ADD COLUMN wallet_mode varchar(32) NOT NULL DEFAULT 'SEAMLESS' AFTER status,
    ADD COLUMN callback_url varchar(500) NOT NULL DEFAULT '' AFTER wallet_mode,
    ADD COLUMN supported_countries varchar(500) NOT NULL DEFAULT '' AFTER callback_url,
    ADD COLUMN supported_currencies varchar(500) NOT NULL DEFAULT '' AFTER supported_countries,
    ADD COLUMN maintenance_message varchar(500) NOT NULL DEFAULT '' AFTER supported_currencies;

ALTER TABLE game_games
    ADD COLUMN name_i18n json NULL AFTER name,
    ADD COLUMN rtp decimal(6,2) DEFAULT NULL AFTER game_type,
    ADD COLUMN volatility varchar(32) NOT NULL DEFAULT '' AFTER rtp,
    ADD COLUMN background_url varchar(500) NOT NULL DEFAULT '' AFTER local_thumbnail_url,
    ADD COLUMN launch_image_url varchar(500) NOT NULL DEFAULT '' AFTER background_url,
    ADD COLUMN maintenance_message varchar(500) NOT NULL DEFAULT '' AFTER launch_image_url;

ALTER TABLE game_game_categories
    ADD COLUMN banner_url varchar(500) NOT NULL DEFAULT '' AFTER icon_url;

CREATE TABLE IF NOT EXISTS game_lobby_sections (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    tenant_id bigint unsigned NOT NULL DEFAULT 0,
    site_id bigint unsigned NOT NULL DEFAULT 0,
    section_code varchar(64) NOT NULL,
    section_type varchar(32) NOT NULL,
    title_zh varchar(128) NOT NULL DEFAULT '',
    title_en varchar(128) NOT NULL DEFAULT '',
    display_style varchar(64) NOT NULL DEFAULT '',
    source_type varchar(32) NOT NULL DEFAULT '',
    source_ref varchar(128) NOT NULL DEFAULT '',
    max_items int NOT NULL DEFAULT 12,
    sort_order int NOT NULL DEFAULT 0,
    enabled tinyint(1) NOT NULL DEFAULT 1,
    supported_countries varchar(500) NOT NULL DEFAULT '',
    supported_currencies varchar(500) NOT NULL DEFAULT '',
    config_json json NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_game_lobby_section_scope_code (tenant_id, site_id, section_code),
    KEY idx_game_lobby_section_display (tenant_id, site_id, enabled, sort_order)
);

INSERT IGNORE INTO game_lobby_sections
(tenant_id, site_id, section_code, section_type, title_zh, title_en, display_style, source_type, source_ref, max_items, sort_order, enabled, supported_countries, supported_currencies, config_json)
VALUES
(0, 0, 'top-banner', 'banner', '顶部 Banner', 'Top Banner', 'carousel', 'banner', 'home', 10, 10, 1, '', '', JSON_OBJECT()),
(0, 0, 'jackpot', 'jackpot', '奖池', 'Jackpot', 'ticker', 'jackpot', '', 1, 20, 1, '', '', JSON_OBJECT()),
(0, 0, 'hot-games', 'game-list', '热门游戏', 'Hot Games', 'horizontal', 'flag', 'hot', 12, 30, 1, '', '', JSON_OBJECT()),
(0, 0, 'new-games', 'game-list', '新游戏', 'New Games', 'horizontal', 'flag', 'new', 12, 40, 1, '', '', JSON_OBJECT()),
(0, 0, 'recommended-games', 'game-list', '推荐游戏', 'Recommended Games', 'horizontal', 'flag', 'featured', 12, 50, 1, '', '', JSON_OBJECT()),
(0, 0, 'continue-playing', 'game-list', '继续游戏', 'Continue Playing', 'horizontal', 'user', 'continue-playing', 12, 60, 1, '', '', JSON_OBJECT()),
(0, 0, 'recently-played', 'game-list', '最近玩过', 'Recently Played', 'horizontal', 'user', 'recently-played', 12, 70, 1, '', '', JSON_OBJECT()),
(0, 0, 'categories', 'category-list', '分类', 'Categories', 'grid', 'category', '', 20, 80, 1, '', '', JSON_OBJECT()),
(0, 0, 'providers', 'provider-list', '平台', 'Providers', 'grid', 'provider', '', 20, 90, 1, '', '', JSON_OBJECT()),
(0, 0, 'footer', 'footer', '底部', 'Footer', 'footer', 'config', '', 1, 100, 1, '', '', JSON_OBJECT());

INSERT IGNORE INTO game_admin_permissions (code, name, created_at, updated_at)
VALUES
('lobby:list', 'List lobby sections', NOW(3), NOW(3)),
('lobby:view', 'View lobby section', NOW(3), NOW(3)),
('lobby:update', 'Update lobby sections', NOW(3), NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, NOW(3), NOW(3)
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('lobby:list', 'lobby:view', 'lobby:update')
WHERE r.code = 'super-admin' AND r.deleted_at IS NULL;
