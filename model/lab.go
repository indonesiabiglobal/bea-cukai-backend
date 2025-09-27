package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Lab maps to the physical table "lab" while keeping a clean domain model name.
// No gorm.Model embedded to avoid conflicting with the existing `id` column.

type Lab struct {
	UID   string          `json:"uid" gorm:"size:64;index:idx_lab_uid"`
	Tgl   time.Time       `json:"tgl" gorm:"type:date;index:idx_lab_tgl"`
	LabID string          `json:"id" gorm:"column:id;size:64;index:idx_lab_id"`
	Nama  string          `json:"nama" gorm:"size:128"`
	Total decimal.Decimal `json:"total" gorm:"type:decimal(20,2);not null;default:0"`
}

// TableName enforces the DB table name.
func (Lab) TableName() string { return "lab" }

// ==========================
// DTOs
// ==========================

// LabRequest defines the payload for create/update operations.
// `tgl` required; `id` refers to the external lab item identifier.

type LabRequest struct {
	UID   string          `json:"uid" validate:"required"`
	Tgl   time.Time       `json:"tgl" validate:"required"`
	LabID string          `json:"id" validate:"required"`
	Nama  string          `json:"nama"`
	Total decimal.Decimal `json:"total"`
}

// LabResponse is returned to clients.

type LabResponse struct {
	UID   string          `json:"uid"`
	Tgl   time.Time       `json:"tgl"`
	LabID string          `json:"id"`
	Nama  string          `json:"nama"`
	Total decimal.Decimal `json:"total"`
}

// Notes:
// - Consider setting a composite primary key if the source ensures uniqueness across (uid, id, tgl):
//   add `gorm:"primaryKey"` to those fields. Otherwise keep plain indexes and use a surrogate key in a wrapper table.
// - If you need timestamps (created_at/updated_at), add fields explicitly since we don't embed gorm.Model here.
