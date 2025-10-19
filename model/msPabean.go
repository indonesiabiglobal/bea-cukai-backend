package model

import (
	"time"
)

type MsPabean struct {
	PabeanCode  string    `json:"pabean_code" gorm:"column:dept_code;type:varchar(255);not null"`
	PabeanName  string    `json:"pabean_name" gorm:"column:dept_name;type:varchar(255);not null"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(255)"`
	CreatedDate time.Time `json:"created_date" gorm:"type:datetime"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:varchar(255)"`
	UpdatedDate time.Time `json:"updated_date" gorm:"type:datetime"`
}

// TableName overrides the default table name.
func (MsPabean) TableName() string { return "ms_pabean_doc" }

type MsPabeanRequest struct {
	PabeanCode string `json:"pabean_code" validate:"required"`
	PabeanName string `json:"pabean_name" validate:"required"`
	Notes      string `json:"notes"`
	CreatedBy  string `json:"created_by"`
	UpdatedBy  string `json:"updated_by"`
}

type MsPabeanResponse struct {
	PabeanCode  string    `json:"pabean_code"`
	PabeanName  string    `json:"pabean_name"`
	Notes       string    `json:"notes"`
	CreatedBy   string    `json:"created_by"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedDate time.Time `json:"updated_date"`
}
