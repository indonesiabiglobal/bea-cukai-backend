-- Migration script untuk menambahkan user logging feature

-- 1. Tambahkan kolom baru ke tabel user
ALTER TABLE `user` 
ADD COLUMN `login_count` INT NOT NULL DEFAULT 0 COMMENT 'Total login count',
ADD COLUMN `last_login_at` DATETIME NULL COMMENT 'Last login timestamp',
ADD COLUMN `last_login_ip` VARCHAR(50) NULL COMMENT 'Last login IP address';

-- 2. Buat tabel user_log untuk menyimpan history aktivitas user
CREATE TABLE IF NOT EXISTS `user_log` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `user_id` VARCHAR(50) NULL COMMENT 'User ID from user table',
  `username` VARCHAR(100) NOT NULL COMMENT 'Username',
  `action` VARCHAR(50) NOT NULL COMMENT 'Action type: login, logout, create, update, delete',
  `ip_address` VARCHAR(50) NULL COMMENT 'IP address of the request',
  `user_agent` TEXT NULL COMMENT 'User agent string from request',
  `status` VARCHAR(20) NOT NULL COMMENT 'Status: success, failed, warning',
  `message` TEXT NULL COMMENT 'Additional message or error description',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Log timestamp',
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_username` (`username`),
  INDEX `idx_action` (`action`),
  INDEX `idx_status` (`status`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User activity log table';

-- 3. Update existing users with default values (optional)
-- UPDATE `user` SET `login_count` = 0 WHERE `login_count` IS NULL;

-- Verification queries (run these to check the migration)
-- SELECT * FROM information_schema.COLUMNS WHERE TABLE_NAME = 'user' AND COLUMN_NAME IN ('login_count', 'last_login_at', 'last_login_ip');
-- SHOW CREATE TABLE user_log;
