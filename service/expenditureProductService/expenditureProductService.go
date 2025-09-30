package expenditureProductService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/expenditureProductRepository"
	"context"
)

// ExpenditureProductService sits on top of the expenditureProductRepository and exposes use-case oriented APIs
// including report aggregations and simple readers.

type ExpenditureProductService struct {
	expenditureProductRepo *expenditureProductRepository.ExpenditureProductRepository
}

func NewExpenditureProductService(expenditureProductRepo *expenditureProductRepository.ExpenditureProductRepository) *ExpenditureProductService {
	return &ExpenditureProductService{expenditureProductRepo: expenditureProductRepo}
}

// ==========================
//  Aggregations
// ==========================

// GetReport retrieves expenditure products with filters and pagination
func (s *ExpenditureProductService) GetReport(filter expenditureProductRepository.GetReportFilter) ([]model.ExpenditureProduct, int64, error) {
	ctx := context.Background()
	return s.expenditureProductRepo.GetReport(ctx, filter)
}
