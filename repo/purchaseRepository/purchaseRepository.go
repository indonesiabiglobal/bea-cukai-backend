package purchaseRepository

import (
	"Dashboard-TRDP/model"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ====== DateRange helper ======
type DateRange struct {
	From time.Time
	To   time.Time // inclusive (tanggal “akhir” yang dipilih)
}

// ====== DTO untuk dashboard ======
type PurchaseKPISummary struct {
	TotalQty      decimal.Decimal `json:"total_qty"`
	TotalSubtotal decimal.Decimal `json:"total_subtotal"`
	UniqueVendors int64           `json:"unique_vendors"`
	UniqueItems   int64           `json:"unique_items"`
	TxCount       int64           `json:"tx_count"`
}

type PurchaseTrendPoint struct {
	Date     time.Time       `json:"date" gorm:"column:date"`
	Qty      decimal.Decimal `json:"qty"`
	Subtotal decimal.Decimal `json:"subtotal"`
}

type TopVendor struct {
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Qty          decimal.Decimal `json:"qty"`
	AveragePrice decimal.Decimal `json:"average_price"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

type TopProduct struct {
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	CategoryCode string          `json:"category_code"`
	CategoryName string          `json:"category_name"`
	Qty          decimal.Decimal `json:"qty"`
	AveragePrice decimal.Decimal `json:"average_price"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

type ByCategory struct {
	CategoryCode string          `json:"category_code"`
	CategoryName string          `json:"category_name"`
	Qty          decimal.Decimal `json:"qty"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

type Vendor struct {
	VendorCode string `json:"vendor_code"`
	VendorName string `json:"vendor_name"`
}
// ====== Repository ======
type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// KPI summary
func (r *PurchaseRepository) GetKPISummary(ctx context.Context, dto model.PurchaseRequestParam) (PurchaseKPISummary, error) {
	var result PurchaseKPISummary

	query := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Joins(`LEFT JOIN master_barang mb ON mb.kode_barang = pembelian.kode_barang`).
		Select(`
        COALESCE(SUM(pembelian.qty), 0)       AS total_qty,
        COALESCE(SUM(pembelian.subtotal), 0)  AS total_subtotal,
        COUNT(DISTINCT pembelian.kode_vendor) AS unique_vendors,
        COUNT(DISTINCT pembelian.kode_barang) AS unique_items,
        COUNT(*)                              AS tx_count
    `).
		Where("pembelian.tanggal_faktur >= ? AND pembelian.tanggal_faktur < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	if vendor := strings.TrimSpace(dto.Vendor); vendor != "" {
		query = query.Where("pembelian.kode_vendor = ?", vendor)
	}

	err := query.Scan(&result).Error

	if err != nil {
		return result, err
	}

	return result, nil
}

// Trend harian
func (r *PurchaseRepository) GetTrend(ctx context.Context, dto model.PurchaseRequestParam) ([]PurchaseTrendPoint, error) {
	var result []PurchaseTrendPoint

	query := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Joins(`LEFT JOIN master_barang mb ON mb.kode_barang = pembelian.kode_barang`).
		Select(`
			DATE_TRUNC('day', tanggal_faktur) AS date,
			COALESCE(SUM(qty), 0)             AS qty,
			COALESCE(SUM(subtotal), 0)        AS subtotal
		`).
		Where("tanggal_faktur >= ? AND tanggal_faktur < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	if vendor := strings.TrimSpace(dto.Vendor); vendor != "" {
		query = query.Where("pembelian.kode_vendor = ?", vendor)
	}

	err := query.
		Group("DATE_TRUNC('day', tanggal_faktur)").
		Order("date ASC").
		Scan(&result).Error

	return result, err
}

// Top vendors (by subtotal DESC)
func (r *PurchaseRepository) GetTopVendors(ctx context.Context, dto model.PurchaseRequestParam) ([]TopVendor, error) {
	var result []TopVendor

	// Jika nama_vendor kosong/null → pakai kode_vendor
	nameExpr := "COALESCE(NULLIF(BTRIM(nama_vendor), ''), kode_vendor)"

	query := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Joins(`LEFT JOIN master_barang mb ON mb.kode_barang = pembelian.kode_barang`).
		Select(`
			kode_vendor AS code,
			`+nameExpr+` AS name,
			COALESCE(SUM(qty), 0)      AS qty,
			COALESCE(AVG(subtotal), 0) / COALESCE(AVG(qty), 1) AS average_price,
			COALESCE(SUM(subtotal), 0) AS subtotal
		`).
		Where("tanggal_faktur >= ? AND tanggal_faktur < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	if vendor := strings.TrimSpace(dto.Vendor); vendor != "" {
		query = query.Where("pembelian.kode_vendor = ?", vendor)
	}

	err := query.
		Group("kode_vendor, " + nameExpr).
		Order("subtotal DESC").
		Scan(&result).Error

	return result, err
}

func (r *PurchaseRepository) GetTopProducts(ctx context.Context, dto model.PurchaseRequestParam) ([]TopProduct, error) {
	var result []TopProduct

	mt := (model.MasterProduct{}).TableName()

	nameExpr := "COALESCE(NULLIF(BTRIM(mp.nama_barang), ''), pembelian.kode_barang)"

	query := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Joins("LEFT JOIN "+mt+" mp ON mp.kode_barang = pembelian.kode_barang").
		Select(fmt.Sprintf(`
            pembelian.kode_barang                   AS code,
            %s                              AS name,
            COALESCE(mp.kode_kategori_barang, '') AS category_code,
            COALESCE(mp.nama_kategori_barang, '') AS category_name,
            COALESCE(SUM(pembelian.qty), 0)         AS qty,
			CASE WHEN COALESCE(SUM(pembelian.qty), 0) = 0 THEN 0
				ELSE (COALESCE(SUM(pembelian.subtotal), 0) / COALESCE(SUM(pembelian.qty), 0))
				END AS average_price,
            COALESCE(SUM(pembelian.subtotal), 0)    AS subtotal
        `, nameExpr)).
		Where("pembelian.tanggal_faktur >= ? AND pembelian.tanggal_faktur < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mp.kode_kategori_barang = ?", category)
	}

	if vendor := strings.TrimSpace(dto.Vendor); vendor != "" {
		query = query.Where("pembelian.kode_vendor = ?", vendor)
	}

	err := query.Group("pembelian.kode_barang, " + nameExpr + ", mp.kode_kategori_barang, mp.nama_kategori_barang").
		Order("subtotal DESC").
		Scan(&result).Error

	return result, err
}

func (r *PurchaseRepository) GetByCategory(ctx context.Context, dto model.PurchaseRequestParam, limit int) ([]ByCategory, error) {
	var result []ByCategory

	mt := (model.MasterProduct{}).TableName()

	query := r.db.WithContext(ctx).
		Model(&model.Purchase{}).
		Joins("LEFT JOIN "+mt+" mp ON mp.kode_barang = pembelian.kode_barang").
		Select(`
            COALESCE(mp.kode_kategori_barang, '') AS category_code,
            COALESCE(mp.nama_kategori_barang, '') AS category_name,
            COALESCE(SUM(pembelian.qty), 0)               AS qty,
            COALESCE(SUM(pembelian.subtotal), 0)          AS subtotal
        `).
		Where("pembelian.tanggal_faktur >= ? AND pembelian.tanggal_faktur < ?", dto.From, dto.To)
		

	if vendor := strings.TrimSpace(dto.Vendor); vendor != "" {
		query = query.Where("pembelian.kode_vendor = ?", vendor)
	}

	err := query.
		Group("mp.kode_kategori_barang, mp.nama_kategori_barang").
		Order("subtotal DESC").
		Limit(limit).
		Scan(&result).Error

	return result, err
}


func (r *PurchaseRepository) GetVendors(ctx context.Context, search string, limit, offset int) ([]Vendor, int64, error) {
	// Normalisasi pencarian
	q := strings.TrimSpace(search)

	// Base query (alias m)
	base := r.db.WithContext(ctx).Table((model.Purchase{}).TableName() + " m")

	// Filter pencarian (opsional)
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		base = base.Where(`
			LOWER(COALESCE(m.kode_vendor, '')) LIKE ? OR
			LOWER(COALESCE(m.nama_vendor, '')) LIKE ?
		`, like, like)
	}

	// Count distinct vendor (butuh subquery)
	var total int64
	countQ := base.
		Select("DISTINCT COALESCE(m.kode_vendor, '') AS vendor_code, COALESCE(m.nama_vendor, '') AS vendor_name")
	err := r.db.Table("(?) AS t", countQ).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Data dengan
	var rows []Vendor
	err = base.
		Select(`
			COALESCE(m.kode_vendor, '')  AS vendor_code,
			COALESCE(m.nama_vendor, '')  AS vendor_name
		`).
		Group("COALESCE(m.kode_vendor,''), COALESCE(m.nama_vendor,'')").
		Order("vendor_name ASC, vendor_code ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	return rows, total, err
}