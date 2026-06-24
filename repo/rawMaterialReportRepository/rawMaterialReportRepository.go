package rawMaterialReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type RawMaterialReportRepository struct {
	db *gorm.DB
}

func NewRawMaterialReportRepository(db *gorm.DB) *RawMaterialReportRepository {
	return &RawMaterialReportRepository{db: db}
}

// ---- DTOs for filters ----

type GetReportFilter struct {
	From     time.Time
	To       time.Time
	ItemCode string
	ItemName string
	Page     int
	Limit    int
}

// getMaxMaterialHarianDate returns the most recent trans_date <= beforeDate.
func (r *RawMaterialReportRepository) getMaxMaterialHarianDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_material_harian_head").
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

// getBothOpnameDates mengambil tglInvAwal dan tglInvAkhir dalam satu DB round trip.
func (r *RawMaterialReportRepository) getBothOpnameDates(ctx context.Context, fromDate, toDate time.Time) (awal, akhir time.Time, err error) {
	var result struct {
		TglAwal  string `gorm:"column:tgl_awal"`
		TglAkhir string `gorm:"column:tgl_akhir"`
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT
			IFNULL(MAX(CASE WHEN trans_date <= ? THEN trans_date END), '2000-01-01') AS tgl_awal,
			IFNULL(MAX(CASE WHEN trans_date <= ? THEN trans_date END), '2000-01-01') AS tgl_akhir
		FROM tr_inv_material_harian_head
		WHERE trans_date <= ?
	`, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"), toDate.Format("2006-01-02")).Scan(&result).Error

	defaultDate, _ := time.Parse("2006-01-02", "2000-01-01")
	if err != nil {
		return defaultDate, defaultDate, err
	}

	awal, errA := time.Parse("2006-01-02", result.TglAwal)
	if errA != nil {
		awal = defaultDate
	}
	akhir, errK := time.Parse("2006-01-02", result.TglAkhir)
	if errK != nil {
		akhir = defaultDate
	}
	return awal, akhir, nil
}

// buildBaseQuery membangun query CTE dan slice argumen yang terurut.
// Pure function — tidak ada DB call, aman untuk unit test.
func buildBaseQuery(tglInvAwal, tglInvAkhir time.Time, filter GetReportFilter) (string, []interface{}) {
	whereConditions := ""
	extraArgs := []interface{}{}

	if filter.ItemCode != "" {
		whereConditions += " AND item.item_code LIKE ?"
		extraArgs = append(extraArgs, "%"+filter.ItemCode+"%")
	}
	if filter.ItemName != "" {
		whereConditions += " AND item.item_name LIKE ?"
		extraArgs = append(extraArgs, "%"+filter.ItemName+"%")
	}

	queryAwal := "(IFNULL(b.awal, 0) + (IFNULL(masuk_awal.trf_in, 0) + IFNULL(movein_awal.movein_after, 0)) - IFNULL(keluar_awal.trf_out, 0) + IFNULL(peny_after_opname.peny, 0))"
	queryMasuk := "IFNULL(c.masuk, 0) + IFNULL(g.movein, 0)"
	akhirExpr := fmt.Sprintf("%s + %s - IFNULL(keluar.keluar,0) + IFNULL(e.peny, 0)", queryAwal, queryMasuk)

	opnameExpr := "IFNULL(f.opname, 0)"
	if tglInvAkhir.Format("2006-01-02") != filter.To.Format("2006-01-02") {
		opnameExpr = akhirExpr
	}

	query := fmt.Sprintf(`
		WITH b AS (
			SELECT b.item_code, SUM(b.wh2) AS awal
			FROM tr_inv_material_harian_head a
			INNER JOIN tr_inv_material_harian_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		),
		c AS (
			SELECT item_code, SUM(qty) AS masuk
			FROM tr_ap_inv_head a
			INNER JOIN tr_ap_inv_det b ON a.trans_no = b.trans_no
			WHERE a.in_date >= ? AND a.in_date <= ?
			GROUP BY item_code
		),
		e AS (
			SELECT b.item_code, SUM(qty) AS peny
			FROM tr_inv_adjust_head a
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no
			LEFT JOIN ms_item c ON b.item_code = c.item_code
			WHERE a.trans_date >= ? AND a.trans_date <= ?
			AND c.item_group = 'MATERIAL'
			GROUP BY b.item_code
		),
		f AS (
			SELECT b.item_code, SUM(b.wh2) AS opname
			FROM tr_inv_material_harian_head a
			INNER JOIN tr_inv_material_harian_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		),
		g AS (
			SELECT item_code, SUM(qty) AS movein
			FROM tr_inv_movein_head moveinhead
			INNER JOIN tr_inv_movein_det moveindet ON moveinhead.trans_no = moveindet.trans_no
			WHERE moveinhead.trans_date >= ? AND moveinhead.trans_date <= ?
			AND moveindet.location_code = 'WH-MAT-2'
			GROUP BY item_code
		),
		keluar AS (
			SELECT apdet.item_code, SUM(rmdet.qty) AS keluar
			FROM tr_inv_rm_head rmhead
			INNER JOIN tr_inv_rm_det rmdet ON rmhead.trans_no = rmdet.trans_no
			INNER JOIN tr_ap_inv_det apdet ON rmdet.data_no = apdet.data_no
			WHERE rmhead.trans_date >= ? AND rmhead.trans_date <= ?
			GROUP BY apdet.item_code
		),
		keluar_awal AS (
			SELECT apdet.item_code, SUM(rmdet.qty) AS trf_out
			FROM tr_inv_rm_head rmhead
			INNER JOIN tr_inv_rm_det rmdet ON rmhead.trans_no = rmdet.trans_no
			INNER JOIN tr_ap_inv_det apdet ON rmdet.data_no = apdet.data_no
			WHERE rmhead.trans_date > ? AND rmhead.trans_date < ?
			GROUP BY apdet.item_code
		),
		masuk_awal AS (
			SELECT apdet.item_code, SUM(apdet.qty) AS trf_in
			FROM tr_ap_inv_head aphead
			INNER JOIN tr_ap_inv_det apdet ON aphead.trans_no = apdet.trans_no
			WHERE aphead.in_date > ? AND aphead.in_date < ?
			GROUP BY apdet.item_code
		),
		movein_awal AS (
			SELECT item_code, SUM(qty) AS movein_after
			FROM tr_inv_movein_head moveinhead
			INNER JOIN tr_inv_movein_det moveindet ON moveinhead.trans_no = moveindet.trans_no
			WHERE moveinhead.trans_date > ? AND moveinhead.trans_date < ?
			AND moveindet.location_code = 'WH-MAT-2'
			GROUP BY item_code
		),
		peny_after_opname AS (
			SELECT b.item_code, SUM(b.qty) AS peny
			FROM tr_inv_adjust_head a
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no
			LEFT JOIN ms_item c ON b.item_code = c.item_code
			WHERE a.trans_date > ? AND a.trans_date < ? 
			AND c.item_group = 'MATERIAL'
			GROUP BY b.item_code
		),
		z AS (
			SELECT
				item.item_code, item.item_name, item.unit_code, item.item_type_code, item.item_group,
				'' AS location_code,
				%s AS awal,
				%s AS masuk,
				IFNULL(keluar.keluar, 0) AS keluar,
				IFNULL(e.peny, 0) AS peny,
				%s AS akhir,
				%s AS opname,
				(%s) - (%s) AS selisih
			FROM ms_item item
			LEFT JOIN b ON item.item_code = b.item_code
			LEFT JOIN c ON item.item_code = c.item_code
			LEFT JOIN e ON item.item_code = e.item_code
			LEFT JOIN f ON item.item_code = f.item_code
			LEFT JOIN g ON item.item_code = g.item_code
			LEFT JOIN keluar ON item.item_code = keluar.item_code
			LEFT JOIN keluar_awal ON item.item_code = keluar_awal.item_code
			LEFT JOIN masuk_awal ON item.item_code = masuk_awal.item_code
			LEFT JOIN movein_awal ON item.item_code = movein_awal.item_code
			LEFT JOIN peny_after_opname ON item.item_code = peny_after_opname.item_code
			WHERE item.item_group = 'MATERIAL' %s
			AND item.item_type_code NOT LIKE 'Recycle%'
		)
		SELECT * FROM z WHERE z.awal <> 0 OR z.opname <> 0 OR z.masuk <> 0 OR z.akhir <> 0 OR z.peny <> 0
	`, queryAwal, queryMasuk, akhirExpr, opnameExpr, akhirExpr, opnameExpr, whereConditions)

	// Urutan args harus sesuai dengan urutan ? di query CTE di atas:
	// b(1) + c(2) + e(2) + f(1) + g(2) + out_after(2) + in_after(2) + movein_after(2) + peny_after(2) = 16
	baseArgs := []interface{}{
		tglInvAwal.Format("2006-01-02"),                                   // b:              awal date
		filter.From.Format("2006-01-02"),                                  // c:              masuk from
		filter.To.Format("2006-01-02"),                                    // c:              masuk to
		filter.From.Format("2006-01-02"),                                  // e:              peny from
		filter.To.Format("2006-01-02"),                                    // e:              peny to
		tglInvAkhir.Format("2006-01-02"),                                  // f:              opname date
		filter.From.Format("2006-01-02"),                                  // g:              movein from
		filter.To.Format("2006-01-02"),                                    // g:              movein to
		filter.From.Format("2006-01-02"),                                  // keluar from
		filter.To.Format("2006-01-02"),                                    // keluar to
		tglInvAwal.Format("2006-01-02"), filter.From.Format("2006-01-02"), // keluar_awal
		tglInvAwal.Format("2006-01-02"), filter.From.Format("2006-01-02"), // masuk_awal
		tglInvAwal.Format("2006-01-02"), filter.From.Format("2006-01-02"), // movein_awal
		tglInvAwal.Format("2006-01-02"), filter.From.Format("2006-01-02"), // peny_after_opname
	}

	return query, append(baseArgs, extraArgs...)
}

// GetReport mengambil laporan bahan baku dengan kalkulasi inventori kompleks.
// Optimasi:
//   - getBothOpnameDates: 2 query tanggal → 1 DB round trip
//   - COUNT(*) OVER():    query count + data → 1 DB round trip (hemat 1 eksekusi CTE penuh)
func (r *RawMaterialReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.RawMaterialReportResponse, int64, error) {
	tglInvAwal, tglInvAkhir, err := r.getBothOpnameDates(ctx, filter.From, filter.To)
	if err != nil {
		return nil, 0, err
	}

	baseQuery, queryArgs := buildBaseQuery(tglInvAwal, tglInvAkhir, filter)

	var (
		results    []model.RawMaterialReportResponse
		totalCount int64
	)

	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}

		// Gunakan window function agar count dan data diperoleh dalam satu pass CTE.
		// Ini menghindari eksekusi ulang seluruh CTE hanya untuk COUNT.
		type rowWithCount struct {
			model.RawMaterialReportResponse
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

		results = make([]model.RawMaterialReportResponse, len(rows))
		for i, row := range rows {
			results[i] = row.RawMaterialReportResponse
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
