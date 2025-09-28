package entryProductService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/entryProductRepository"
	"context"
	"time"
)

// EntryProductService sits on top of the entryProductRepository and exposes use-case oriented APIs
// including report aggregations and simple readers.

type EntryProductService struct {
	entryProductRepo *entryProductRepository.EntryProductRepository
}

func NewEntryProductService(entryProductRepo *entryProductRepository.EntryProductRepository) *EntryProductService {
	return &EntryProductService{entryProductRepo: entryProductRepo}
}

// ==========================
//  Aggregations
// ==========================

// GetReport retrieves entry products with filters and pagination
func (s *EntryProductService) GetReport(from, to time.Time, pabeanType, productGroup, noPabean, productCode, productName string, page, limit int) ([]model.EntryProduct, int64, error) {
	ctx := context.Background()
	filter := entryProductRepository.GetReportFilter{
		From:         from,
		To:           to,
		PabeanType:   pabeanType,
		ProductGroup: productGroup,
		NoPabean:     noPabean,
		ProductCode:  productCode,
		ProductName:  productName,
		Page:         page,
		Limit:        limit,
	}
	return s.entryProductRepo.GetReport(ctx, filter)
}
