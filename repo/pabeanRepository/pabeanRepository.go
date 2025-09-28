package pabeanRepository

import (
	"Bea-Cukai/model"
	"context"

	"gorm.io/gorm"
)

// ---- Constructor ----

type PabeanRepository struct {
	db *gorm.DB
}

func NewPabeanRepository(db *gorm.DB) *PabeanRepository {
	return &PabeanRepository{db: db}
}

// ---- Core operations ----

// GetAll retrieves all pabean documents
func (r *PabeanRepository) GetAll(ctx context.Context) ([]model.MsPabean, error) {
	var results []model.MsPabean
	err := r.db.WithContext(ctx).
		Find(&results).Error
	return results, err
}
