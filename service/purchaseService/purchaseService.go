package purchaseService

import (
	"context"

	"Dashboard-TRDP/model"
	"Dashboard-TRDP/repo/purchaseRepository"
)

type PurchaseService struct {
	repo *purchaseRepository.PurchaseRepository
}

type VendorsResult struct {
	Items []purchaseRepository.Vendor `json:"items"`
	Total int64                       `json:"total"`
}

const (
	defaultLimit = 10
	maxLimit     = 200
	allBatchSize = 500
	hardCapAll   = 10000
)

func NewPurchaseService(repo *purchaseRepository.PurchaseRepository) *PurchaseService {
	return &PurchaseService{repo: repo}
}

// Summary KPI
func (s *PurchaseService) GetKPISummary(dto model.PurchaseRequestParam) (purchaseRepository.PurchaseKPISummary, error) {
	return s.repo.GetKPISummary(context.Background(), dto)
}

// Trend harian
func (s *PurchaseService) GetTrend(dto model.PurchaseRequestParam) ([]purchaseRepository.PurchaseTrendPoint, error) {
	return s.repo.GetTrend(context.Background(), dto)
}

// Top Vendors
func (s *PurchaseService) GetTopVendors(dto model.PurchaseRequestParam) ([]purchaseRepository.TopVendor, error) {
	return s.repo.GetTopVendors(context.Background(), dto)
}

// Top Products
func (s *PurchaseService) GetTopProducts(dto model.PurchaseRequestParam) ([]purchaseRepository.TopProduct, error) {
	return s.repo.GetTopProducts(context.Background(), dto)
}

// By Category
func (s *PurchaseService) GetByCategory(dto model.PurchaseRequestParam, limit int) ([]purchaseRepository.ByCategory, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetByCategory(context.Background(), dto, limit)
}

func (s *PurchaseService) GetVendors(
	ctx context.Context, search string, page, limit int,
) (VendorsResult, error) {
	// sentinel: -1 = minta semua
	if limit == -1 {
		items, err := s.GetAllVendors(ctx, search)
		if err != nil {
			return VendorsResult{}, err
		}
		return VendorsResult{Items: items, Total: int64(len(items))}, nil
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

	items, total, err := s.repo.GetVendors(ctx, search, limit, offset)
	if err != nil {
		return VendorsResult{}, err
	}
	return VendorsResult{Items: items, Total: total}, nil
}

func (s *PurchaseService) GetAllVendors(
	ctx context.Context, search string,
) ([]purchaseRepository.Vendor, error) {
	var all []purchaseRepository.Vendor

	page := 1
	for {
		offset := (page - 1) * allBatchSize
		batch, total, err := s.repo.GetVendors(ctx, search, allBatchSize, offset)
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