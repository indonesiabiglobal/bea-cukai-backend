package model

import "time"

type UserLog struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	UserId    string    `json:"user_id" gorm:"column:user_id;not null"`
	Username  string    `json:"username" gorm:"column:username;not null"`
	Action    string    `json:"action" gorm:"column:action;not null"` // login, logout, create, update, delete
	IpAddress string    `json:"ip_address" gorm:"column:ip_address"`
	UserAgent string    `json:"user_agent" gorm:"column:user_agent"`
	Status    string    `json:"status" gorm:"column:status"` // success, failed
	Message   string    `json:"message" gorm:"column:message"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName specifies the table name for GORM
func (UserLog) TableName() string {
	return "user_log"
}

type UserLogRequest struct {
	UserId    string `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action" validate:"required"`
	IpAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Status    string `json:"status" validate:"required"`
	Message   string `json:"message"`
}

type UserLogResponse struct {
	Id        int       `json:"id"`
	UserId    string    `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	IpAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type UserLogListRequest struct {
	UserId    string `json:"user_id" form:"user_id"`
	Username  string `json:"username" form:"username"`
	Action    string `json:"action" form:"action"`
	Status    string `json:"status" form:"status"`
	StartDate string `json:"start_date" form:"start_date"` // format: 2006-01-02
	EndDate   string `json:"end_date" form:"end_date"`     // format: 2006-01-02
	Page      int    `json:"page" form:"page"`
	Limit     int    `json:"limit" form:"limit"`
}
