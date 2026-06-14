package finishedProductReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type FinishedProductReportRepository struct {
	db *gorm.DB
}

func NewFinishedProductReportRepository(db *gorm.DB) *FinishedProductReportRepository {
	return &FinishedProductReportRepository{db: db}
}

// ---- DTOs ----

type GetReportFilter struct {
	From     time.Time
	To       time.Time
	ItemCode string
	ItemName string
	Page     int
	Limit    int
}

// productOpnameDates menyimpan tanggal opname per tipe head yang dibutuhkan laporan.
//
// Struktur data di tr_inv_produk_harian_head:
//   - opname_gudang2 = 1  → det berisi wh2 saja  (wh1/mesin/qc = 0)
//   - opname_gudang2 = 0  → det berisi wh1+mesin+qc (wh2 = 0)
//
// Kedua tipe bisa memiliki trans_date berbeda sehingga masing-masing butuh
// tanggal referensinya sendiri agar tidak ada data yang terlewat.
type productOpnameDates struct {
	TglAwalGudang2  time.Time // MAX(trans_date WHERE opname_gudang2=1 AND <= From) — CTE b(gudang2=1), c, d, e
	TglAwalGudang0  time.Time // MAX(trans_date WHERE opname_gudang2=0 AND <= From) — CTE b(gudang2=0)
	TglAkhirGudang2 time.Time // MAX(trans_date WHERE opname_gudang2=1 AND <= To)   — CTE f(gudang2=1)
	TglAkhirGudang0 time.Time // MAX(trans_date WHERE opname_gudang2=0 AND <= To)   — CTE f(gudang2=0)
}

// getMaxProductHarianDate returns the most recent trans_date <= beforeDate.
// Dipertahankan untuk backward compatibility / testing.
func (r *FinishedProductReportRepository) getMaxProductHarianDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_produk_harian_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ?", beforeDate.Format("2006-01-02")).
		Scan(&result).Error

	defaultDate, _ := time.Parse("2006-01-02", "2000-01-01")
	if err != nil {
		return defaultDate, err
	}

	parsedDate, err := time.Parse("2006-01-02", result.TransDate)
	if err != nil {
		return defaultDate, err
	}
	return parsedDate, nil
}

// getAllProductOpnameDates mengambil 4 tanggal opname dalam satu DB round trip.
// Masing-masing tanggal difilter per opname_gudang2 karena kedua tipe head
// bisa memiliki trans_date yang berbeda.
func (r *FinishedProductReportRepository) getAllProductOpnameDates(ctx context.Context, fromDate, toDate time.Time) (productOpnameDates, error) {
	var result struct {
		TglAwalGudang2  string `gorm:"column:tgl_awal_gudang2"`
		TglAkhirGudang2 string `gorm:"column:tgl_akhir_gudang2"`
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			IFNULL(MAX(CASE WHEN opname_gudang2 = 1 AND trans_date < ? THEN trans_date END), '2000-01-01') AS tgl_awal_gudang2,
			IFNULL(MAX(CASE WHEN opname_gudang2 = 1 AND trans_date <= ? THEN trans_date END), '2000-01-01') AS tgl_akhir_gudang2
		FROM tr_inv_produk_harian_head
		WHERE trans_date <= ?
	`, fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
		toDate.Format("2006-01-02")).Scan(&result).Error

	defaultDate, _ := time.Parse("2006-01-02", "2000-01-01")
	dates := productOpnameDates{
		TglAwalGudang2:  defaultDate,
		TglAkhirGudang2: defaultDate,
	}

	if err != nil {
		return dates, err
	}

	if t, e := time.Parse("2006-01-02", result.TglAwalGudang2); e == nil {
		dates.TglAwalGudang2 = t
	}
	if t, e := time.Parse("2006-01-02", result.TglAkhirGudang2); e == nil {
		dates.TglAkhirGudang2 = t
	}

	return dates, nil
}

// CTE c/d/e menangani periode laporan (filter.From → filter.To) untuk kolom masuk/keluar/peny.
func buildBaseQuery(dates productOpnameDates, filter GetReportFilter) (string, []any) {
	whereConditions := ""
	extraArgs := []any{}

	if filter.ItemCode != "" {
		whereConditions += " AND a.item_code LIKE ?"
		extraArgs = append(extraArgs, "%"+filter.ItemCode+"%")
	}
	if filter.ItemName != "" {
		whereConditions += " AND a.item_name LIKE ?"
		extraArgs = append(extraArgs, "%"+filter.ItemName+"%")
	}

	awalExpr := "(IFNULL(b.awal, 0) + IFNULL(masuk_awal.masuk, 0) - IFNULL(keluar_awal.keluar, 0) + IFNULL(peny_awal.peny, 0))"

	akhirExpr := fmt.Sprintf("%s + IFNULL(c.masuk, 0) - IFNULL(d.keluar, 0) + IFNULL(e.peny, 0)", awalExpr)

	opnameExpr := "IFNULL(f.opname, 0)"
	if dates.TglAkhirGudang2.Format("2006-01-02") != filter.To.Format("2006-01-02") {
		opnameExpr = akhirExpr
	}

	query := fmt.Sprintf(`
		WITH b AS (
			SELECT det.item_code,
				SUM(
					CASE WHEN head.opname_gudang2 = 1 THEN det.wh2 ELSE 0 END
				) AS awal
			FROM tr_inv_produk_harian_head head
			INNER JOIN tr_inv_produk_harian_det det ON head.trans_no = det.trans_no
			WHERE (head.opname_gudang2 = 1 AND head.trans_date = ?)
			GROUP BY det.item_code
		),
		masuk_awal AS (
			SELECT no_produk, SUM(isi_palet) AS masuk
			FROM tr_produk_in_head
			WHERE tgl_proses > ? AND tgl_proses < ?
			GROUP BY no_produk
		),
		keluar_awal AS (
			SELECT b.no_produk, SUM(isi_palet) AS keluar
			FROM tr_export_head a
			INNER JOIN tr_export_det b ON a.trans_no = b.trans_no
			WHERE a.tgl_ekspor > ? AND a.tgl_ekspor < ?
			GROUP BY b.no_produk
		),
		peny_awal AS (
			SELECT b.item_code, SUM(qty) AS peny
			FROM tr_inv_adjust_head a
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no
			LEFT JOIN ms_item c ON b.item_code = c.item_code
			WHERE a.trans_date > ? AND a.trans_date < ? AND c.item_group = 'PRODUCT'
			GROUP BY b.item_code
		),
		c AS (
			SELECT no_produk, SUM(isi_palet) AS masuk
			FROM tr_produk_in_head
			WHERE tgl_proses >= ? AND tgl_proses <= ?
			GROUP BY no_produk
		),
		d AS (
			SELECT b.no_produk, SUM(isi_palet) AS keluar
			FROM tr_export_head a
			INNER JOIN tr_export_det b ON a.trans_no = b.trans_no
			WHERE a.tgl_ekspor >= ? AND a.tgl_ekspor <= ?
			GROUP BY b.no_produk
		),
		e AS (
			SELECT b.item_code, SUM(qty) AS peny
			FROM tr_inv_adjust_head a
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no
			LEFT JOIN ms_item c ON b.item_code = c.item_code
			WHERE a.trans_date >= ? AND a.trans_date <= ? AND c.item_group = 'PRODUCT'
			GROUP BY b.item_code
		),
		f AS (
			SELECT det.item_code,
				SUM(
					CASE WHEN head.opname_gudang2 = 1 THEN det.wh2 ELSE 0 END
				) AS opname
			FROM tr_inv_produk_harian_head head
			INNER JOIN tr_inv_produk_harian_det det ON head.trans_no = det.trans_no
			WHERE (head.opname_gudang2 = 1 AND head.trans_date = ?)
			GROUP BY det.item_code
		),
		a AS (
			SELECT
				a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group,
				'' AS location_code,
				%s AS awal,
				IFNULL(c.masuk, 0) AS masuk,
				IFNULL(d.keluar, 0) AS keluar,
				IFNULL(e.peny, 0) AS peny,
				%s AS akhir,
				%s AS opname,
				0 AS selisih
			FROM ms_item a
			LEFT JOIN b ON a.item_code = b.item_code
			LEFT JOIN masuk_awal ON a.item_code = masuk_awal.no_produk
			LEFT JOIN keluar_awal ON a.item_code = keluar_awal.no_produk
			LEFT JOIN peny_awal ON a.item_code = peny_awal.item_code
			LEFT JOIN c ON a.item_code = c.no_produk
			LEFT JOIN d ON a.item_code = d.no_produk
			LEFT JOIN e ON a.item_code = e.item_code
			LEFT JOIN f ON a.item_code = f.item_code
			WHERE a.item_group = 'PRODUCT' %s
		)
		SELECT a.*
		FROM a
		WHERE a.awal <> 0 OR a.opname <> 0 OR a.keluar <> 0 OR a.peny <> 0 OR akhir <> 0
	`, awalExpr, akhirExpr, opnameExpr, whereConditions)

	fmt.Println("dates.TglAwalGudang2", dates.TglAwalGudang2)
	baseArgs := []any{
		dates.TglAwalGudang2.Format("2006-01-02"),  // b:           opname_gudang2=1 AND trans_date = ?
		dates.TglAwalGudang2.Format("2006-01-02"),  // masuk_awal:  tgl_proses > ?
		filter.From.Format("2006-01-02"),           // masuk_awal:  tgl_proses <= ?
		dates.TglAwalGudang2.Format("2006-01-02"),  // keluar_awal: tgl_ekspor > ?
		filter.From.Format("2006-01-02"),           // keluar_awal: tgl_ekspor <= ?
		dates.TglAwalGudang2.Format("2006-01-02"),  // peny_awal:   trans_date > ?
		filter.From.Format("2006-01-02"),           // peny_awal:   trans_date <= ?
		filter.From.Format("2006-01-02"),           // c:           tgl_proses > ?
		filter.To.Format("2006-01-02"),             // c:           tgl_proses <= ?
		filter.From.Format("2006-01-02"),           // d:           tgl_ekspor > ?
		filter.To.Format("2006-01-02"),             // d:           tgl_ekspor <= ?
		filter.From.Format("2006-01-02"),           // e:           trans_date > ?
		filter.To.Format("2006-01-02"),             // e:           trans_date <= ?
		dates.TglAkhirGudang2.Format("2006-01-02"), // f:           opname_gudang2=1 AND trans_date = ?
	}

	return query, append(baseArgs, extraArgs...)
}

// GetReport mengambil laporan produk jadi dengan kalkulasi inventori kompleks.
// Optimasi:
//   - getAllProductOpnameDates: 4 tanggal → 1 DB round trip
//   - COUNT(*) OVER():          query count + data → 1 DB round trip
func (r *FinishedProductReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.FinishedProductReportResponse, int64, error) {
	dates, err := r.getAllProductOpnameDates(ctx, filter.From, filter.To)
	if err != nil {
		return nil, 0, err
	}

	baseQuery, queryArgs := buildBaseQuery(dates, filter)

	var (
		results    []model.FinishedProductReportResponse
		totalCount int64
	)

	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}

		type rowWithCount struct {
			model.FinishedProductReportResponse
			TotalCount int64 `gorm:"column:_total_count"`
		}

		paginatedQuery := fmt.Sprintf(`
			SELECT inner_q.*, COUNT(*) OVER() AS _total_count
			FROM (%s) AS inner_q
			LIMIT %d OFFSET %d
		`, baseQuery, filter.Limit, offset)

		var rows []rowWithCount
		if err = r.db.WithContext(ctx).Raw(paginatedQuery, queryArgs...).Scan(&rows).Error; err != nil {
			return nil, 0, err
		}

		results = make([]model.FinishedProductReportResponse, len(rows))
		for i, row := range rows {
			results[i] = row.FinishedProductReportResponse
		}
		if len(rows) > 0 {
			totalCount = rows[0].TotalCount
		}
	} else {
		if err = r.db.WithContext(ctx).Raw(baseQuery, queryArgs...).Scan(&results).Error; err != nil {
			return nil, 0, err
		}
		totalCount = int64(len(results))
	}

	return results, totalCount, nil
}
