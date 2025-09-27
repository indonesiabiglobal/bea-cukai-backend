// service/saleService/sale_service.go
package saleService

import (
	"context"

	"Dashboard-TRDP/model"
	"Dashboard-TRDP/repo/saleRepository"
)

// SaleService adalah lapisan bisnis untuk penjualan (sales)
type SaleService struct {
	repo *saleRepository.SaleRepository
}

// NewSaleService membuat instance SaleService
func NewSaleService(repo *saleRepository.SaleRepository) *SaleService {
	return &SaleService{repo: repo}
}

// GetKPISummary mengembalikan ringkasan KPI penjualan untuk rentang tanggal
func (s *SaleService) GetKPISummary(dto model.SaleRequestParam) (saleRepository.SalesKPISummary, error) {
	return s.repo.GetKPISummary(context.Background(), dto)
}

// GetTrend mengembalikan tren harian Qty & Subtotal
func (s *SaleService) GetTrend(dto model.SaleRequestParam) ([]saleRepository.SalesTrendPoint, error) {
	return s.repo.GetTrend(context.Background(), dto)
}

// GetTopProducts mengembalikan produk teratas berdasarkan subtotal (desc)
// limit <= 0 akan di-normalisasi ke 10
func (s *SaleService) GetTopProducts(dto model.SaleRequestParam) ([]saleRepository.SalesTopProduct, error) {
	return s.repo.GetTopProducts(context.Background(), dto)
}

// GetByCategory mengembalikan agregasi subtotal & qty per kategori
// limit <= 0 akan di-normalisasi ke 10
func (s *SaleService) GetByCategory(dto model.SaleRequestParam, limit int) ([]saleRepository.SalesByCategory, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetByCategory(context.Background(), dto, limit)
}
