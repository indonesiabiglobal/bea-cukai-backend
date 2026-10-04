package entryProductRepository

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---- Test helpers ----

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func newMockMySQL(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm.Open mysql mock: %v", err)
	}
	return gormDB, mock
}

func newMockPostgres(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm.Open postgres mock: %v", err)
	}
	return gormDB, mock
}

var entryCols = []string{"jenis_pabean", "no_pabean", "tgl_pabean", "trans_no", "item_code", "item_name", "rcv_qty", "pch_unit", "seq_num"}

// ============================================================
// GetReport — gabungan FKI (MySQL) + Itory (Postgres)
// ============================================================

// Test 1: hasil dua DB digabung, diurutkan tgl_pabean DESC lalu trans_no DESC,
// urutan seq_num per transaksi tetap, dan pch_unit terisi dari kedua sumber.
func TestGetReport_MergesBothSourcesSorted(t *testing.T) {
	fkiDB, fkiMock := newMockMySQL(t)
	itoryDB, itoryMock := newMockPostgres(t)
	repo := NewEntryProductRepository(fkiDB, itoryDB)

	d1, d2 := mustParseDate("2026-09-01"), mustParseDate("2026-09-02")

	fkiMock.ExpectQuery(regexp.QuoteMeta("AND cukai_date BETWEEN ? AND ? AND jenis_pabean = ?")).
		WithArgs("2026-09-01", "2026-09-30", "BC 2.3").
		WillReturnRows(sqlmock.NewRows(entryCols).
			AddRow("BC 2.3", "F1", d1, "FKI-001", "A", "Item A", "1.5", "PCS", 1).
			AddRow("BC 2.3", "F1", d1, "FKI-001", "B", "Item B", "2", "KG", 2))

	// Cek duplikat ke Itory (tidak ada yang cocok)
	itoryMock.ExpectQuery(regexp.QuoteMeta("string_to_array")).
		WillReturnRows(sqlmock.NewRows([]string{"ref_key", "bc_no", "bc_date", "vendor_code"}))

	// Postgres harus memakai placeholder $n, bukan ?
	itoryMock.ExpectQuery(regexp.QuoteMeta("AND bc_date BETWEEN $1 AND $2 AND bc_type = $3")).
		WithArgs("2026-09-01", "2026-09-30", "BC 2.3").
		WillReturnRows(sqlmock.NewRows(entryCols).
			AddRow("BC 2.3", "I1", d2, "ITR-001", "C", "Item C", "3", "BOX", 1))

	res, total, err := repo.GetReport(context.Background(), GetReportFilter{
		From: d1, To: mustParseDate("2026-09-30"), PabeanType: "BC 2.3",
	})
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if total != 3 || len(res) != 3 {
		t.Fatalf("want 3 rows, got total=%d len=%d", total, len(res))
	}
	want := []string{"C", "A", "B"} // d2 (Itory) dulu, lalu FKI-001 seq 1, 2
	for i, code := range want {
		if res[i].ItemCode != code {
			t.Errorf("row %d: want item %s, got %s", i, code, res[i].ItemCode)
		}
	}
	if res[0].PchUnit != "BOX" || res[1].PchUnit != "PCS" {
		t.Errorf("pch_unit tidak ter-mapping: %q, %q", res[0].PchUnit, res[1].PchUnit)
	}
	if res[1].RcvQty.String() != "1.5" {
		t.Errorf("rcv_qty: want 1.5, got %s", res[1].RcvQty)
	}
	if err := fkiMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if err := itoryMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// Test 2: dokumen yang ada di FKI dan Itory hanya diambil dari Itory. Cocok kalau invoice sama
// (abaikan spasi & huruf besar/kecil; invoice kosong tak pernah cocok), atau BC no + tgl + vendor sama
// selama dokumen Itory itu belum jadi pasangan invoice dokumen FKI lain.
func TestGetReport_FKICopyOfItoryIsDropped(t *testing.T) {
	fkiDB, fkiMock := newMockMySQL(t)
	itoryDB, itoryMock := newMockPostgres(t)
	repo := NewEntryProductRepository(fkiDB, itoryDB)

	d1, d2 := mustParseDate("2026-09-01"), mustParseDate("2026-09-02")
	d1WIB := time.Date(2026, 9, 1, 0, 0, 0, 0, time.FixedZone("WIB", 7*3600))
	cols := append(entryCols, "vend_dlv_no", "vendor_code")

	fkiMock.ExpectQuery(regexp.QuoteMeta("FROM tr_ap_inv_det_direct det")).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("BC 2.3", "F1", d1, "FKI-001", "A", "Item A", "1", "PCS", 1, " inv-1 ", "V1").     // invoice sama
			AddRow("BC 2.3", "F2", d1, "FKI-002", "B", "Item B", "1", "PCS", 1, "INV-2", "V1").       // hanya di FKI
			AddRow("BC 2.3", "F3", d1, "FKI-003", "C", "Item C", "1", "PCS", 1, "", "V1").            // tanpa invoice
			AddRow("BC 2.3", "F4", d1, "FKI-004", "D", "Item D", "1", "PCS", 1, "INV-4 (MB)", "V1").  // BC sama
			AddRow("BC 2.3", "F4", d1, "FKI-005", "E", "Item E", "1", "PCS", 1, "INV-4 (MB)", "V2").  // BC sama, vendor beda
			AddRow("BC 2.3", "F6", d1, "FKI-006", "F", "Item F", "1", "PCS", 1, "INV-6", "V1").       // invoice sama
			AddRow("BC 2.3", "F6", d1, "FKI-007", "G", "Item G", "1", "PCS", 1, "INV-6 (GIT)", "V1"). // BC sama, Itory sudah dipasangkan
			AddRow("BC 2.3", "F4", d2, "FKI-008", "H", "Item H", "1", "PCS", 1, "INV-8", "V1").       // BC no sama, tanggal beda
			AddRow("BC 2.3", "F9", d1WIB, "FKI-009", "I", "Item I", "1", "PCS", 1, "INV-9", "V1"))    // hari sama: WIB (MySQL) vs UTC (pgx)

	itoryMock.ExpectQuery(regexp.QuoteMeta("= ANY(string_to_array($1, chr(31))) OR TRIM(no_pabean) = ANY(string_to_array($2, chr(31)))")).
		WithArgs("INV-1\x1fINV-2\x1fINV-4 (MB)\x1fINV-6\x1fINV-6 (GIT)\x1fINV-8\x1fINV-9", "F1\x1fF2\x1fF3\x1fF4\x1fF6\x1fF9").
		WillReturnRows(sqlmock.NewRows([]string{"ref_key", "bc_no", "bc_date", "vendor_code"}).
			AddRow("INV-1", "X1", d1, "V1").
			AddRow("INV-4", "F4", d1, "V1").
			AddRow("INV-6", "F6", d1, "V1").
			AddRow("X9", "F9", d1, "V1"))

	itoryMock.ExpectQuery(regexp.QuoteMeta("FROM tr_ap_inv_det det")).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("BC 2.3", "I1", d1, "ITR-001", "A", "Item A", "1", "PCS", 1, "INV-1", "V1"))

	res, total, err := repo.GetReport(context.Background(), GetReportFilter{From: d1, To: d2})
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	got := map[string]bool{}
	for _, r := range res {
		got[r.TransNo] = true
	}
	want := map[string]bool{"ITR-001": true, "FKI-002": true, "FKI-003": true, "FKI-005": true, "FKI-007": true, "FKI-008": true}
	if total != int64(len(want)) || len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing %s (got %v)", k, got)
		}
	}
	if err := fkiMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if err := itoryMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// Regresi performa: FKI = MySQL 5.5, yang me-materialize derived table secara penuh
// (scan seluruh histori, ~17 detik). Filter harus masuk ke WHERE query sumber, bukan dibungkus.
func TestBuildQuery_FiltersInsideSourceWhere(t *testing.T) {
	for _, src := range []entrySource{fkiSource, itorySource} {
		q, args := buildQuery(src, GetReportFilter{
			From: mustParseDate("2026-09-01"), To: mustParseDate("2026-09-30"), ProductName: "baut",
		})
		if strings.Contains(q, "SELECT * FROM (") {
			t.Errorf("query tidak boleh dibungkus derived table:\n%s", q)
		}
		if !strings.Contains(q, "AND "+src.tgl+" BETWEEN ? AND ?") {
			t.Errorf("filter tanggal harus pakai kolom asli %s:\n%s", src.tgl, q)
		}
		if len(args) != 3 || args[2] != "%baut%" {
			t.Errorf("args: %v", args)
		}
	}
}

// ============================================================
// mergeAndPage — pure unit tests
// ============================================================

func rowsOn(dates ...string) []entryRow {
	rows := make([]entryRow, len(dates))
	for i, d := range dates {
		rows[i].TglPabean = mustParseDate(d)
		rows[i].TransNo = d
	}
	return rows
}

// Test 3: export diurutkan ASC
func TestMergeAndPage_ExportAscending(t *testing.T) {
	res, _ := mergeAndPage(rowsOn("2026-09-03", "2026-09-01", "2026-09-02"), GetReportFilter{IsExport: true})
	if res[0].TransNo != "2026-09-01" || res[2].TransNo != "2026-09-03" {
		t.Errorf("export harus ASC, got %s..%s", res[0].TransNo, res[2].TransNo)
	}
}

// Test 4: paginasi memotong hasil gabungan; total tetap jumlah semua baris
func TestMergeAndPage_Pagination(t *testing.T) {
	rows := rowsOn("2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04", "2026-09-05")

	res, total := mergeAndPage(rows, GetReportFilter{Page: 2, Limit: 2})
	if total != 5 || len(res) != 2 || res[0].TransNo != "2026-09-03" {
		t.Errorf("page 2: total=%d len=%d first=%v", total, len(res), res)
	}

	res, _ = mergeAndPage(rows, GetReportFilter{Page: 3, Limit: 2})
	if len(res) != 1 {
		t.Errorf("page 3: want 1 row, got %d", len(res))
	}

	res, total = mergeAndPage(rows, GetReportFilter{Page: 9, Limit: 2})
	if len(res) != 0 || total != 5 {
		t.Errorf("page di luar range harus kosong, got len=%d total=%d", len(res), total)
	}
}

func TestMergeAndPage_HugePageLimitNoPanic(t *testing.T) {
	rows := rowsOn("2026-09-01", "2026-09-02", "2026-09-03")
	maxInt := int(^uint(0) >> 1)

	if res, _ := mergeAndPage(rows, GetReportFilter{Page: 2, Limit: maxInt}); len(res) != 0 {
		t.Errorf("page 2, limit max: want 0 rows, got %d", len(res))
	}
	if res, _ := mergeAndPage(rows, GetReportFilter{Page: maxInt, Limit: 2}); len(res) != 0 {
		t.Errorf("page max: want 0 rows, got %d", len(res))
	}
	if res, _ := mergeAndPage(rows, GetReportFilter{Page: 1, Limit: maxInt}); len(res) != 3 {
		t.Errorf("page 1, limit max: want 3 rows, got %d", len(res))
	}
}

func TestLocalDate_SameDayAcrossSourcesSortsByTransNo(t *testing.T) {
	fki := entryRow{}
	fki.TglPabean = time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local) // MySQL loc=Local
	fki.TransNo = "A"
	itory := entryRow{}
	itory.TglPabean = localDate(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) // pgx DATE
	itory.TransNo = "B"

	res, _ := mergeAndPage([]entryRow{fki, itory}, GetReportFilter{})
	if res[0].TransNo != "B" || !res[0].TglPabean.Equal(res[1].TglPabean) {
		t.Errorf("same day must tie on date and order by trans_no DESC, got %v / %v", res[0], res[1])
	}
}

// QtyByTransDate memfilter trans_date (FKI) / in_date (Itory) dan menjumlah rcv_qty per item,
// dengan kode item dinormalisasi seperti perbandingan MySQL (case-insensitive, spasi diabaikan).
func TestQtyByTransDate_SumsPerItemOnTransDate(t *testing.T) {
	fkiDB, fkiMock := newMockMySQL(t)
	itoryDB, itoryMock := newMockPostgres(t)
	repo := NewEntryProductRepository(fkiDB, itoryDB)
	d1 := mustParseDate("2026-09-01")

	fkiMock.ExpectQuery(regexp.QuoteMeta("AND trans_date BETWEEN ? AND ?\nORDER BY")).
		WithArgs("2026-09-01", "2026-09-30").
		WillReturnRows(sqlmock.NewRows(entryCols).
			AddRow("BC 2.3", "F1", d1, "FKI-001", "box01", "Box", "1.5", "PCS", 1).
			AddRow("BC 2.3", "F1", d1, "FKI-001", "INK1", "Ink", "2", "KG", 2))
	itoryMock.ExpectQuery(regexp.QuoteMeta("string_to_array")).
		WillReturnRows(sqlmock.NewRows([]string{"ref_key", "bc_no", "bc_date", "vendor_code"}))
	itoryMock.ExpectQuery(regexp.QuoteMeta("AND in_date BETWEEN $1 AND $2\nORDER BY")).
		WithArgs("2026-09-01", "2026-09-30").
		WillReturnRows(sqlmock.NewRows(entryCols).
			AddRow("BC 2.3", "I1", d1, "ITR-001", "BOX01 ", "Box", "3", "PCS", 1))

	qty, err := repo.QtyByTransDate(context.Background(), d1, mustParseDate("2026-09-30"))
	if err != nil {
		t.Fatal(err)
	}
	if len(qty) != 2 || qty["BOX01"].String() != "4.5" || qty["INK1"].String() != "2" {
		t.Errorf("got %v", qty)
	}
	if err := fkiMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if err := itoryMock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
