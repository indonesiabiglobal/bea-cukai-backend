package finishedProductReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/finishedProductReportRepository"
	"context"
)

// FinishedProductReportService sits on top of the finishedProductReportRepository and exposes use-case oriented APIs
// for finished product reporting.

type FinishedProductReportService struct {
	repo *finishedProductReportRepository.FinishedProductReportRepository
}

func NewFinishedProductReportService(repo *finishedProductReportRepository.FinishedProductReportRepository) *FinishedProductReportService {
	return &FinishedProductReportService{repo: repo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves finished product report with filters and pagination
func (s *FinishedProductReportService) GetReport(filter finishedProductReportRepository.GetReportFilter) ([]model.FinishedProductReportResponse, int64, error) {
	ctx := context.Background()
	return s.repo.GetReport(ctx, filter)
}
