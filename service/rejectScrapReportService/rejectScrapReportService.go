package rejectScrapReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/rejectScrapReportRepository"
	"context"
)

// RejectScrapReportService sits on top of the rejectScrapReportRepository and exposes use-case oriented APIs
// for reject and scrap reporting.

type RejectScrapReportService struct {
	repo *rejectScrapReportRepository.RejectScrapReportRepository
}

func NewRejectScrapReportService(repo *rejectScrapReportRepository.RejectScrapReportRepository) *RejectScrapReportService {
	return &RejectScrapReportService{repo: repo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves reject and scrap report with filters and pagination
func (s *RejectScrapReportService) GetReport(filter rejectScrapReportRepository.GetReportFilter) ([]model.RejectScrapReportResponse, int64, error) {
	ctx := context.Background()
	return s.repo.GetReport(ctx, filter)
}
