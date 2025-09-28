package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type WipPositionReport struct {
	ItemCode     string          `json:"item_code"`
	ItemName     string          `json:"item_name"`
	UnitCode     string          `json:"unit_code"`
	ItemTypeCode string          `json:"item_type_code"`
	ItemGroup    string          `json:"item_group"`
	LocationCode string          `json:"location_code"`
	Awal         decimal.Decimal `json:"awal"`
	Masuk        decimal.Decimal `json:"masuk"`
	Keluar       decimal.Decimal `json:"keluar"`
	Peny         decimal.Decimal `json:"peny"`
	Akhir        decimal.Decimal `json:"akhir"`
	Opname       decimal.Decimal `json:"opname"`
}

type WipPositionReportRequest struct {
	TglAwal  time.Time `json:"tgl_awal" validate:"required"`
	TglAkhir time.Time `json:"tgl_akhir" validate:"required"`
	ItemCode string    `json:"item_code"`
	ItemName string    `json:"item_name"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type WipPositionReportResponse struct {
	ItemCode string `json:"kode_barang"`
	ItemName string `json:"nama_barang"`
	UnitCode string `json:"sat"`
	Jumlah   string `json:"jumlah"`
}
