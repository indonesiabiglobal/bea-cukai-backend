package entryProductService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/entryProductRepository"
	"context"
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
func (s *EntryProductService) GetReport(filter entryProductRepository.GetReportFilter) ([]model.EntryProduct, int64, error) {
	ctx := context.Background()
	return s.entryProductRepo.GetReport(ctx, filter)
}
