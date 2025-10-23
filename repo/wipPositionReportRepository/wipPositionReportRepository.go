package wipPositionReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ---- Constructor ----

type WipPositionReportRepository struct {
	db *gorm.DB
}

func NewWipPositionReportRepository(db *gorm.DB) *WipPositionReportRepository {
	return &WipPositionReportRepository{db: db}
}

// ---- DTOs for filters ----
type GetReportFilter struct {
	TglAwal  time.Time
	TglAkhir time.Time
	ItemCode string
	ItemName string
	Page     int
	Limit    int
}

// Helper function to get max inventory opname date
func (r *WipPositionReportRepository) getMaxOpnameDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_opname_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ? AND item_group = 'WIP'", beforeDate.Format("2006-01-02")).
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
} // GetReport retrieves WIP position report with complex inventory calculations
func (r *WipPositionReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.WipPositionReportResponse, int64, error) {
	// Get inventory opname dates similar to PHP logic
	tglInvAwal, err := r.getMaxOpnameDate(ctx, filter.TglAwal.AddDate(0, 0, -1))
	if err != nil {
		return nil, 0, err
	}

	tglInvAkhir, err := r.getMaxOpnameDate(ctx, filter.TglAkhir)
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

	// Complex SQL query equivalent to the PHP version
	baseQuery := fmt.Sprintf(`
		SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code,
			IFNULL(b.awal, IFNULL(x.awal, 0)) as awal,
			IFNULL(c.masuk, 0) as masuk,
			IFNULL(d.keluar, 0) as keluar,
			0 as peny,
			0 as akhir,
			IFNULL(f.opname, IFNULL(y.opname, 0)) as opname
		FROM ms_item a
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as awal
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		) as b ON a.item_code = b.item_code
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as awal
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		) as x ON a.item_code = x.item_code
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as masuk
			FROM tr_inv_movein_head a 
			INNER JOIN tr_inv_movein_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date BETWEEN ? AND ?
			GROUP BY b.item_code
		) as c ON a.item_code = c.item_code
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as keluar
			FROM tr_inv_moveout_head a 
			INNER JOIN tr_inv_moveout_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date BETWEEN ? AND ?
			GROUP BY b.item_code
		) as d ON a.item_code = d.item_code
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as opname
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		) as f ON a.item_code = f.item_code
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as opname
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
			WHERE a.trans_date = ?
			GROUP BY b.item_code
		) as y ON a.item_code = y.item_code
		WHERE a.item_group = 'WIP' %s
	`, whereConditions)

	// Prepare arguments for the complex query
	queryArgs := []interface{}{
		filter.TglAwal.AddDate(0, 0, -1).Format("2006-01-02"), // DATE_SUB(tgl_awal, INTERVAL 1 DAY)
		tglInvAwal.Format("2006-01-02"),
		filter.TglAwal.Format("2006-01-02"),
		filter.TglAkhir.Format("2006-01-02"),
		filter.TglAwal.Format("2006-01-02"),
		filter.TglAkhir.Format("2006-01-02"),
		filter.TglAkhir.Format("2006-01-02"),
		tglInvAkhir.Format("2006-01-02"),
	}
	queryArgs = append(queryArgs, args...)

	// Get total count first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery WHERE subquery.awal <> 0", baseQuery)
	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, queryArgs...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Add LIMIT and OFFSET for pagination
	finalQuery := fmt.Sprintf("SELECT * FROM (%s) as subquery WHERE subquery.awal <> 0", baseQuery)
	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}
		finalQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)
	}

	// Execute the final query
	var rawResults []struct {
		ItemCode     string          `gorm:"column:item_code"`
		ItemName     string          `gorm:"column:item_name"`
		UnitCode     string          `gorm:"column:unit_code"`
		ItemTypeCode string          `gorm:"column:item_type_code"`
		ItemGroup    string          `gorm:"column:item_group"`
		LocationCode string          `gorm:"column:location_code"`
		Awal         decimal.Decimal `gorm:"column:awal"`
		Masuk        decimal.Decimal `gorm:"column:masuk"`
		Keluar       decimal.Decimal `gorm:"column:keluar"`
		Peny         decimal.Decimal `gorm:"column:peny"`
		Akhir        decimal.Decimal `gorm:"column:akhir"`
		Opname       decimal.Decimal `gorm:"column:opname"`
	}

	err = r.db.WithContext(ctx).Raw(finalQuery, queryArgs...).Scan(&rawResults).Error
	if err != nil {
		return nil, 0, err
	}

	// Transform to response format (kode_barang, nama_barang, sat, jumlah)
	var results []model.WipPositionReportResponse
	for _, raw := range rawResults {
		response := model.WipPositionReportResponse{
			ItemCode: raw.ItemCode,
			ItemName: raw.ItemName,
			UnitCode: raw.UnitCode,
			Jumlah:   raw.Awal.String(),
		}
		results = append(results, response)
	}

	return results, totalCount, nil
}
