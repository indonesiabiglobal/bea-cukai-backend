# User Logging System Documentation

## Overview
Sistem logging user yang lengkap untuk mencatat semua aktivitas user termasuk login, create, update, delete dengan informasi IP address, user agent, timestamp, dan status.

## Database Changes

### Table: `user`
Ditambahkan 3 kolom baru:
- `login_count` (INT): Total berapa kali user sudah login
- `last_login_at` (DATETIME): Timestamp login terakhir
- `last_login_ip` (VARCHAR): IP address login terakhir

### Table: `user_log` (New)
Tabel baru untuk menyimpan history aktivitas user:
- `id` (INT, PK): Auto increment ID
- `user_id` (VARCHAR): ID user dari tabel user
- `username` (VARCHAR): Username
- `action` (VARCHAR): Jenis aksi (login, logout, create, update, delete)
- `ip_address` (VARCHAR): IP address dari request
- `user_agent` (TEXT): User agent browser/client
- `status` (VARCHAR): Status aksi (success, failed, warning)
- `message` (TEXT): Pesan tambahan atau deskripsi error
- `created_at` (DATETIME): Timestamp log dibuat

## Migration
Jalankan script SQL untuk migration:
```bash
mysql -u username -p database_name < database/migration_user_logging.sql
```

## API Endpoints

### 1. Get All User Logs (Admin)
**GET** `/user-logs`

**Headers:**
- Authorization: Bearer {token}

**Query Parameters:**
- `user_id` (optional): Filter by user ID
- `username` (optional): Filter by username (partial match)
- `action` (optional): Filter by action (login, logout, create, update, delete)
- `status` (optional): Filter by status (success, failed, warning)
- `start_date` (optional): Filter by start date (format: 2006-01-02)
- `end_date` (optional): Filter by end date (format: 2006-01-02)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "message": "User logs retrieved successfully",
  "data": [
    {
      "id": 1,
      "user_id": "USR001",
      "username": "admin",
      "action": "login",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "status": "success",
      "message": "Login successful",
      "created_at": "2025-11-07T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_count": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  },
  "total": 150
}
```

### 2. Get My Logs (Current User)
**GET** `/user-logs/my-logs`

**Headers:**
- Authorization: Bearer {token}

**Query Parameters:**
- `limit` (optional): Number of logs to retrieve (default: 50)

**Response:**
```json
{
  "message": "User logs retrieved successfully",
  "data": [
    {
      "id": 1,
      "user_id": "USR001",
      "username": "admin",
      "action": "login",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "status": "success",
      "message": "Login successful",
      "created_at": "2025-11-07T10:30:00Z"
    }
  ],
  "total": 25
}
```

### 3. Get Logs by User ID (Admin)
**GET** `/user-logs/:user_id`

**Headers:**
- Authorization: Bearer {token}

**Path Parameters:**
- `user_id`: User ID to get logs for

**Query Parameters:**
- `limit` (optional): Number of logs to retrieve (default: 50)

**Response:**
```json
{
  "message": "User logs retrieved successfully",
  "data": [
    {
      "id": 1,
      "user_id": "USR001",
      "username": "admin",
      "action": "login",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "status": "success",
      "message": "Login successful",
      "created_at": "2025-11-07T10:30:00Z"
    }
  ],
  "total": 25
}
```

### 4. Get User Profile (Updated)
**GET** `/users/profile`

Sekarang response profile juga include informasi login:
```json
{
  "message": "User profile retrieved successfully",
  "data": {
    "id": "USR001",
    "username": "admin",
    "level": "admin",
    "login_count": 45,
    "last_login_at": "2025-11-07T10:30:00Z",
    "last_login_ip": "192.168.1.100"
  }
}
```

## Log Actions

### Login
- **Action**: `login`
- **Status**: `success` | `failed`
- Dicatat setiap kali user login (success atau gagal)
- Update `login_count`, `last_login_at`, `last_login_ip` di table user jika sukses

### Create User
- **Action**: `create`
- **Status**: `success` | `failed`
- Dicatat saat membuat user baru

### Update User
- **Action**: `update`
- **Status**: `success` | `failed`
- Dicatat saat update data user (termasuk update password)

### Delete User
- **Action**: `delete`
- **Status**: `success` | `failed`
- Dicatat saat menghapus user

## Security Features

1. **Password Hashing**: Semua password di-hash menggunakan bcrypt (cost 14)
2. **IP Tracking**: Setiap aktivitas mencatat IP address
3. **User Agent Tracking**: Mencatat device/browser yang digunakan
4. **Failed Login Tracking**: Login gagal juga dicatat untuk security monitoring
5. **Authentication Required**: Semua endpoint user logs memerlukan authentication

## Helper Functions

### GetIPAddress(ctx *gin.Context)
Mengambil IP address dari request, support:
- X-Forwarded-For header (untuk proxy/load balancer)
- X-Real-IP header
- Direct client IP

### GetUserAgent(ctx *gin.Context)
Mengambil User-Agent string dari request header

### HashPassword(password string)
Hash password menggunakan bcrypt

### CheckPasswordHash(password, hash string)
Verifikasi password dengan hash yang tersimpan

## Example Usage

### Login dengan logging
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Inventory@2511"
  }'
```

Log yang dicatat:
- Action: login
- IP: Client IP address
- User Agent: Client browser/app
- Status: success/failed
- Update login_count di table user (jika sukses)

### Lihat log aktivitas sendiri
```bash
curl -X GET "http://localhost:8080/user-logs/my-logs?limit=10" \
  -H "Authorization: Bearer {token}"
```

### Admin melihat semua log
```bash
curl -X GET "http://localhost:8080/user-logs?action=login&status=failed&page=1&limit=20" \
  -H "Authorization: Bearer {token}"
```

## Maintenance

### Clean Old Logs
Untuk menghapus log lama (misalnya lebih dari 90 hari), bisa menggunakan function di repository:

```go
err := userLogRepo.DeleteOldLogs(90) // hapus log lebih dari 90 hari
```

Atau langsung via SQL:
```sql
DELETE FROM user_log WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
```

## Notes

1. Semua endpoint user logs memerlukan authentication token
2. Log disimpan otomatis tanpa perlu action manual
3. Gunakan filtering untuk mencari log spesifik
4. Log diurutkan dari yang terbaru (DESC created_at)
5. Pertimbangkan untuk membuat scheduled task untuk cleanup log lama
