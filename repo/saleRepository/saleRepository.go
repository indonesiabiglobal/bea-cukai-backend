// repo/saleRepository/sale_repository.go
package saleRepository

import (
	"Dashboard-TRDP/model"
	"context"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ====== DTO untuk dashboard sales ======
type SalesKPISummary struct {
	TotalQty      decimal.Decimal `json:"total_qty"`
	TotalSubtotal decimal.Decimal `json:"total_subtotal"`
	UniqueItems   int64           `json:"unique_items"`
	TxCount       int64           `json:"tx_count"`
}

type SalesTrendPoint struct {
	Date     time.Time       `json:"date" gorm:"column:date"`
	Qty      decimal.Decimal `json:"qty"`
	Subtotal decimal.Decimal `json:"subtotal"`
}

type SalesTopProduct struct {
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	CategoryCode string          `json:"category_code"`
	CategoryName string          `json:"category_name"`
	Qty          decimal.Decimal `json:"qty"`
	AveragePrice decimal.Decimal `json:"average_price"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

type SalesByCategory struct {
	CategoryCode string          `json:"category_code"`
	CategoryName string          `json:"category_name"`
	Qty          decimal.Decimal `json:"qty"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

// ====== Repository ======
type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// GetKPISummary: total qty, total subtotal, unique items, transaksi
func (r *SaleRepository) GetKPISummary(ctx context.Context, dto model.SaleRequestParam) (SalesKPISummary, error) {
	var result SalesKPISummary

	query := r.db.WithContext(ctx).
		Table("penjualan s").
		Joins(`LEFT JOIN master_barang mb ON mb.kode_barang = s.kode_barang`).
		Select(`
			COALESCE(SUM(s.qty), 0)       AS total_qty,
			COALESCE(SUM(s.subtotal), 0)  AS total_subtotal,
			COUNT(DISTINCT s.kode_barang) AS unique_items,
			COUNT(*)                       AS tx_count
		`).
		Where("s.tanggal_penjualan >= ? AND s.tanggal_penjualan < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	err := query.
		Scan(&result).Error

	if err != nil {
		return result, err
	}
	return result, nil
}

// GetTrend: agregasi harian qty & subtotal (DATE_TRUNC)
func (r *SaleRepository) GetTrend(ctx context.Context, dto model.SaleRequestParam) ([]SalesTrendPoint, error) {
	var result []SalesTrendPoint

	query := r.db.WithContext(ctx).
		Table("penjualan s").
		Joins(`LEFT JOIN master_barang mb ON mb.kode_barang = s.kode_barang`).
		Select(`
			DATE_TRUNC('day', s.tanggal_penjualan) AS date,
			COALESCE(SUM(s.qty), 0)                AS qty,
			COALESCE(SUM(s.subtotal), 0)           AS subtotal
		`).
		Where("s.tanggal_penjualan >= ? AND s.tanggal_penjualan < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	err := query.
		Group("DATE_TRUNC('day', s.tanggal_penjualan)").
		Order("date ASC").
		Scan(&result).Error

	return result, err
}

// GetTopProducts: ranking produk (subtotal desc) + join master product utk nama/kategori
func (r *SaleRepository) GetTopProducts(ctx context.Context, dto model.SaleRequestParam) ([]SalesTopProduct, error) {
	var result []SalesTopProduct

	saleTable := (model.Sale{}).TableName()
	masterProductTable := (model.MasterProduct{}).TableName()

	// Nama aman: pakai nama_barang kalau ada/trim != '', selain itu pakai kode_barang
	nameExpr := "COALESCE(NULLIF(BTRIM(mp.nama_barang), ''), s.kode_barang)"

	query := r.db.WithContext(ctx).
		Table(saleTable+" s").
		Joins("LEFT JOIN "+masterProductTable+" mp ON mp.kode_barang = s.kode_barang").
		Select(`
			s.kode_barang                       AS code,
			`+nameExpr+`                        AS name,
			COALESCE(mp.kode_kategori_barang,'') AS category_code,
			COALESCE(mp.nama_kategori_barang,'') AS category_name,
			COALESCE(SUM(s.qty), 0)            AS qty,
			CASE WHEN COALESCE(SUM(s.qty), 0) = 0 THEN 0
				ELSE (COALESCE(SUM(s.subtotal), 0) / COALESCE(SUM(s.qty), 0))
				END AS average_price,
			COALESCE(SUM(s.subtotal), 0)       AS subtotal
		`).
		Where("s.tanggal_penjualan >= ? AND s.tanggal_penjualan < ?", dto.From, dto.To)

	if category := strings.TrimSpace(dto.Category); category != "" {
		query = query.Where("mb.kode_kategori_barang = ?", category)
	}

	err := query.
		Group("s.kode_barang, " + nameExpr + ", mp.kode_kategori_barang, mp.nama_kategori_barang").
		Order("subtotal DESC").
		Scan(&result).Error

	return result, err
}

// GetByCategory: agregasi subtotal & qty per kategori (dari master product)
func (r *SaleRepository) GetByCategory(ctx context.Context, dto model.SaleRequestParam, limit int) ([]SalesByCategory, error) {
	var result []SalesByCategory

	saleTable := (model.Sale{}).TableName()
	masterProductTable := (model.MasterProduct{}).TableName()

	err := r.db.WithContext(ctx).
		Table(saleTable+" s").
		Joins("LEFT JOIN "+masterProductTable+" mp ON mp.kode_barang = s.kode_barang").
		Select(`
			COALESCE(mp.kode_kategori_barang,'') AS category_code,
			COALESCE(mp.nama_kategori_barang,'') AS category_name,
			COALESCE(SUM(s.qty), 0)              AS qty,
			COALESCE(SUM(s.subtotal), 0)         AS subtotal
		`).
		Where("s.tanggal_penjualan >= ? AND s.tanggal_penjualan < ?", dto.From, dto.To).
		Group("mp.kode_kategori_barang, mp.nama_kategori_barang").
		Order("subtotal DESC").
		Limit(limit).
		Scan(&result).Error

	return result, err
}
