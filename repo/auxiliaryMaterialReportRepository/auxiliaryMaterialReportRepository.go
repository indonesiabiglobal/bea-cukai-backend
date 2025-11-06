package auxiliaryMaterialReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type AuxiliaryMaterialReportRepository struct {
	db *gorm.DB
}

func NewAuxiliaryMaterialReportRepository(db *gorm.DB) *AuxiliaryMaterialReportRepository {
	return &AuxiliaryMaterialReportRepository{db: db}
}

// ---- DTOs for filters ----
type GetReportFilter struct {
	From     time.Time
	To       time.Time
	ItemCode string
	ItemName string
	Lap      string // Dynamic item group parameter
	Page     int
	Limit    int
}

// Helper function to get max opname date for specific item group (lap parameter)
func (r *AuxiliaryMaterialReportRepository) getMaxOpnameDateByGroup(ctx context.Context, beforeDate time.Time, itemGroup string) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_opname_head").
		Select("IFNULL(MAX(trans_date), '2000-01-01') as trans_date").
		Where("trans_date <= ? AND item_group = ?", beforeDate, itemGroup).
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

// GetReport retrieves auxiliary material report with complex inventory calculations
func (r *AuxiliaryMaterialReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.AuxiliaryMaterialReportResponse, int64, error) {
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

	// Complex query equivalent to the PHP version for auxiliary materials
	// Note: This uses different data sources than other reports
	baseQuery := fmt.Sprintf(`
		SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code, 
			IFNULL(b.awal, 0) as awal,
			IFNULL(c.masuk, 0) as masuk,
			IFNULL(e.peny, 0) as peny,
			0 as akhir,
			IFNULL(f.opname, 0) as opname,
			(IFNULL(b.awal, 0) + IFNULL(c.masuk, 0) - IFNULL(f.opname, 0)) as keluar
		FROM ms_item a 
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as awal 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
			INNER JOIN ms_item item ON item.item_code = b.item_code 
			WHERE a.trans_date = ? AND
			item.item_group = ?
			GROUP BY b.item_code
		) as b ON a.item_code = b.item_code 
		LEFT JOIN (
			SELECT
				pb.item_code,
				SUM( pb.rcv_qty ) AS masuk,
				item.item_name 
			FROM
				tr_pemasukan_barang pb
				INNER JOIN ms_item item ON item.item_code = pb.item_code 
			WHERE
				item.item_group = ? AND
				pb.trans_date BETWEEN ? AND ? 
			GROUP BY
				item_code
			ORDER BY 
				item_code
		) as c ON a.item_code = c.item_code 
		LEFT JOIN (
			SELECT b.item_code, SUM(qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			INNER JOIN ms_item c ON b.item_code = c.item_code 
			WHERE c.item_group = ? AND a.trans_date BETWEEN ? AND ?
			GROUP BY b.item_code
		) as e ON a.item_code = e.item_code 
		LEFT JOIN (
			SELECT b.item_code, SUM(b.qty) as opname 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no 
			WHERE a.item_group = ? AND a.trans_date = ?
			GROUP BY b.item_code
		) as f ON a.item_code = f.item_code 
		WHERE a.item_group = ? %s
		AND a.item_code NOT IN ('IK0107', 'TL0001', 'IT0105')
		HAVING awal <> 0 OR masuk <> 0 OR keluar <> 0 OR peny <> 0 OR akhir <> 0 OR opname <> 0
	`, whereConditions)

	// Prepare arguments for the complex query
	queryArgs := []interface{}{
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // DATE_SUB for beginning balance
		filter.Lap,                       // Item group for awal query
		filter.Lap,                       // Item group for masuk query
		filter.From.Format("2006-01-02"), // Masuk start date
		filter.To.Format("2006-01-02"),   // Masuk end date
		filter.Lap,                       // Item group for peny query
		filter.From.Format("2006-01-02"), // Peny start date
		filter.To.Format("2006-01-02"),   // Peny end date
		filter.Lap,                       // Item group for opname query
		filter.To.Format("2006-01-02"),   // End opname exact date
		filter.Lap,                       // Main WHERE item group
	}
	queryArgs = append(queryArgs, args...)

	// Get total count first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery", baseQuery)
	var totalCount int64
	err := r.db.WithContext(ctx).Raw(countQuery, queryArgs...).Scan(&totalCount).Error
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
	}

	var rawResults []RawResult
	err = r.db.WithContext(ctx).Raw(paginatedQuery, queryArgs...).Scan(&rawResults).Error
	if err != nil {
		return nil, 0, err
	}

	// Process results like PHP does - apply number formatting
	var results []model.AuxiliaryMaterialReportResponse
	for _, raw := range rawResults {
		result := model.AuxiliaryMaterialReportResponse{
			ItemCode:     raw.ItemCode,
			ItemName:     raw.ItemName,
			UnitCode:     raw.UnitCode,
			ItemTypeCode: raw.ItemTypeCode,
			ItemGroup:    raw.ItemGroup,
			LocationCode: raw.LocationCode,
			Awal:         formatNumber(raw.Awal, 2),
			Masuk:        formatNumber(raw.Masuk, 2),
			Keluar:       formatNumber(raw.Keluar, 2), // Keluar sudah dihitung di SQL
			Peny:         formatNumber(raw.Peny, 2),
			Akhir:        strconv.FormatFloat(raw.Akhir, 'f', -1, 64),
			Opname:       formatNumber(raw.Opname, 2),
			Selisih:      "0", // Always 0 as in PHP
		}
		results = append(results, result)
	}

	return results, totalCount, nil
}

// formatNumber formats a float64 to string with specified decimal places, matching PHP's number_format
func formatNumber(value float64, decimals int) string {
	return strconv.FormatFloat(value, 'f', decimals, 64)
}
