package transactionLogRepository

import (
	"Bea-Cukai/model"
	"fmt"

	"gorm.io/gorm"
)

type TransactionLogRepository struct {
	db *gorm.DB
}

func NewTransactionLogRepository(db *gorm.DB) *TransactionLogRepository {
	return &TransactionLogRepository{
		db: db,
	}
}

// GetAll - get all transaction logs with filtering and pagination
func (r *TransactionLogRepository) GetAll(req model.TransactionLogRequest) ([]model.TransactionLog, int64, error) {
	var logs []model.TransactionLog
	var total int64

	// Build the complex query combining multiple tables
	whereClause := "1=1"
	if req.UserName != "" {
		whereClause = fmt.Sprintf("1=1 AND user_name LIKE '%%%s%%' ", req.UserName)
	}

	// Date filter - only add if both dates are provided
	var dateWhereClause1, dateWhereClause2 string
	if req.StartDate != "" && req.EndDate != "" {
		dateWhereClause1 = fmt.Sprintf("DATE(a.created_date) BETWEEN '%s' AND '%s'", req.StartDate, req.EndDate)
		dateWhereClause2 = fmt.Sprintf("DATE(a.updated_date) BETWEEN '%s' AND '%s'", req.StartDate, req.EndDate)
	} else {
		dateWhereClause1 = "1=1"
		dateWhereClause2 = "1=1"
	}

	// Complex UNION query combining all transaction tables
	query := fmt.Sprintf(`
		WITH transaction_logs AS (
			SELECT COALESCE(a.created_date, NULL) trans_date, 
				COALESCE(a.created_by, '') user_name, 
				CONCAT('Mst Item ', a.item_group) module, 
				a.item_code action_code, 'Created Master' activity_log  
			FROM ms_item a 
			WHERE %s
			
			UNION ALL
			
			SELECT COALESCE(a.updated_date, NULL) trans_date, 
				COALESCE(a.updated_by, '') user_name, 
				CONCAT('Mst Item ', a.item_group) module, 
				a.item_code action_code, 'Updated Master' activity_log  
			FROM ms_item a 
			WHERE %s AND a.updated_date != a.created_date
		)
		SELECT * FROM transaction_logs 
		WHERE %s
		ORDER BY trans_date DESC
	`,
		dateWhereClause1,
		dateWhereClause2,
		whereClause,
	)

	// Count total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			%s
		) AS total_count
	`, query)

	err := r.db.Raw(countQuery).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, req.Limit, offset)
	}

	// Execute query
	err = r.db.Raw(query).Scan(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
