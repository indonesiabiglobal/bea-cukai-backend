package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Sale maps to the physical table "sale".
// We don't embed gorm.Model because the schema provides its own primary key (uid)
// and doesn't include created_at/updated_at.

// Notes:
// - Use decimal for qty/subtotal to avoid floating-point rounding errors.
// - Indexes on tanggal_penjualan and kode_barang for common queries.

type Sale struct {
	UID              int32           `json:"uid" gorm:"primaryKey;column:uid"`
	TanggalPenjualan time.Time       `json:"tanggal_penjualan" gorm:"column:tanggal_penjualan;type:timestamp;index:idx_sale_tanggal"`
	KodeBarang       string          `json:"kode_barang" gorm:"column:kode_barang;size:64;index:idx_sale_barang"`
	Qty              decimal.Decimal `json:"qty" gorm:"column:qty;type:decimal(20,3);not null;default:0"`
	Subtotal         decimal.Decimal `json:"subtotal" gorm:"column:subtotal;type:decimal(20,2);not null;default:0"`
}

// TableName enforces the DB table name.
func (Sale) TableName() string { return "penjualan" }

// ==========================
// DTOs
// ==========================
type SaleRequestParam struct {
	From     time.Time `json:"from" form:"from" query:"from"`
	To       time.Time `json:"to" form:"to" query:"to"`
	Category string    `json:"category" form:"category" query:"category"`
}

// SaleRequest defines payload for create/update operations.

type SaleRequest struct {
	UID              int32           `json:"uid" validate:"required"`
	TanggalPenjualan time.Time       `json:"tanggal_penjualan" validate:"required"`
	KodeBarang       string          `json:"kode_barang"`
	Qty              decimal.Decimal `json:"qty"`
	Subtotal         decimal.Decimal `json:"subtotal"`
}

// SaleResponse is returned to clients.

type SaleResponse struct {
	UID              int32           `json:"uid"`
	TanggalPenjualan time.Time       `json:"tanggal_penjualan"`
	KodeBarang       string          `json:"kode_barang"`
	Qty              decimal.Decimal `json:"qty"`
	Subtotal         decimal.Decimal `json:"subtotal"`
}

// If you plan to bulk-import, ensure UID is unique in the source. If not,
// consider using a surrogate key or composite unique (kode_barang, tanggal_penjualan, uid)
// depending on your data guarantees.
