package incomeService

import (
	"context"
	"time"

	"Dashboard-TRDP/model"
	"Dashboard-TRDP/repo/incomeRepository"

	"github.com/jinzhu/copier"
)

// IncomeService sits on top of the incomeRepository and exposes use-case oriented APIs
// including dashboard aggregations and simple readers.

type IncomeService struct {
	incomeRepo *incomeRepository.IncomeRepository
}

func NewIncomeService(incomeRepo *incomeRepository.IncomeRepository) *IncomeService {
	return &IncomeService{incomeRepo: incomeRepo}
}

// ==========================
// Simple Readers (template-compatible)
// ==========================

// GetAllIncomes returns raw income rows (optionally pass a date range via params later if needed).
// NOTE: The original template filtered by userID via a relation that does not
// exist on the pendapatan table. We ignore userID here and return all data.
func (s *IncomeService) GetAllIncomes(userID uint) ([]model.IncomeResponse, error) {
	ctx := context.Background()
	rows, err := s.incomeRepo.GetAllIncomes(ctx, nil)
	if err != nil {
		return nil, err
	}
	var out []model.IncomeResponse
	if err := copier.Copy(&out, &rows); err != nil {
		return nil, err
	}
	return out, nil
}

// GetIncomeByID returns a single income by primary id.
func (s *IncomeService) GetIncomeByID(incomeID uint) (model.IncomeResponse, error) {
	ctx := context.Background()
	row, err := s.incomeRepo.GetIncomeByID(ctx, incomeID)
	if err != nil {
		return model.IncomeResponse{}, err
	}
	var out model.IncomeResponse
	if err := copier.Copy(&out, &row); err != nil {
		return model.IncomeResponse{}, err
	}
	return out, nil
}

// ==========================
// Dashboard Aggregations
// ==========================

// internal helper to build a repository DateRange
func makeRange(from, to time.Time) incomeRepository.DateRange {
	return incomeRepository.DateRange{From: from, To: to}
}

// KPI Summary (totals, unique counts)
func (s *IncomeService) GetKPISummary(from, to time.Time) (incomeRepository.KPISummary, error) {
	ctx := context.Background()
	return s.incomeRepo.GetKPISummary(ctx, makeRange(from, to))
}

// Daily trend of Net (and Debit/Credit)
func (s *IncomeService) GetRevenueTrend(from, to time.Time) ([]incomeRepository.RevenuePoint, error) {
	ctx := context.Background()
	return s.incomeRepo.GetRevenueTrend(ctx, makeRange(from, to))
}

// Top Units by Net
func (s *IncomeService) GetTopUnits(from, to time.Time, limit int) ([]incomeRepository.TopKeyAmount, error) {
	ctx := context.Background()
	return s.incomeRepo.GetTopUnits(ctx, makeRange(from, to), limit)
}

// Top Providers by Net
func (s *IncomeService) GetTopProviders(from, to time.Time, limit int) ([]incomeRepository.TopKeyAmount, error) {
	ctx := context.Background()
	return s.incomeRepo.GetTopProviders(ctx, makeRange(from, to), limit)
}

// Top Guarantors by Net
func (s *IncomeService) GetTopGuarantors(from, to time.Time, limit int) ([]incomeRepository.TopKeyAmount, error) {
	ctx := context.Background()
	return s.incomeRepo.GetTopGuarantors(ctx, makeRange(from, to), limit)
}

// Top Guarantor Groups by Net
func (s *IncomeService) GetTopGuarantorGroups(from, to time.Time, limit int) ([]incomeRepository.TopKeyAmount, error) {
	ctx := context.Background()
	return s.incomeRepo.GetTopGuarantorGroups(ctx, makeRange(from, to), limit)
}

// Revenue by Layanan
func (s *IncomeService) GetRevenueByService(from, to time.Time) ([]incomeRepository.TopKeyAmount, error) {
	ctx := context.Background()
	return s.incomeRepo.GetRevenueByService(ctx, makeRange(from, to))
}

// Mix IP/OP
func (s *IncomeService) GetRevenueByIPOP(from, to time.Time) ([]incomeRepository.MixIPOP, error) {
	ctx := context.Background()
	return s.incomeRepo.GetRevenueByIPOP(ctx, makeRange(from, to))
}

// Revenue grouped by Day-of-Week
func (s *IncomeService) GetRevenueByDOW(from, to time.Time) ([]incomeRepository.DOWPoint, error) {
	ctx := context.Background()
	return s.incomeRepo.GetRevenueByDOW(ctx, makeRange(from, to))
}
