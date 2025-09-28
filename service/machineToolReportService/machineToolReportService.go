package machineToolReportService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/machineToolReportRepository"
	"context"
	"time"
)

// MachineToolReportService sits on top of the machineToolReportRepository and exposes use-case oriented APIs
// for machine and tool reporting.

type MachineToolReportService struct {
	repo *machineToolReportRepository.MachineToolReportRepository
}

func NewMachineToolReportService(repo *machineToolReportRepository.MachineToolReportRepository) *MachineToolReportService {
	return &MachineToolReportService{repo: repo}
}

// ==========================
// Business Operations
// ==========================

// GetReport retrieves machine and tool report with filters and pagination
func (s *MachineToolReportService) GetReport(from, to time.Time, itemCode, itemName string, page, limit int) ([]model.MachineToolReportResponse, int64, error) {
	ctx := context.Background()
	filter := machineToolReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	return s.repo.GetReport(ctx, filter)
}
