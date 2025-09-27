package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Purchase struct {
	UID           int32           `json:"uid" gorm:"primaryKey;column:uid"`
	TanggalFaktur time.Time       `json:"tanggal_faktur" gorm:"column:tanggal_faktur;type:timestamp;index:idx_purchase_tanggal"`
	KodeVendor    string          `json:"kode_vendor" gorm:"column:kode_vendor;size:64;index:idx_purchase_vendor"`
	NamaVendor    string          `json:"nama_vendor" gorm:"column:nama_vendor;size:255"`
	KodeBarang    string          `json:"kode_barang" gorm:"column:kode_barang;size:64;index:idx_purchase_barang"`
	Qty           decimal.Decimal `json:"qty" gorm:"column:qty;type:decimal(20,3);not null;default:0"`
	Subtotal      decimal.Decimal `json:"subtotal" gorm:"column:subtotal;type:decimal(20,2);not null;default:0"`
	Product       MasterProduct   `json:"product,omitempty" gorm:"foreignKey:KodeBarang;references:KodeBarang"`
}

// TableName enforces the DB table name.
func (Purchase) TableName() string { return "pembelian" }

// ==========================
// DTOs
// ==========================

type PurchaseRequestParam struct {
	From     time.Time `json:"from" form:"from" query:"from"`
	To       time.Time `json:"to" form:"to" query:"to"`
	Category string    `json:"category" form:"category" query:"category"`
	Vendor   string    `json:"vendor" form:"vendor" query:"vendor"`
}

// PurchaseRequest defines payload for create/update operations.
type PurchaseRequest struct {
	UID           int32           `json:"uid" validate:"required"`
	TanggalFaktur time.Time       `json:"tanggal_faktur" validate:"required"`
	KodeVendor    string          `json:"kode_vendor"`
	NamaVendor    string          `json:"nama_vendor"`
	KodeBarang    string          `json:"kode_barang"`
	Qty           decimal.Decimal `json:"qty"`
	Subtotal      decimal.Decimal `json:"subtotal"`
}

// PurchaseResponse is returned to clients.

type PurchaseResponse struct {
	UID           int32           `json:"uid"`
	TanggalFaktur time.Time       `json:"tanggal_faktur"`
	KodeVendor    string          `json:"kode_vendor"`
	NamaVendor    string          `json:"nama_vendor"`
	KodeBarang    string          `json:"kode_barang"`
	Qty           decimal.Decimal `json:"qty"`
	Subtotal      decimal.Decimal `json:"subtotal"`
}

// If you plan to import bulk data, consider creating a composite unique index
// (uid) or (kode_vendor, kode_barang, tanggal_faktur) depending on your source
// guarantees. Adjust tags accordingly.
