CREATE TABLE IF NOT EXISTS `game_h5_navigation_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `menu_code` varchar(64) NOT NULL,
  `label_zh` varchar(64) NOT NULL,
  `label_en` varchar(64) NOT NULL,
  `action_type` varchar(16) NOT NULL,
  `page_route` varchar(255) NOT NULL DEFAULT '',
  `provider_code` varchar(64) NOT NULL DEFAULT '',
  `game_id` bigint unsigned NOT NULL DEFAULT 0,
  `provider_game_code` varchar(128) NOT NULL DEFAULT '',
  `game_name` varchar(128) NOT NULL DEFAULT '',
  `sort_order` int NOT NULL DEFAULT 0,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_h5_navigation_code` (`menu_code`),
  KEY `idx_h5_navigation_status_sort` (`action_type`,`enabled`,`sort_order`),
  KEY `idx_h5_navigation_game` (`game_id`),
  KEY `idx_game_h5_navigation_items_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT IGNORE INTO `game_h5_navigation_items`
(`menu_code`,`label_zh`,`label_en`,`action_type`,`page_route`,`sort_order`,`enabled`,`created_at`,`updated_at`)
VALUES
('home','首页','Home','PAGE','/',10,1,NOW(3),NOW(3)),
('offers','优惠','Offers','PAGE','/offers',20,1,NOW(3),NOW(3)),
('share','分享','Share','PAGE','/share',40,1,NOW(3),NOW(3)),
('profile','我的','Profile','PAGE','/profile',50,1,NOW(3),NOW(3));
