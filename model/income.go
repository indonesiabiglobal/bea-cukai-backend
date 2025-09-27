package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Income struct {
	NoTransaksi          string          `json:"no_transaksi" gorm:"size:64;not null;index:idx_no_transaksi,unique"`
	PatID                string          `json:"patid" gorm:"size:64;index:idx_patid"`
	Episode              string          `json:"episode" gorm:"size:64;index:idx_episode"`
	TglTransaksi         time.Time       `json:"tgl_transaksi" gorm:"type:date;index:idx_tgl_transaksi"`
	IPOP                 string          `json:"ipop" gorm:"size:8;index:idx_ipop;column:ipop"`
	KodeUnit             string          `json:"kode_unit" gorm:"size:64;index:idx_unit"`
	NamaUnit             string          `json:"nama_unit" gorm:"size:128"`
	ProvID               string          `json:"provid" gorm:"column:provid;size:64;index:idx_provid"`
	ProvName             string          `json:"provname" gorm:"column:provname;size:128"`
	Debit                decimal.Decimal `json:"debit" gorm:"type:decimal(20,2);not null;default:0"`
	Credit               decimal.Decimal `json:"credit" gorm:"type:decimal(20,2);not null;default:0"`
	KodeLayanan          string          `json:"kode_layanan" gorm:"size:64;index:idx_kode_layanan"`
	KodePenjamin         string          `json:"kode_penjamin" gorm:"size:64;index:idx_kode_penjamin"`
	NamaPenjamin         string          `json:"nama_penjamin" gorm:"size:128"`
	KodeKelompokPenjamin string          `json:"kode_kelompok_penjamin" gorm:"size:64;index:idx_kelompok_penjamin"`
	NamaKelompokPenjamin string          `json:"nama_kelompok_penjamin" gorm:"size:128"`
	DateIdx              int             `json:"dateidx" gorm:"column:dateidx;index:idx_dateidx;not null"`
}

// TableName overrides the default table name.
func (Income) TableName() string { return "pendapatan" }

type IncomeRequest struct {
	NoTransaksi          string          `json:"no_transaksi" validate:"required"`
	PatID                string          `json:"patid"`
	Episode              string          `json:"episode"`
	TglTransaksi         time.Time       `json:"tgl_transaksi" validate:"required"`
	IPOP                 string          `json:"ipop"`
	KodeUnit             string          `json:"kode_unit"`
	NamaUnit             string          `json:"nama_unit"`
	ProvID               string          `json:"provid"`
	ProvName             string          `json:"provname"`
	Debit                decimal.Decimal `json:"debit"`
	Credit               decimal.Decimal `json:"credit"`
	KodeLayanan          string          `json:"kode_layanan"`
	KodePenjamin         string          `json:"kode_penjamin"`
	NamaPenjamin         string          `json:"nama_penjamin"`
	KodeKelompokPenjamin string          `json:"kode_kelompok_penjamin"`
	NamaKelompokPenjamin string          `json:"nama_kelompok_penjamin"`
}

type IncomeResponse struct {
	ID                   uint            `json:"id"`
	NoTransaksi          string          `json:"no_transaksi"`
	PatID                string          `json:"patid"`
	Episode              string          `json:"episode"`
	TglTransaksi         time.Time       `json:"tgl_transaksi"`
	IPOP                 string          `json:"ipop"`
	KodeUnit             string          `json:"kode_unit"`
	NamaUnit             string          `json:"nama_unit"`
	ProvID               string          `json:"provid"`
	ProvName             string          `json:"provname"`
	Debit                decimal.Decimal `json:"debit"`
	Credit               decimal.Decimal `json:"credit"`
	KodeLayanan          string          `json:"kode_layanan"`
	KodePenjamin         string          `json:"kode_penjamin"`
	NamaPenjamin         string          `json:"nama_penjamin"`
	KodeKelompokPenjamin string          `json:"kode_kelompok_penjamin"`
	NamaKelompokPenjamin string          `json:"nama_kelompok_penjamin"`
	DateIdx              int             `json:"dateidx"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
