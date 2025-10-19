package entryProductRepository

import (
	"Bea-Cukai/model"
	"context"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type EntryProductRepository struct {
	db *gorm.DB
}

func NewEntryProductRepository(db *gorm.DB) *EntryProductRepository {
	return &EntryProductRepository{db: db}
}

// ---- Core aggregations ----

// Filter struct for GetReport method
type GetReportFilter struct {
	From         time.Time
	To           time.Time
	PabeanType   string
	ProductGroup string
	NoPabean     string
	ProductCode  string
	ProductName  string
	Page         int
	Limit        int
	IsExport     bool
}

// GetReport retrieves entry products with filters and pagination
func (c *EntryProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.EntryProduct, int64, error) {
	from, to := filter.From, filter.To

	query := c.db.WithContext(ctx).Model(&model.EntryProduct{}).
		Where("tgl_pabean BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02"))

	// Apply filters if provided
	if filter.PabeanType != "" {
		query = query.Where("jenis_pabean = ?", filter.PabeanType)
	}
	if filter.NoPabean != "" {
		query = query.Where("no_pabean = ?", filter.NoPabean)
	}
	if filter.ProductCode != "" {
		query = query.Where("item_code = ?", filter.ProductCode)
	}
	if filter.ProductName != "" {
		query = query.Where("item_name LIKE ?", "%"+filter.ProductName+"%")
	}
	// Note: ProductGroup filter might need adjustment based on your data structure
	// since the current model doesn't have a product_group field directly

	// Get total count before applying pagination
	var totalCount int64
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination only if both page and limit are provided (> 0)
	if filter.Limit > 0 && filter.Page > 0 {
		query = query.Limit(filter.Limit)
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset)
	}

	var results []model.EntryProduct
	if filter.IsExport {
		query = query.Order("tgl_pabean ASC")
	} else {
		query = query.Order("tgl_pabean DESC")
	}
	err = query.Find(&results).Error
	return results, totalCount, err
}
