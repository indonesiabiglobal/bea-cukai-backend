package productService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/productRepository"
	"math"
)

type ProductService struct {
	repo productRepository.ProductRepository
}

func NewProductService(repo productRepository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetAll(req model.ProductRequest) ([]model.Product, int64, map[string]interface{}, error) {
	// Set defaults for pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Get products from repository
	products, total, err := s.repo.GetAll(req)
	if err != nil {
		return nil, 0, nil, err
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	// Prepare metadata
	meta := map[string]interface{}{
		"page":        req.Page,
		"limit":       req.Limit,
		"total_count": total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}

	return products, total, meta, nil
}

func (s *ProductService) GetByCode(code string) (*model.Product, error) {
	return s.repo.GetByCode(code)
}