# Fix untuk Windows - Sync Controller

## Masalah
Error: `fork/exec ./test_sync_local.sh: %1 is not a valid Win32 application`

Terjadi karena Windows tidak bisa menjalankan bash script (.sh) secara langsung.

## Solusi yang Diterapkan

### 1. Update `syncController.go`

Semua command sekarang menggunakan `bash` wrapper:

**Before:**
```go
cmd := exec.Command(scriptPath)
cmd := exec.Command("test", "-f", lockFile)
cmd := exec.Command("tail", "-n", "100", path)
```

**After:**
```go
cmd := exec.Command("bash", scriptPath)
cmd := exec.Command("bash", "-c", "test -f "+lockFile)
cmd := exec.Command("bash", "-c", "tail -n 100 "+path)
```

### 2. Prerequisites

Pastikan **Git Bash** terinstall di Windows dan ada di PATH.

### 3. Testing

Setelah restart server, coba lagi:

```bash
curl -X POST http://localhost:8000/sync/run \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected Response:
```json
{
  "status": "success",
  "message": "Sinkronisasi database berhasil.",
  "output": "=== Starting Test Database Sync ===\n..."
}
```

## Cara Install Git Bash (jika belum ada)

1. Download Git for Windows: https://git-scm.com/download/win
2. Install dengan default settings
3. Restart terminal/command prompt
4. Test: `bash --version`

## Restart Server

Setelah update code, restart server:

```bash
# Jika menggunakan air
Ctrl+C (di terminal air)
# Server akan auto-restart

# Jika menggunakan go run
Ctrl+C
go run main.go
```

## Compatibility

✅ Windows (Git Bash)
✅ Linux
✅ macOS
✅ WSL (Windows Subsystem for Linux)

Sekarang sync API akan bekerja di semua platform!
