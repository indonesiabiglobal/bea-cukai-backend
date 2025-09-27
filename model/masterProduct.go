package model

// MasterProduct maps to the physical table "master_product" while the code keeps
// using the domain name MasterProduct. We do not embed gorm.Model because the
// schema you provided doesn't include surrogate id or timestamps.
//
// Notes:
// - Boolean flags are modeled as bool. If your DB stores them as 0/1, GORM will
//   map bool <-> TINYINT(1)/BOOLEAN accordingly. If they are stored as 'Y'/'N',
//   we can switch to a custom type later.
// - "status_barang" is kept as string; if you prefer enum/bool, say the allowed
//   values and we'll tighten it.

// GORM tags: size and indexes are conservative defaults.

type MasterProduct struct {
	KodeBarang         string `json:"kode_barang" gorm:"primaryKey;size:64"`
	NamaBarang         string `json:"nama_barang" gorm:"size:255"`
	KodeKategoriBarang string `json:"kode_kategori_barang" gorm:"size:64;index:idx_mp_kategori"`
	NamaKategoriBarang string `json:"nama_kategori_barang" gorm:"size:128"`
	KodeAkunBarang     string `json:"kode_akun_barang" gorm:"size:64;index:idx_mp_akun"`
	NamaAkunBarang     string `json:"nama_akun_barang" gorm:"size:128"`
	KodePrinciple      string `json:"kode_principle" gorm:"size:64;index:idx_mp_principle"`
	NamaPrinciple      string `json:"nama_principle" gorm:"size:128"`
	StatusBarang       string `json:"status_barang" gorm:"size:32;index:idx_mp_status"`
	ObatGenerik        bool   `json:"obat_generik" gorm:"not null;default:false"`
	ObatKeras          bool   `json:"obat_keras" gorm:"not null;default:false"`
	Narkotik           bool   `json:"narkotik" gorm:"not null;default:false"`
	Psikotropika       bool   `json:"psikotropika" gorm:"not null;default:false"`
}

// TableName enforces the DB table name.
func (MasterProduct) TableName() string { return "master_barang" }

// ==========================
// DTOs
// ==========================

// MasterProductRequest is used for create/update payloads.

type MasterProductRequest struct {
	KodeBarang         string `json:"kode_barang" validate:"required"`
	NamaBarang         string `json:"nama_barang"`
	KodeKategoriBarang string `json:"kode_kategori_barang"`
	NamaKategoriBarang string `json:"nama_kategori_barang"`
	KodeAkunBarang     string `json:"kode_akun_barang"`
	NamaAkunBarang     string `json:"nama_akun_barang"`
	KodePrinciple      string `json:"kode_principle"`
	NamaPrinciple      string `json:"nama_principle"`
	StatusBarang       string `json:"status_barang"`
	ObatGenerik        bool   `json:"obat_generik"`
	ObatKeras          bool   `json:"obat_keras"`
	Narkotik           bool   `json:"narkotik"`
	Psikotropika       bool   `json:"psikotropika"`
}

// MasterProductResponse is returned to clients.

type MasterProductResponse struct {
	KodeBarang         string `json:"kode_barang"`
	NamaBarang         string `json:"nama_barang"`
	KodeKategoriBarang string `json:"kode_kategori_barang"`
	NamaKategoriBarang string `json:"nama_kategori_barang"`
	KodeAkunBarang     string `json:"kode_akun_barang"`
	NamaAkunBarang     string `json:"nama_akun_barang"`
	KodePrinciple      string `json:"kode_principle"`
	NamaPrinciple      string `json:"nama_principle"`
	StatusBarang       string `json:"status_barang"`
	ObatGenerik        bool   `json:"obat_generik"`
	ObatKeras          bool   `json:"obat_keras"`
	Narkotik           bool   `json:"narkotik"`
	Psikotropika       bool   `json:"psikotropika"`
}
