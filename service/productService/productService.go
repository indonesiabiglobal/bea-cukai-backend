package productService

import (
	"context"

	"Dashboard-TRDP/model"
	"Dashboard-TRDP/repo/productRepository"
)

type ProductService struct {
	repo *productRepository.ProductRepository
}

func NewProductService(repo *productRepository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

type CategoriesResult struct {
	Items []productRepository.MPCategory `json:"items"`
	Total int64                          `json:"total"`
}

const (
	defaultLimit = 10
	maxLimit     = 200
	allBatchSize = 500
	hardCapAll   = 10000
)

func (s *ProductService) GetCategories(
	ctx context.Context, search string, page, limit int,
) (CategoriesResult, error) {
	// sentinel: -1 = minta semua
	if limit == -1 {
		items, err := s.GetAllCategories(ctx, search)
		if err != nil {
			return CategoriesResult{}, err
		}
		return CategoriesResult{Items: items, Total: int64(len(items))}, nil
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	items, total, err := s.repo.GetCategories(ctx, search, limit, offset)
	if err != nil {
		return CategoriesResult{}, err
	}
	return CategoriesResult{Items: items, Total: total}, nil
}

func (s *ProductService) GetAllCategories(
	ctx context.Context, search string,
) ([]productRepository.MPCategory, error) {
	var all []productRepository.MPCategory

	page := 1
	for {
		offset := (page - 1) * allBatchSize
		batch, total, err := s.repo.GetCategories(ctx, search, allBatchSize, offset)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		all = append(all, batch...)

		// selesai jika sudah mencapai total
		if int64(len(all)) >= total {
			break
		}
		page++
	}
	return all, nil
}

type ProductsResult struct {
	Items []model.MasterProduct `json:"items"`
	Total int64                 `json:"total"`
}

func (s *ProductService) GetProducts(ctx context.Context, f productRepository.ProductFilter, page, limit int) (ProductsResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	items, total, err := s.repo.GetProducts(ctx, f, limit, offset)
	if err != nil {
		return ProductsResult{}, err
	}
	return ProductsResult{Items: items, Total: total}, nil
}
