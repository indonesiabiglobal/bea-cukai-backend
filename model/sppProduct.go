package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// SPPProduct maps to the physical table "spp_product".
// We don't embed gorm.Model because the schema defines its own primary key (id)
// and doesn't include created_at/updated_at.

type SPPProduct struct {
	ID           int32           `json:"id" gorm:"primaryKey;column:id"`
	Tgl          time.Time       `json:"tgl" gorm:"column:tgl;type:timestamp;index:idx_spp_tgl"`
	KodeBarang   string          `json:"kode_barang" gorm:"column:kode_barang;size:64;index:idx_spp_barang"`
	Qty          decimal.Decimal `json:"qty" gorm:"column:qty;type:decimal(20,3);not null;default:0"`
	Total        decimal.Decimal `json:"total" gorm:"column:total;type:decimal(20,2);not null;default:0"`
	LokasiAsal   string          `json:"lokasi_asal" gorm:"column:lokasi_asal;size:128;index:idx_spp_asal"`
	LokasiTujuan string          `json:"lokasi_tujuan" gorm:"column:lokasi_tujuan;size:128;index:idx_spp_tujuan"`
}

// TableName enforces the DB table name.
func (SPPProduct) TableName() string { return "spp_barang" }

// ==========================
// DTOs
// ==========================

// SPPProductRequest defines payload for create/update operations.

type SPPProductRequest struct {
	ID           int32           `json:"id" validate:"required"`
	Tgl          time.Time       `json:"tgl" validate:"required"`
	KodeBarang   string          `json:"kode_barang"`
	Qty          decimal.Decimal `json:"qty"`
	Total        decimal.Decimal `json:"total"`
	LokasiAsal   string          `json:"lokasi_asal"`
	LokasiTujuan string          `json:"lokasi_tujuan"`
}

// SPPProductResponse is returned to clients.

type SPPProductResponse struct {
	ID           int32           `json:"id"`
	Tgl          time.Time       `json:"tgl"`
	KodeBarang   string          `json:"kode_barang"`
	Qty          decimal.Decimal `json:"qty"`
	Total        decimal.Decimal `json:"total"`
	LokasiAsal   string          `json:"lokasi_asal"`
	LokasiTujuan string          `json:"lokasi_tujuan"`
}

// Notes:
// - Consider a composite unique index on (id) if it's guaranteed unique from source.
// - If transfers can have multiple detail lines per id, change primary key to a surrogate
//   or make (id, kode_barang) unique.
