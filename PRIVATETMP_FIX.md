# Solusi: Log File Tidak Ditemukan - PrivateTmp Issue

## 🔍 Root Cause

Backend service `bea-cukai-backend` berjalan dengan systemd **PrivateTmp=yes**, yang membuat:
- Service memiliki **namespace `/tmp` tersendiri** yang terisolasi
- Script yang dijalankan dari service menulis ke `/tmp` **private namespace**
- Controller API yang membaca log mencari di `/tmp` **host namespace**
- Hasilnya: **File tidak ditemukan meskipun benar-benar ada!**

## 🔍 Bukti Masalah

Di `/tmp` server terlihat:
```bash
systemd-private-...-bea-cukai-backend.service-Xey6RV
sync_fkk_20251112_144759.log  ✅ Ada di host namespace
sync_fkk_latest.log            ✅ Ada di host namespace
```

Tapi ketika API hit `/sync/log`:
```json
{
  "status": "error",
  "message": "Log file tidak ditemukan."
}
```

## ✅ Solusi yang Diterapkan

### 1. **Pindah Log ke /var/log/bea-cukai/**

Directory `/var/log` **tidak terpengaruh** oleh PrivateTmp, jadi bisa diakses dari semua namespace.

**Changes in `sync_fkk_db.sh`:**
```bash
# Sebelum:
WORKDIR="/tmp"
LOG_FILE="${WORKDIR}/sync_fkk_${DATESTR}.log"

# Sesudah:
WORKDIR="/tmp"          # Untuk dump files (boleh private)
LOG_DIR="/var/log/bea-cukai"  # Untuk log files (shared)
LOG_FILE="${LOG_DIR}/sync_fkk_${DATESTR}.log"

# Create log directory
mkdir -p "${LOG_DIR}"
chmod 755 "${LOG_DIR}"
```

### 2. **Update Controller Priority**

**Changes in `syncController.go`:**
```go
// Priority order:
// 1. /var/log/bea-cukai/sync_fkk_latest.log (NEW - shared namespace)
// 2. /var/log/bea-cukai/sync_fkk_*.log     (NEW - shared namespace)
// 3. /tmp/sync_fkk_latest.log              (Fallback - might be private)
// 4. ./sync_test.log                       (Test only)
```

## 📝 Setup di Server

### 1. Create Log Directory
```bash
sudo mkdir -p /var/log/bea-cukai
sudo chown srdbfki-03:srdbfki-03 /var/log/bea-cukai
sudo chmod 755 /var/log/bea-cukai
```

### 2. Update Script di Server
```bash
cd /opt/bea-cukai-app
# Upload sync_fkk_db.sh yang sudah diupdate
chmod +x sync_fkk_db.sh
```

### 3. Restart Backend Service
```bash
sudo systemctl restart bea-cukai-backend
```

### 4. Test Sync
```bash
# Via API
curl -X POST http://localhost:8000/sync/run \
  -H "Authorization: Bearer YOUR_TOKEN"

# Check log
curl -X GET http://localhost:8000/sync/log \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Verify Log Location
```bash
ls -lh /var/log/bea-cukai/
# Should see:
# sync_fkk_20251112_xxxxxx.log
# sync_fkk_latest.log -> sync_fkk_20251112_xxxxxx.log
```

## 🔧 Alternative Solution (Jika tidak mau ubah script)

Disable PrivateTmp di systemd service:

```bash
# Edit service file
sudo systemctl edit bea-cukai-backend.service

# Add:
[Service]
PrivateTmp=no

# Reload
sudo systemctl daemon-reload
sudo systemctl restart bea-cukai-backend
```

**⚠️ Warning:** Ini mengurangi isolasi security, tidak direkomendasikan untuk production!

## 📊 Comparison

| Aspek | /tmp (PrivateTmp) | /var/log/bea-cukai |
|-------|-------------------|-------------------|
| **Isolation** | ✅ Per-service | ❌ Shared |
| **API Access** | ❌ Tidak bisa | ✅ Bisa |
| **Security** | ✅ Lebih aman | ⚠️ Kurang aman |
| **Persistence** | ⚠️ Terhapus saat restart | ✅ Persistent |
| **Rotation** | Manual | ✅ logrotate ready |

## ✅ Benefits of /var/log/bea-cukai/

1. **Accessible from all namespaces** - API controller bisa akses
2. **Persistent** - Tidak terhapus saat service restart
3. **Standard location** - Sesuai Linux FHS (Filesystem Hierarchy Standard)
4. **Logrotate ready** - Bisa setup automatic log rotation
5. **Better organization** - Terpisah dari temporary files

## 🎯 Expected Behavior After Fix

### Before:
```bash
# Di server /tmp
ls /tmp/sync_fkk_*.log
# sync_fkk_20251112_144759.log ✅

# API Response
GET /sync/log
# "Log file tidak ditemukan" ❌
```

### After:
```bash
# Di server /var/log/bea-cukai
ls /var/log/bea-cukai/
# sync_fkk_20251112_xxxxxx.log ✅
# sync_fkk_latest.log -> sync_fkk_20251112_xxxxxx.log ✅

# API Response
GET /sync/log
# {
#   "status": "success",
#   "logFile": "/var/log/bea-cukai/sync_fkk_latest.log",
#   "content": "[2025-11-12 14:47:59] ==="
# } ✅
```

## 🧹 Cleanup Old Logs

Script automatically keeps only 10 most recent logs:
```bash
find /var/log/bea-cukai -name "sync_fkk_*.log" ! -name "sync_fkk_latest.log" -mtime +0 | sort -r | tail -n +11 | xargs rm -f
```

## 📚 References

- [Systemd PrivateTmp Documentation](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#PrivateTmp=)
- [Linux FHS - /var/log](https://refspecs.linuxfoundation.org/FHS_3.0/fhs/ch05s08.html)
