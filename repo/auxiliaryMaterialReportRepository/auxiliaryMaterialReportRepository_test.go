package auxiliaryMaterialReportRepository

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWithMasuk_DerivesKeluarAndDropsIdleRows(t *testing.T) {
	rows := []rawResult{
		{ItemCode: "box01", Awal: 10, Opname: 4}, // masuk via normalized key
		{ItemCode: "IDLE"},                       // no movement -> dropped
		{ItemCode: "ONLYIN"},                     // only masuk -> kept
		{ItemCode: "PENY", Peny: -2},             // only adjustment -> kept
	}
	masuk := map[string]decimal.Decimal{"BOX01": decimal.NewFromFloat(5.5), "ONLYIN": decimal.NewFromInt(3)}

	got := withMasuk(rows, masuk)
	if len(got) != 3 || got[0].ItemCode != "box01" || got[1].ItemCode != "ONLYIN" || got[2].ItemCode != "PENY" {
		t.Fatalf("got %+v", got)
	}
	if got[0].Masuk != 5.5 || got[0].Keluar != 11.5 { // 10 + 5.5 - 4
		t.Errorf("box01 masuk=%v keluar=%v", got[0].Masuk, got[0].Keluar)
	}
	if got[1].Keluar != 3 {
		t.Errorf("ONLYIN keluar=%v", got[1].Keluar)
	}
}

func TestPage(t *testing.T) {
	rows := make([]rawResult, 25)
	maxInt := int(^uint(0) >> 1)
	cases := []struct{ page, limit, want int }{
		{1, 10, 10}, {3, 10, 5}, {4, 10, 0}, {0, 10, 10}, {1, 0, 25}, {maxInt, maxInt, 0},
	}
	for _, c := range cases {
		if got := len(page(rows, c.page, c.limit)); got != c.want {
			t.Errorf("page(%d,%d) = %d rows, want %d", c.page, c.limit, got, c.want)
		}
	}
}

func TestGetReport_MergesLiveMasukIntoItoryStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`AND a.item_name ILIKE $9
	ORDER BY UPPER(a.item_code) COLLATE "C", a.item_code`)).
		WithArgs("2026-08-31", "BOX", "BOX", "2026-09-01", "2026-09-30", "BOX", "2026-09-30", "BOX", "%kardus%").
		WillReturnRows(sqlmock.NewRows([]string{"item_code", "item_name", "awal", "peny", "opname"}).
			AddRow("XX01", "Kardus A", "2", "0", "1").
			AddRow("XX02", "Kardus B", "0", "0", "0"))

	var gotFrom, gotTo time.Time
	masuk := func(_ context.Context, from, to time.Time) (map[string]decimal.Decimal, error) {
		gotFrom, gotTo = from, to
		return map[string]decimal.Decimal{"XX01": decimal.NewFromInt(4)}, nil
	}
	from, to := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local), time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)
	res, total, err := NewAuxiliaryMaterialReportRepository(gormDB, masuk).GetReport(context.Background(),
		GetReportFilter{From: from, To: to, Lap: "BOX", ItemName: "kardus", Page: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(res) != 1 || res[0].Masuk != "4.00" || res[0].Keluar != "5.00" || res[0].Awal != "2.00" {
		t.Errorf("total=%d res=%+v", total, res)
	}
	if !gotFrom.Equal(from) || !gotTo.Equal(to) {
		t.Errorf("masuk called with %v..%v", gotFrom, gotTo)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
