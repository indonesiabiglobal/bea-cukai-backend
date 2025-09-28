package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ExpenditureProduct struct {
	Idx         int             `json:"idx" gorm:"primaryKey;autoIncrement"`
	JenisPabean string          `json:"jenis_pabean" gorm:"type:varchar(255)"`
	NoPabean    string          `json:"no_pabean" gorm:"type:varchar(255)"`
	TglPabean   time.Time       `json:"tgl_pabean" gorm:"type:date"`
	TransNo     string          `json:"trans_no" gorm:"type:varchar(255)"`
	TransDate   time.Time       `json:"trans_date" gorm:"type:date"`
	CustCode    string          `json:"cust_code" gorm:"type:varchar(255)"`
	CustName    string          `json:"cust_name" gorm:"type:varchar(255)"`
	ItemCode    string          `json:"item_code" gorm:"type:varchar(255)"`
	ItemName    string          `json:"item_name" gorm:"type:varchar(255)"`
	DlvQty      decimal.Decimal `json:"dlv_qty" gorm:"type:decimal(20,2);not null;default:0"`
	SalesUnit   string          `json:"sales_unit" gorm:"type:varchar(255)"`
	CurrCode    string          `json:"curr_code" gorm:"type:varchar(255)"`
	NetPrice    decimal.Decimal `json:"net_price" gorm:"type:decimal(20,2);not null;default:0"`
	NetAmount   decimal.Decimal `json:"net_amount" gorm:"type:decimal(20,2);not null;default:0"`
}

// TableName overrides the default table name.
func (ExpenditureProduct) TableName() string { return "tr_pengeluaran_barang" }

type ExpenditureProductRequest struct {
	JenisPabean string          `json:"jenis_pabean"`
	NoPabean    string          `json:"no_pabean"`
	TglPabean   time.Time       `json:"tgl_pabean"`
	TransNo     string          `json:"trans_no"`
	TransDate   time.Time       `json:"trans_date"`
	CustCode    string          `json:"cust_code"`
	CustName    string          `json:"cust_name"`
	ItemCode    string          `json:"item_code"`
	ItemName    string          `json:"item_name"`
	DlvQty      decimal.Decimal `json:"dlv_qty"`
	SalesUnit   string          `json:"sales_unit"`
	CurrCode    string          `json:"curr_code"`
	NetPrice    decimal.Decimal `json:"net_price"`
	NetAmount   decimal.Decimal `json:"net_amount"`
}

type ExpenditureProductResponse struct {
	Idx         int             `json:"idx"`
	JenisPabean string          `json:"jenis_pabean"`
	NoPabean    string          `json:"no_pabean"`
	TglPabean   time.Time       `json:"tgl_pabean"`
	TransNo     string          `json:"trans_no"`
	TransDate   time.Time       `json:"trans_date"`
	CustCode    string          `json:"cust_code"`
	CustName    string          `json:"cust_name"`
	ItemCode    string          `json:"item_code"`
	ItemName    string          `json:"item_name"`
	DlvQty      decimal.Decimal `json:"dlv_qty"`
	SalesUnit   string          `json:"sales_unit"`
	CurrCode    string          `json:"curr_code"`
	NetPrice    decimal.Decimal `json:"net_price"`
	NetAmount   decimal.Decimal `json:"net_amount"`
}
