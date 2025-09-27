package patientService

import (
	"context"
	"time"

	"Dashboard-TRDP/repo/patientRepository"
)

type PatientService struct {
	repo *patientRepository.PatientRepository
}

func NewPatientService(repo *patientRepository.PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

/* ====== Result wrappers ====== */

type InpatientMonitoringResult struct {
	Items []patientRepository.InpatientMonitoringRow `json:"items"`
	Total int64                                      `json:"total"`
}

func (s *PatientService) MonitoringInpatient(ctx context.Context, page, limit int) (InpatientMonitoringResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	items, total, err := s.repo.MonitoringInpatient(ctx, limit, offset)
	if err != nil {
		return InpatientMonitoringResult{}, err
	}
	return InpatientMonitoringResult{Items: items, Total: total}, nil
}

type DischargedResult struct {
	Items   []patientRepository.DischargedRow   `json:"items"`
	Summary patientRepository.DischargedSummary `json:"summary"`
}

func (s *PatientService) DischargedPatients(ctx context.Context, from, to time.Time, page, limit int) (DischargedResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	dr := patientRepository.DateRange{From: from, To: to}

	items, summary, err := s.repo.DischargedPatients(ctx, dr, limit, offset)
	if err != nil {
		return DischargedResult{}, err
	}
	return DischargedResult{Items: items, Summary: summary}, nil
}
