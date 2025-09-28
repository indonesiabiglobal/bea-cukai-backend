package auxiliaryMaterialReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/auxiliaryMaterialReportRepository"
	"context"
	"time"
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
func (s *AuxiliaryMaterialReportService) GetReport(from, to time.Time, itemCode, itemName, lap string, page, limit int) ([]model.AuxiliaryMaterialReportResponse, int64, error) {
	ctx := context.Background()
	filter := auxiliaryMaterialReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Lap:      lap,
		Page:     page,
		Limit:    limit,
	}
	return s.repo.GetReport(ctx, filter)
}
