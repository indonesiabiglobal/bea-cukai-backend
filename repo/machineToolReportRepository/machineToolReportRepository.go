package machineToolReportRepository

import (
	"Bea-Cukai/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type MachineToolReportRepository struct {
	db *gorm.DB
}

func NewMachineToolReportRepository(db *gorm.DB) *MachineToolReportRepository {
	return &MachineToolReportRepository{db: db}
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

// Helper function to get max opname date for MESIN items
func (r *MachineToolReportRepository) getMaxOpnameDate(ctx context.Context, beforeDate time.Time) (time.Time, error) {
	var result struct {
		TransDate string `gorm:"column:trans_date"`
	}

	err := r.db.WithContext(ctx).
		Table("tr_inv_opname_head").
		Select("COALESCE(MAX(trans_date)::text, '2000-01-01') as trans_date").
		Where("trans_date <= ? AND item_group = 'MESIN'", beforeDate.Format("2006-01-02")).
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

// GetReport retrieves machine and tool report with complex inventory calculations
func (r *MachineToolReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.MachineToolReportResponse, int64, error) {
	// Get opname dates similar to PHP logic
	tglInvAwal, err := r.getMaxOpnameDate(ctx, filter.From.AddDate(0, 0, -1))
	if err != nil {
		return nil, 0, err
	}

	tglInvAkhir, err := r.getMaxOpnameDate(ctx, filter.To)
	if err != nil {
		return nil, 0, err
	}

	// Build WHERE conditions for filters
	whereConditions := ""
	args := []interface{}{}

	if filter.ItemCode != "" {
		whereConditions += " AND a.item_code ILIKE ?"
		args = append(args, "%"+filter.ItemCode+"%")
	}

	if filter.ItemName != "" {
		whereConditions += " AND a.item_name ILIKE ?"
		args = append(args, "%"+filter.ItemName+"%")
	}

	// Complex query equivalent to the PHP version for MESIN items.
	// Defaults are 0.00 so empty values print "0.00" like MySQL's DECIMAL(16,2) (scanned straight into strings).
	baseQuery := fmt.Sprintf(`
		SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code, 
			COALESCE(b.awal, COALESCE(x.awal, 0.00)) as awal,  
			COALESCE(c.masuk, 0.00) as masuk, 
			COALESCE(d.keluar, 0.00) as keluar, 
			COALESCE(e.peny, 0.00) as peny,  
			(COALESCE(b.awal, COALESCE(x.awal, 0.00)) + COALESCE(c.masuk, 0.00) - COALESCE(d.keluar, 0.00) + COALESCE(e.peny, 0.00)) as akhir, 
			COALESCE(f.opname, COALESCE(y.opname, 0.00)) as opname, 
			0 as selisih, 
			(COALESCE(d.keluar, 0.00) + COALESCE(f.opname, COALESCE(y.opname, 0.00))) - (COALESCE(b.awal, COALESCE(x.awal, 0.00)) + COALESCE(c.masuk, 0.00) - COALESCE(d.keluar, 0.00) + COALESCE(e.peny, 0.00)) as kel,
			to_char(COALESCE(b.awal, COALESCE(x.awal, 0.00)), 'FM999,999,999,990') as awl,
			to_char(COALESCE(c.masuk, 0.00), 'FM999,999,999,990') as msk,
			to_char(COALESCE(d.keluar, 0.00), 'FM999,999,999,990') as kl,
			to_char(COALESCE(e.peny, 0.00), 'FM999,999,999,990') as pen,
			to_char(COALESCE(f.opname, COALESCE(y.opname, 0.00)), 'FM999,999,999,990') as opm
		FROM ms_item a  
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as awal 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date = ?
			GROUP BY UPPER(b.item_code)
		) as b ON UPPER(a.item_code) = b.item_code 
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as awal 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date = ?
			GROUP BY UPPER(b.item_code)
		) as x ON UPPER(a.item_code) = x.item_code
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as masuk 
			FROM tr_inv_movein_head a 
			INNER JOIN tr_inv_movein_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date BETWEEN ? AND ?
			GROUP BY UPPER(b.item_code)
		) as c ON UPPER(a.item_code) = c.item_code  
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as keluar 
			FROM tr_inv_moveout_head a 
			INNER JOIN tr_inv_moveout_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date BETWEEN ? AND ?
			GROUP BY UPPER(b.item_code)
		) as d ON UPPER(a.item_code) = d.item_code  
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(qty) as peny 
			FROM tr_inv_adjust_head a 
			INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no 
			LEFT JOIN ms_item c ON UPPER(b.item_code) = UPPER(c.item_code)  
			WHERE a.trans_date BETWEEN ? AND ?
			GROUP BY UPPER(b.item_code)
		) as e ON UPPER(a.item_code) = e.item_code  
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as opname 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date = ?
			GROUP BY UPPER(b.item_code)
		) as f ON UPPER(a.item_code) = f.item_code 
		LEFT JOIN (
			SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as opname 
			FROM tr_inv_opname_head a 
			INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no  
			WHERE a.trans_date = ?
			GROUP BY UPPER(b.item_code)
		) as y ON UPPER(a.item_code) = y.item_code
		WHERE a.item_group = 'MESIN' %s
	`, whereConditions)

	// Filter out zero records as in PHP
	finalQuery := fmt.Sprintf(`
		SELECT * FROM (%s) as a 
		WHERE a.awal <> 0 OR a.opname <> 0 OR a.peny <> 0 OR a.masuk <> 0 OR a.keluar <> 0
	`, baseQuery)

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as subquery", finalQuery)
	var totalCount int64
	err = r.db.WithContext(ctx).Raw(countQuery, queryArgs...).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Add LIMIT and OFFSET for pagination
	// Explicit order: MySQL returned ms_item order (case-insensitive item_code, i.e. uppercase weights); Postgres needs it for stable paging.
	paginatedQuery := finalQuery + ` ORDER BY UPPER(item_code) COLLATE "C", item_code`
	if filter.Limit > 0 {
		offset := 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}
		paginatedQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)
	}

	// Execute the final query
	var results []model.MachineToolReportResponse
	err = r.db.WithContext(ctx).Raw(paginatedQuery, queryArgs...).Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}
