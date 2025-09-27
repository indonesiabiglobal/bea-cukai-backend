package incomeRepository

import (
	"context"
	"fmt"
	"time"

	"Dashboard-TRDP/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ---- Constructor ----

type IncomeRepository struct {
	db *gorm.DB
}

func NewIncomeRepository(db *gorm.DB) *IncomeRepository { return &IncomeRepository{db: db} }

// ---- Helpers ----

type DateRange struct{ From, To time.Time } // inclusive range by [From, To]

func (r DateRange) norm() (time.Time, time.Time) {
	from := time.Date(r.From.Year(), r.From.Month(), r.From.Day(), 0, 0, 0, 0, r.From.Location())
	// make To inclusive end-of-day
	toEnd := time.Date(r.To.Year(), r.To.Month(), r.To.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), r.To.Location())
	return from, toEnd
}

// ---- DTOs for dashboard results ----

type RevenuePoint struct {
	Date   time.Time       `json:"date" gorm:"column:d"`
	Debit  decimal.Decimal `json:"debit"`
	Credit decimal.Decimal `json:"credit"`
	Net    decimal.Decimal `json:"net"`
}

type TopKeyAmount struct {
	Code   string          `json:"code"`
	Name   string          `json:"name"`
	Debit  decimal.Decimal `json:"debit"`
	Credit decimal.Decimal `json:"credit"`
	Net    decimal.Decimal `json:"net"`
}

type MixIPOP struct {
	Ipop   string          `json:"ipop"`
	Debit  decimal.Decimal `json:"debit"`
	Credit decimal.Decimal `json:"credit"`
	Net    decimal.Decimal `json:"net"`
}

type KPISummary struct {
	TotalDebit     decimal.Decimal `json:"total_debit"`
	TotalCredit    decimal.Decimal `json:"total_credit"`
	Net            decimal.Decimal `json:"net"`
	UniquePatients int64           `json:"unique_patients"`
	UniqueEpisodes int64           `json:"unique_episodes"`
	TxCount        int64           `json:"tx_count"`
}

type DOWPoint struct {
	DOW int             `json:"dow"` // 0=Sunday .. 6=Saturday (Postgres EXTRACT(DOW))
	Net decimal.Decimal `json:"net"`
}

// ---- Core aggregations ----

// GetKPISummary computes totals and distinct counts.
func (c *IncomeRepository) GetKPISummary(ctx context.Context, dr DateRange) (KPISummary, error) {
	from, to := dr.norm()
	var out KPISummary
	type row struct {
		TotalDebit  decimal.Decimal
		TotalCredit decimal.Decimal
		Net         decimal.Decimal
		Patients    int64
		Episodes    int64
		TxCount     int64
	}
	var result row
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("COALESCE(SUM(debit),0) as total_debit, COALESCE(SUM(credit),0) as total_credit, COALESCE(SUM(debit - credit),0) as net, COUNT(DISTINCT patid) as patients, COUNT(DISTINCT episode) as episodes, COUNT(*) as tx_count").
		Scan(&result).Error
	if err != nil {
		return out, err
	}
	out = KPISummary{TotalDebit: result.TotalDebit, TotalCredit: result.TotalCredit, Net: result.Net, UniquePatients: result.Patients, UniqueEpisodes: result.Episodes, TxCount: result.TxCount}
	return out, nil
}

// GetRevenueTrend returns daily sums over a date range.
func (c *IncomeRepository) GetRevenueTrend(ctx context.Context, dr DateRange) ([]RevenuePoint, error) {
	from, to := dr.norm()
	rows := []RevenuePoint{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("date_trunc('day', tgl_transaksi) as d, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("d").
		Order("d").
		Scan(&rows).Error
	return rows, err
}

// GetTopUnits returns top N units by Net within date range.
func (c *IncomeRepository) GetTopUnits(ctx context.Context, dr DateRange, limit int) ([]TopKeyAmount, error) {
	from, to := dr.norm()
	rows := []TopKeyAmount{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("kode_unit as code, COALESCE(nama_unit, kode_unit) as name, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("code, name").
		Order("net DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetTopProviders returns top N provider names (provname) by Net.
func (c *IncomeRepository) GetTopProviders(ctx context.Context, dr DateRange, limit int) ([]TopKeyAmount, error) {
	from, to := dr.norm()
	rows := []TopKeyAmount{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("provid as code, COALESCE(provname, provid) as name, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("code, name").
		Order("net DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetTopGuarantors returns top N guarantors (penjamin) by Net.
func (c *IncomeRepository) GetTopGuarantors(ctx context.Context, dr DateRange, limit int) ([]TopKeyAmount, error) {
	from, to := dr.norm()
	rows := []TopKeyAmount{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("kode_penjamin as code, COALESCE(nama_penjamin, kode_penjamin) as name, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("code, name").
		Order("net DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetTopGuarantorGroups returns top N guarantor groups by Net.
func (c *IncomeRepository) GetTopGuarantorGroups(ctx context.Context, dr DateRange, limit int) ([]TopKeyAmount, error) {
	from, to := dr.norm()
	rows := []TopKeyAmount{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("kode_kelompok_penjamin as code, COALESCE(nama_kelompok_penjamin, kode_kelompok_penjamin) as name, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("code, name").
		Order("net DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetMixIPOP returns net revenue split by IPOP (e.g., IP/OP).
func (c *IncomeRepository) GetRevenueByIPOP(ctx context.Context, dr DateRange) ([]MixIPOP, error) {
	from, to := dr.norm()
	rows := []MixIPOP{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Select("COALESCE(ipop,'') as ipop, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Group("ipop").
		Order("net DESC").
		Scan(&rows).Error

	fmt.Println(rows)
	return rows, err
}

// GetRevenueByService returns top N services by Net.
func (c *IncomeRepository) GetRevenueByService(ctx context.Context, dr DateRange) ([]TopKeyAmount, error) {
	from, to := dr.norm()
	rows := []TopKeyAmount{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("kode_layanan as code, kode_layanan as name, SUM(debit) as debit, SUM(credit) as credit, SUM(debit - credit) as net").
		Group("code, name").
		Order("net DESC").
		Scan(&rows).Error
	return rows, err
}

// GetRevenueByDOW returns net revenue grouped by day-of-week (0=Sun ... 6=Sat).
func (c *IncomeRepository) GetRevenueByDOW(ctx context.Context, dr DateRange) ([]DOWPoint, error) {
	from, to := dr.norm()
	rows := []DOWPoint{}
	err := c.db.WithContext(ctx).
		Model(&model.Income{}).
		Where("tgl_transaksi BETWEEN ? AND ?", from, to).
		Select("CAST(EXTRACT(DOW FROM tgl_transaksi) AS INT) as dow, SUM(debit - credit) as net").
		Group("dow").
		Order("dow").
		Scan(&rows).Error
	return rows, err
}

// OPTIONAL: Raw list endpoints (adapt to your model types)
// NOTE: Your original template referenced model.IncomeGetModel & Preload("User").
//       The pendapatan table has no user relation, so below we return plain model.Income.

func (c *IncomeRepository) GetAllIncomes(ctx context.Context, dr *DateRange) ([]model.Income, error) {
	q := c.db.WithContext(ctx).Model(&model.Income{})
	if dr != nil {
		from, to := dr.norm()
		q = q.Where("tgl_transaksi BETWEEN ? AND ?", from, to)
	}
	var rows []model.Income
	if err := q.Order("tgl_transaksi DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *IncomeRepository) GetIncomeByID(ctx context.Context, id uint) (model.Income, error) {
	var row model.Income
	err := c.db.WithContext(ctx).First(&row, id).Error
	return row, err
}
