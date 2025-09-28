package model

import (
	"time"
)

type RawMaterialReport struct {
	ItemCode     string `json:"item_code"`
	ItemName     string `json:"item_name"`
	UnitCode     string `json:"unit_code"`
	ItemTypeCode string `json:"item_type_code"`
	ItemGroup    string `json:"item_group"`
	LocationCode string `json:"location_code"`
	Awal         string `json:"awal"`
	Masuk        string `json:"masuk"`
	Keluar       string `json:"keluar"`
	Peny         string `json:"peny"`
	Akhir        string `json:"akhir"`
	Opname       string `json:"opname"`
	Selisih      string `json:"selisih"`
}

type RawMaterialReportRequest struct {
	TglAwal  time.Time `json:"tgl_awal" validate:"required"`
	TglAkhir time.Time `json:"tgl_akhir" validate:"required"`
	ItemCode string    `json:"item_code"`
	ItemName string    `json:"item_name"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type RawMaterialReportResponse struct {
	ItemCode     string `json:"item_code"`
	ItemName     string `json:"item_name"`
	UnitCode     string `json:"unit_code"`
	ItemTypeCode string `json:"item_type_code"`
	ItemGroup    string `json:"item_group"`
	LocationCode string `json:"location_code"`
	Awal         string `json:"awal"`
	Masuk        string `json:"masuk"`
	Keluar       string `json:"keluar"`
	Peny         string `json:"peny"`
	Akhir        string `json:"akhir"`
	Opname       string `json:"opname"`
	Selisih      string `json:"selisih"`
}
