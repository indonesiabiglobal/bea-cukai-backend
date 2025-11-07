-- Script untuk cek dan debugging table user_log

-- 1. Cek apakah table user_log sudah ada
SELECT 
    TABLE_NAME,
    TABLE_ROWS,
    CREATE_TIME,
    UPDATE_TIME
FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'user_log';

-- 2. Cek struktur table user_log
DESCRIBE user_log;

-- 3. Cek apakah ada data di user_log
SELECT COUNT(*) as total_logs FROM user_log;

-- 4. Lihat 10 log terakhir
SELECT * FROM user_log ORDER BY created_at DESC LIMIT 10;

-- 5. Cek kolom baru di table user
DESCRIBE user;

-- 6. Cek data user dengan kolom login baru
SELECT id, username, level, login_count, last_login_at, last_login_ip 
FROM user 
LIMIT 5;

-- 7. Test insert manual ke user_log (untuk debugging)
-- INSERT INTO user_log (user_id, username, action, ip_address, user_agent, status, message) 
-- VALUES ('TEST', 'test_user', 'login', '127.0.0.1', 'Test Agent', 'success', 'Manual test insert');

-- 8. Cek apakah ada error di MySQL
SHOW WARNINGS;
