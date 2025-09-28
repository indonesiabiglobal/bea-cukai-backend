package rejectScrapReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type RejectScrapReportRepository struct {
	db *gorm.DB
}

func NewRejectScrapReportRepository(db *gorm.DB) *RejectScrapReportRepository {
	return &RejectScrapReportRepository{db: db}
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

// Helper function to get max opname date for SCRAP items
func (r *RejectScrapReportRepository) getMaxScrapOpnameDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_opname_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ? AND item_group = 'SCRAP'", beforeDate).
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

// GetReport retrieves reject and scrap report with complex inventory calculations
func (r *RejectScrapReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.RejectScrapReportResponse, int64, error) {
	// Get opname dates similar to PHP logic
	tglInvAwal, err := r.getMaxScrapOpnameDate(ctx, filter.From.AddDate(0, 0, -1))
	if err != nil {
		return nil, 0, err
	}

	tglInvAkhir, err := r.getMaxScrapOpnameDate(ctx, filter.To)
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

	// Complex query equivalent to the PHP version for SCRAP items
	baseQuery := fmt.Sprintf(`
		SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code, 
			IFNULL(b.awal, IFNULL(x.awal, 0)) as awal,  
			IFNULL(c.masuk, 0) as masuk, 
			IFNULL(d.keluar, 0) as keluar, 
			IFNULL(e.peny, 0) as peny,  
			(IFNULL(b.awal, IFNULL(x.awal, 0)) + IFNULL(c.masuk, 0) - IFNULL(d.keluar, 0) + IFNULL(e.peny, 0)) as akhir, 
			IFNULL(f.opname, IFNULL(y.opname, 0)) as opname,
			0 as selisih
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
			SELECT b.item_code, SUM(qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			LEFT JOIN ms_item c ON b.item_code = c.item_code  
			WHERE a.trans_date BETWEEN ? AND ? AND c.item_group = 'WIP'
			GROUP BY b.item_code
		) as e ON a.item_code = e.item_code  
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
		WHERE a.item_group = 'SCRAP' %s
	`, whereConditions)

	// Prepare arguments for the complex query
	queryArgs := []interface{}{
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // DATE_SUB for beginning balance
		tglInvAwal.Format("2006-01-02"),                    // Max opname date before start
		filter.From.Format("2006-01-02"),                   // Move in start date
		filter.To.Format("2006-01-02"),                     // Move in end date
		filter.From.Format("2006-01-02"),                   // Move out start date
		filter.To.Format("2006-01-02"),                     // Move out end date
		filter.From.Format("2006-01-02"),                   // Adjustment start date
		filter.To.Format("2006-01-02"),                     // Adjustment end date
		filter.To.Format("2006-01-02"),                     // End opname exact date
		tglInvAkhir.Format("2006-01-02"),                   // Max opname date at end
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
	paginatedQuery := baseQuery
	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}
		paginatedQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)
	}

	// Execute the final query to get raw data
	type RawResult struct {
		ItemCode     string  `gorm:"column:item_code"`
		ItemName     string  `gorm:"column:item_name"`
		UnitCode     string  `gorm:"column:unit_code"`
		ItemTypeCode string  `gorm:"column:item_type_code"`
		ItemGroup    string  `gorm:"column:item_group"`
		LocationCode string  `gorm:"column:location_code"`
		Awal         float64 `gorm:"column:awal"`
		Masuk        float64 `gorm:"column:masuk"`
		Keluar       float64 `gorm:"column:keluar"`
		Peny         float64 `gorm:"column:peny"`
		Akhir        float64 `gorm:"column:akhir"`
		Opname       float64 `gorm:"column:opname"`
		Selisih      float64 `gorm:"column:selisih"`
	}

	var rawResults []RawResult
	err = r.db.WithContext(ctx).Raw(paginatedQuery, queryArgs...).Scan(&rawResults).Error
	if err != nil {
		return nil, 0, err
	}

	// Process results like PHP does - apply number formatting and keluar calculation
	var results []model.RejectScrapReportResponse
	for _, raw := range rawResults {
		// Calculate keluar as in PHP: keluar + opname - akhir
		calculatedKeluar := raw.Keluar + raw.Opname - raw.Akhir

		result := model.RejectScrapReportResponse{
			ItemCode:     raw.ItemCode,
			ItemName:     raw.ItemName,
			UnitCode:     raw.UnitCode,
			ItemTypeCode: raw.ItemTypeCode,
			ItemGroup:    raw.ItemGroup,
			LocationCode: raw.LocationCode,
			Awal:         formatNumber(raw.Awal, 2),
			Masuk:        formatNumber(raw.Masuk, 2),
			Keluar:       formatNumber(calculatedKeluar, 2),
			Peny:         formatNumber(raw.Peny, 2),
			Akhir:        strconv.FormatFloat(raw.Akhir, 'f', -1, 64),
			Opname:       formatNumber(raw.Opname, 2),
			Selisih:      strconv.FormatFloat(raw.Selisih, 'f', -1, 64),
		}
		results = append(results, result)
	}

	return results, totalCount, nil
}

// formatNumber formats a float64 to string with specified decimal places, matching PHP's number_format
func formatNumber(value float64, decimals int) string {
	return strconv.FormatFloat(value, 'f', decimals, 64)
}
