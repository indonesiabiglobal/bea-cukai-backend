package expenditureProductRepository

import (
	"Bea-Cukai/model"
	"context"
	"time"

	"gorm.io/gorm"
)

// ---- Constructor ----

type ExpenditureProductRepository struct {
	db *gorm.DB
}

func NewExpenditureProductRepository(db *gorm.DB) *ExpenditureProductRepository {
	return &ExpenditureProductRepository{db: db}
}

// ---- Helpers ----

type DateRange struct{ From, To time.Time } // inclusive range by [From, To]

func (r DateRange) norm() (time.Time, time.Time) {
	from := time.Date(r.From.Year(), r.From.Month(), r.From.Day(), 0, 0, 0, 0, r.From.Location())
	// make To inclusive end-of-day
	toEnd := time.Date(r.To.Year(), r.To.Month(), r.To.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), r.To.Location())
	return from, toEnd
}

// ---- Core operations ----

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
}

// GetReport retrieves expenditure products with filters and pagination
func (c *ExpenditureProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.ExpenditureProduct, int64, error) {
	from, to := DateRange{From: filter.From, To: filter.To}.norm()

	query := c.db.WithContext(ctx).Model(&model.ExpenditureProduct{}).
		Where("trans_date BETWEEN ? AND ?", from, to)

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

	var results []model.ExpenditureProduct
	err = query.Find(&results).Error
	return results, totalCount, err
}
