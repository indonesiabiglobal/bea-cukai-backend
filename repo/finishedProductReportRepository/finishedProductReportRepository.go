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

// ---- DTOs for filters ----
type GetReportFilter struct {
	From     time.Time
	To       time.Time
	ItemCode string
	ItemName string
	Page     int
	Limit    int
}

// Helper function to get max product harian date
func (r *FinishedProductReportRepository) getMaxProductHarianDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_produk_harian_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ?", beforeDate).
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

// GetReport retrieves finished product report with complex inventory calculations
func (r *FinishedProductReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.FinishedProductReportResponse, int64, error) {
	// Get product harian dates similar to PHP logic
	tglInvAwal := filter.From.AddDate(0, 0, -1) // DATE_SUB(tgl_awal, INTERVAL 1 DAY)

	tglInvAkhir, err := r.getMaxProductHarianDate(ctx, filter.To)
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

	// Complex CTE query equivalent to the PHP version
	// Note: Different CTEs for finished products vs raw materials
	baseQuery := fmt.Sprintf(`
		WITH b AS (
			SELECT b.item_code, SUM(b.wh2 + b.wh1 + b.mesin + b.qc) as awal 
			FROM tr_inv_produk_harian_head a 
			INNER JOIN tr_inv_produk_harian_det b ON a.trans_no = b.trans_no 
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		), 
		c AS (
			SELECT no_produk, SUM(isi_palet) as masuk 
			FROM tr_produk_in_head 
			WHERE tgl_proses BETWEEN ? AND ? 
			GROUP BY no_produk
		), 
		d AS (
			SELECT b.no_produk, SUM(isi_palet) as keluar 
			FROM tr_export_head a 
			INNER JOIN tr_export_det b ON a.trans_no = b.trans_no 
			WHERE a.tgl_ekspor BETWEEN ? AND ? 
			GROUP BY b.no_produk
		), 
		e AS (
			SELECT b.item_code, SUM(qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			LEFT JOIN ms_item c ON b.item_code = c.item_code 
			WHERE a.trans_date BETWEEN ? AND ? AND c.item_group = 'PRODUCT' 
			GROUP BY b.item_code
		), 
		f AS (
			SELECT b.item_code, SUM(b.wh2 + b.wh1 + b.mesin + b.qc) as opname 
			FROM tr_inv_produk_harian_head a 
			INNER JOIN tr_inv_produk_harian_det b ON a.trans_no = b.trans_no 
			WHERE a.trans_date = ? 
			GROUP BY b.item_code
		), 
		a AS (
			SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code,
				IFNULL(b.awal, 0) as awal,
				IFNULL(c.masuk, 0) as masuk,
				IFNULL(d.keluar, 0) as keluar,
				IFNULL(e.peny, 0) as peny,
				(IFNULL(b.awal, 0) + IFNULL(c.masuk, 0) - IFNULL(d.keluar, 0) + IFNULL(e.peny, 0)) as akhir,
				IFNULL(f.opname, 0) as opname,
				0 as selisih
			FROM ms_item a 
			LEFT JOIN b ON a.item_code = b.item_code 
			LEFT JOIN c ON a.item_code = c.no_produk 
			LEFT JOIN d ON a.item_code = d.no_produk 
			LEFT JOIN e ON a.item_code = e.item_code 
			LEFT JOIN f ON a.item_code = f.item_code 
			WHERE a.item_group = 'PRODUCT' %s
		)
		SELECT a.*, opname as akhr, masuk - (akhir - opname) as msk 
		FROM a 
		WHERE a.awal <> 0 OR a.opname <> 0 OR a.keluar <> 0 OR a.peny <> 0 OR akhir <> 0 OR opname <> 0
	`, whereConditions)

	// Prepare arguments for the complex query
	queryArgs := []interface{}{
		tglInvAwal.Format("2006-01-02"),
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		filter.From.Format("2006-01-02"),
		filter.To.Format("2006-01-02"),
		tglInvAkhir.Format("2006-01-02"),
	}
	fmt.Println(queryArgs...)
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
	var results []model.FinishedProductReportResponse
	err = r.db.WithContext(ctx).Raw(finalQuery, queryArgs...).Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}
