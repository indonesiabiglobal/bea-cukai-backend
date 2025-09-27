package model

// MasterDoctor maps to the physical table "master_docter" while the domain name
// stays readable in code. No gorm.Model embedded because the schema has no
// surrogate id or timestamps.

// Suggested statuses (adjust as needed): "AKTIF", "NONAKTIF".
// Add validation at service layer if you want to enforce an enum.

type MasterDoctor struct {
	KodeDokter    string `json:"kode_dokter" gorm:"primaryKey;size:64"`
	Prefix        string `json:"prefix" gorm:"size:32"`
	NamaDokter    string `json:"nama_dokter" gorm:"size:128;index:idx_md_nama"`
	Suffix        string `json:"suffix" gorm:"size:32"`
	Status        string `json:"status" gorm:"size:32;index:idx_md_status"`
	KodeSpesialis string `json:"kode_spesialis" gorm:"size:64;index:idx_md_spesialis"`
	NamaSpesialis string `json:"nama_spesialis" gorm:"size:128"`
}

func (MasterDoctor) TableName() string { return "master_dokter" }

// ==========================
// DTOs
// ==========================

// MasterDoctorRequest for create/update operations.

type MasterDoctorRequest struct {
	KodeDokter    string `json:"kode_dokter" validate:"required"`
	Prefix        string `json:"prefix"`
	NamaDokter    string `json:"nama_dokter" validate:"required"`
	Suffix        string `json:"suffix"`
	Status        string `json:"status"`
	KodeSpesialis string `json:"kode_spesialis"`
	NamaSpesialis string `json:"nama_spesialis"`
}

// MasterDoctorResponse returned to clients.

type MasterDoctorResponse struct {
	KodeDokter    string `json:"kode_dokter"`
	Prefix        string `json:"prefix"`
	NamaDokter    string `json:"nama_dokter"`
	Suffix        string `json:"suffix"`
	Status        string `json:"status"`
	KodeSpesialis string `json:"kode_spesialis"`
	NamaSpesialis string `json:"nama_spesialis"`
}
