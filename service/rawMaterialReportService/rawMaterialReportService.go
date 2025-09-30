package rawMaterialReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/rawMaterialReportRepository"
	"context"
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
func (s *RawMaterialReportService) GetReport(filter rawMaterialReportRepository.GetReportFilter) ([]model.RawMaterialReportResponse, int64, error) {
	ctx := context.Background()
	return s.repo.GetReport(ctx, filter)
}
