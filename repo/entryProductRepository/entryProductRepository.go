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
WHERE cukai_no <> '' -- BC number required; old (2012-2014) docs have it without jenis_pabean
	AND approved_date <> '2012-01-01'
	AND canceled_date < approved_date`

// ponytail: non-empty bc_type/bc_no mirrors FKI's cukai_no filter; Itory has no approved/canceled filter yet.
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
WHERE bc_type <> '' AND bc_no <> ''`

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

// bcKey identifies one customs document: BC number + BC date + vendor.
func bcKey(noPabean string, tglPabean time.Time, vendorCode string) string {
	return strings.TrimSpace(noPabean) + "|" + tglPabean.Format("2006-01-02") + "|" + vendorCode
}

// itoryHead is an Itory report document that may be the same as an FKI one.
type itoryHead struct {
	RefKey, BcNo, VendorCode string
	BcDate                   time.Time
}

// keyList joins keys for string_to_array(?, chr(31)): one bound parameter, so no 65535-param cap.
func keyList(seen map[string]bool) string {
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "\x1f")
}

// itoryCopiesDropped removes FKI rows whose document is also in Itory, so Itory wins. A match is:
//   - same invoice (FKI invoice_no = Itory ref_no, normalized; empty never matches), or
//   - same BC no + BC date + vendor, when that Itory doc isn't already the invoice match of another
//     FKI doc (FKI often adds " (MB)", " (GIT)", "_MB" etc. or a typo to the invoice).
//
// Checked against the whole Itory source, not the report filter; an Itory copy without bc_type/bc_no
// is not in the report, so it doesn't count and the FKI row stays.
// The wrapping subquery is fine here: Postgres inlines it (unlike FKI's MySQL 5.5).
// ponytail: "already matched" only sees the fetched FKI rows; an item filter can hide the matching
// doc (rare GIT pairs under one BC) - query FKI heads for the BC keys if that ever matters.
func (c *EntryProductRepository) itoryCopiesDropped(ctx context.Context, fkiRows []entryRow) ([]entryRow, error) {
	invoices, bcNos := map[string]bool{}, map[string]bool{}
	for _, r := range fkiRows {
		if k := dlvKey(r.VendDlvNo); k != "" {
			invoices[k] = true
		}
		if no := strings.TrimSpace(r.NoPabean); no != "" {
			bcNos[no] = true
		}
	}
	if len(invoices) == 0 && len(bcNos) == 0 {
		return fkiRows, nil
	}

	var heads []itoryHead
	err := c.itoryDB.WithContext(ctx).Raw(
		"SELECT DISTINCT UPPER(TRIM(vend_dlv_no)) AS ref_key, TRIM(no_pabean) AS bc_no, tgl_pabean AS bc_date, vendor_code\n"+
			"FROM ("+itoryEntrySQL+"\n) src\n"+
			"WHERE UPPER(TRIM(vend_dlv_no)) = ANY(string_to_array(?, chr(31))) OR TRIM(no_pabean) = ANY(string_to_array(?, chr(31)))",
		keyList(invoices), keyList(bcNos)).Scan(&heads).Error
	if err != nil {
		return nil, err
	}

	itoryInvoices, freeBC := map[string]bool{}, map[string]bool{}
	for _, h := range heads {
		if h.RefKey != "" {
			itoryInvoices[h.RefKey] = true
		}
		if h.RefKey == "" || !invoices[h.RefKey] {
			freeBC[bcKey(h.BcNo, h.BcDate, h.VendorCode)] = true
		}
	}

	kept := fkiRows[:0]
	for _, r := range fkiRows {
		if !itoryInvoices[dlvKey(r.VendDlvNo)] && !freeBC[bcKey(r.NoPabean, r.TglPabean, r.VendorCode)] {
			kept = append(kept, r)
		}
	}
	return kept, nil
}

// GetReport retrieves entry products from FKI + Itory with filters and pagination.
// A document present in both sources (see itoryCopiesDropped) is taken from Itory only.
// ponytail: both sources are fully loaded for the date range then paged in memory;
// if ranges grow too large, fetch only offset+limit rows per source and merge.
func (c *EntryProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.EntryProduct, int64, error) {
	fkiRows, err := fetch(ctx, c.fkiDB, fkiSource, filter)
	if err != nil {
		return nil, 0, err
	}
	fkiRows, err = c.itoryCopiesDropped(ctx, fkiRows)
	if err != nil {
		return nil, 0, err
	}

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
