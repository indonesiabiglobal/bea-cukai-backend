package rawMaterialReportRepository

import (
	"context"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open mock: %v", err)
	}
	return gormDB, mock
}

// ============================================================
// buildBaseQuery — pure unit tests, tidak butuh DB
// ============================================================

// Test 1: saat tglInvAkhir == filter.To, opname harus pakai COALESCE(f.opname, 0)
func TestBuildBaseQuery_OpnameUsesF_WhenAkhirEqualsTo(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"), // sama dengan tglInvAkhir
	}

	query, _ := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	if !strings.Contains(query, "COALESCE(f.opname, 0) AS opname") {
		t.Error("opname harus COALESCE(f.opname, 0) ketika tglInvAkhir == filter.To")
	}
	if strings.Contains(query, "IFNULL") {
		t.Error("IFNULL tidak ada di Postgres")
	}
}

// Test 2: saat tglInvAkhir != filter.To, opname harus pakai akhirExpr (bukan f.opname)
func TestBuildBaseQuery_OpnameUsesAkhir_WhenAkhirNotEqualsTo(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-20") // berbeda dari filter.To
	filter := GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	}

	query, _ := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	if strings.Contains(query, "COALESCE(f.opname, 0) AS opname") {
		t.Error("opname TIDAK boleh pakai COALESCE(f.opname, 0) ketika tglInvAkhir != filter.To")
	}
}

// Test 3: jumlah base args harus tepat 18 (tanpa filter item_code/item_name)
func TestBuildBaseQuery_ArgsCount_NoFilter(t *testing.T) {
	_, args := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	})

	// b(1)+c(2)+e(2)+f(1)+g(2)+keluar(2)+keluar_awal(2)+masuk_awal(2)+movein_awal(2)+peny_after_opname(2) = 18
	const wantCount = 18
	if len(args) != wantCount {
		t.Errorf("jumlah args: want %d, got %d", wantCount, len(args))
	}
}

// Test 4: filter item_code dan item_name menambah 2 args ekstra
func TestBuildBaseQuery_ArgsCount_WithBothFilters(t *testing.T) {
	_, args := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemCode: "MAT001",
		ItemName: "Bahan A",
	})

	const wantCount = 20 // 18 base + 2 filter
	if len(args) != wantCount {
		t.Errorf("jumlah args dengan filter: want %d, got %d", wantCount, len(args))
	}
}

// Test 5: urutan args harus persis sesuai urutan ? di CTE
// Ini adalah test paling kritis — urutan args salah = data salah
func TestBuildBaseQuery_ArgsOrder(t *testing.T) {
	_, args := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	})

	// Rentang *_awal: trans_date > tglInvAwal AND trans_date < filter.From (eksklusif)
	want := []string{
		"2024-01-01",               // [0]  b:           tglInvAwal
		"2024-01-15", "2024-01-31", // [1-2]  c:         filter.From..To
		"2024-01-15", "2024-01-31", // [3-4]  e:         filter.From..To
		"2024-01-31",               // [5]  f:           tglInvAkhir
		"2024-01-15", "2024-01-31", // [6-7]  g:         filter.From..To
		"2024-01-15", "2024-01-31", // [8-9]  keluar:    filter.From..To
		"2024-01-01", "2024-01-15", // [10-11] keluar_awal
		"2024-01-01", "2024-01-15", // [12-13] masuk_awal
		"2024-01-01", "2024-01-15", // [14-15] movein_awal
		"2024-01-01", "2024-01-15", // [16-17] peny_after_opname
	}

	if len(args) != len(want) {
		t.Fatalf("len(args): want %d, got %d", len(want), len(args))
	}
	for i, wantVal := range want {
		gotVal, ok := args[i].(string)
		if !ok {
			t.Fatalf("args[%d] bukan string: %T", i, args[i])
		}
		if gotVal != wantVal {
			t.Errorf("args[%d]: want %q, got %q", i, wantVal, gotVal)
		}
	}
}

// Test 6: rentang *_awal memakai batas eksklusif > tglInvAwal dan < filter.From
func TestBuildBaseQuery_AfterOpnameRangeIsExclusive(t *testing.T) {
	query, _ := buildBaseQuery(mustParseDate("2024-01-05"), mustParseDate("2024-01-31"), GetReportFilter{
		From: mustParseDate("2024-01-20"),
		To:   mustParseDate("2024-01-31"),
	})

	for _, want := range []string{
		"rmhead.trans_date > ? AND rmhead.trans_date < ?",
		"aphead.in_date > ? AND aphead.in_date < ?",
		"moveinhead.trans_date > ? AND moveinhead.trans_date < ?",
		"a.trans_date > ? AND a.trans_date < ?",
	} {
		if !strings.Contains(query, want) {
			t.Errorf("query harus mengandung %q", want)
		}
	}
}

// Test 7: filter item_code ditambahkan sebagai ILIKE pattern di akhir args
func TestBuildBaseQuery_ItemCodeLikePattern(t *testing.T) {
	query, args := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemCode: "MAT-001",
	})

	if !strings.Contains(query, "a.item_code ILIKE ?") {
		t.Error("query harus mengandung 'a.item_code ILIKE ?' (case-insensitive seperti MySQL)")
	}
	if lastArg, _ := args[len(args)-1].(string); lastArg != "%MAT-001%" {
		t.Errorf("item_code arg: want %%MAT-001%%, got %q", lastArg)
	}
}

// Test 8: filter item_name ditambahkan sebagai ILIKE pattern di akhir args
func TestBuildBaseQuery_ItemNameLikePattern(t *testing.T) {
	query, args := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemName: "Bahan Baku",
	})

	if !strings.Contains(query, "a.item_name ILIKE ?") {
		t.Error("query harus mengandung 'a.item_name ILIKE ?' (case-insensitive seperti MySQL)")
	}
	if lastArg, _ := args[len(args)-1].(string); lastArg != "%Bahan Baku%" {
		t.Errorf("item_name arg: want %%Bahan Baku%%, got %q", lastArg)
	}
}

// Test 9: saat tglInvAwal == filter.From, rentang *_awal kosong (> X AND < X)
func TestBuildBaseQuery_AfterOpnameEmptyRange_WhenAwalEqualsFrom(t *testing.T) {
	sameDate := mustParseDate("2024-01-15")
	_, args := buildBaseQuery(sameDate, mustParseDate("2024-01-31"), GetReportFilter{
		From: sameDate,
		To:   mustParseDate("2024-01-31"),
	})

	for i := 10; i < 18; i += 2 {
		if args[i] != args[i+1] {
			t.Errorf("args[%d..%d] = %v..%v, want sama (rentang eksklusif kosong)", i, i+1, args[i], args[i+1])
		}
	}
}

// ============================================================
// getBothOpnameDates — DB tests menggunakan sqlmock
// ============================================================

// Test 10: mengembalikan dua tanggal yang benar dari DB
func TestGetBothOpnameDates_ReturnsCorrectDates(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	rows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2024-01-01", "2024-01-31")

	mock.ExpectQuery("tr_inv_material_harian_head").
		WithArgs("2024-01-15", "2024-01-31", "2024-01-31").
		WillReturnRows(rows)

	awal, akhir, err := repo.getBothOpnameDates(
		context.Background(),
		mustParseDate("2024-01-15"),
		mustParseDate("2024-01-31"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantAwal := mustParseDate("2024-01-01")
	wantAkhir := mustParseDate("2024-01-31")

	if !awal.Equal(wantAwal) {
		t.Errorf("tglInvAwal: want %s, got %s", wantAwal.Format("2006-01-02"), awal.Format("2006-01-02"))
	}
	if !akhir.Equal(wantAkhir) {
		t.Errorf("tglInvAkhir: want %s, got %s", wantAkhir.Format("2006-01-02"), akhir.Format("2006-01-02"))
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations not met: %v", err)
	}
}

// Test 11: mengembalikan default "2000-01-01" saat tabel kosong (COALESCE menjamin ini)
func TestGetBothOpnameDates_DefaultWhenNoData(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	rows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2000-01-01", "2000-01-01")

	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(rows)

	defaultDate := mustParseDate("2000-01-01")

	awal, akhir, err := repo.getBothOpnameDates(
		context.Background(),
		mustParseDate("2024-01-15"),
		mustParseDate("2024-01-31"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !awal.Equal(defaultDate) {
		t.Errorf("awal default: want %s, got %s", defaultDate.Format("2006-01-02"), awal.Format("2006-01-02"))
	}
	if !akhir.Equal(defaultDate) {
		t.Errorf("akhir default: want %s, got %s", defaultDate.Format("2006-01-02"), akhir.Format("2006-01-02"))
	}
}

// Test 12: awal dan akhir bisa berbeda (fromDate lebih kecil dari toDate)
func TestGetBothOpnameDates_AwalAndAkhirCanDiffer(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	rows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2023-12-31", "2024-01-28") // tanggal berbeda

	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(rows)

	awal, akhir, err := repo.getBothOpnameDates(
		context.Background(),
		mustParseDate("2024-01-01"),
		mustParseDate("2024-01-31"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !awal.Equal(mustParseDate("2023-12-31")) {
		t.Errorf("awal: want 2023-12-31, got %s", awal.Format("2006-01-02"))
	}
	if !akhir.Equal(mustParseDate("2024-01-28")) {
		t.Errorf("akhir: want 2024-01-28, got %s", akhir.Format("2006-01-02"))
	}
}

// ============================================================
// getMaxMaterialHarianDate — DB tests menggunakan sqlmock
// ============================================================

// Test 13: mengembalikan tanggal maksimum yang benar
func TestGetMaxMaterialHarianDate_ReturnsMaxDate(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	rows := sqlmock.NewRows([]string{"trans_date"}).
		AddRow("2024-01-25")

	mock.ExpectQuery("tr_inv_material_harian_head").
		WithArgs("2024-01-31").
		WillReturnRows(rows)

	got, err := repo.getMaxMaterialHarianDate(context.Background(), mustParseDate("2024-01-31"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := mustParseDate("2024-01-25")
	if !got.Equal(want) {
		t.Errorf("want %s, got %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

// Test 14: mengembalikan "2000-01-01" saat tidak ada data (nilai COALESCE)
func TestGetMaxMaterialHarianDate_DefaultDate(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	rows := sqlmock.NewRows([]string{"trans_date"}).
		AddRow("2000-01-01")

	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(rows)

	got, err := repo.getMaxMaterialHarianDate(context.Background(), mustParseDate("2024-01-31"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := mustParseDate("2000-01-01")
	if !got.Equal(want) {
		t.Errorf("default date: want %s, got %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

// ============================================================
// GetReport — integration tests menggunakan sqlmock
// ============================================================

// Test 15: GetReport dengan pagination mengembalikan data dan total count
func TestGetReport_Paginated_ReturnsResultsAndCount(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	// Mock 1: getBothOpnameDates
	dateRows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2024-01-01", "2024-01-31")
	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(dateRows)

	// Mock 2: paginated query dengan COUNT(*) OVER()
	dataRows := sqlmock.NewRows([]string{
		"item_code", "item_name", "unit_code", "item_type_code", "item_group",
		"location_code", "awal", "masuk", "keluar", "peny", "akhir", "opname", "selisih",
		"_total_count",
	}).
		AddRow("MAT001", "Bahan A", "KG", "RAW", "MATERIAL", "", "100", "50", "30", "0", "120", "120", "0", int64(2)).
		AddRow("MAT002", "Bahan B", "PCS", "RAW", "MATERIAL", "", "200", "10", "5", "0", "205", "205", "0", int64(2))
	mock.ExpectQuery("_total_count").
		WillReturnRows(dataRows)

	results, total, err := repo.GetReport(context.Background(), GetReportFilter{
		From:  mustParseDate("2024-01-15"),
		To:    mustParseDate("2024-01-31"),
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total count: want 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("len(results): want 2, got %d", len(results))
	}
	if results[0].ItemCode != "MAT001" {
		t.Errorf("results[0].ItemCode: want MAT001, got %s", results[0].ItemCode)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

// Test 16: GetReport tanpa pagination mengembalikan semua data, total = len(results)
func TestGetReport_NoPagination_TotalEqualsLen(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	dateRows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2024-01-01", "2024-01-31")
	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(dateRows)

	dataRows := sqlmock.NewRows([]string{
		"item_code", "item_name", "unit_code", "item_type_code", "item_group",
		"location_code", "awal", "masuk", "keluar", "peny", "akhir", "opname", "selisih",
	}).
		AddRow("MAT001", "Bahan A", "KG", "RAW", "MATERIAL", "", "100", "50", "30", "0", "120", "120", "0").
		AddRow("MAT002", "Bahan B", "PCS", "RAW", "MATERIAL", "", "200", "10", "5", "0", "205", "205", "0").
		AddRow("MAT003", "Bahan C", "LTR", "RAW", "MATERIAL", "", "50", "20", "10", "0", "60", "60", "0")
	mock.ExpectQuery("SELECT").
		WillReturnRows(dataRows)

	results, total, err := repo.GetReport(context.Background(), GetReportFilter{
		From:  mustParseDate("2024-01-15"),
		To:    mustParseDate("2024-01-31"),
		Limit: 0, // no pagination
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(total) != len(results) {
		t.Errorf("total (%d) harus sama dengan len(results) (%d) saat no pagination", total, len(results))
	}
	if len(results) != 3 {
		t.Errorf("len(results): want 3, got %d", len(results))
	}
}

// Test 17: GetReport page 2 menghasilkan offset = (page-1)*limit
func TestGetReport_PageOffset(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewRawMaterialReportRepository(db)

	dateRows := sqlmock.NewRows([]string{"tgl_awal", "tgl_akhir"}).
		AddRow("2024-01-01", "2024-01-31")
	mock.ExpectQuery("tr_inv_material_harian_head").
		WillReturnRows(dateRows)

	// Verifikasi query mengandung LIMIT 10 OFFSET 10 (page=2, limit=10)
	dataRows := sqlmock.NewRows([]string{
		"item_code", "item_name", "unit_code", "item_type_code", "item_group",
		"location_code", "awal", "masuk", "keluar", "peny", "akhir", "opname", "selisih",
		"_total_count",
	}) // empty page
	mock.ExpectQuery("LIMIT 10 OFFSET 10").
		WillReturnRows(dataRows)

	results, total, err := repo.GetReport(context.Background(), GetReportFilter{
		From:  mustParseDate("2024-01-15"),
		To:    mustParseDate("2024-01-31"),
		Page:  2,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("total: want 0, got %d", total)
	}
	if len(results) != 0 {
		t.Errorf("len(results): want 0, got %d", len(results))
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

// Test 10: Itory menyimpan kode detail dengan huruf campur (mm313 vs MM313); MySQL dulu
// mencocokkan case-insensitive, jadi semua GROUP BY / join kode harus lewat UPPER.
func TestBuildBaseQuery_ItemCodesMatchCaseInsensitively(t *testing.T) {
	query, _ := buildBaseQuery(mustParseDate("2024-01-01"), mustParseDate("2024-01-31"), GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	})
	for _, line := range strings.Split(query, "\n") {
		line = strings.TrimSpace(line)
		if (strings.HasPrefix(line, "GROUP BY") || strings.Contains(line, " ON ") && strings.Contains(line, "item_code")) && !strings.Contains(line, "UPPER(") {
			t.Errorf("kode item dibandingkan case-sensitive: %q", line)
		}
	}
}
