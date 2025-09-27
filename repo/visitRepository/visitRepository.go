package visitRepository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ===== DateRange helper (sama pola dengan repo lain) =====
type DateRange struct {
	From time.Time
	To   time.Time // inclusive; akan dinormalisasi ke [From 00:00, To+1d 00:00)
}

func (dr DateRange) norm() (time.Time, time.Time) {
	from := time.Date(dr.From.Year(), dr.From.Month(), dr.From.Day(), 0, 0, 0, 0, dr.From.Location())
	toExclusive := time.Date(dr.To.Year(), dr.To.Month(), dr.To.Day(), 0, 0, 0, 0, dr.To.Location()).Add(24 * time.Hour)
	return from, toExclusive
}

// ===== DTO untuk Dashboard =====

type VisitKPISummary struct {
	TotalVisits      int64   `json:"total_visits"`
	UniquePatients   int64   `json:"unique_patients"`
	InpatientVisits  int64   `json:"inpatient_visits"`
	OutpatientVisits int64   `json:"outpatient_visits"`
	AdmittedFromOP   int64   `json:"admitted_from_op"` // contoh: dari IGD/OP direncanakan masuk RI (tipe_discharge "2. Rawat Inap")
	Discharges       int64   `json:"discharges"`       // jumlah pulang RI pada periode (tgl_keluar in range)
	ALOSDays         float64 `json:"alos_days"`        // rata-rata LOS (hari) utk discharges pada periode
	MedianLOSDays    float64 `json:"median_los_days"`  // median LOS (hari) utk discharges pada periode
}

type VisitTrendPoint struct {
	Date       time.Time `json:"date" gorm:"column:date"`
	Visits     int64     `json:"visits"`
	Inpatient  int64     `json:"inpatient"`
	Outpatient int64     `json:"outpatient"`
	Admits     int64     `json:"admits"` // flag "2. Rawat Inap" di hari kunjungan
}

type TopService struct {
	TypeCode string `json:"type_code"`
	Name     string `json:"name"`
	Visits   int64  `json:"visits"`
}

type TopGuarantor struct {
	Name   string `json:"name"`
	Visits int64  `json:"visits"`
}

type ByDOW struct {
	DOW    int32  `json:"dow"`   // 0=Sun .. 6=Sat
	Label  string `json:"label"` // Sun..Sat
	Visits int64  `json:"visits"`
}

type ByRegion struct {
	Kota   string `json:"kota"`
	Visits int64  `json:"visits"`
}

type IPOPBreakdown struct {
	Ipop   string `json:"ipop"   gorm:"column:ipop"` // 'I' atau 'O'
	Visits int64  `json:"visits"`
}

type LOSBucket struct {
	Bucket string `json:"bucket"` // '<=0.5d','0.5-1d','1-2d','2-3d','3+d'
	Count  int64  `json:"count"`
}

// ===== Repository =====

type VisitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

// ===== helper SQL exprs =====

// IP/OP dari kolom "kelas"
const ipopExpr = `
CASE
  WHEN UPPER(COALESCE(k.kelas,'')) = 'RAWAT JALAN' THEN 'O'
  ELSE 'I'
END
`

// LOS dalam hari (dengan pecahan)
const losDaysExpr = `
EXTRACT(EPOCH FROM (k.tgl_keluar::timestamp - k.tgl_masuk::timestamp))/86400.0
`

// ===== KPI Summary =====
// Visits / Unique patients by tgl_masuk; Discharges & LOS untuk RI dengan tgl_keluar pada periode.
func (r *VisitRepository) GetKPISummary(ctx context.Context, dr DateRange) (VisitKPISummary, error) {
	from, to := dr.norm()
	var out VisitKPISummary

	// 1) Counts by tgl_masuk
	type countsRow struct {
		Total    int64
		Patients int64
		IP       int64
		OP       int64
		Admits   int64
	}
	var cr countsRow
	if err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			COUNT(*) AS total,
			COUNT(DISTINCT k.patid) AS patients,
			SUM(CASE WHEN `+ipopExpr+`='I' THEN 1 ELSE 0 END) AS ip,
			SUM(CASE WHEN `+ipopExpr+`='O' THEN 1 ELSE 0 END) AS op,
			SUM(CASE WHEN k.tipe_discharge ILIKE '2.%' THEN 1 ELSE 0 END) AS admits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Scan(&cr).Error; err != nil {
		return out, err
	}

	// 2) Discharges & LOS utk RI dengan tgl_keluar di periode
	type losRow struct {
		Discharges int64
		ALOS       float64
		Median     float64
	}
	var lr losRow
	if err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			COUNT(*) AS discharges,
			COALESCE(AVG(`+losDaysExpr+`), 0) AS alos,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY `+losDaysExpr+`), 0) AS median
		`).
		Where("k.tgl_keluar IS NOT NULL AND "+ipopExpr+"='I' AND k.tgl_keluar >= ? AND k.tgl_keluar < ?", from, to).
		Scan(&lr).Error; err != nil {
		return out, err
	}

	out = VisitKPISummary{
		TotalVisits:      cr.Total,
		UniquePatients:   cr.Patients,
		InpatientVisits:  cr.IP,
		OutpatientVisits: cr.OP,
		AdmittedFromOP:   cr.Admits,
		Discharges:       lr.Discharges,
		ALOSDays:         lr.ALOS,
		MedianLOSDays:    lr.Median,
	}
	return out, nil
}

// ===== Trend Harian (berdasarkan tgl_masuk) =====
func (r *VisitRepository) GetTrend(ctx context.Context, dr DateRange) ([]VisitTrendPoint, error) {
	from, to := dr.norm()
	rows := []VisitTrendPoint{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			DATE_TRUNC('day', k.tgl_masuk) AS date,
			COUNT(*) AS visits,
			SUM(CASE WHEN `+ipopExpr+`='I' THEN 1 ELSE 0 END) AS inpatient,
			SUM(CASE WHEN `+ipopExpr+`='O' THEN 1 ELSE 0 END) AS outpatient,
			SUM(CASE WHEN k.tipe_discharge ILIKE '2.%' THEN 1 ELSE 0 END) AS admits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("DATE_TRUNC('day', k.tgl_masuk)").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

// ===== Top Services (nama_admisi_layanan) =====
func (r *VisitRepository) GetTopServices(ctx context.Context, dr DateRange) ([]TopService, error) {
	from, to := dr.norm()
	rows := []TopService{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			COALESCE(k.tipe_admisi_layanan::text, '') AS type_code,
			COALESCE(NULLIF(BTRIM(k.nama_admisi_layanan), ''), 'UNKNOWN') AS name,
			COUNT(*) AS visits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("type_code, name").
		Order("visits DESC, name ASC").
		Scan(&rows).Error
	return rows, err
}

// ===== Top Guarantors (nama_penjamin) =====
func (r *VisitRepository) GetTopGuarantors(ctx context.Context, dr DateRange) ([]TopGuarantor, error) {
	from, to := dr.norm()
	rows := []TopGuarantor{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			COALESCE(NULLIF(BTRIM(k.nama_penjamin), ''), 'UNKNOWN') AS name,
			COUNT(*) AS visits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("name").
		Order("visits DESC, name ASC").
		Scan(&rows).Error
	return rows, err
}

// ===== Distribusi per Hari (DOW) =====
func (r *VisitRepository) GetByDOW(ctx context.Context, dr DateRange) ([]ByDOW, error) {
	from, to := dr.norm()
	rows := []ByDOW{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			CAST(EXTRACT(DOW FROM k.tgl_masuk)::int AS int) AS dow,
			CASE CAST(EXTRACT(DOW FROM k.tgl_masuk)::int AS int)
				WHEN 0 THEN 'Sun'
				WHEN 1 THEN 'Mon'
				WHEN 2 THEN 'Tue'
				WHEN 3 THEN 'Wed'
				WHEN 4 THEN 'Thu'
				WHEN 5 THEN 'Fri'
				WHEN 6 THEN 'Sat'
				ELSE 'NA'
			END AS label,
			COUNT(*) AS visits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("dow, label").
		Order("dow ASC").
		Scan(&rows).Error
	return rows, err
}

// ===== Top Region (kota) =====
func (r *VisitRepository) GetByRegionKota(ctx context.Context, dr DateRange, limit int) ([]ByRegion, error) {
	from, to := dr.norm()
	if limit <= 0 {
		limit = 10
	}
	rows := []ByRegion{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			COALESCE(NULLIF(BTRIM(k.kota), ''), 'UNKNOWN') AS kota,
			COUNT(*) AS visits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("kota").
		Order("visits DESC, kota ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ===== Mix IP/OP =====
func (r *VisitRepository) GetMixIPOP(ctx context.Context, dr DateRange) ([]IPOPBreakdown, error) {
	from, to := dr.norm()
	rows := []IPOPBreakdown{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			CASE
				WHEN UPPER(COALESCE(k.kelas,'')) = 'RAWAT JALAN' THEN 'O'
				ELSE 'I'
			END AS ipop,
			COUNT(*) AS visits
		`).
		Where("k.tgl_masuk >= ? AND k.tgl_masuk < ?", from, to).
		Group("ipop").
		Order("visits DESC").
		Scan(&rows).Error
	return rows, err
}

// ===== Histogram LOS (RI saja; tgl_keluar in range) =====
func (r *VisitRepository) GetLOSBuckets(ctx context.Context, dr DateRange) ([]LOSBucket, error) {
	from, to := dr.norm()
	rows := []LOSBucket{}
	err := r.db.WithContext(ctx).
		Table("kunjungan k").
		Select(`
			CASE
				WHEN `+losDaysExpr+` <= 0.5 THEN '<=0.5d'
				WHEN `+losDaysExpr+` <= 1   THEN '0.5-1d'
				WHEN `+losDaysExpr+` <= 2   THEN '1-2d'
				WHEN `+losDaysExpr+` <= 3   THEN '2-3d'
				ELSE '3+d'
			END AS bucket,
			COUNT(*) AS count
		`).
		Where("k.tgl_keluar IS NOT NULL AND "+ipopExpr+"='I' AND k.tgl_keluar >= ? AND k.tgl_keluar < ?", from, to).
		Group("bucket").
		Scan(&rows).Error
	return rows, err
}
