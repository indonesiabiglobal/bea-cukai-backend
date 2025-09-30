package wipPositionReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/wipPositionReportRepository"
	"context"
)

// WipPositionReportService sits on top of the wipPositionReportRepository and exposes use-case oriented APIs
// for WIP position reporting.

type WipPositionReportService struct {
	wipRepo *wipPositionReportRepository.WipPositionReportRepository
}

func NewWipPositionReportService(wipRepo *wipPositionReportRepository.WipPositionReportRepository) *WipPositionReportService {
	return &WipPositionReportService{wipRepo: wipRepo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves WIP position report with filters and pagination
func (s *WipPositionReportService) GetReport(filter wipPositionReportRepository.GetReportFilter) ([]model.WipPositionReportResponse, int64, error) {
	ctx := context.Background()
	return s.wipRepo.GetReport(ctx, filter)
}
