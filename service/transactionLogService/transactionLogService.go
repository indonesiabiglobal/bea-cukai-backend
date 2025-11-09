package transactionLogService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/transactionLogRepository"
	"math"
)

type TransactionLogService struct {
	repo *transactionLogRepository.TransactionLogRepository
}

func NewTransactionLogService(repo *transactionLogRepository.TransactionLogRepository) *TransactionLogService {
	return &TransactionLogService{
		repo: repo,
	}
}

// GetAll - get all transaction logs with filtering and pagination
func (s *TransactionLogService) GetAll(req model.TransactionLogRequest) (model.TransactionLogResponse, error) {
	// Set default pagination if not provided
	if req.Page < 0 {
		req.Page = 1
	}
	if req.Limit < 0 {
		req.Limit = 10
	}

	// No default date range - allow empty dates to get all data

	logs, total, err := s.repo.GetAll(req)
	if err != nil {
		return model.TransactionLogResponse{}, err
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	response := model.TransactionLogResponse{
		Data:        logs,
		Total:       int(total),
		Page:        req.Page,
		Limit:       req.Limit,
		TotalPages:  totalPages,
		HasNextPage: req.Page < totalPages,
		HasPrevPage: req.Page > 1,
	}

	return response, nil
}
