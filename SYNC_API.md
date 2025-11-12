# Database Sync API Documentation

## Overview
API endpoints untuk menjalankan dan memonitor script sinkronisasi database.

## Prerequisites
1. Set environment variable `SYNC_SCRIPT_PATH` di file `.env`:
   ```
   SYNC_SCRIPT_PATH=/opt/bea-cukai-app/sync_fkk_db.sh
   ```

2. Pastikan script sync memiliki permission execute:
   ```bash
   chmod +x /opt/bea-cukai-app/sync_fkk_db.sh
   ```

3. Script harus menggunakan lock file untuk mencegah eksekusi bersamaan
4. Script harus memiliki logging yang baik

## Authentication
Semua endpoint sync memerlukan authentication token di header:
```
Authorization: Bearer <token>
```

## Endpoints

### 1. Run Sync Script
Menjalankan script sinkronisasi database.

**Endpoint:** `POST /sync/run`

**Request:**
```bash
curl -X POST http://localhost:8000/sync/run \
  -H "Authorization: Bearer <your-token>"
```

**Response Success (200):**
```json
{
  "status": "success",
  "message": "Sync script executed successfully",
  "output": "Script output here..."
}
```

**Response Already Running (409):**
```json
{
  "status": "running",
  "message": "Sync is already running. Please wait for it to complete."
}
```

**Response Error:**
```json
{
  "status": "error",
  "message": "Error message here",
  "output": "Script error output..."
}
```

### 2. Get Sync Status
Mengecek status apakah sync sedang berjalan.

**Endpoint:** `GET /sync/status`

**Request:**
```bash
curl -X GET http://localhost:8000/sync/status \
  -H "Authorization: Bearer <your-token>"
```

**Response - Sync Running:**
```json
{
  "status": "running",
  "message": "Sync is currently running"
}
```

**Response - Sync Not Running:**
```json
{
  "status": "idle",
  "message": "No sync is currently running"
}
```

### 3. Get Sync Log
Mendapatkan 100 baris terakhir dari log sync.

**Endpoint:** `GET /sync/log`

**Request:**
```bash
curl -X GET http://localhost:8000/sync/log \
  -H "Authorization: Bearer <your-token>"
```

**Response Success:**
```json
{
  "status": "success",
  "log": "Last 100 lines of log...\nLine 2...\nLine 3..."
}
```

**Response Error:**
```json
{
  "status": "error",
  "message": "Error reading log file"
}
```

## Script Requirements

Script sync (`sync_fkk_db.sh`) harus memenuhi kriteria berikut:

### 1. Lock File Mechanism
Script harus menggunakan lock file untuk mencegah multiple execution:

```bash
#!/bin/bash

LOCK_FILE="/tmp/sync_fkk_db.lock"
LOG_FILE="/var/log/sync_fkk_db.log"

# Check if already running
if [ -f "$LOCK_FILE" ]; then
    echo "Sync is already running"
    exit 2  # Exit code 2 = already running
fi

# Create lock file
touch "$LOCK_FILE"

# Cleanup on exit
cleanup() {
    rm -f "$LOCK_FILE"
}
trap cleanup EXIT

# Your sync logic here
# ...

exit 0  # Exit code 0 = success
```

### 2. Exit Codes
- `0`: Success
- `2`: Already running (lock file exists)
- Other: Error

### 3. Logging
Script harus menulis log ke file yang dapat dibaca oleh API:

```bash
LOG_FILE="/var/log/sync_fkk_db.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting database sync..."
# Your sync operations
log "Sync completed successfully"
```

## Example Script Implementation

```bash
#!/bin/bash

# Configuration
LOCK_FILE="/tmp/sync_fkk_db.lock"
LOG_FILE="/var/log/sync_fkk_db.log"
SOURCE_DB="source_database"
TARGET_DB="target_database"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if already running
if [ -f "$LOCK_FILE" ]; then
    log "ERROR: Sync is already running"
    exit 2
fi

# Create lock file
touch "$LOCK_FILE"

# Cleanup function
cleanup() {
    rm -f "$LOCK_FILE"
    log "Cleanup completed"
}
trap cleanup EXIT

# Start sync
log "=== Starting database sync ==="

# Your database sync logic here
# Example: mysqldump and mysql restore
log "Dumping source database..."
# mysqldump commands...

log "Importing to target database..."
# mysql import commands...

log "=== Sync completed successfully ==="
exit 0
```

## Usage Example

```javascript
// Run sync
fetch('http://localhost:8000/sync/run', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
.then(res => res.json())
.then(data => {
  console.log(data);
  if (data.status === 'running') {
    // Poll status
    checkStatus();
  }
});

// Check status
function checkStatus() {
  fetch('http://localhost:8000/sync/status', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  })
  .then(res => res.json())
  .then(data => {
    if (data.status === 'running') {
      // Still running, check again after 5 seconds
      setTimeout(checkStatus, 5000);
    } else {
      // Finished, get logs
      getLogs();
    }
  });
}

// Get logs
function getLogs() {
  fetch('http://localhost:8000/sync/log', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  })
  .then(res => res.json())
  .then(data => {
    console.log(data.log);
  });
}
```

## Security Considerations

1. **Authentication Required**: Semua endpoint memerlukan valid JWT token
2. **Script Path**: Path script diambil dari environment variable untuk security
3. **Lock File**: Mencegah multiple execution yang bisa menyebabkan data corruption
4. **Log File Permissions**: Pastikan log file hanya bisa dibaca oleh user yang authorized
5. **Script Permissions**: Script harus memiliki permission yang tepat (750 atau 755)

## Troubleshooting

### Error: "SYNC_SCRIPT_PATH environment variable is not set"
**Solusi:** Tambahkan `SYNC_SCRIPT_PATH` di file `.env`

### Error: "Failed to execute sync script"
**Solusi:** 
- Cek apakah script path benar
- Cek permission execute pada script
- Cek log untuk detail error

### Sync stuck di status "running"
**Solusi:**
- Cek apakah script benar-benar masih berjalan: `ps aux | grep sync_fkk_db.sh`
- Jika tidak berjalan, hapus lock file manual: `rm /tmp/sync_fkk_db.lock`

### Log tidak muncul
**Solusi:**
- Cek apakah script menulis ke log file yang benar
- Cek permission read pada log file
- Cek apakah path log file sesuai dengan yang didefinisikan di script
