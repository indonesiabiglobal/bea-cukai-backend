package rejectScrapReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/rejectScrapReportRepository"
	"context"
	"time"
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
func (s *RejectScrapReportService) GetReport(from, to time.Time, itemCode, itemName string, page, limit int) ([]model.RejectScrapReportResponse, int64, error) {
	ctx := context.Background()
	filter := rejectScrapReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	return s.repo.GetReport(ctx, filter)
}
