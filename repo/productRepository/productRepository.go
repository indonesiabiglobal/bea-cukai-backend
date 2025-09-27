package productRepository

import (
	"context"
	"strings"

	"Dashboard-TRDP/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

/* ========= DTO & Filter ========= */

type MPCategory struct {
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name"`
	ProductCount int64  `json:"product_count"`
}

type ProductFilter struct {
	Search        string
	CategoryCode  string
	Status        string
	KodeAkun      string
	KodePrinciple string
	ObatGenerik   *bool
	ObatKeras     *bool
	Narkotik      *bool
	Psikotropika  *bool
}

/* ========= Categories (grouped) ========= */

// GetCategories: daftar kategori (code+name) + total produk per kategori.
// Support pencarian by code/name kategori, juga ikut nama barang jika mau (opsional).
func (r *ProductRepository) GetCategories(ctx context.Context, search string, limit, offset int) ([]MPCategory, int64, error) {
	// Normalisasi pencarian
	q := strings.TrimSpace(search)

	// Base query (alias m)
	base := r.db.WithContext(ctx).Table(model.MasterProduct{}.TableName() + " m")

	// Filter pencarian (opsional)
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		base = base.Where(`
			LOWER(COALESCE(m.kode_kategori_barang, '')) LIKE ? OR
			LOWER(COALESCE(m.nama_kategori_barang, '')) LIKE ?
		`, like, like)
	}

	// Count distinct kategori (butuh subquery)
	var total int64
	countQ := base.
		Select("DISTINCT COALESCE(m.kode_kategori_barang, '') AS category_code, COALESCE(m.nama_kategori_barang, '') AS category_name")
	err := r.db.Table("(?) AS t", countQ).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Data dengan agregasi count
	var rows []MPCategory
	err = base.
		Select(`
			COALESCE(m.kode_kategori_barang, '')  AS category_code,
			COALESCE(m.nama_kategori_barang, '')  AS category_name,
			COUNT(*)                               AS product_count
		`).
		Group("COALESCE(m.kode_kategori_barang,''), COALESCE(m.nama_kategori_barang,'')").
		Order("category_name ASC, category_code ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	return rows, total, err
}

/* ========= Products (list) ========= */

// GetProducts: daftar produk dengan filter + pagination; kembalikan total untuk paging.
func (r *ProductRepository) GetProducts(ctx context.Context, f ProductFilter, limit, offset int) ([]model.MasterProduct, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MasterProduct{})

	// Search (kode/nama barang)
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where(`
			LOWER(kode_barang) LIKE ? OR
			LOWER(nama_barang) LIKE ?
		`, like, like)
	}

	if f.CategoryCode != "" {
		q = q.Where("kode_kategori_barang = ?", f.CategoryCode)
	}
	if f.Status != "" {
		q = q.Where("status_barang = ?", f.Status)
	}
	if f.KodeAkun != "" {
		q = q.Where("kode_akun_barang = ?", f.KodeAkun)
	}
	if f.KodePrinciple != "" {
		q = q.Where("kode_principle = ?", f.KodePrinciple)
	}
	if f.ObatGenerik != nil {
		q = q.Where("obat_generik = ?", *f.ObatGenerik)
	}
	if f.ObatKeras != nil {
		q = q.Where("obat_keras = ?", *f.ObatKeras)
	}
	if f.Narkotik != nil {
		q = q.Where("narkotik = ?", *f.Narkotik)
	}
	if f.Psikotropika != nil {
		q = q.Where("psikotropika = ?", *f.Psikotropika)
	}

	// Count total
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data
	var rows []model.MasterProduct
	err := q.
		Order("nama_barang ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error

	return rows, total, err
}
