package expenditureProductRepository

import (
	"Bea-Cukai/model"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

// ExpenditureProductRepository reads pengeluaran barang live from FKI (MySQL).
// Itory has no sales invoices (tr_ar_inv_* is empty), so FKI is the only source.
type ExpenditureProductRepository struct {
	db *gorm.DB
}

func NewExpenditureProductRepository(fkiDB *gorm.DB) *ExpenditureProductRepository {
	return &ExpenditureProductRepository{db: fkiDB}
}

// Filter struct for GetReport method
type GetReportFilter struct {
	From         time.Time
	To           time.Time
	PabeanType   string
	ProductGroup string
	NoPabean     string
	ProductCode  string
	ProductName  string
	Page         int
	Limit        int
	IsExport     bool
}

// ---- Source query ----
// Column aliases = model.ExpenditureProduct fields. Filters are appended to this query's own WHERE
// using the real column names: FKI runs MySQL 5.5, which materializes a wrapping derived table.

// approved_date defaults to 2012-01-01 (not NULL) when unapproved; canceled_date < approved_date
// keeps docs re-approved after a cancel and drops those canceled after approval.
const fkiExpenditureSQL = `
SELECT
	jenis_pabean,
	cukai_no AS no_pabean,
	cukai_date AS tgl_pabean,
	invoice_no AS trans_no,
	trans_date,
	cust.cust_code,
	cust.cust_name,
	CONCAT('1', det.item_code) AS item_code,
	det.item_name,
	COALESCE(qty, 0) AS dlv_qty,
	det.unit AS sales_unit,
	curr_code,
	COALESCE(price, 0) AS net_price,
	COALESCE(ROUND(qty * price, 4), 0) AS net_amount
FROM tr_ar_inv_head head
INNER JOIN tr_ar_inv_det_direct det ON head.trans_no = det.trans_no
INNER JOIN ms_cust cust ON head.cust_code = cust.cust_code
WHERE (jenis_pabean <> '' OR cukai_no <> '') -- old (2012-2014) docs have cukai_no but no jenis_pabean
	AND approved_date > '2012-01-01'
	AND canceled_date < approved_date`

// ---- Query building ----

// rawCodes returns the det.item_code values matching a productCode filter. The report's item_code is
// '1' + det.item_code, filtered like the old "item_code = ? OR SUBSTR(item_code, 2) = ?" ("1" alone = empty code).
func rawCodes(code string) []string {
	if code == "1" {
		return []string{""}
	}
	codes := []string{code}
	if rest, ok := strings.CutPrefix(code, "1"); ok {
		codes = append(codes, rest)
	}
	return codes
}

// buildQuery appends the report filters to the source query's own WHERE.
func buildQuery(filter GetReportFilter) (string, []any) {
	where := []string{"cukai_date BETWEEN ? AND ?"}
	args := []any{filter.From.Format("2006-01-02"), filter.To.Format("2006-01-02")}

	if filter.PabeanType != "" {
		where = append(where, "jenis_pabean = ?")
		args = append(args, filter.PabeanType)
	}
	if filter.NoPabean != "" {
		where = append(where, "cukai_no = ?")
		args = append(args, filter.NoPabean)
	}
	if filter.ProductCode != "" {
		where = append(where, "det.item_code IN ?")
		args = append(args, rawCodes(filter.ProductCode))
	}
	if filter.ProductName != "" {
		where = append(where, "det.item_name LIKE ?")
		args = append(args, "%"+filter.ProductName+"%")
	}
	// Note: ProductGroup is accepted but not applied (no product_group column in the source)

	return fkiExpenditureSQL + "\n\tAND " + strings.Join(where, " AND "), args
}
func normalizeItemCode(itemCode string) string {
	if itemCode == "1" {
		return itemCode
	}

	return strings.TrimPrefix(itemCode, "1")
}

// GetReport retrieves expenditure products with filters and pagination (paged in SQL).
func (c *ExpenditureProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.ExpenditureProduct, int64, error) {
	base, args := buildQuery(filter)
	db := c.db.WithContext(ctx)

	dir := "DESC"
	if filter.IsExport {
		dir = "ASC"
	}
	// ponytail: the key isn't unique (same item/qty twice on one invoice can tie); add seq_num if paging ever shifts.
	query := base + "\nORDER BY tgl_pabean " + dir + ", no_pabean DESC, trans_no DESC, item_name ASC, dlv_qty ASC"

	// Apply pagination only if both page and limit are provided (> 0); only then is a separate COUNT needed.
	// Compared via division so a huge page/limit from the query string can't overflow.
	paged := filter.Limit > 0 && filter.Page > 0
	var totalCount int64
	if paged {
		if err := db.Raw("SELECT COUNT(*) FROM ("+base+"\n) x", args...).Scan(&totalCount).Error; err != nil {
			return nil, 0, err
		}
		if int64(filter.Page-1) > totalCount/int64(filter.Limit) {
			return []model.ExpenditureProduct{}, totalCount, nil
		}
		query += "\nLIMIT ? OFFSET ?"
		args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	}

	// Non-nil so an empty result serializes as [] (Raw+Scan leaves dest untouched on zero rows).
	results := []model.ExpenditureProduct{}
	if err := db.Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}
	if !paged {
		totalCount = int64(len(results))
	}
	for i := range results {
		results[i].ItemCode = normalizeItemCode(results[i].ItemCode)
	}
	return results, totalCount, nil
}
