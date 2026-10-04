package auxiliaryMaterialReportRepository

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/entryProductRepository"
	"context"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ---- Constructor ----

// MasukFunc returns the received qty per entryProductRepository.ItemKey for entries with trans_date in [from, to].
type MasukFunc func(ctx context.Context, from, to time.Time) (map[string]decimal.Decimal, error)

// AuxiliaryMaterialReportRepository reads stock (opname/adjust) from Itory; "masuk" comes from the
// live pemasukan sources (FKI + Itory), which can't be joined in one SQL query.
type AuxiliaryMaterialReportRepository struct {
	db    *gorm.DB
	masuk MasukFunc
}

func NewAuxiliaryMaterialReportRepository(itoryDB *gorm.DB, masuk MasukFunc) *AuxiliaryMaterialReportRepository {
	return &AuxiliaryMaterialReportRepository{db: itoryDB, masuk: masuk}
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

type rawResult struct {
	ItemCode     string  `gorm:"column:item_code"`
	ItemName     string  `gorm:"column:item_name"`
	UnitCode     string  `gorm:"column:unit_code"`
	ItemTypeCode string  `gorm:"column:item_type_code"`
	ItemGroup    string  `gorm:"column:item_group"`
	LocationCode string  `gorm:"column:location_code"`
	Awal         float64 `gorm:"column:awal"`
	Peny         float64 `gorm:"column:peny"`
	Opname       float64 `gorm:"column:opname"`
	Masuk        float64 `gorm:"-"`
	Keluar       float64 `gorm:"-"`
}

// Every item of the group with its opening balance, adjustments and closing opname.
const stockSQL = `
	SELECT a.item_code, a.item_name, a.unit_code, a.item_type_code, a.item_group, '' as location_code,
		COALESCE(b.awal, 0) as awal,
		COALESCE(e.peny, 0) as peny,
		COALESCE(f.opname, 0) as opname
	FROM ms_item a
	LEFT JOIN (
		SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as awal
		FROM tr_inv_opname_head a
		INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
		INNER JOIN ms_item item ON UPPER(item.item_code) = UPPER(b.item_code)
		WHERE a.trans_date = ? AND
		item.item_group = ?
		GROUP BY UPPER(b.item_code)
	) as b ON UPPER(a.item_code) = b.item_code
	LEFT JOIN (
		SELECT UPPER(b.item_code) AS item_code, SUM(qty) as peny
		FROM tr_inv_adjust_head a
		INNER JOIN tr_inv_adjust_det b ON a.trans_no = b.trans_no
		INNER JOIN ms_item c ON UPPER(b.item_code) = UPPER(c.item_code)
		WHERE c.item_group = ? AND a.trans_date BETWEEN ? AND ?
		GROUP BY UPPER(b.item_code)
	) as e ON UPPER(a.item_code) = e.item_code
	LEFT JOIN (
		SELECT UPPER(b.item_code) AS item_code, SUM(b.qty) as opname
		FROM tr_inv_opname_head a
		INNER JOIN tr_inv_opname_det b ON a.trans_no = b.trans_no
		WHERE a.item_group = ? AND a.trans_date = ?
		GROUP BY UPPER(b.item_code)
	) as f ON UPPER(a.item_code) = f.item_code
	WHERE a.item_group = ?
	AND a.item_code NOT IN ('IK0107', 'TL0001', 'IT0105')`

// GetReport retrieves auxiliary material report with complex inventory calculations
func (r *AuxiliaryMaterialReportRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.AuxiliaryMaterialReportResponse, int64, error) {
	query := stockSQL
	args := []any{
		filter.From.AddDate(0, 0, -1).Format("2006-01-02"), // beginning balance = opname the day before
		filter.Lap,
		filter.Lap, filter.From.Format("2006-01-02"), filter.To.Format("2006-01-02"), // peny
		filter.Lap, filter.To.Format("2006-01-02"), // closing opname
		filter.Lap,
	}
	if filter.ItemCode != "" {
		query += " AND a.item_code ILIKE ?"
		args = append(args, "%"+filter.ItemCode+"%")
	}
	if filter.ItemName != "" {
		query += " AND a.item_name ILIKE ?"
		args = append(args, "%"+filter.ItemName+"%")
	}
	// MySQL returned ms_item order (case-insensitive item_code, i.e. uppercase weights); Postgres needs it explicit.
	query += `
	ORDER BY UPPER(a.item_code) COLLATE "C", a.item_code`

	var rows []rawResult
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	masuk, err := r.masuk(ctx, filter.From, filter.To)
	if err != nil {
		return nil, 0, err
	}

	rows = withMasuk(rows, masuk)
	total := int64(len(rows))
	return format(page(rows, filter.Page, filter.Limit)), total, nil
}

// withMasuk fills masuk, derives keluar = awal + masuk - opname, and keeps only rows with movement
// (the old HAVING; akhir is always 0).
func withMasuk(rows []rawResult, masuk map[string]decimal.Decimal) []rawResult {
	kept := rows[:0]
	for _, row := range rows {
		row.Masuk = masuk[entryProductRepository.ItemKey(row.ItemCode)].InexactFloat64()
		row.Keluar = row.Awal + row.Masuk - row.Opname
		if row.Awal != 0 || row.Masuk != 0 || row.Keluar != 0 || row.Peny != 0 || row.Opname != 0 {
			kept = append(kept, row)
		}
	}
	return kept
}

// page applies LIMIT/OFFSET (page < 1 = first page, limit <= 0 = all). Compared via division so a
// huge page/limit from the query string can't overflow.
func page(rows []rawResult, pageNo, limit int) []rawResult {
	if limit <= 0 {
		return rows
	}
	pageNo = max(pageNo, 1)
	if pageNo-1 > len(rows)/limit {
		return rows[:0]
	}
	start := min((pageNo-1)*limit, len(rows))
	return rows[start : start+min(limit, len(rows)-start)]
}

// format applies number formatting like the old PHP report.
func format(rows []rawResult) []model.AuxiliaryMaterialReportResponse {
	var results []model.AuxiliaryMaterialReportResponse
	for _, raw := range rows {
		results = append(results, model.AuxiliaryMaterialReportResponse{
			ItemCode:     raw.ItemCode,
			ItemName:     raw.ItemName,
			UnitCode:     raw.UnitCode,
			ItemTypeCode: raw.ItemTypeCode,
			ItemGroup:    raw.ItemGroup,
			LocationCode: raw.LocationCode,
			Awal:         formatNumber(raw.Awal, 2),
			Masuk:        formatNumber(raw.Masuk, 2),
			Keluar:       formatNumber(raw.Keluar, 2),
			Peny:         formatNumber(raw.Peny, 2),
			Akhir:        "0",
			Opname:       formatNumber(raw.Opname, 2),
			Selisih:      "0",
		})
	}
	return results
}

// formatNumber formats a float64 to string with specified decimal places, matching PHP's number_format
func formatNumber(value float64, decimals int) string {
	return strconv.FormatFloat(value, 'f', decimals, 64)
}
