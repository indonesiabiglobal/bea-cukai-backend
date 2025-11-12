# Quick Start - Testing Sync API di Local

## ⚠️ Prerequisites untuk Windows

**PENTING**: Pastikan Git Bash sudah terinstall dan ada di PATH!

```bash
# Test apakah bash tersedia
bash --version

# Jika belum ada, install Git for Windows
# Download: https://git-scm.com/download/win
```

## 🚀 Quick Setup (3 Langkah)

### 1. Berikan Permission pada Script
```bash
cd "d:/R/Code/Project/Indonesia BI Global/fukusuke/bea-cukai/bea-cukai-app/bea-cukai-backend"
chmod +x test_sync_local.sh
chmod +x test_sync_api.sh
```

### 2. Test Script Manual
```bash
./test_sync_local.sh
```

**Expected Output:**
```
=== Starting Test Database Sync ===
Step 1: Checking database connection...
✓ Database connection successful
...
=== Test Sync Completed Successfully ===
```

### 3. Start API Server
```bash
# Di terminal lain
go run main.go
```

## 🧪 Test API (Simple)

### Dapatkan Token
```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your_password"}'
```

Copy token dari response.

### Test Endpoint

**1. Check Status:**
```bash
curl -X GET http://localhost:8000/sync/status \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**2. Run Sync:**
```bash
curl -X POST http://localhost:8000/sync/run \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**3. Get Log:**
```bash
curl -X GET http://localhost:8000/sync/log \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🤖 Automated Testing

1. **Edit `test_sync_api.sh`**, isi token:
```bash
TOKEN="paste_your_token_here"
```

2. **Run:**
```bash
./test_sync_api.sh
```

Script akan otomatis test semua skenario!

## ✅ Checklist

- [ ] Script `test_sync_local.sh` executable dan bisa jalan manual
- [ ] File `.env` sudah diupdate dengan `SYNC_SCRIPT_PATH=./test_sync_local.sh`
- [ ] API server running di port 8000
- [ ] Berhasil login dan dapat token
- [ ] Test manual 3 endpoint berhasil
- [ ] Test automated script berhasil

## 📝 Files yang Dibuat

1. `test_sync_local.sh` - Script simulasi sync untuk testing
2. `test_sync_api.sh` - Script automated testing API
3. `TESTING_GUIDE.md` - Dokumentasi lengkap testing
4. `SYNC_API.md` - Dokumentasi API endpoints
5. `sync_test.log` - Log file (akan dibuat saat run script)

## 🔧 Troubleshooting Cepat

**Error: bash: command not found**
- Install Git for Windows dari https://git-scm.com/download/win
- Pastikan Git Bash ada di PATH

**Error: %1 is not a valid Win32 application**
- Server sudah diperbaiki untuk menggunakan bash command
- Restart server (Ctrl+C lalu `go run main.go` atau `air` lagi)

**Error: Permission Denied**
```bash
chmod +x test_sync_local.sh
```

**Error: SYNC_SCRIPT_PATH not set**
- Restart aplikasi setelah edit `.env`

**Error: Script not found**
- Pastikan path di `.env` benar: `SYNC_SCRIPT_PATH=./test_sync_local.sh`

**Error: 401 Unauthorized**
- Token salah atau expired, login ulang

## 📖 Dokumentasi Lengkap

Lihat `TESTING_GUIDE.md` untuk:
- Testing scenarios lengkap
- Testing dengan Postman
- Troubleshooting detail
- Production deployment guide
