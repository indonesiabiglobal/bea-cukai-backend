package wipPositionReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/wipPositionReportRepository"
	"context"
	"time"
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
func (s *WipPositionReportService) GetReport(tglAwal, tglAkhir time.Time, itemCode, itemName string, page, limit int) ([]model.WipPositionReportResponse, int64, error) {
	ctx := context.Background()
	filter := wipPositionReportRepository.GetReportFilter{
		TglAwal:  tglAwal,
		TglAkhir: tglAkhir,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	return s.wipRepo.GetReport(ctx, filter)
}
