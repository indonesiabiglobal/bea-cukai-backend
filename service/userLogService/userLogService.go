package userLogService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/userLogRepository"

	"github.com/jinzhu/copier"
)

type UserLogService struct {
	userLogRepo *userLogRepository.UserLogRepository
}

func NewUserLogService(userLogRepository *userLogRepository.UserLogRepository) *UserLogService {
	return &UserLogService{
		userLogRepo: userLogRepository,
	}
}

// CreateLog - create new user log entry
func (s *UserLogService) CreateLog(logRequest model.UserLogRequest) (model.UserLogResponse, error) {
	// Call repository to save log
	createdLog, err := s.userLogRepo.CreateLog(logRequest)
	if err != nil {
		return model.UserLogResponse{}, err
	}

	var logResponse model.UserLogResponse
	err = copier.Copy(&logResponse, &createdLog)
	if err != nil {
		return model.UserLogResponse{}, err
	}

	return logResponse, nil
}

// GetAll - get all user logs with filtering and pagination
func (s *UserLogService) GetAll(req model.UserLogListRequest) ([]model.UserLogResponse, int64, map[string]interface{}, error) {
	// Set defaults for pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// Get logs from repository
	logs, total, err := s.userLogRepo.GetAll(req)
	if err != nil {
		return nil, 0, nil, err
	}

	// Convert to response format
	var logResponses []model.UserLogResponse
	for _, log := range logs {
		logResponse := model.UserLogResponse{
			Id:        log.Id,
			UserId:    log.UserId,
			Username:  log.Username,
			Action:    log.Action,
			IpAddress: log.IpAddress,
			UserAgent: log.UserAgent,
			Status:    log.Status,
			Message:   log.Message,
			CreatedAt: log.CreatedAt,
		}
		logResponses = append(logResponses, logResponse)
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	// Prepare metadata
	meta := map[string]interface{}{
		"page":        req.Page,
		"limit":       req.Limit,
		"total_count": total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}

	return logResponses, total, meta, nil
}

// GetByUserId - get user logs by user id
func (s *UserLogService) GetByUserId(userId string, limit int) ([]model.UserLogResponse, error) {
	logs, err := s.userLogRepo.GetByUserId(userId, limit)
	if err != nil {
		return nil, err
	}

	var logResponses []model.UserLogResponse
	for _, log := range logs {
		logResponse := model.UserLogResponse{
			Id:        log.Id,
			UserId:    log.UserId,
			Username:  log.Username,
			Action:    log.Action,
			IpAddress: log.IpAddress,
			UserAgent: log.UserAgent,
			Status:    log.Status,
			Message:   log.Message,
			CreatedAt: log.CreatedAt,
		}
		logResponses = append(logResponses, logResponse)
	}

	return logResponses, nil
}

// DeleteOldLogs - delete logs older than specified days
func (s *UserLogService) DeleteOldLogs(days int) error {
	return s.userLogRepo.DeleteOldLogs(days)
}
