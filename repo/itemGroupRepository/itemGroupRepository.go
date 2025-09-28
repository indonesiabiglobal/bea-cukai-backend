package itemGroupRepository

import (
	"Bea-Cukai/model"
	"context"

	"gorm.io/gorm"
)

// ---- Constructor ----

type ItemGroupRepository struct {
	db *gorm.DB
}

func NewItemGroupRepository(db *gorm.DB) *ItemGroupRepository {
	return &ItemGroupRepository{db: db}
}

// ---- Core operations ----

// GetAll retrieves all item groups
func (r *ItemGroupRepository) GetAll(ctx context.Context) ([]model.ItemGroup, error) {
	var results []model.ItemGroup
	err := r.db.WithContext(ctx).
		Order("idx ASC").
		Find(&results).Error
	return results, err
}
