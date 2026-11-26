CREATE TABLE IF NOT EXISTS `game_operation_music` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  `code` varchar(64) NOT NULL,
  `name_zh` varchar(128) NOT NULL,
  `name_en` varchar(128) NOT NULL,
  `audio_url` varchar(500) NOT NULL,
  `scene` varchar(32) NOT NULL,
  `loop_play` tinyint(1) NOT NULL DEFAULT 1,
  `volume_percent` tinyint unsigned NOT NULL DEFAULT 70,
  `sort_order` int NOT NULL DEFAULT 0,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `start_at` datetime(3) DEFAULT NULL,
  `end_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_game_operation_music_code` (`code`),
  KEY `idx_game_operation_music_scene` (`scene`),
  KEY `idx_game_operation_music_enabled` (`enabled`),
  KEY `idx_game_operation_music_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT IGNORE INTO `game_operation_music`
(`code`, `name_zh`, `name_en`, `audio_url`, `scene`, `loop_play`, `volume_percent`, `sort_order`, `enabled`)
VALUES
('lobby-bgm', '大厅背景音乐', 'Lobby Background Music', '/static/h5/audio/backgound_music.mp3', 'LOBBY', 1, 70, 10, 1);

INSERT IGNORE INTO `game_admin_permissions` (`code`, `name`, `name_zh`, `name_en`, `module`, `permission_type`, `enabled`, `system_protected`)
VALUES
('operation:music:list', 'Music List', '音乐列表', 'Music List', 'operation', 'API', 1, 1),
('operation:music:create', 'Create Music', '创建音乐', 'Create Music', 'operation', 'API', 1, 1),
('operation:music:update', 'Update Music', '更新音乐', 'Update Music', 'operation', 'API', 1, 1),
('operation:music:delete', 'Delete Music', '删除音乐', 'Delete Music', 'operation', 'API', 1, 1),
('operation:asset:upload', 'Upload Operation Asset', '上传运营资源', 'Upload Operation Asset', 'operation', 'API', 1, 1);

INSERT IGNORE INTO `game_admin_role_permissions` (`role_id`, `permission_id`, `created_at`, `updated_at`)
SELECT r.id, p.id, NOW(3), NOW(3)
FROM `game_admin_roles` r
JOIN `game_admin_permissions` p ON p.code IN ('operation:music:list','operation:music:create','operation:music:update','operation:music:delete','operation:asset:upload')
WHERE r.code = 'super-admin';
