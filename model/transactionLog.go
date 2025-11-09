package model

import "time"

type TransactionLog struct {
	TransDate   time.Time `json:"trans_date" gorm:"column:trans_date"`
	UserName    string    `json:"user_name" gorm:"column:user_name"`
	Module      string    `json:"module" gorm:"column:module"`
	ActionCode  string    `json:"action_code" gorm:"column:action_code"`
	ActivityLog string    `json:"activity_log" gorm:"column:activity_log"`
}

// TableName is not needed as this is a view/query result

type TransactionLogRequest struct {
	StartDate string `json:"start_date" form:"start_date"` // format: 2006-01-02, optional
	EndDate   string `json:"end_date" form:"end_date"`     // format: 2006-01-02, optional
	UserName  string `json:"user_name" form:"user_name"`   // optional filter
	Page      int    `json:"page" form:"page"`
	Limit     int    `json:"limit" form:"limit"`
}

type TransactionLogResponse struct {
	Data        []TransactionLog `json:"data"`
	Total       int              `json:"total"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
	TotalPages  int              `json:"total_pages"`
	HasNextPage bool             `json:"has_next_page"`
	HasPrevPage bool             `json:"has_prev_page"`
}
