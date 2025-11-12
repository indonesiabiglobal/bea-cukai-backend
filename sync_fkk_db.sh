#!/usr/bin/env bash
set -euo pipefail

LOCKFILE="/tmp/sync_fkk_db.lock"

if [ -f "$LOCKFILE" ]; then
  echo "[WARN] Sync sedang berjalan. Batalkan."
  exit 2
fi

trap "rm -f $LOCKFILE" EXIT
touch $LOCKFILE

# =============== KONFIG (EDIT JIKA PERLU) ===============
SRC_HOST="192.168.1.100"
SRC_PORT="3306"
SRC_USER="user-sync"
SRC_PASS="R@h4514Umum!"
SRC_DB="fkk"

DEST_HOST="192.168.100.100"    # running di server ini
DEST_PORT="3306"
DEST_USER="fukusuke-3"
DEST_PASS="R@h4514Umum"
STG_DB="fkk_temp"
FINAL_DB="fukusuke_fkk"

WORKDIR="/tmp"
LOG_DIR="/var/log/bea-cukai"
DATESTR="$(date +%Y%m%d_%H%M%S)"
DUMP_SRC="${WORKDIR}/${SRC_DB}_${DATESTR}.sql"
DUMP_STG="${WORKDIR}/${STG_DB}_${DATESTR}.sql"
LOG_FILE="${LOG_DIR}/sync_fkk_${DATESTR}.log"
LOG_LATEST="${LOG_DIR}/sync_fkk_latest.log"

# Create log directory if not exists
mkdir -p "${LOG_DIR}"
chmod 755 "${LOG_DIR}" 2>/dev/null || true

MYSQL_LOCAL="mysql -h ${DEST_HOST} -P ${DEST_PORT} -u ${DEST_USER} --password=${DEST_PASS}"
MYSQL_SRC="mysql -h ${SRC_HOST} -P ${SRC_PORT} -u ${SRC_USER} --password=${SRC_PASS}"
MYSQLDUMP_SRC="mysqldump -h ${SRC_HOST} -P ${SRC_PORT} -u ${SRC_USER} --password=${SRC_PASS}"
MYSQLDUMP_LOCAL="mysqldump -h ${DEST_HOST} -P ${DEST_PORT} -u ${DEST_USER} --password=${DEST_PASS}"

# Create log file and symlink immediately
touch "$LOG_FILE"
chmod 666 "$LOG_FILE" 2>/dev/null || true
ln -sf "$LOG_FILE" "$LOG_LATEST"
chmod 666 "$LOG_LATEST" 2>/dev/null || true

log(){ echo "[$(date '+%F %T')] $*" | tee -a "$LOG_FILE" ; }

# Log start message with environment info
log "===== MULAI SINKRONISASI DATABASE ====="
log "Log file: ${LOG_FILE}"
log "Symlink: ${LOG_LATEST}"
log "User: $(whoami)"
log "UID: $(id -u)"
log "PWD: $(pwd)"
log "WORKDIR: ${WORKDIR}"

# Verify log file is writable
if [ -w "$LOG_FILE" ]; then
  log "✓ Log file writable"
else
  echo "[ERROR] Log file not writable: $LOG_FILE" >&2
  ls -l "$LOG_FILE" >&2
fi

# =============== CEK KONEKSI ===============
log "Cek koneksi SOURCE ${SRC_HOST}..."
$MYSQL_SRC -e "SELECT @@version AS src_version\G" >/dev/null

log "Cek koneksi DEST ${DEST_HOST}..."
$MYSQL_LOCAL -e "SELECT @@version AS dest_version\G" >/dev/null

# =============== 1) DUMP DARI SOURCE ===============
log "Dump ${SRC_DB} dari ${SRC_HOST} -> ${DUMP_SRC}"
$MYSQLDUMP_SRC \
  --single-transaction --routines --triggers --events \
  --set-gtid-purged=OFF \
  "${SRC_DB}" > "${DUMP_SRC}"

# =============== 1b) HAPUS DEFINER DARI DUMP ===============
log "Hapus DEFINER dari dump SQL..."
sed -i -E 's/DEFINER=`[^`]+`@`[^`]+`//g' "${DUMP_SRC}"

# =============== 2) RECREATE STAGING ===============
log "Drop & create staging DB ${STG_DB} di ${DEST_HOST}"
$MYSQL_LOCAL -e "DROP DATABASE IF EXISTS \`${STG_DB}\`; CREATE DATABASE \`${STG_DB}\`;"

# =============== 3) IMPORT KE STAGING DENGAN SQL_MODE DILONGGARKAN ===============
log "Import ke ${STG_DB} (longgarkan sql_mode untuk zero date & DDL lama)"
if command -v pv >/dev/null 2>&1; then
  # Gunakan pv untuk progress bar jika tersedia
  pv -pteba "${DUMP_SRC}" | \
  mysql -h ${DEST_HOST} -P ${DEST_PORT} -u ${DEST_USER} --password=${DEST_PASS} \
    --database="${STG_DB}" \
    --init-command="SET SESSION sql_mode=REPLACE(@@sql_mode,'STRICT_TRANS_TABLES','');
                    SET SESSION sql_mode=REPLACE(@@sql_mode,'NO_ZERO_DATE','');
                    SET SESSION sql_mode=REPLACE(@@sql_mode,'NO_ZERO_IN_DATE','');
                    SET SESSION sql_mode=CONCAT(@@sql_mode, ',ALLOW_INVALID_DATES');
                    SET SESSION foreign_key_checks=0;
                    SET SESSION unique_checks=0;
                    SET SESSION sql_log_bin=0;
                    SET autocommit=0;"
else
  # Fallback tanpa pv
  $MYSQL_LOCAL \
    --database="${STG_DB}" \
    --init-command="SET SESSION sql_mode=REPLACE(@@sql_mode,'STRICT_TRANS_TABLES','');
                    SET SESSION sql_mode=REPLACE(@@sql_mode,'NO_ZERO_DATE','');
                    SET SESSION sql_mode=REPLACE(@@sql_mode,'NO_ZERO_IN_DATE','');
                    SET SESSION sql_mode=CONCAT(@@sql_mode, ',ALLOW_INVALID_DATES');
                    SET SESSION foreign_key_checks=0;
                    SET SESSION unique_checks=0;
                    SET SESSION sql_log_bin=0;
                    SET autocommit=0;" \
    < "${DUMP_SRC}"
fi

# =============== 4) ALTER DEFAULT ZERO -> NULL (DATETIME/TIMESTAMP/DATE) ===============
log "Generate ALTER kolom DATETIME/TIMESTAMP default nol -> DEFAULT NULL"
$MYSQL_LOCAL -N -e "
SELECT CONCAT(
  'ALTER TABLE \`${STG_DB}\`.\`', TABLE_NAME, '\` MODIFY \`', COLUMN_NAME, '\` ',
  UPPER(DATA_TYPE), ' NULL DEFAULT NULL;'
)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = '${STG_DB}'
  AND DATA_TYPE IN ('datetime','timestamp')
  AND COLUMN_DEFAULT = '0000-00-00 00:00:00';
" > "${WORKDIR}/alter_zero_dt_${DATESTR}.sql"

log "Generate ALTER kolom DATE default nol -> DEFAULT NULL"
$MYSQL_LOCAL -N -e "
SELECT CONCAT(
  'ALTER TABLE \`${STG_DB}\`.\`', TABLE_NAME, '\` MODIFY \`', COLUMN_NAME, '\` DATE NULL DEFAULT NULL;'
)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = '${STG_DB}'
  AND DATA_TYPE = 'date'
  AND COLUMN_DEFAULT = '0000-00-00';
" > "${WORKDIR}/alter_zero_date_${DATESTR}.sql"

if [ -s "${WORKDIR}/alter_zero_dt_${DATESTR}.sql" ]; then
  log "Jalankan ALTER DATETIME/TIMESTAMP (ignore error jika ada)..."
  $MYSQL_LOCAL < "${WORKDIR}/alter_zero_dt_${DATESTR}.sql" 2>/dev/null || log "Beberapa ALTER gagal, akan dilanjutkan dengan UPDATE..."
fi
if [ -s "${WORKDIR}/alter_zero_date_${DATESTR}.sql" ]; then
  log "Jalankan ALTER DATE (ignore error jika ada)..."
  $MYSQL_LOCAL < "${WORKDIR}/alter_zero_date_${DATESTR}.sql" 2>/dev/null || log "Beberapa ALTER gagal, akan dilanjutkan dengan UPDATE..."
fi

# =============== 5) UPDATE NILAI ZERO -> NULL ===============
log "Jalankan UPDATE spesifik (daftar yang kamu berikan)..."
$MYSQL_LOCAL --database="${STG_DB}" <<'SQLUPD'
-- Longgarkan sql_mode untuk session ini
SET @OLD_SQL_MODE := @@SESSION.sql_mode;
SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'STRICT_TRANS_TABLES','');
SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_DATE','');
SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_IN_DATE','');
SET SESSION sql_mode = CONCAT(@@SESSION.sql_mode, ',ALLOW_INVALID_DATES');

-- ==== DATE ====
UPDATE tr_ap_inv_det             SET tgl_po = NULL WHERE tgl_po = '0000-00-00';
UPDATE tr_ap_inv_head            SET in_date  = NULL WHERE in_date  = '0000-00-00';
UPDATE tr_ap_inv_head            SET bl_date  = NULL WHERE bl_date  = '0000-00-00';
UPDATE tr_ap_inv_head            SET due_date = NULL WHERE due_date = '0000-00-00';
UPDATE tr_ap_inv_head            SET bc_date  = NULL WHERE bc_date  = '0000-00-00';
UPDATE tr_ap_payment_head        SET cheque_due_date = NULL WHERE cheque_due_date = '0000-00-00';
UPDATE tr_ar_payment_head        SET cheque_due_date = NULL WHERE cheque_due_date = '0000-00-00';
UPDATE tr_export_head            SET trans_date = NULL WHERE trans_date = '0000-00-00';
UPDATE tr_export_head            SET tgl_ekspor = NULL WHERE tgl_ekspor = '0000-00-00';
UPDATE tr_export_head            SET custom_date = NULL WHERE custom_date = '0000-00-00';
UPDATE tr_export_head            SET tgl_doc_keluar = NULL WHERE tgl_doc_keluar = '0000-00-00';
UPDATE tr_inv_adjust_head        SET trans_date = NULL WHERE trans_date = '0000-00-00';
UPDATE tr_inv_rm_head            SET trans_date = NULL WHERE trans_date = '0000-00-00';
UPDATE tr_pengeluaran_barang_new SET tgl_pabean = NULL WHERE tgl_pabean = '0000-00-00';
UPDATE tr_pengeluaran_barang_new SET trans_date = NULL WHERE trans_date = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_proses   = NULL WHERE tgl_proses   = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_produksi = NULL WHERE tgl_produksi = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_ekspor1  = NULL WHERE tgl_ekspor1  = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_ekspor2  = NULL WHERE tgl_ekspor2  = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_ekspor3  = NULL WHERE tgl_ekspor3  = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_ekspor4  = NULL WHERE tgl_ekspor4  = '0000-00-00';
UPDATE tr_produk_in_head         SET tgl_koreksi  = NULL WHERE tgl_koreksi  = '0000-00-00';
UPDATE tr_spe_entry_head         SET trans_date   = NULL WHERE trans_date   = '0000-00-00';

-- ==== DATETIME ====
UPDATE ms_item                SET created_date = NULL WHERE created_date = '0000-00-00 00:00:00';
UPDATE ms_item                SET updated_date = NULL WHERE updated_date = '0000-00-00 00:00:00';
UPDATE tr_ar_inv_head         SET created_date = NULL WHERE created_date = '0000-00-00 00:00:00';
UPDATE tr_ar_inv_head         SET updated_date = NULL WHERE updated_date = '0000-00-00 00:00:00';
UPDATE tr_ar_payment_head     SET created_date = NULL WHERE created_date = '0000-00-00 00:00:00';
UPDATE tr_ar_payment_head     SET updated_date = NULL WHERE updated_date = '0000-00-00 00:00:00';
UPDATE tr_export_head         SET created_date = NULL WHERE created_date = '0000-00-00 00:00:00';
UPDATE tr_export_head         SET updated_date = NULL WHERE updated_date = '0000-00-00 00:00:00';
UPDATE tr_inv_rm_head         SET created_date  = NULL WHERE created_date  = '0000-00-00 00:00:00';
UPDATE tr_inv_rm_head         SET updated_date  = NULL WHERE updated_date  = '0000-00-00 00:00:00';
UPDATE tr_inv_rm_head         SET approved_date = NULL WHERE approved_date = '0000-00-00 00:00:00';
UPDATE tr_inv_rm_head         SET canceled_date = NULL WHERE canceled_date = '0000-00-00 00:00:00';
UPDATE tr_produk_in_head      SET created_date  = NULL WHERE created_date  = '0000-00-00 00:00:00';
UPDATE tr_produk_in_head      SET updated_date  = NULL WHERE updated_date  = '0000-00-00 00:00:00';
UPDATE tr_produk_in_head      SET approved_date = NULL WHERE approved_date = '0000-00-00 00:00:00';
UPDATE tr_produk_in_head      SET canceled_date = NULL WHERE canceled_date = '0000-00-00 00:00:00';
UPDATE tr_spe_entry_head      SET created_date = NULL WHERE created_date = '0000-00-00 00:00:00';
UPDATE tr_spe_entry_head      SET updated_date = NULL WHERE updated_date = '0000-00-00 00:00:00';

-- Kembalikan sql_mode
SET SESSION sql_mode := @OLD_SQL_MODE;
SQLUPD

# (Jaga-jaga) generator universal supaya tidak ada yang terlewat:
log "Generator universal UPDATE nol->NULL (DATETIME/TIMESTAMP)"
$MYSQL_LOCAL -N -e "
SELECT CONCAT(
  'UPDATE \`${STG_DB}\`.\`', TABLE_NAME, '\` SET \`', COLUMN_NAME, '\`=NULL WHERE \`',
  COLUMN_NAME, \"\`='0000-00-00 00:00:00';\"
)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA='${STG_DB}' AND DATA_TYPE IN ('datetime','timestamp');
" > "${WORKDIR}/upd_all_dt_${DATESTR}.sql"

if [ -s "${WORKDIR}/upd_all_dt_${DATESTR}.sql" ]; then
  log "Eksekusi UPDATE universal DATETIME/TIMESTAMP..."
  {
    echo "SET @OLD_SQL_MODE := @@SESSION.sql_mode;"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'STRICT_TRANS_TABLES','');"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_DATE','');"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_IN_DATE','');"
    echo "SET SESSION sql_mode = CONCAT(@@SESSION.sql_mode, ',ALLOW_INVALID_DATES');"
    cat "${WORKDIR}/upd_all_dt_${DATESTR}.sql"
    echo "SET SESSION sql_mode := @OLD_SQL_MODE;"
  } | $MYSQL_LOCAL --database="${STG_DB}"
fi

log "Generator universal UPDATE nol->NULL (DATE)"
$MYSQL_LOCAL -N -e "
SELECT CONCAT(
  'UPDATE \`${STG_DB}\`.\`', TABLE_NAME, '\` SET \`', COLUMN_NAME, '\`=NULL WHERE \`',
  COLUMN_NAME, \"\`='0000-00-00';\"
)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA='${STG_DB}' AND DATA_TYPE='date';
" > "${WORKDIR}/upd_all_date_${DATESTR}.sql"

if [ -s "${WORKDIR}/upd_all_date_${DATESTR}.sql" ]; then
  log "Eksekusi UPDATE universal DATE..."
  {
    echo "SET @OLD_SQL_MODE := @@SESSION.sql_mode;"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'STRICT_TRANS_TABLES','');"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_DATE','');"
    echo "SET SESSION sql_mode = REPLACE(@@SESSION.sql_mode,'NO_ZERO_IN_DATE','');"
    echo "SET SESSION sql_mode = CONCAT(@@SESSION.sql_mode, ',ALLOW_INVALID_DATES');"
    cat "${WORKDIR}/upd_all_date_${DATESTR}.sql"
    echo "SET SESSION sql_mode := @OLD_SQL_MODE;"
  } | $MYSQL_LOCAL --database="${STG_DB}"
fi

# =============== 6) DUMP DARI STAGING YANG SUDAH BERSIH ===============
log "Dump staging ${STG_DB} -> ${DUMP_STG}"
$MYSQLDUMP_LOCAL \
  --single-transaction --routines --triggers --events \
  --set-gtid-purged=OFF \
  "${STG_DB}" > "${DUMP_STG}"

# =============== 6b) MODIFIKASI DUMP UNTUK TIDAK DROP DATABASE ===============
log "Hapus DROP DATABASE dari dump staging (agar tabel existing tidak hilang)..."
sed -i '/^DROP DATABASE/d' "${DUMP_STG}"
sed -i '/^CREATE DATABASE/d' "${DUMP_STG}"
sed -i '/^USE /d' "${DUMP_STG}"

# =============== 7) PUBLISH KE DB FINAL DI 100.100 ===============
log "Pastikan DB final ada..."
$MYSQL_LOCAL -e "CREATE DATABASE IF NOT EXISTS \`${FINAL_DB}\`;"

log "Backup tabel yang akan dipertahankan..."
$MYSQL_LOCAL --database="${FINAL_DB}" -e "
CREATE TABLE IF NOT EXISTS tr_pemasukan_barang_backup LIKE tr_pemasukan_barang;
CREATE TABLE IF NOT EXISTS tr_pengeluaran_barang_backup LIKE tr_pengeluaran_barang;
CREATE TABLE IF NOT EXISTS tr_ap_inv_det_direct_fki_backup LIKE tr_ap_inv_det_direct_fki;
CREATE TABLE IF NOT EXISTS tr_ap_inv_head_fki_backup LIKE tr_ap_inv_head_fki;
CREATE TABLE IF NOT EXISTS tr_ar_inv_det_direct_fki_backup LIKE tr_ar_inv_det_direct_fki;
CREATE TABLE IF NOT EXISTS tr_ar_inv_head_fki_backup LIKE tr_ar_inv_head_fki;
CREATE TABLE IF NOT EXISTS user_backup LIKE user;
CREATE TABLE IF NOT EXISTS user_log_backup LIKE user_log;
INSERT INTO tr_pemasukan_barang_backup SELECT * FROM tr_pemasukan_barang;
INSERT INTO tr_pengeluaran_barang_backup SELECT * FROM tr_pengeluaran_barang;
INSERT INTO tr_ap_inv_det_direct_fki_backup SELECT * FROM tr_ap_inv_det_direct_fki;
INSERT INTO tr_ap_inv_head_fki_backup SELECT * FROM tr_ap_inv_head_fki;
INSERT INTO tr_ar_inv_det_direct_fki_backup SELECT * FROM tr_ar_inv_det_direct_fki;
INSERT INTO tr_ar_inv_head_fki_backup SELECT * FROM tr_ar_inv_head_fki;
INSERT INTO user_backup SELECT * FROM user;
INSERT INTO user_log_backup SELECT * FROM user_log;
" 2>/dev/null || log "Tabel backup mungkin belum ada, skip backup..."

log "Import ke ${FINAL_DB} (DROP TABLE terlebih dahulu kecuali tabel yang dipertahankan)..."
# Drop semua tabel kecuali tabel yang ingin dipertahankan
$MYSQL_LOCAL -N -e "
SELECT CONCAT('DROP TABLE IF EXISTS \`${FINAL_DB}\`.\`', TABLE_NAME, '\`;')
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = '${FINAL_DB}'
  AND TABLE_NAME NOT IN ('tr_pemasukan_barang', 'tr_pengeluaran_barang',
                          'tr_ap_inv_det_direct_fki', 'tr_ap_inv_head_fki',
                          'tr_ar_inv_det_direct_fki', 'tr_ar_inv_head_fki', 'user', 'user_log',
                          'tr_pemasukan_barang_backup', 'tr_pengeluaran_barang_backup',
                          'tr_ap_inv_det_direct_fki_backup', 'tr_ap_inv_head_fki_backup',
                          'tr_ar_inv_det_direct_fki_backup', 'tr_ar_inv_head_fki_backup', 
                          'user_backup', 'user_log_backup');
" | $MYSQL_LOCAL 2>/dev/null || true

log "Import data dari staging ke final..."
if command -v pv >/dev/null 2>&1; then
  pv -pteba "${DUMP_STG}" | $MYSQL_LOCAL --database="${FINAL_DB}"
else
  $MYSQL_LOCAL --database="${FINAL_DB}" < "${DUMP_STG}"
fi

log "Restore tabel yang dipertahankan (jika ada perubahan)..."
$MYSQL_LOCAL --database="${FINAL_DB}" -e "
DROP TABLE IF EXISTS tr_pemasukan_barang_backup;
DROP TABLE IF EXISTS tr_pengeluaran_barang_backup;
DROP TABLE IF EXISTS tr_ap_inv_det_direct_fki_backup;
DROP TABLE IF EXISTS tr_ap_inv_head_fki_backup;
DROP TABLE IF EXISTS tr_ar_inv_det_direct_fki_backup;
DROP TABLE IF EXISTS tr_ar_inv_head_fki_backup;
DROP TABLE IF EXISTS user_backup;
DROP TABLE IF EXISTS user_log_backup;
" 2>/dev/null || true

# =============== 8) CLEANUP FILE DUMP LAMA ===============
log "Cleanup file dump lama (pertahankan hanya yang terbaru)..."
# Hapus dump source yang lebih lama dari file saat ini
find "${WORKDIR}" -name "${SRC_DB}_*.sql" -type f ! -name "$(basename ${DUMP_SRC})" -mtime +0 -delete 2>/dev/null || true
# Hapus dump staging yang lebih lama dari file saat ini
find "${WORKDIR}" -name "${STG_DB}_*.sql" -type f ! -name "$(basename ${DUMP_STG})" -mtime +0 -delete 2>/dev/null || true
# Hapus file ALTER dan UPDATE yang lebih lama
find "${WORKDIR}" -name "alter_zero_*.sql" -type f -mtime +0 -delete 2>/dev/null || true
find "${WORKDIR}" -name "upd_all_*.sql" -type f -mtime +0 -delete 2>/dev/null || true

# Hapus log file lama (keep 10 terbaru) dari LOG_DIR bukan WORKDIR
find "${LOG_DIR}" -name "sync_fkk_*.log" -type f ! -name "sync_fkk_latest.log" -mtime +0 2>/dev/null | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true

log "SELESAI. Log: ${LOG_FILE}"

# Ensure log file is flushed and symlink is updated
sync
ln -sf "$LOG_FILE" "$LOG_LATEST" 2>/dev/null || true
