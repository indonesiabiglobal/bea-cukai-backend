package model

// MasterGuarantor maps to the physical table "master_guarantor".
// We keep the code clean (no gorm.Model) since your schema doesn't include
// surrogate id or timestamps. Primary key uses kode_penjamin.

// Suggested values for Status (adjust in service/validator):
// - "AKTIF", "NONAKTIF"

// Indexes are added for common lookups: nama_penjamin, kota, status.

type MasterGuarantor struct {
	KodePenjamin string `json:"kode_penjamin" gorm:"primaryKey;size:64"`
	NamaPenjamin string `json:"nama_penjamin" gorm:"size:255;index:idx_mg_nama"`
	Alamat       string `json:"alamat" gorm:"size:255"`
	Kota         string `json:"kota" gorm:"size:128;index:idx_mg_kota"`
	NoTelp       string `json:"no_telp" gorm:"size:64"`
	NoFax        string `json:"no_fax" gorm:"size:64"`
	Kontak       string `json:"kontak" gorm:"size:128"`
	Status       string `json:"status" gorm:"size:32;index:idx_mg_status"`
}

// TableName enforces the DB table name.
func (MasterGuarantor) TableName() string { return "master_penjamin" }

// ==========================
// DTOs
// ==========================

// MasterGuarantorRequest defines the payload for create/update operations.

type MasterGuarantorRequest struct {
	KodePenjamin string `json:"kode_penjamin" validate:"required"`
	NamaPenjamin string `json:"nama_penjamin"`
	Alamat       string `json:"alamat"`
	Kota         string `json:"kota"`
	NoTelp       string `json:"no_telp"`
	NoFax        string `json:"no_fax"`
	Kontak       string `json:"kontak"`
	Status       string `json:"status"`
}

// MasterGuarantorResponse is used for API responses.

type MasterGuarantorResponse struct {
	KodePenjamin string `json:"kode_penjamin"`
	NamaPenjamin string `json:"nama_penjamin"`
	Alamat       string `json:"alamat"`
	Kota         string `json:"kota"`
	NoTelp       string `json:"no_telp"`
	NoFax        string `json:"no_fax"`
	Kontak       string `json:"kontak"`
	Status       string `json:"status"`
}

// Notes:
// - Phone numbers kept as string to preserve leading zeros and symbols.
// - If you need uniqueness on nama_penjamin within kota, consider a composite unique index.
