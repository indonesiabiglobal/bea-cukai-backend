package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// MySQLBit handles MySQL BIT type conversion
type MySQLBit int

// Scan implements the Scanner interface for database/sql
func (b *MySQLBit) Scan(value interface{}) error {
	if value == nil {
		*b = 0
		return nil
	}

	switch v := value.(type) {
	case []uint8:
		// MySQL BIT returns []uint8
		if len(v) > 0 && v[0] == 1 {
			*b = 1
		} else {
			*b = 0
		}
	case int64:
		*b = MySQLBit(v)
	case int:
		*b = MySQLBit(v)
	case bool:
		if v {
			*b = 1
		} else {
			*b = 0
		}
	default:
		return fmt.Errorf("cannot scan %T into MySQLBit", value)
	}
	return nil
}

// Value implements the driver.Valuer interface
func (b MySQLBit) Value() (driver.Value, error) {
	return int64(b), nil
}

type Product struct {
	ItemCode     string          `json:"item_code" gorm:"primaryKey;column:item_code"`
	ItemName     string          `json:"item_name" gorm:"column:item_name"`
	UnitCode     string          `json:"unit_code" gorm:"column:unit_code"`
	ItemTypeCode string          `json:"item_type_code" gorm:"column:item_type_code"`
	ItemGroup    string          `json:"item_group" gorm:"column:item_group"`
	SafetyStock  decimal.Decimal `json:"safety_stock" gorm:"column:safety_stock"`
	PchUnit      string          `json:"pch_unit" gorm:"column:pch_unit"`
	PchPrice     decimal.Decimal `json:"pch_price" gorm:"column:pch_price"`
	SalesUnit    string          `json:"sales_unit" gorm:"column:sales_unit"`
	SalesPrice   decimal.Decimal `json:"sales_price" gorm:"column:sales_price"`
	LastCost     decimal.Decimal `json:"last_cost" gorm:"column:last_cost"`
	JenisBahan   string          `json:"jenis_bahan" gorm:"column:jenis_bahan"`
	Tebal        decimal.Decimal `json:"tebal" gorm:"column:tebal"`
	Lebar        decimal.Decimal `json:"lebar" gorm:"column:lebar"`
	Panjang      decimal.Decimal `json:"panjang" gorm:"column:panjang"`
	Corona       MySQLBit        `json:"corona" gorm:"column:corona"` // 0=false, 1=true (MySQL BIT compatibility)
	Embos        MySQLBit        `json:"embos" gorm:"column:embos"`   // 0=false, 1=true (MySQL BIT compatibility)
	IsiPerGaiso  decimal.Decimal `json:"isi_per_gaiso" gorm:"column:isi_per_gaiso"`
	JmlGaiso     decimal.Decimal `json:"jml_gaiso" gorm:"column:jml_gaiso"`
	Seal         string          `json:"seal" gorm:"column:seal"`
	LebarHagata  decimal.Decimal `json:"lebar_hagata" gorm:"column:lebar_hagata"`
	DalamHagata  decimal.Decimal `json:"dalam_hagata" gorm:"column:dalam_hagata"`
	TipeHagata   string          `json:"tipe_hagata" gorm:"column:tipe_hagata"`
	IsiPerPalet  decimal.Decimal `json:"isi_per_palet" gorm:"column:isi_per_palet"`
	GrUnit       decimal.Decimal `json:"gr_unit" gorm:"column:gr_unit"`
	NoHan        int             `json:"no_han" gorm:"column:no_han"`
	Inactive     int             `json:"inactive" gorm:"column:inactive"`
	CreatedBy    string          `json:"created_by" gorm:"column:created_by"`
	CreatedDate  *time.Time      `json:"created_date" gorm:"column:created_date"`
	UpdatedBy    string          `json:"updated_by" gorm:"column:updated_by"`
	UpdatedDate  *time.Time      `json:"updated_date" gorm:"column:updated_date"`
}

// TableName specifies the table name for GORM
func (Product) TableName() string {
	return "ms_item"
}

// ProductRequest for filtering and pagination
type ProductRequest struct {
	ItemCode     string `json:"item_code" form:"item_code"`
	ItemName     string `json:"item_name" form:"item_name"`
	ItemGroup    string `json:"item_group" form:"item_group"`
	ItemTypeCode string `json:"item_type_code" form:"item_type_code"`
	Page         int    `json:"page" form:"page"`
	Limit        int    `json:"limit" form:"limit"`
}

// ProductResponse type alias for consistency
type ProductResponse = Product
