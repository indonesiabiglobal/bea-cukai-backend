package expenditureProductService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/expenditureProductRepository"
	"context"
	"time"
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
func (s *ExpenditureProductService) GetReport(from, to time.Time, pabeanType, productGroup, noPabean, productCode, productName string, page, limit int) ([]model.ExpenditureProduct, int64, error) {
	ctx := context.Background()
	filter := expenditureProductRepository.GetReportFilter{
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
	return s.expenditureProductRepo.GetReport(ctx, filter)
}
