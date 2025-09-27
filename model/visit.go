package model

import (
	"time"

	"gorm.io/gorm"
)

// Visit is the GORM model mapping to the physical table "visits".
// Keep domain name "Visit" while the DB table stays plural via TableName().
// Fields and json tags follow your provided schema.

type Visit struct {
	gorm.Model

	UID                  string     `json:"uid" gorm:"size:64;uniqueIndex:idx_uid"`
	PatID                string     `json:"patid" gorm:"size:64;index:idx_patid"`
	Episode              string     `json:"episode" gorm:"size:64;index:idx_episode"`
	NamaPasien           string     `json:"nama_pasien" gorm:"size:128"`
	TipeAdmisiLayanan    string     `json:"tipe_admisi_layanan" gorm:"size:64"`
	NamaAdmisiLayanan    string     `json:"nama_admisi_layanan" gorm:"size:128"`
	NamaPenjamin         string     `json:"nama_penjamin" gorm:"size:128;index:idx_penjamin"`
	TipeDischarge        string     `json:"tipe_discharge" gorm:"size:64"`
	KondisiDischarge     string     `json:"kondisi_discharge" gorm:"size:64"`
	TglMasuk             time.Time  `json:"tgl_masuk" gorm:"type:date;index:idx_tgl_masuk"`
	TglKeluar            *time.Time `json:"tgl_keluar" gorm:"type:date;index:idx_tgl_keluar"`
	Kecamatan            string     `json:"kecamatan" gorm:"size:128"`
	Kota                 string     `json:"kota" gorm:"size:128"`
	AdmisiMasuk          string     `json:"admisi_masuk" gorm:"size:128"`
	AlasanAdmisi         string     `json:"alasan_admisi" gorm:"type:text"`
	KodeDPJP             string     `json:"kode_dpjp" gorm:"size:64;index:idx_kode_dpjp"`
	NamaDPJP             string     `json:"nama_dpjp" gorm:"size:128"`
	NoBed                string     `json:"no_bed" gorm:"size:64"`
	Kelas                string     `json:"kelas" gorm:"size:64"`
	PatsvcMrhPriDiag     string     `json:"patsvcmrhpridiag" gorm:"size:255"`
	PatsvcMrhPriDiagICD  string     `json:"patsvcmrhpridiagicd" gorm:"size:64"`
	PatsvcMrhPriDiagDesc string     `json:"patsvcmrhpridiagicddesc" gorm:"size:255"`
	NamaYangMerujuk      string     `json:"nama_yang_merujuk" gorm:"size:128"`
	PatDOB               time.Time  `json:"patdob" gorm:"type:date;index:idx_patdob"`
	KategoriPasien       string     `json:"kategori_pasien" gorm:"size:64;index:idx_kategori"`
}

// TableName enforces the DB table name.
func (Visit) TableName() string { return "kunjungan" }

// ==========================
// DTOs
// ==========================

// VisitRequest is used for create/update payloads.
// Use pointer for TglKeluar to allow null (patient not yet discharged).

type VisitRequest struct {
	UID                  string     `json:"uid" validate:"required"`
	PatID                string     `json:"patid" validate:"required"`
	Episode              string     `json:"episode" validate:"required"`
	NamaPasien           string     `json:"nama_pasien"`
	TipeAdmisiLayanan    string     `json:"tipe_admisi_layanan"`
	NamaAdmisiLayanan    string     `json:"nama_admisi_layanan"`
	NamaPenjamin         string     `json:"nama_penjamin"`
	TipeDischarge        string     `json:"tipe_discharge"`
	KondisiDischarge     string     `json:"kondisi_discharge"`
	TglMasuk             time.Time  `json:"tgl_masuk" validate:"required"`
	TglKeluar            *time.Time `json:"tgl_keluar"`
	Kecamatan            string     `json:"kecamatan"`
	Kota                 string     `json:"kota"`
	AdmisiMasuk          string     `json:"admisi_masuk"`
	AlasanAdmisi         string     `json:"alasan_admisi"`
	KodeDPJP             string     `json:"kode_dpjp"`
	NamaDPJP             string     `json:"nama_dpjp"`
	NoBed                string     `json:"no_bed"`
	Kelas                string     `json:"kelas"`
	PatsvcMrhPriDiag     string     `json:"patsvcmrhpridiag"`
	PatsvcMrhPriDiagICD  string     `json:"patsvcmrhpridiagicd"`
	PatsvcMrhPriDiagDesc string     `json:"patsvcmrhpridiagicddesc"`
	NamaYangMerujuk      string     `json:"nama_yang_merujuk"`
	PatDOB               time.Time  `json:"patdob"`
	KategoriPasien       string     `json:"kategori_pasien"`
}

// VisitResponse is returned to clients.

type VisitResponse struct {
	ID                   uint       `json:"id"`
	UID                  string     `json:"uid"`
	PatID                string     `json:"patid"`
	Episode              string     `json:"episode"`
	NamaPasien           string     `json:"nama_pasien"`
	TipeAdmisiLayanan    string     `json:"tipe_admisi_layanan"`
	NamaAdmisiLayanan    string     `json:"nama_admisi_layanan"`
	NamaPenjamin         string     `json:"nama_penjamin"`
	TipeDischarge        string     `json:"tipe_discharge"`
	KondisiDischarge     string     `json:"kondisi_discharge"`
	TglMasuk             time.Time  `json:"tgl_masuk"`
	TglKeluar            *time.Time `json:"tgl_keluar"`
	Kecamatan            string     `json:"kecamatan"`
	Kota                 string     `json:"kota"`
	AdmisiMasuk          string     `json:"admisi_masuk"`
	AlasanAdmisi         string     `json:"alasan_admisi"`
	KodeDPJP             string     `json:"kode_dpjp"`
	NamaDPJP             string     `json:"nama_dpjp"`
	NoBed                string     `json:"no_bed"`
	Kelas                string     `json:"kelas"`
	PatsvcMrhPriDiag     string     `json:"patsvcmrhpridiag"`
	PatsvcMrhPriDiagICD  string     `json:"patsvcmrhpridiagicd"`
	PatsvcMrhPriDiagDesc string     `json:"patsvcmrhpridiagicddesc"`
	NamaYangMerujuk      string     `json:"nama_yang_merujuk"`
	PatDOB               time.Time  `json:"patdob"`
	KategoriPasien       string     `json:"kategori_pasien"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
