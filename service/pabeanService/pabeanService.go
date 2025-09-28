package pabeanService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/pabeanRepository"
	"context"
)

// PabeanService sits on top of the pabeanRepository and exposes use-case oriented APIs.

type PabeanService struct {
	pabeanRepo *pabeanRepository.PabeanRepository
}

func NewPabeanService(pabeanRepo *pabeanRepository.PabeanRepository) *PabeanService {
	return &PabeanService{pabeanRepo: pabeanRepo}
}

// ==========================
// Business Operations
// ==========================

// GetAll retrieves all pabean documents
func (s *PabeanService) GetAll() ([]model.MsPabean, error) {
	ctx := context.Background()
	return s.pabeanRepo.GetAll(ctx)
}
