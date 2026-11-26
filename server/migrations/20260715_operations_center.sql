ALTER TABLE game_banners
    ADD COLUMN title_zh varchar(128) NOT NULL DEFAULT '' AFTER title,
    ADD COLUMN title_en varchar(128) NOT NULL DEFAULT '' AFTER title_zh,
    ADD COLUMN mobile_image_url varchar(500) NOT NULL DEFAULT '' AFTER image_url,
    ADD COLUMN target_type varchar(32) NOT NULL DEFAULT 'NONE' AFTER link_type,
    ADD COLUMN target_value varchar(500) NOT NULL DEFAULT '' AFTER target_type,
    ADD COLUMN platform varchar(32) NOT NULL DEFAULT 'ALL' AFTER position,
    ADD COLUMN countries varchar(500) NOT NULL DEFAULT '' AFTER platform,
    ADD COLUMN languages varchar(128) NOT NULL DEFAULT 'zh-CN,en-US' AFTER countries,
    ADD COLUMN enabled tinyint(1) NOT NULL DEFAULT 1 AFTER status,
    ADD COLUMN sort_order int NOT NULL DEFAULT 0 AFTER sort;

UPDATE game_banners SET title_zh=title, title_en=title, target_type=UPPER(link_type), target_value=link_url,
    enabled=(status=1), sort_order=sort WHERE title_zh='';

ALTER TABLE game_notices
    ADD COLUMN title_zh varchar(128) NOT NULL DEFAULT '' AFTER title,
    ADD COLUMN title_en varchar(128) NOT NULL DEFAULT '' AFTER title_zh,
    ADD COLUMN content_zh mediumtext NULL AFTER content_html,
    ADD COLUMN content_en mediumtext NULL AFTER content_zh,
    ADD COLUMN priority int NOT NULL DEFAULT 0 AFTER notice_type,
    ADD COLUMN countries varchar(500) NOT NULL DEFAULT '' AFTER priority,
    ADD COLUMN platforms varchar(128) NOT NULL DEFAULT 'ALL' AFTER countries,
    ADD COLUMN languages varchar(128) NOT NULL DEFAULT 'zh-CN,en-US' AFTER platforms,
    ADD COLUMN enabled tinyint(1) NOT NULL DEFAULT 1 AFTER status;

UPDATE game_notices SET title_zh=title, title_en=title, content_zh=content_html, content_en=content_html,
    enabled=(status=1) WHERE title_zh='';

ALTER TABLE game_activity_events
    ADD COLUMN name_zh varchar(128) NOT NULL DEFAULT '' AFTER title,
    ADD COLUMN name_en varchar(128) NOT NULL DEFAULT '' AFTER name_zh,
    ADD COLUMN description_zh text NULL AFTER subtitle,
    ADD COLUMN description_en text NULL AFTER description_zh,
    ADD COLUMN promotion_type varchar(48) NOT NULL DEFAULT 'DEPOSIT_BONUS' AFTER kind,
    ADD COLUMN countries varchar(500) NOT NULL DEFAULT '' AFTER promotion_type,
    ADD COLUMN currencies varchar(128) NOT NULL DEFAULT 'PKR' AFTER countries,
    ADD COLUMN status_code varchar(32) NOT NULL DEFAULT 'DRAFT' AFTER status,
    ADD COLUMN total_budget decimal(18,2) NOT NULL DEFAULT 0.00 AFTER status_code,
    ADD COLUMN per_user_limit int unsigned NOT NULL DEFAULT 1 AFTER total_budget,
    ADD COLUMN total_claim_limit int unsigned NOT NULL DEFAULT 0 AFTER per_user_limit,
    ADD COLUMN claim_count int unsigned NOT NULL DEFAULT 0 AFTER total_claim_limit,
    ADD COLUMN reward_amount decimal(18,2) NOT NULL DEFAULT 0.00 AFTER claim_count;

UPDATE game_activity_events SET name_zh=title, name_en=title,
    promotion_type=UPPER(REPLACE(kind, '-', '_')),
    status_code=CASE status WHEN 1 THEN 'ACTIVE' WHEN 3 THEN 'ENDED' ELSE 'PAUSED' END
WHERE name_zh='';

CREATE TABLE IF NOT EXISTS game_promotion_participation_rules (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    promotion_id bigint unsigned NOT NULL,
    rule_type varchar(48) NOT NULL,
    comparison_operator varchar(16) NOT NULL DEFAULT 'EQ',
    numeric_value decimal(18,2) DEFAULT NULL,
    text_value varchar(500) NOT NULL DEFAULT '',
    enabled tinyint(1) NOT NULL DEFAULT 1,
    sort_order int NOT NULL DEFAULT 0,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), KEY idx_promotion_participation (promotion_id, enabled)
);

CREATE TABLE IF NOT EXISTS game_promotion_reward_rules (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    promotion_id bigint unsigned NOT NULL,
    reward_type varchar(32) NOT NULL,
    reward_amount decimal(18,2) NOT NULL DEFAULT 0.00,
    currency varchar(8) NOT NULL DEFAULT 'PKR',
    wagering_multiplier decimal(8,2) NOT NULL DEFAULT 1.00,
    minimum_amount decimal(18,2) NOT NULL DEFAULT 0.00,
    maximum_reward decimal(18,2) NOT NULL DEFAULT 0.00,
    enabled tinyint(1) NOT NULL DEFAULT 1,
    sort_order int NOT NULL DEFAULT 0,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), KEY idx_promotion_reward (promotion_id, enabled)
);

CREATE TABLE IF NOT EXISTS game_promotion_eligible_games (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    promotion_id bigint unsigned NOT NULL,
    game_id bigint unsigned NOT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_promotion_game (promotion_id, game_id), KEY idx_promotion_game_game (game_id)
);

ALTER TABLE game_bonus_records
    ADD COLUMN bonus_id varchar(64) NOT NULL DEFAULT '' AFTER order_no,
    ADD COLUMN promotion_id bigint unsigned NOT NULL DEFAULT 0 AFTER source_id,
    ADD COLUMN bonus_type varchar(32) NOT NULL DEFAULT 'PROMOTION' AFTER promotion_id,
    ADD COLUMN currency varchar(8) NOT NULL DEFAULT 'PKR' AFTER amount,
    ADD COLUMN current_wagering decimal(18,2) NOT NULL DEFAULT 0.00 AFTER completed_valid_bet,
    ADD COLUMN remaining_wagering decimal(18,2) NOT NULL DEFAULT 0.00 AFTER current_wagering,
    ADD COLUMN withdrawable_amount decimal(18,2) NOT NULL DEFAULT 0.00 AFTER remaining_wagering,
    ADD COLUMN status_code varchar(32) NOT NULL DEFAULT 'PENDING' AFTER status,
    ADD COLUMN expires_at datetime(3) DEFAULT NULL AFTER unlocked_at;

UPDATE game_bonus_records SET bonus_id=order_no, promotion_id=source_id,
    current_wagering=completed_valid_bet,
    remaining_wagering=GREATEST(required_valid_bet-completed_valid_bet,0),
    status_code=CASE status WHEN 1 THEN 'PENDING' WHEN 2 THEN 'COMPLETED' WHEN 3 THEN 'EXPIRED' ELSE 'CANCELLED' END
WHERE bonus_id='';

CREATE TABLE IF NOT EXISTS game_bonus_status_histories (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    bonus_record_id bigint unsigned NOT NULL,
    before_status varchar(32) NOT NULL,
    after_status varchar(32) NOT NULL,
    operator_id bigint unsigned NOT NULL DEFAULT 0,
    operator_username varchar(64) NOT NULL DEFAULT '',
    reason varchar(500) NOT NULL DEFAULT '',
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), KEY idx_bonus_history (bonus_record_id, created_at)
);

ALTER TABLE game_vip_levels
    ADD COLUMN code varchar(32) NOT NULL DEFAULT '' AFTER level_no,
    ADD COLUMN name_zh varchar(64) NOT NULL DEFAULT '' AFTER name,
    ADD COLUMN name_en varchar(64) NOT NULL DEFAULT '' AFTER name_zh,
    ADD COLUMN required_experience bigint unsigned NOT NULL DEFAULT 0 AFTER icon_url,
    ADD COLUMN required_deposit decimal(18,2) NOT NULL DEFAULT 0.00 AFTER required_experience,
    ADD COLUMN required_valid_bet decimal(18,2) NOT NULL DEFAULT 0.00 AFTER required_deposit,
    ADD COLUMN withdrawal_limit decimal(18,2) NOT NULL DEFAULT 0.00 AFTER required_valid_bet,
    ADD COLUMN cashback_rate decimal(8,4) NOT NULL DEFAULT 0.0000 AFTER withdrawal_limit,
    ADD COLUMN birthday_reward decimal(18,2) NOT NULL DEFAULT 0.00 AFTER cashback_rate,
    ADD COLUMN sort_order int NOT NULL DEFAULT 0 AFTER monthly_reward,
    ADD COLUMN enabled tinyint(1) NOT NULL DEFAULT 1 AFTER status;

UPDATE game_vip_levels SET code=CONCAT('VIP',level_no), name_zh=name, name_en=name,
    sort_order=level_no, enabled=(status=1) WHERE code='';

CREATE TABLE IF NOT EXISTS game_vip_benefits (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    vip_level_id bigint unsigned NOT NULL,
    benefit_code varchar(48) NOT NULL,
    name_zh varchar(128) NOT NULL DEFAULT '',
    name_en varchar(128) NOT NULL DEFAULT '',
    benefit_value decimal(18,4) NOT NULL DEFAULT 0.0000,
    value_unit varchar(24) NOT NULL DEFAULT 'AMOUNT',
    enabled tinyint(1) NOT NULL DEFAULT 1,
    sort_order int NOT NULL DEFAULT 0,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_vip_benefit (vip_level_id, benefit_code)
);

CREATE TABLE IF NOT EXISTS game_vip_adjustment_logs (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    user_id bigint unsigned NOT NULL,
    before_level int unsigned NOT NULL,
    after_level int unsigned NOT NULL,
    operator_id bigint unsigned NOT NULL,
    operator_username varchar(64) NOT NULL,
    reason varchar(500) NOT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), KEY idx_vip_adjustment_user (user_id, created_at)
);

CREATE TABLE IF NOT EXISTS game_promo_codes (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    code varchar(64) NOT NULL,
    promotion_id bigint unsigned NOT NULL DEFAULT 0,
    reward_type varchar(32) NOT NULL,
    reward_value decimal(18,2) NOT NULL DEFAULT 0.00,
    currency varchar(8) NOT NULL DEFAULT 'PKR',
    total_usage_limit int unsigned NOT NULL DEFAULT 0,
    used_count int unsigned NOT NULL DEFAULT 0,
    per_user_limit int unsigned NOT NULL DEFAULT 1,
    countries varchar(500) NOT NULL DEFAULT '',
    new_users_only tinyint(1) NOT NULL DEFAULT 0,
    enabled tinyint(1) NOT NULL DEFAULT 1,
    start_at datetime(3) DEFAULT NULL,
    expires_at datetime(3) DEFAULT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_promo_code (code), KEY idx_promo_status_time (enabled, start_at, expires_at)
);

CREATE TABLE IF NOT EXISTS game_promo_code_redemptions (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    promo_code_id bigint unsigned NOT NULL,
    user_id bigint unsigned NOT NULL,
    idempotency_key varchar(128) NOT NULL,
    reward_type varchar(32) NOT NULL,
    reward_value decimal(18,2) NOT NULL DEFAULT 0.00,
    currency varchar(8) NOT NULL,
    wallet_transaction_id bigint unsigned NOT NULL DEFAULT 0,
    status varchar(32) NOT NULL DEFAULT 'COMPLETED',
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_promo_redemption_idempotency (idempotency_key),
    KEY idx_promo_redemption_user (promo_code_id, user_id, created_at)
);

CREATE TABLE IF NOT EXISTS game_notifications (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    notification_id varchar(64) NOT NULL,
    title_zh varchar(128) NOT NULL DEFAULT '',
    title_en varchar(128) NOT NULL DEFAULT '',
    content_zh text NULL,
    content_en text NULL,
    channel varchar(24) NOT NULL,
    audience_type varchar(32) NOT NULL,
    audience_value varchar(500) NOT NULL DEFAULT '',
    status varchar(32) NOT NULL DEFAULT 'DRAFT',
    target_user_count int unsigned NOT NULL DEFAULT 0,
    success_count int unsigned NOT NULL DEFAULT 0,
    failure_count int unsigned NOT NULL DEFAULT 0,
    scheduled_at datetime(3) DEFAULT NULL,
    sent_at datetime(3) DEFAULT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_notification_id (notification_id), KEY idx_notification_status_time (status, scheduled_at)
);

CREATE TABLE IF NOT EXISTS game_notification_jobs (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    notification_id bigint unsigned NOT NULL,
    job_type varchar(24) NOT NULL DEFAULT 'SEND',
    status varchar(24) NOT NULL DEFAULT 'PENDING',
    scheduled_at datetime(3) DEFAULT NULL,
    attempts int unsigned NOT NULL DEFAULT 0,
    last_error varchar(500) NOT NULL DEFAULT '',
    idempotency_key varchar(128) NOT NULL,
    created_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at datetime(3) DEFAULT NULL,
    PRIMARY KEY (id), UNIQUE KEY uk_notification_job_idempotency (idempotency_key), KEY idx_notification_job_pending (status, scheduled_at)
);

ALTER TABLE game_home_popup_dialogs
    ADD COLUMN title_zh varchar(128) NOT NULL DEFAULT '' AFTER title,
    ADD COLUMN title_en varchar(128) NOT NULL DEFAULT '' AFTER title_zh,
    ADD COLUMN content_zh text NULL AFTER title_en,
    ADD COLUMN content_en text NULL AFTER content_zh,
    ADD COLUMN page_position varchar(32) NOT NULL DEFAULT 'LOBBY' AFTER popup_type,
    ADD COLUMN countries varchar(500) NOT NULL DEFAULT '' AFTER page_position,
    ADD COLUMN platforms varchar(128) NOT NULL DEFAULT 'ALL' AFTER countries,
    ADD COLUMN priority int NOT NULL DEFAULT 0 AFTER sort,
    ADD COLUMN enabled tinyint(1) NOT NULL DEFAULT 1 AFTER status,
    ADD COLUMN show_frequency varchar(32) NOT NULL DEFAULT 'EVERY_LOGIN' AFTER enabled,
    ADD COLUMN maximum_display_count int unsigned NOT NULL DEFAULT 0 AFTER show_frequency,
    ADD COLUMN login_required tinyint(1) NOT NULL DEFAULT 0 AFTER maximum_display_count,
    ADD COLUMN new_users_only tinyint(1) NOT NULL DEFAULT 0 AFTER login_required,
    ADD COLUMN vip_levels varchar(128) NOT NULL DEFAULT '' AFTER new_users_only;

UPDATE game_home_popup_dialogs SET title_zh=title, title_en=title, priority=sort, enabled=(status=1)
WHERE title_zh='';

INSERT IGNORE INTO game_admin_permissions (code, name, created_at, updated_at) VALUES
('operation:banner:list','List banners',NOW(3),NOW(3)),('operation:banner:create','Create banners',NOW(3),NOW(3)),('operation:banner:update','Update banners',NOW(3),NOW(3)),('operation:banner:delete','Delete banners',NOW(3),NOW(3)),
('operation:announcement:list','List announcements',NOW(3),NOW(3)),('operation:announcement:create','Create announcements',NOW(3),NOW(3)),('operation:announcement:update','Update announcements',NOW(3),NOW(3)),('operation:announcement:delete','Delete announcements',NOW(3),NOW(3)),
('operation:promotion:list','List promotions',NOW(3),NOW(3)),('operation:promotion:create','Create promotions',NOW(3),NOW(3)),('operation:promotion:update','Update promotions',NOW(3),NOW(3)),('operation:promotion:status','Change promotion status',NOW(3),NOW(3)),
('operation:bonus:list','List bonuses',NOW(3),NOW(3)),('operation:bonus:view','View bonuses',NOW(3),NOW(3)),('operation:bonus:activate','Activate bonuses',NOW(3),NOW(3)),('operation:bonus:cancel','Cancel bonuses',NOW(3),NOW(3)),('operation:bonus:forfeit','Forfeit bonuses',NOW(3),NOW(3)),
('operation:vip:list','List VIP configuration',NOW(3),NOW(3)),('operation:vip:update','Update VIP configuration',NOW(3),NOW(3)),('operation:vip:adjust','Adjust user VIP',NOW(3),NOW(3)),
('operation:promo-code:list','List promo codes',NOW(3),NOW(3)),('operation:promo-code:create','Create promo codes',NOW(3),NOW(3)),('operation:promo-code:update','Update promo codes',NOW(3),NOW(3)),('operation:promo-code:delete','Delete promo codes',NOW(3),NOW(3)),
('operation:notification:list','List notifications',NOW(3),NOW(3)),('operation:notification:create','Create notifications',NOW(3),NOW(3)),('operation:notification:update','Update notifications',NOW(3),NOW(3)),('operation:notification:send','Send notifications',NOW(3),NOW(3)),('operation:notification:cancel','Cancel notifications',NOW(3),NOW(3)),
('operation:popup:list','List popups',NOW(3),NOW(3)),('operation:popup:create','Create popups',NOW(3),NOW(3)),('operation:popup:update','Update popups',NOW(3),NOW(3)),('operation:popup:delete','Delete popups',NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id,p.id,NOW(3),NOW(3) FROM game_admin_roles r JOIN game_admin_permissions p ON p.code LIKE 'operation:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL;
