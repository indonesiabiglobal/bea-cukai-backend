package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type FinishedProductReport struct {
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
	Selisih      decimal.Decimal `json:"selisih"`
	Akhr         decimal.Decimal `json:"akhr"`
	Msk          decimal.Decimal `json:"msk"`
}

type FinishedProductReportRequest struct {
	TglAwal  time.Time `json:"tgl_awal" validate:"required"`
	TglAkhir time.Time `json:"tgl_akhir" validate:"required"`
	ItemCode string    `json:"item_code"`
	ItemName string    `json:"item_name"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// Response type alias for consistency
type FinishedProductReportResponse = FinishedProductReport
