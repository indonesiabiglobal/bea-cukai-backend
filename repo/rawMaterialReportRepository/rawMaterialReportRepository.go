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

// Helper function to get max material harian date
func (r *RawMaterialReportRepository) getMaxMaterialHarianDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_material_harian_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ?", beforeDate.Format("2006-01-02")).
		Scan(&result).Error

	if err != nil {
		// Return default date if error
		defaultDate, _ := time.Parse("2006-01-02", "2000-01-01")
		return defaultDate, err
	}

	// Parse the string date to time.Time
	parsedDate, err := time.Parse("2006-01-02", result.TransDate)
	if err != nil {
		// Return default date if parsing fails
		defaultDate, _ := time.Parse("2006-01-02", "2000-01-01")
		return defaultDate, err
	}

	return parsedDate, nil
}

// GetReport retrieves raw material report with complex inventory calculations
func (r *RawMaterialReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.RawMaterialReportResponse, int64, error) {
	tglInvAwal, err := r.getMaxMaterialHarianDate(ctx, filter.From)
	if err != nil {
		return nil, 0, err
	}

	tglInvAkhir, err := r.getMaxMaterialHarianDate(ctx, filter.To)
	if err != nil {
		return nil, 0, err
	}

	// Build WHERE conditions for filters
	whereConditions := ""
	args := []interface{}{}

	if filter.ItemCode != "" {
		whereConditions += " AND a.item_code LIKE ?"
		args = append(args, "%"+filter.ItemCode+"%")
	}

	if filter.ItemName != "" {
		whereConditions += " AND a.item_name LIKE ?"
		args = append(args, "%"+filter.ItemName+"%")
	}

	queryAwal := "(IFNULL(b.awal, 0) + (IFNULL(in_after_opname.trf_in, 0) + IFNULL(movein_after_opname.movein_after, 0)) - IFNULL(out_after_opname.trf_out, 0) + IFNULL(peny_after_opname.peny, 0))"
	queryMasuk := "IFNULL(c.masuk, 0) + IFNULL(g.movein, 0)"
	// Tentukan ekspresi opname: jika tglInvAkhir == filter.To, gunakan nilai akhir (computed); jika tidak, gunakan data opname dari tabel
	akhirExpr := fmt.Sprintf("%s + %s - 0 + IFNULL(e.peny, 0)", queryAwal, queryMasuk)
	opnameExpr := "IFNULL(f.opname, 0)"
	if tglInvAkhir.Format("2006-01-02") != filter.To.Format("2006-01-02") {
		opnameExpr = akhirExpr
	}

	// Complex CTE query equivalent to the PHP version
	baseQuery := fmt.Sprintf(`
		WITH b AS (
			-- SELECT b.item_code, SUM(b.wh2 + b.wh1 + b.mesin) as awal 
			SELECT b.item_code, SUM(b.wh2) as awal 
			FROM tr_inv_material_harian_head a 
			INNER JOIN tr_inv_material_harian_det b ON a.trans_no = b.trans_no 
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		), 
		c AS (
			SELECT item_code, SUM(qty) as masuk 
			FROM tr_ap_inv_head a 
			INNER JOIN tr_ap_inv_det b ON a.trans_no = b.trans_no 
			WHERE a.in_date BETWEEN ? AND ? 
			GROUP BY item_code
		), 
		e AS (
			SELECT b.item_code, SUM(qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			LEFT JOIN ms_item c ON b.item_code = c.item_code 
			WHERE a.trans_date BETWEEN ? AND ? AND c.item_group = 'MATERIAL' 
			GROUP BY b.item_code
		), 
		f AS (
			-- SELECT b.item_code, SUM(b.wh2 + b.wh1 + b.mesin) as opname 
			SELECT b.item_code, SUM(b.wh2) as opname 
			FROM tr_inv_material_harian_head a 
			INNER JOIN tr_inv_material_harian_det b ON a.trans_no = b.trans_no 
			WHERE a.trans_date = ? 
			GROUP BY b.item_code
		), 
		g AS (
			SELECT item_code, SUM(qty) as movein 
			FROM tr_inv_movein_head moveinhead
			INNER JOIN tr_inv_movein_det moveindet ON moveinhead.trans_no = moveindet.trans_no 
			WHERE moveinhead.trans_date BETWEEN ? AND ? 
			AND moveindet.location_code = 'WH-MAT-2'
			GROUP BY item_code
		),
		out_after_opname AS (
			SELECT apdet.item_code, SUM(rmdet.qty) as trf_out 
			FROM tr_inv_rm_head rmhead
			INNER JOIN tr_inv_rm_det rmdet ON rmhead.trans_no = rmdet.trans_no
			INNER JOIN tr_ap_inv_det apdet ON rmdet.data_no = apdet.data_no
			WHERE rmhead.trans_date BETWEEN ? AND ? 
			GROUP BY apdet.item_code
		),
		in_after_opname AS (
			SELECT apdet.item_code, SUM(apdet.qty) as trf_in 
			FROM tr_ap_inv_head aphead
			INNER JOIN tr_ap_inv_det apdet ON aphead.trans_no = apdet.trans_no
			WHERE aphead.in_date BETWEEN ? AND ? 
			GROUP BY apdet.item_code
		),
		movein_after_opname AS (
			SELECT item_code, SUM(qty) as movein_after 
			FROM tr_inv_movein_head moveinhead
			INNER JOIN tr_inv_movein_det moveindet ON moveinhead.trans_no = moveindet.trans_no 
			WHERE moveinhead.trans_date BETWEEN ? AND ? 
			AND moveindet.location_code = 'WH-MAT-2'
			GROUP BY item_code
		),
		peny_after_opname AS (
			SELECT b.item_code, SUM(b.qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			LEFT JOIN ms_item c ON b.item_code = c.item_code 
			WHERE a.trans_date BETWEEN ? AND ? AND c.item_group = 'MATERIAL' 
			GROUP BY b.item_code
		),
		z AS (
			SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code,
				%s as awal,
				%s as masuk,
				%s - %s as keluar,
				IFNULL(e.peny, 0) as peny,
				%s as akhir,
				%s as opname,
				0 as selisih
			FROM ms_item a
			LEFT JOIN b ON a.item_code = b.item_code
			LEFT JOIN c ON a.item_code = c.item_code
			LEFT JOIN e ON a.item_code = e.item_code
			LEFT JOIN f ON a.item_code = f.item_code
			LEFT JOIN g ON a.item_code = g.item_code
			LEFT JOIN out_after_opname ON a.item_code = out_after_opname.item_code
			LEFT JOIN in_after_opname ON a.item_code = in_after_opname.item_code
			LEFT JOIN movein_after_opname ON a.item_code = movein_after_opname.item_code
			LEFT JOIN peny_after_opname ON a.item_code = peny_after_opname.item_code
			WHERE a.item_group = 'MATERIAL' %s
		)
		SELECT * FROM z WHERE z.awal <> 0 OR z.opname <> 0 OR z.masuk <> 0 OR z.akhir <> 0 OR z.peny <> 0
	`, queryAwal, queryMasuk, akhirExpr, opnameExpr, akhirExpr, opnameExpr, whereConditions)
	fmt.Println("akhirExpr:", akhirExpr)
	fmt.Println("opnameExpr:", opnameExpr)

	// Prepare arguments for the complex query
	queryArgs := []interface{}{
		tglInvAwal.Format("2006-01-02"),
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		tglInvAkhir.Format("2006-01-02"), // opname date
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		tglInvAwal.AddDate(0, 0, 1).Format("2006-01-02"),   // trf_out start date
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // trf_out end date
		tglInvAwal.AddDate(0, 0, 1).Format("2006-01-02"),   // trf_in start date
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // trf_in end date
		tglInvAwal.AddDate(0, 0, 1).Format("2006-01-02"),   // movein_after start date
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // movein_after end date
		tglInvAwal.AddDate(0, 0, 1).Format("2006-01-02"),   // peny_after_opname start date
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // peny_after_opname end date
	}

	queryArgs = append(queryArgs, args...)

	// Get total count first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery", baseQuery)
	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, queryArgs...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Add LIMIT and OFFSET for pagination
	finalQuery := baseQuery
	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}
		finalQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)
	}

	// Execute the final query
	var results []model.RawMaterialReportResponse
	err = r.db.WithContext(ctx).Raw(finalQuery, queryArgs...).Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}
