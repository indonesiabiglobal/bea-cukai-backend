package visitService

import (
	"context"
	"time"

	"Dashboard-TRDP/repo/visitRepository"
)

type VisitService struct {
	repo *visitRepository.VisitRepository
}

func NewVisitService(repo *visitRepository.VisitRepository) *VisitService {
	return &VisitService{repo: repo}
}

// ===== Helpers =====
func dr(from, to time.Time) visitRepository.DateRange {
	return visitRepository.DateRange{From: from, To: to}
}

// ===== Services =====

func (s *VisitService) GetKPISummary(from, to time.Time) (visitRepository.VisitKPISummary, error) {
	return s.repo.GetKPISummary(context.Background(), dr(from, to))
}

func (s *VisitService) GetTrend(from, to time.Time) ([]visitRepository.VisitTrendPoint, error) {
	return s.repo.GetTrend(context.Background(), dr(from, to))
}

func (s *VisitService) GetTopServices(from, to time.Time) ([]visitRepository.TopService, error) {
	return s.repo.GetTopServices(context.Background(), dr(from, to))
}

func (s *VisitService) GetTopGuarantors(from, to time.Time) ([]visitRepository.TopGuarantor, error) {
	return s.repo.GetTopGuarantors(context.Background(), dr(from, to))
}

func (s *VisitService) GetByDOW(from, to time.Time) ([]visitRepository.ByDOW, error) {
	return s.repo.GetByDOW(context.Background(), dr(from, to))
}

func (s *VisitService) GetByRegionKota(from, to time.Time, limit int) ([]visitRepository.ByRegion, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetByRegionKota(context.Background(), dr(from, to), limit)
}

func (s *VisitService) GetMixIPOP(from, to time.Time) ([]visitRepository.IPOPBreakdown, error) {
	return s.repo.GetMixIPOP(context.Background(), dr(from, to))
}

func (s *VisitService) GetLOSBuckets(from, to time.Time) ([]visitRepository.LOSBucket, error) {
	return s.repo.GetLOSBuckets(context.Background(), dr(from, to))
}
