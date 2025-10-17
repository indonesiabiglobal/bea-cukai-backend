package entryProductRepository

import (
	"Bea-Cukai/model"
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ---- Constructor ----

type EntryProductRepository struct {
	db *gorm.DB
}

func NewEntryProductRepository(db *gorm.DB) *EntryProductRepository {
	return &EntryProductRepository{db: db}
}

// ---- DTOs for report results ----
type KPISummary struct {
	TotalRcvQty    decimal.Decimal `json:"total_rcv_qty"`
	TotalNetPrice  decimal.Decimal `json:"total_net_price"`
	TotalNetAmount decimal.Decimal `json:"total_net_amount"`
	UniqueVendors  int64           `json:"unique_vendors"`
	UniqueItems    int64           `json:"unique_items"`
	TxCount        int64           `json:"tx_count"`
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
}

// GetReport retrieves entry products with filters and pagination
func (c *EntryProductRepository) GetReport(ctx context.Context, filter GetReportFilter) ([]model.EntryProduct, int64, error) {
	from, to := filter.From, filter.To

	query := c.db.WithContext(ctx).Model(&model.EntryProduct{}).
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

	var results []model.EntryProduct
	err = query.Order("tgl_pabean DESC").Find(&results).Error
	return results, totalCount, err
}
