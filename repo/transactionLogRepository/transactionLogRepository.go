package transactionLogRepository

import (
	"Bea-Cukai/model"

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

// buildQuery returns the ms_item created/updated activity log query; filters are bound parameters.
func buildQuery(req model.TransactionLogRequest) (string, []any) {
	createdWhere, updatedWhere, where := "TRUE", "TRUE", "TRUE"
	var createdArgs, updatedArgs, whereArgs []any

	// Date filter - only add if both dates are provided
	if req.StartDate != "" && req.EndDate != "" {
		createdWhere = "a.created_date::date BETWEEN ? AND ?"
		updatedWhere = "a.updated_date::date BETWEEN ? AND ?"
		createdArgs = []any{req.StartDate, req.EndDate}
		updatedArgs = []any{req.StartDate, req.EndDate}
	}
	if req.UserName != "" {
		where = "user_name ILIKE ?"
		whereArgs = []any{"%" + req.UserName + "%"}
	}

	query := `
		WITH transaction_logs AS (
			SELECT a.created_date trans_date,
				COALESCE(a.created_by, '') user_name,
				CONCAT('Mst Item ', a.item_group) module,
				a.item_code action_code, 'Created Master' activity_log
			FROM ms_item a
			WHERE ` + createdWhere + `

			UNION ALL

			SELECT a.updated_date trans_date,
				COALESCE(a.updated_by, '') user_name,
				CONCAT('Mst Item ', a.item_group) module,
				a.item_code action_code, 'Updated Master' activity_log
			FROM ms_item a
			WHERE ` + updatedWhere + ` AND a.updated_date != a.created_date
		)
		SELECT * FROM transaction_logs
		WHERE ` + where + `
		ORDER BY trans_date DESC NULLS LAST` // NULLS LAST = MySQL's DESC order

	args := append(append(createdArgs, updatedArgs...), whereArgs...)
	return query, args
}

// GetAll - get all transaction logs with filtering and pagination
func (r *TransactionLogRepository) GetAll(req model.TransactionLogRequest) ([]model.TransactionLog, int64, error) {
	var logs []model.TransactionLog
	var total int64

	query, args := buildQuery(req)

	err := r.db.Raw("SELECT COUNT(*) FROM ("+query+"\n) AS total_count", args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Page > 0 && req.Limit > 0 {
		query += "\nLIMIT ? OFFSET ?"
		args = append(args, req.Limit, (req.Page-1)*req.Limit)
	}

	err = r.db.Raw(query, args...).Scan(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
