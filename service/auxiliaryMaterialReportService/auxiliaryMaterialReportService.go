package auxiliaryMaterialReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/auxiliaryMaterialReportRepository"
	"context"
)

// AuxiliaryMaterialReportService sits on top of the auxiliaryMaterialReportRepository and exposes use-case oriented APIs
// for auxiliary material reporting.

type AuxiliaryMaterialReportService struct {
	repo *auxiliaryMaterialReportRepository.AuxiliaryMaterialReportRepository
}

func NewAuxiliaryMaterialReportService(repo *auxiliaryMaterialReportRepository.AuxiliaryMaterialReportRepository) *AuxiliaryMaterialReportService {
	return &AuxiliaryMaterialReportService{repo: repo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves auxiliary material report with filters and pagination
func (s *AuxiliaryMaterialReportService) GetReport(filter auxiliaryMaterialReportRepository.GetReportFilter) ([]model.AuxiliaryMaterialReportResponse, int64, error) {
	ctx := context.Background()
	
	return s.repo.GetReport(ctx, filter)
}
