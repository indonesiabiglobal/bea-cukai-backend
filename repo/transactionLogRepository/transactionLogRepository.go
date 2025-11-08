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
	whereClause := ""
	if req.UserName != "" {
		whereClause = fmt.Sprintf(" AND user_name LIKE '%%%s%%' ", req.UserName)
	}

	betweenClause := fmt.Sprintf(" BETWEEN '%s' AND '%s'", req.StartDate, req.EndDate)

	// Complex UNION query combining all transaction tables
	query := fmt.Sprintf(`
		WITH transaction_logs AS (
			-- Stock Opname
			SELECT a.created_date trans_date, a.created_by user_name, 
				CONCAT('STOCK OPNAME ', a.item_group) module, 
				a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_opname_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				CONCAT('STOCK OPNAME ', a.item_group) module, 
				a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_opname_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Inv Move In
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Inv MOVE IN' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_movein_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Inv MOVE IN' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_movein_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Inv Move Out
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Inv MOVE Out' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_moveout_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'inv MOVE Out' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_moveout_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Master Item
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				CONCAT('Mst Item ', a.item_group) module, 
				a.item_code action_code, 'Created Master' activity_log  
			FROM ms_item a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				CONCAT('Mst Item ', a.item_group) module, 
				a.item_code action_code, 'Updated Master' activity_log  
			FROM ms_item a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Inv Material Harian
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Inv Mat Harian' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_material_harian_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Inv Mat Harian' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_material_harian_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Purchase Incoming
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Pch Incoming' module, a.ref_no action_code, 'Created Transaction' activity_log  
			FROM tr_ap_inv_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Pch Incoming' module, a.ref_no action_code, 'Updated Transaction' activity_log  
			FROM tr_ap_inv_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Inv Adjust
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Inv Adjust' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_adjust_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Inv Adjust' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_adjust_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Inv Produk Harian
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Inv Prod Harian' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_inv_produk_harian_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Inv Prod Harian' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_inv_produk_harian_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Produk In
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Prod. In' module, a.trans_no action_code, 'Created Transaction' activity_log  
			FROM tr_produk_in_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Prod. In' module, a.trans_no action_code, 'Updated Transaction' activity_log  
			FROM tr_produk_in_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
			
			-- Sales Delivery
			UNION ALL
			SELECT a.created_date trans_date, a.created_by user_name, 
				'Sls Dlv' module, a.spe_no action_code, 'Created Transaction' activity_log  
			FROM tr_export_head a 
			WHERE DATE(a.created_date) %s
			
			UNION ALL
			SELECT a.updated_date trans_date, a.updated_by user_name, 
				'Sls Dlv' module, a.spe_no action_code, 'Updated Transaction' activity_log  
			FROM tr_export_head a 
			WHERE DATE(a.updated_date) %s AND a.updated_date != a.created_date
		)
		SELECT * FROM transaction_logs 
		WHERE 1=1 %s
		ORDER BY trans_date DESC
	`,
		betweenClause, betweenClause, // Stock Opname
		betweenClause, betweenClause, // Inv Move In
		betweenClause, betweenClause, // Inv Move Out
		betweenClause, betweenClause, // Master Item
		betweenClause, betweenClause, // Inv Material Harian
		betweenClause, betweenClause, // Purchase Incoming
		betweenClause, betweenClause, // Inv Adjust
		betweenClause, betweenClause, // Inv Produk Harian
		betweenClause, betweenClause, // Produk In
		betweenClause, betweenClause, // Sales Delivery
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
