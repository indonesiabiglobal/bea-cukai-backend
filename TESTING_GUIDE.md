# Testing Sync API di Local Laptop

## Prerequisites

1. **Install Git Bash** (untuk Windows) atau gunakan terminal bash di Linux/Mac
2. **Install jq** (optional, untuk format JSON yang rapi):
   - Windows (Git Bash): Download dari https://stedolan.github.io/jq/download/
   - Linux: `sudo apt install jq`
   - Mac: `brew install jq`

## Setup Testing

### 1. Update Environment Variable

Edit file `.env` dan ubah `SYNC_SCRIPT_PATH` ke script test local:

```env
# Untuk testing di local (Windows)
SYNC_SCRIPT_PATH=D:/R/Code/Project/Indonesia BI Global/fukusuke/bea-cukai/bea-cukai-app/bea-cukai-backend/test_sync_local.sh

# Atau jika menggunakan WSL/Git Bash, gunakan format Unix:
SYNC_SCRIPT_PATH=./test_sync_local.sh
```

### 2. Berikan Permission Execute pada Script

```bash
chmod +x test_sync_local.sh
chmod +x test_sync_api.sh
```

### 3. Test Script Sync Secara Manual

Test dulu apakah script sync bisa jalan:

```bash
./test_sync_local.sh
```

Output yang diharapkan:
```
=== Starting Test Database Sync ===
Step 1: Checking database connection...
✓ Database connection successful
Step 2: Creating backup...
✓ Backup created successfully
Step 3: Syncing data...
✓ Data synced: 1000 records processed
Step 4: Verifying integrity...
✓ Data integrity verified
Step 5: Optimizing tables...
✓ Tables optimized
=== Test Sync Completed Successfully ===
```

File log akan dibuat: `sync_test.log`

## Testing API Endpoints

### Cara 1: Menggunakan Test Script (Otomatis)

1. **Dapatkan JWT Token terlebih dahulu:**

```bash
# Login untuk mendapatkan token
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your_password"}'
```

Response akan berisi token:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

2. **Edit file `test_sync_api.sh`:**

Buka file dan isi variable `TOKEN`:
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # Paste token dari login
```

3. **Jalankan Test Script:**

```bash
./test_sync_api.sh
```

Script akan otomatis test semua endpoint:
- Check status awal
- Get log
- Run sync
- Check status saat running
- Try run lagi (harusnya fail)
- Wait sampai selesai
- Check status akhir
- Get final log

### Cara 2: Manual Testing dengan cURL

#### 1. Check Sync Status

```bash
curl -X GET http://localhost:8000/sync/status \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected Response (idle):
```json
{
  "status": "idle",
  "message": "No sync is currently running"
}
```

#### 2. Run Sync Script

```bash
curl -X POST http://localhost:8000/sync/run \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected Response (success):
```json
{
  "status": "success",
  "message": "Sync script executed successfully",
  "output": "=== Starting Test Database Sync ===\n..."
}
```

Expected Response (already running):
```json
{
  "status": "running",
  "message": "Sync is already running. Please wait for it to complete."
}
```

#### 3. Get Sync Log

```bash
curl -X GET http://localhost:8000/sync/log \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected Response:
```json
{
  "status": "success",
  "log": "[2025-11-12 10:30:45] === Starting Test Database Sync ===\n..."
}
```

### Cara 3: Testing dengan Postman

1. **Import Collection:**

Buat collection baru dengan 3 requests:

**Request 1: Check Status**
- Method: GET
- URL: `http://localhost:8000/sync/status`
- Headers: `Authorization: Bearer YOUR_TOKEN`

**Request 2: Run Sync**
- Method: POST
- URL: `http://localhost:8000/sync/run`
- Headers: `Authorization: Bearer YOUR_TOKEN`

**Request 3: Get Log**
- Method: GET
- URL: `http://localhost:8000/sync/log`
- Headers: `Authorization: Bearer YOUR_TOKEN`

2. **Test Scenario:**
   - Run Request 1 → Harus idle
   - Run Request 2 → Harus success
   - Run Request 1 lagi → Harus running
   - Run Request 2 lagi → Harus error (already running)
   - Tunggu 8 detik
   - Run Request 1 lagi → Harus idle
   - Run Request 3 → Lihat log lengkap

## Test Scenarios

### Scenario 1: Normal Sync Flow
```bash
# 1. Check initial status (should be idle)
curl -X GET http://localhost:8000/sync/status -H "Authorization: Bearer TOKEN"

# 2. Run sync
curl -X POST http://localhost:8000/sync/run -H "Authorization: Bearer TOKEN"

# 3. Wait 8 seconds
sleep 8

# 4. Check status again (should be idle)
curl -X GET http://localhost:8000/sync/status -H "Authorization: Bearer TOKEN"

# 5. Get log
curl -X GET http://localhost:8000/sync/log -H "Authorization: Bearer TOKEN"
```

### Scenario 2: Concurrent Execution Prevention
```bash
# 1. Run sync in background
curl -X POST http://localhost:8000/sync/run -H "Authorization: Bearer TOKEN" &

# 2. Immediately try to run again (should fail)
sleep 1
curl -X POST http://localhost:8000/sync/run -H "Authorization: Bearer TOKEN"

# Expected: Second request returns status "running" with 409 status code
```

### Scenario 3: Error Handling
```bash
# Test without authentication (should fail with 401)
curl -X POST http://localhost:8000/sync/run

# Test with invalid token (should fail with 401)
curl -X POST http://localhost:8000/sync/run -H "Authorization: Bearer INVALID"
```

## Troubleshooting

### Error: "SYNC_SCRIPT_PATH environment variable is not set"
**Solusi:**
1. Pastikan file `.env` sudah di-update dengan path yang benar
2. Restart aplikasi (stop dan start lagi)

### Error: "Failed to execute sync script"
**Solusi:**
1. Cek path script di `.env` apakah benar
2. Cek permission: `ls -l test_sync_local.sh` (harus ada `x`)
3. Jika di Windows, pastikan menggunakan Git Bash atau WSL

### Error: "fork/exec ... no such file or directory"
**Solusi:**
1. Gunakan absolute path di `.env`
2. Atau gunakan relative path `./test_sync_local.sh`

### Lock file tidak terhapus
**Solusi:**
```bash
# Hapus manual lock file
rm /tmp/sync_fkk_db_test.lock
```

### Log tidak muncul
**Solusi:**
1. Jalankan script manual dulu: `./test_sync_local.sh`
2. Cek apakah file `sync_test.log` terbuat
3. Cek permission read pada file log

## Expected Results

### Test Script Output (`./test_sync_local.sh`)
```
=== Starting Test Database Sync ===
Step 1: Checking database connection...
✓ Database connection successful
Step 2: Creating backup...
✓ Backup created successfully
Step 3: Syncing data...
✓ Data synced: 1000 records processed
Step 4: Verifying integrity...
✓ Data integrity verified
Step 5: Optimizing tables...
✓ Tables optimized
=== Test Sync Completed Successfully ===
```

### Log File Content (`sync_test.log`)
```
[2025-11-12 10:30:45] Lock file created: /tmp/sync_fkk_db_test.lock
[2025-11-12 10:30:45] === Starting Test Database Sync ===
[2025-11-12 10:30:45] Step 1: Checking database connection...
[2025-11-12 10:30:46] ✓ Database connection successful
[2025-11-12 10:30:46] Step 2: Creating backup...
[2025-11-12 10:30:47] ✓ Backup created successfully
[2025-11-12 10:30:47] Step 3: Syncing data...
[2025-11-12 10:30:49] ✓ Data synced: 1000 records processed
[2025-11-12 10:30:49] Step 4: Verifying integrity...
[2025-11-12 10:30:50] ✓ Data integrity verified
[2025-11-12 10:30:50] Step 5: Optimizing tables...
[2025-11-12 10:30:51] ✓ Tables optimized
[2025-11-12 10:30:51] === Test Sync Completed Successfully ===
[2025-11-12 10:30:51] Total execution time: ~6 seconds
[2025-11-12 10:30:51] Lock file removed: /tmp/sync_fkk_db_test.lock
```

## Clean Up

Setelah testing selesai:

```bash
# Hapus log file
rm sync_test.log

# Hapus lock file jika masih ada
rm /tmp/sync_fkk_db_test.lock
```

## Next Steps

Setelah testing berhasil di local:

1. **Deploy ke Server:**
   - Copy script production `sync_fkk_db.sh` ke server
   - Update `.env` di server dengan path production
   - Test di server

2. **Monitor:**
   - Setup cron job untuk monitoring
   - Setup alert jika sync gagal
   - Review log secara berkala

3. **Production Script:**
   - Ganti `test_sync_local.sh` dengan script production yang real
   - Implementasi database sync logic
   - Tambah error handling yang lebih robust
