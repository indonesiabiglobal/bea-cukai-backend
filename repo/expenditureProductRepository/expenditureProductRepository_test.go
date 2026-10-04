package expenditureProductRepository

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

var sept = GetReportFilter{
	From: time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local),
	To:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local),
}

func TestBuildQuery_FiltersInsideSourceWhere(t *testing.T) {
	f := sept
	f.PabeanType, f.NoPabean, f.ProductName = "BC 3.0", "123", "bolt"
	q, args := buildQuery(f)

	if strings.Contains(q, "SELECT * FROM (") {
		t.Error("filters must not wrap a derived table (MySQL 5.5 materializes it)")
	}
	want := "AND canceled_date < approved_date\n\tAND cukai_date BETWEEN ? AND ? AND jenis_pabean = ? AND cukai_no = ? AND det.item_name LIKE ?"
	if !strings.HasSuffix(q, want) {
		t.Errorf("query must end with %q, got:\n%s", want, q)
	}
	if len(args) != 5 || args[0] != "2026-09-01" || args[4] != "%bolt%" {
		t.Errorf("unexpected args %v", args)
	}
}

func TestRawCodes_MatchesOldSyncSemantics(t *testing.T) {
	cases := map[string][]string{
		"ABC":  {"ABC"},         // reported "1ABC"
		"1ABC": {"1ABC", "ABC"}, // reported "11ABC" or "1ABC"
		"1":    {""},            // empty item code
	}
	for code, want := range cases {
		if got := rawCodes(code); strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%q: got %v want %v", code, got, want)
		}
	}
}

func TestGetReport_PagesInSQLAndNormalizesItemCode(t *testing.T) {
	db, mock := newMockMySQL(t)
	repo := NewExpenditureProductRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM (\nSELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY tgl_pabean DESC, no_pabean DESC, invoice_no DESC, head.trans_no DESC, det.seq_num ASC\nLIMIT ? OFFSET ?")).
		WithArgs("2026-09-01", "2026-09-30", 10, 10).
		WillReturnRows(sqlmock.NewRows([]string{"trans_no", "item_code"}).AddRow("INV-1", "1ABC").AddRow("INV-2", "1"))

	f := sept
	f.Page, f.Limit = 2, 10
	res, total, err := repo.GetReport(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if total != 25 || len(res) != 2 || res[0].ItemCode != "ABC" || res[1].ItemCode != "1" {
		t.Errorf("got total=%d res=%+v", total, res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetReport_HugePageSkipsQueryWithoutOverflow(t *testing.T) {
	db, mock := newMockMySQL(t)
	repo := NewExpenditureProductRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	maxInt := int(^uint(0) >> 1)
	f := sept
	f.Page, f.Limit = maxInt, maxInt
	res, total, err := repo.GetReport(context.Background(), f)
	if err != nil || total != 5 || len(res) != 0 {
		t.Errorf("got res=%d total=%d err=%v", len(res), total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetReport_ExportSkipsCountAndProductCodeExpandsIn(t *testing.T) {
	db, mock := newMockMySQL(t)
	repo := NewExpenditureProductRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("det.item_code IN (?,?)\nORDER BY tgl_pabean ASC, no_pabean ASC, invoice_no ASC, head.trans_no ASC, det.seq_num ASC")).
		WithArgs("2026-09-01", "2026-09-30", "1ABC", "ABC").
		WillReturnRows(sqlmock.NewRows([]string{"trans_no"}))

	f := sept
	f.ProductCode, f.IsExport = "1ABC", true
	res, total, err := repo.GetReport(context.Background(), f)
	if err != nil || res == nil || len(res) != 0 || total != 0 {
		t.Errorf("empty result must be non-nil []: res=%v total=%d err=%v", res, total, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
