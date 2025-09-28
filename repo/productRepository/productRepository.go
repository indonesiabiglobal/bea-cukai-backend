package productRepository

import (
	"Bea-Cukai/model"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetAll(req model.ProductRequest) ([]model.Product, int64, error)
	GetByCode(code string) (*model.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(req model.ProductRequest) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	// Build the base query
	query := r.db.Model(&model.Product{})

	// Apply filters
	if req.ItemCode != "" {
		query = query.Where("item_code LIKE ?", "%"+req.ItemCode+"%")
	}
	if req.ItemName != "" {
		query = query.Where("item_name LIKE ?", "%"+req.ItemName+"%")
	}
	if req.ItemGroup != "" {
		query = query.Where("item_group = ?", req.ItemGroup)
	}
	if req.ItemTypeCode != "" {
		query = query.Where("item_type_code = ?", req.ItemTypeCode)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply pagination
	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Execute query with ordering
	if err := query.Order("item_code ASC").Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}

	return products, total, nil
}

func (r *productRepository) GetByCode(code string) (*model.Product, error) {
	var product model.Product
	
	// Clean the code parameter
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("item code cannot be empty")
	}

	if err := r.db.Where("item_code = ?", code).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product with code %s not found", code)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}