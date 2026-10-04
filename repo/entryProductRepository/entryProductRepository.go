package entryProductRepository

import (
	"Bea-Cukai/model"
	"context"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

// EntryProductRepository reads pemasukan barang live from two sources:
// FKI (MySQL) and Itory (PostgreSQL), merged in Go.
type EntryProductRepository struct {
	fkiDB   *gorm.DB
	itoryDB *gorm.DB
}

func NewEntryProductRepository(fkiDB, itoryDB *gorm.DB) *EntryProductRepository {
	return &EntryProductRepository{fkiDB: fkiDB, itoryDB: itoryDB}
}

// ---- Source queries ----
// Both produce identical column aliases (= model.EntryProduct fields) plus seq_num for ordering.
// Filters are appended to each query's own WHERE using the source's real column names:
// FKI runs MySQL 5.5, which materializes a wrapping derived table (full history scan, ~17s).

type entrySource struct {
	sql                                      string
	tgl, jenis, noPabean, itemCode, itemName string
}

var fkiSource = entrySource{
	sql: fkiEntrySQL, tgl: "cukai_date", jenis: "jenis_pabean", noPabean: "cukai_no",
	itemCode: "det.item_code", itemName: "det.item_name",
}

var itorySource = entrySource{
	sql: itoryEntrySQL, tgl: "bc_date", jenis: "bc_type", noPabean: "bc_no",
	itemCode: "det.item_code", itemName: "item.item_name",
}

const fkiEntrySQL = `
SELECT
	jenis_pabean,
	cukai_no AS no_pabean,
	cukai_date AS tgl_pabean,
	head.trans_no,
	invoice_no AS vend_dlv_no,
	trans_date,
	vend.vendor_code,
	vend.vendor_name,
	det.item_code,
	det.item_name,
	COALESCE(qty, 0) AS rcv_qty,
	det.unit AS pch_unit,
	curr_code,
	COALESCE(price, 0) AS net_price,
	-- gross value like the old sync table (amount here is net of disc_value)
	COALESCE(ROUND(qty * price, 4), 0) AS net_amount,
	det.seq_num
FROM tr_ap_inv_det_direct det
INNER JOIN tr_ap_inv_head head ON head.trans_no = det.trans_no
INNER JOIN ms_vendor vend ON head.vendor_code = vend.vendor_code
WHERE (jenis_pabean <> '' OR cukai_no <> '') -- old (2012-2014) docs have cukai_no but no jenis_pabean
	AND approved_date <> '2012-01-01'
	AND canceled_date < approved_date`

// ponytail: the non-empty bc_type filter mirrors FKI's jenis_pabean filter; Itory has no approved/canceled filter yet.
const itoryEntrySQL = `
SELECT
	bc_type AS jenis_pabean,
	bc_no AS no_pabean,
	bc_date AS tgl_pabean,
	head.trans_no,
	ref_no AS vend_dlv_no,
	in_date AS trans_date,
	vend.vendor_code,
	vend.vendor_name,
	det.item_code,
	item.item_name,
	COALESCE(qty, 0) AS rcv_qty,
	item.unit_code AS pch_unit,
	curr_code,
	COALESCE(price, 0) AS net_price,
	-- gross value like the old sync table (net_amount here is rounded to 2 dp and net of disc_value)
	COALESCE(ROUND(qty * price, 4), 0) AS net_amount,
	det.seq_num
FROM tr_ap_inv_det det
INNER JOIN tr_ap_inv_head head ON head.trans_no = det.trans_no
INNER JOIN ms_vendor vend ON head.vendor_code = vend.vendor_code
INNER JOIN ms_item item ON det.item_code = item.item_code
WHERE bc_type <> ''`

// ---- Core aggregations ----

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

// buildQuery appends the report filters and ordering to a source query.
// GORM rewrites "?" into the dialect's placeholder ($n for Postgres).
func buildQuery(src entrySource, filter GetReportFilter) (string, []any) {
	where := []string{src.tgl + " BETWEEN ? AND ?"}
	args := []any{filter.From.Format("2006-01-02"), filter.To.Format("2006-01-02")}

	if filter.PabeanType != "" {
		where = append(where, src.jenis+" = ?")
		args = append(args, filter.PabeanType)
	}
	if filter.NoPabean != "" {
		where = append(where, src.noPabean+" = ?")
		args = append(args, filter.NoPabean)
	}
	if filter.ProductCode != "" {
		where = append(where, src.itemCode+" = ?")
		args = append(args, filter.ProductCode)
	}
	if filter.ProductName != "" {
		// LOWER on both sides: MySQL LIKE is case-insensitive, Postgres LIKE is not.
		where = append(where, "LOWER("+src.itemName+") LIKE LOWER(?)")
		args = append(args, "%"+filter.ProductName+"%")
	}
	// Note: ProductGroup is accepted but not applied (no product_group column in either source)

	query := src.sql + "\n\tAND " + strings.Join(where, " AND ") +
		"\nORDER BY " + src.tgl + ", head.trans_no DESC, det.seq_num ASC"
	return query, args
}

// entryRow carries seq_num only for SQL ordering; the API returns model.EntryProduct.
type entryRow struct {
	model.EntryProduct
	SeqNum int
}

func fetch(ctx context.Context, db *gorm.DB, src entrySource, filter GetReportFilter) ([]entryRow, error) {
	query, args := buildQuery(src, filter)
	var rows []entryRow
	err := db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	return rows, err
}

// mergeAndPage sorts the combined rows by tgl_pabean (ASC for export, DESC otherwise),
// then trans_no DESC; the stable sort keeps each source's seq_num order within a transaction.
func mergeAndPage(rows []entryRow, filter GetReportFilter) ([]model.EntryProduct, int64) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if !a.TglPabean.Equal(b.TglPabean) {
			if filter.IsExport {
				return a.TglPabean.Before(b.TglPabean)
			}
			return a.TglPabean.After(b.TglPabean)
		}
		return a.TransNo > b.TransNo
	})

	total := int64(len(rows))
	start, end := 0, len(rows)
	// Apply pagination only if both page and limit are provided (> 0)
	// Compared via division so a huge page/limit from the query string can't overflow.
	if filter.Limit > 0 && filter.Page > 0 {
		start = len(rows)
		if filter.Page-1 <= len(rows)/filter.Limit {
			start = min((filter.Page-1)*filter.Limit, len(rows))
		}
		end = start + min(filter.Limit, len(rows)-start)
	}

	results := make([]model.EntryProduct, 0, end-start)
	for _, r := range rows[start:end] {
		results = append(results, r.EntryProduct)
	}
	return results, total
}

// localDate re-reads a DATE column's calendar day in time.Local. pgx returns DATE as UTC midnight,
// while FKI (MySQL, loc=Local) returns local midnight; without this the same day sorts as two instants.
func localDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// dlvKey normalizes vend_dlv_no for FKI↔Itory matching (FKI invoice_no = Itory ref_no).
func dlvKey(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

// dropDuplicates removes rows whose vend_dlv_no key is in dup.
func dropDuplicates(rows []entryRow, dup map[string]bool) []entryRow {
	kept := rows[:0]
	for _, r := range rows {
		if !dup[dlvKey(r.VendDlvNo)] {
			kept = append(kept, r)
		}
	}
	return kept
}

// itoryDuplicates returns which FKI vend_dlv_no keys also exist in Itory, so Itory wins.
// Checked against the whole Itory source (not the report filter): an Itory copy with a
// different bc_date/item still counts as a duplicate. Empty vend_dlv_no never matches.
// The wrapping subquery is fine here: Postgres inlines it (unlike FKI's MySQL 5.5).
// ponytail: one IN list; Postgres caps at 65535 params, chunk if a range ever exceeds that.
func (c *EntryProductRepository) itoryDuplicates(ctx context.Context, fkiRows []entryRow) (map[string]bool, error) {
	seen := map[string]bool{}
	keys := []string{}
	for _, r := range fkiRows {
		if k := dlvKey(r.VendDlvNo); k != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	dup := map[string]bool{}
	if len(keys) == 0 {
		return dup, nil
	}

	var found []string
	err := c.itoryDB.WithContext(ctx).Raw(
		"SELECT DISTINCT UPPER(TRIM(vend_dlv_no)) FROM ("+itoryEntrySQL+"\n) src WHERE UPPER(TRIM(vend_dlv_no)) IN ?",
		keys).Scan(&found).Error
	for _, k := range found {
		dup[k] = true
	}
	return dup, err
}

// GetReport retrieves entry products from FKI + Itory with filters and pagination.
// A document present in both sources (same vend_dlv_no) is taken from Itory only.
// ponytail: both sources are fully loaded for the date range then paged in memory;
// if ranges grow too large, fetch only offset+limit rows per source and merge.
func (c *EntryProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.EntryProduct, int64, error) {
	fkiRows, err := fetch(ctx, c.fkiDB, fkiSource, filter)
	if err != nil {
		return nil, 0, err
	}
	dup, err := c.itoryDuplicates(ctx, fkiRows)
	if err != nil {
		return nil, 0, err
	}
	fkiRows = dropDuplicates(fkiRows, dup)

	itoryRows, err := fetch(ctx, c.itoryDB, itorySource, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range itoryRows {
		itoryRows[i].TglPabean = localDate(itoryRows[i].TglPabean)
		itoryRows[i].TransDate = localDate(itoryRows[i].TransDate)
	}

	results, total := mergeAndPage(append(fkiRows, itoryRows...), filter)
	return results, total, nil
}
