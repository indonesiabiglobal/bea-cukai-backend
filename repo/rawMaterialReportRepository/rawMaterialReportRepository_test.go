package rawMaterialReportRepository

import (
	"context"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
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
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
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

// Test 1: saat tglInvAkhir == filter.To, opname harus pakai IFNULL(f.opname, 0)
func TestBuildBaseQuery_OpnameUsesF_WhenAkhirEqualsTo(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"), // sama dengan tglInvAkhir
	}

	query, _ := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	if !strings.Contains(query, "IFNULL(f.opname, 0) AS opname") {
		t.Error("opname harus IFNULL(f.opname, 0) ketika tglInvAkhir == filter.To")
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

	if strings.Contains(query, "IFNULL(f.opname, 0) AS opname") {
		t.Error("opname TIDAK boleh pakai IFNULL(f.opname, 0) ketika tglInvAkhir != filter.To")
	}
}

// Test 3: jumlah base args harus tepat 16 (tanpa filter item_code/item_name)
func TestBuildBaseQuery_ArgsCount_NoFilter(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	}

	_, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	// b(1)+c(2)+e(2)+f(1)+g(2)+out_after(2)+in_after(2)+movein_after(2)+peny_after(2) = 16
	const wantCount = 16
	if len(args) != wantCount {
		t.Errorf("jumlah args: want %d, got %d", wantCount, len(args))
	}
}

// Test 4: filter item_code dan item_name menambah 2 args ekstra
func TestBuildBaseQuery_ArgsCount_WithBothFilters(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemCode: "MAT001",
		ItemName: "Bahan A",
	}

	_, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	const wantCount = 18 // 16 base + 2 filter
	if len(args) != wantCount {
		t.Errorf("jumlah args dengan filter: want %d, got %d", wantCount, len(args))
	}
}

// Test 5: urutan args harus persis sesuai urutan ? di CTE
// Ini adalah test paling kritis — urutan args salah = data salah
func TestBuildBaseQuery_ArgsOrder(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: mustParseDate("2024-01-15"),
		To:   mustParseDate("2024-01-31"),
	}

	_, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	// afterStart = tglInvAwal+1 = "2024-01-02"
	// afterEnd   = filter.From-1 = "2024-01-14"
	want := []string{
		"2024-01-01", // [0]  b:              tglInvAwal
		"2024-01-15", // [1]  c:              filter.From
		"2024-01-31", // [2]  c:              filter.To
		"2024-01-15", // [3]  e:              filter.From
		"2024-01-31", // [4]  e:              filter.To
		"2024-01-31", // [5]  f:              tglInvAkhir
		"2024-01-15", // [6]  g:              filter.From
		"2024-01-31", // [7]  g:              filter.To
		"2024-01-02", // [8]  out_after start (tglInvAwal+1)
		"2024-01-14", // [9]  out_after end   (filter.From-1)
		"2024-01-02", // [10] in_after start
		"2024-01-14", // [11] in_after end
		"2024-01-02", // [12] movein_after start
		"2024-01-14", // [13] movein_after end
		"2024-01-02", // [14] peny_after start
		"2024-01-14", // [15] peny_after end
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

// Test 6: tanggal after_opname dihitung +1/-1 hari dari tglInvAwal/filter.From
func TestBuildBaseQuery_AfterOpnameDateBoundary(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-05")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: mustParseDate("2024-01-20"),
		To:   mustParseDate("2024-01-31"),
	}

	_, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	// out_after: args[8] = tglInvAwal+1 = "2024-01-06", args[9] = filter.From-1 = "2024-01-19"
	tests := []struct {
		idx  int
		want string
		desc string
	}{
		{8, "2024-01-06", "out_after start (tglInvAwal+1)"},
		{9, "2024-01-19", "out_after end   (filter.From-1)"},
		{10, "2024-01-06", "in_after start"},
		{11, "2024-01-19", "in_after end"},
		{12, "2024-01-06", "movein_after start"},
		{13, "2024-01-19", "movein_after end"},
		{14, "2024-01-06", "peny_after start"},
		{15, "2024-01-19", "peny_after end"},
	}
	for _, tc := range tests {
		got, ok := args[tc.idx].(string)
		if !ok {
			t.Fatalf("args[%d] bukan string", tc.idx)
		}
		if got != tc.want {
			t.Errorf("%s [args[%d]]: want %q, got %q", tc.desc, tc.idx, tc.want, got)
		}
	}
}

// Test 7: filter item_code ditambahkan sebagai LIKE pattern di akhir args
func TestBuildBaseQuery_ItemCodeLikePattern(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemCode: "MAT-001",
	}

	query, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	if !strings.Contains(query, "a.item_code LIKE ?") {
		t.Error("query harus mengandung 'a.item_code LIKE ?'")
	}
	lastArg, ok := args[len(args)-1].(string)
	if !ok {
		t.Fatal("arg item_code bukan string")
	}
	if lastArg != "%MAT-001%" {
		t.Errorf("item_code arg: want %%MAT-001%%, got %q", lastArg)
	}
}

// Test 8: filter item_name ditambahkan sebagai LIKE pattern di akhir args
func TestBuildBaseQuery_ItemNameLikePattern(t *testing.T) {
	tglInvAwal := mustParseDate("2024-01-01")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From:     mustParseDate("2024-01-15"),
		To:       mustParseDate("2024-01-31"),
		ItemName: "Bahan Baku",
	}

	query, args := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	if !strings.Contains(query, "a.item_name LIKE ?") {
		t.Error("query harus mengandung 'a.item_name LIKE ?'")
	}
	lastArg, ok := args[len(args)-1].(string)
	if !ok {
		t.Fatal("arg item_name bukan string")
	}
	if lastArg != "%Bahan Baku%" {
		t.Errorf("item_name arg: want %%Bahan Baku%%, got %q", lastArg)
	}
}

// Test 9: saat tglInvAwal == filter.From, range after_opname kosong (start > end)
// Ini valid: tidak ada transaksi untuk disesuaikan
func TestBuildBaseQuery_AfterOpnameEmptyRange_WhenAwalEqualsFrom(t *testing.T) {
	sameDate := mustParseDate("2024-01-15")
	tglInvAkhir := mustParseDate("2024-01-31")
	filter := GetReportFilter{
		From: sameDate,
		To:   mustParseDate("2024-01-31"),
	}

	_, args := buildBaseQuery(sameDate, tglInvAkhir, filter)

	afterStart := args[8].(string) // tglInvAwal+1 = "2024-01-16"
	afterEnd := args[9].(string)   // filter.From-1 = "2024-01-14"

	// afterStart > afterEnd → BETWEEN akan return 0 rows (behavior yang benar)
	start, _ := time.Parse("2006-01-02", afterStart)
	end, _ := time.Parse("2006-01-02", afterEnd)
	if !start.After(end) {
		t.Errorf("expected afterStart (%s) > afterEnd (%s) ketika tglInvAwal == filter.From", afterStart, afterEnd)
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

// Test 11: mengembalikan default "2000-01-01" saat tabel kosong (IFNULL menjamin ini)
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

// Test 14: mengembalikan "2000-01-01" saat tidak ada data (nilai IFNULL)
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
