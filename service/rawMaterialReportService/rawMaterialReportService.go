package rawMaterialReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/rawMaterialReportRepository"
	"context"
	"time"
)

// RawMaterialReportService sits on top of the rawMaterialReportRepository and exposes use-case oriented APIs
// for raw material reporting.

type RawMaterialReportService struct {
	repo *rawMaterialReportRepository.RawMaterialReportRepository
}

func NewRawMaterialReportService(repo *rawMaterialReportRepository.RawMaterialReportRepository) *RawMaterialReportService {
	return &RawMaterialReportService{repo: repo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves raw material report with filters and pagination
func (s *RawMaterialReportService) GetReport(from, to time.Time, itemCode, itemName string, page, limit int) ([]model.RawMaterialReportResponse, int64, error) {
	ctx := context.Background()
	filter := rawMaterialReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	return s.repo.GetReport(ctx, filter)
}
