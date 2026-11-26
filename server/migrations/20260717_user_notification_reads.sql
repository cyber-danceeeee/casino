CREATE TABLE IF NOT EXISTS game_user_notification_reads (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  notification_id BIGINT UNSIGNED NOT NULL,
  read_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_notification_read (user_id, notification_id),
  KEY idx_notification_read_user_time (user_id, read_at),
  KEY idx_notification_read_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
