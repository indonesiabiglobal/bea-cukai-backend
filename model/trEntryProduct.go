package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type EntryProduct struct {
	Idx         int             `json:"idx" gorm:"not null;index:idx_idx"`
	JenisPabean string          `json:"jenis_pabean" gorm:"type:varchar(255)"`
	NoPabean    string          `json:"no_pabean" gorm:"type:varchar(255)"`
	TglPabean   time.Time       `json:"tgl_pabean" gorm:"type:date"`
	TransNo     string          `json:"trans_no" gorm:"type:varchar(255)"`
	VendDlvNo   string          `json:"vend_dlv_no" gorm:"type:varchar(255)"`
	TransDate   time.Time       `json:"trans_date" gorm:"type:date"`
	VendorCode  string          `json:"vendor_code" gorm:"type:varchar(255)"`
	VendorName  string          `json:"vendor_name" gorm:"type:varchar(255)"`
	ItemCode    string          `json:"item_code" gorm:"type:varchar(255)"`
	ItemName    string          `json:"item_name" gorm:"type:varchar(255)"`
	RcvQty      decimal.Decimal `json:"rcv_qty" gorm:"type:decimal(20,2);not null;default:0"`
	PchUnit     string          `json:"pch_unit" gorm:"type:varchar(255)"`
	CurrCode    string          `json:"curr_code" gorm:"type:varchar(255)"`
	NetPrice    decimal.Decimal `json:"net_price" gorm:"type:decimal(20,2);not null;default:0"`
	NetAmount   decimal.Decimal `json:"net_amount" gorm:"type:decimal(20,2);not null;default:0"`
}

// TableName overrides the default table name.
func (EntryProduct) TableName() string { return "tr_pemasukan_barang" }

type EntryProductRequest struct {
	Idx         int             `json:"idx" validate:"required"`
	JenisPabean string          `json:"jenis_pabean"`
	NoPabean    string          `json:"no_pabean"`
	TglPabean   time.Time       `json:"tgl_pabean"`
	TransNo     string          `json:"trans_no"`
	VendDlvNo   string          `json:"vend_dlv_no"`
	TransDate   time.Time       `json:"trans_date"`
	VendorCode  string          `json:"vendor_code"`
	VendorName  string          `json:"vendor_name"`
	ItemCode    string          `json:"item_code"`
	ItemName    string          `json:"item_name"`
	RcvQty      decimal.Decimal `json:"rcv_qty"`
	PchUnit     string          `json:"pch_unit"`
	CurrCode    string          `json:"curr_code"`
	NetPrice    decimal.Decimal `json:"net_price"`
	NetAmount   decimal.Decimal `json:"net_amount"`
}

type EntryProductResponse struct {
	ID          uint            `json:"id"`
	Idx         int             `json:"idx"`
	JenisPabean string          `json:"jenis_pabean"`
	NoPabean    string          `json:"no_pabean"`
	TglPabean   time.Time       `json:"tgl_pabean"`
	TransNo     string          `json:"trans_no"`
	VendDlvNo   string          `json:"vend_dlv_no"`
	TransDate   time.Time       `json:"trans_date"`
	VendorCode  string          `json:"vendor_code"`
	VendorName  string          `json:"vendor_name"`
	ItemCode    string          `json:"item_code"`
	ItemName    string          `json:"item_name"`
	RcvQty      decimal.Decimal `json:"rcv_qty"`
	PchUnit     string          `json:"pch_unit"`
	CurrCode    string          `json:"curr_code"`
	NetPrice    decimal.Decimal `json:"net_price"`
	NetAmount   decimal.Decimal `json:"net_amount"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
