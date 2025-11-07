package userLogRepository

import (
	"Bea-Cukai/model"
	"time"

	"gorm.io/gorm"
)

type UserLogRepository struct {
	db *gorm.DB
}

func NewUserLogRepository(db *gorm.DB) *UserLogRepository {
	return &UserLogRepository{
		db: db,
	}
}

// CreateLog - create new user log entry
func (r *UserLogRepository) CreateLog(logRequest model.UserLogRequest) (model.UserLog, error) {
	logModel := model.UserLog{
		UserId:    logRequest.UserId,
		Username:  logRequest.Username,
		Action:    logRequest.Action,
		IpAddress: logRequest.IpAddress,
		UserAgent: logRequest.UserAgent,
		Status:    logRequest.Status,
		Message:   logRequest.Message,
	}

	err := r.db.Create(&logModel).Error
	if err != nil {
		return model.UserLog{}, err
	}

	return logModel, nil
}

// GetAll - get all user logs with filtering and pagination
func (r *UserLogRepository) GetAll(req model.UserLogListRequest) ([]model.UserLog, int64, error) {
	var logs []model.UserLog
	var total int64

	// Build the base query
	query := r.db.Model(&model.UserLog{})

	// Apply filters
	if req.UserId != "" {
		query = query.Where("user_id = ?", req.UserId)
	}
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Date range filter
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startDate)
		}
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// Add 1 day to include the end date
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endDate)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Execute query with ordering (latest first)
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserId - get user logs by user id
func (r *UserLogRepository) GetByUserId(userId string, limit int) ([]model.UserLog, error) {
	var logs []model.UserLog

	query := r.db.Where("user_id = ?", userId).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// DeleteOldLogs - delete logs older than specified days
func (r *UserLogRepository) DeleteOldLogs(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := r.db.Where("created_at < ?", cutoffDate).Delete(&model.UserLog{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
