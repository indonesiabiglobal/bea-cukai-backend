package patientRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

/* ====== DTOs ====== */

type InpatientMonitoringRow struct {
	PatientID            string          `json:"patient_id"`
	PatientName          string          `json:"patient_name"`
	Episode              int64           `json:"episode"`
	AdmissionServiceType string          `json:"admission_service_type"`
	Diagnosis            string          `json:"diagnosis"`
	AdmissionReason      string          `json:"admission_reason"`
	DPJP                 string          `json:"dpjp"`
	NamaPenjamin         string          `json:"nama_penjamin"`
	StartDate            time.Time       `json:"start_date"`
	LOSDays              int64           `json:"los_days"`
	LOSHoursExact        float64         `json:"los_hours_exact"`
	TotalDebit           decimal.Decimal `json:"total_debit"`
	TotalCredit          decimal.Decimal `json:"total_credit"`
}

type DischargedRow struct {
	PatientID            string          `json:"patient_id"`
	PatientName          string          `json:"patient_name"`
	Episode              int64           `json:"episode"`
	AdmissionServiceType string          `json:"admission_service_type"`
	Diagnosis            string          `json:"diagnosis"`
	AdmissionReason      string          `json:"admission_reason"`
	DPJP                 string          `json:"dpjp"`
	NamaPenjamin         string          `json:"nama_penjamin"`
	StartDate            time.Time       `json:"start_date"`
	EndDate              time.Time       `json:"end_date"`
	LOSDays              int64           `json:"los_days"`
	LOSHoursExact        float64         `json:"los_hours_exact"`
	TotalDebit           decimal.Decimal `json:"total_debit"`
	TotalCredit          decimal.Decimal `json:"total_credit"`
}

type DischargedSummary struct {
	Total       int64           `json:"total"`
	AverageLos  float64         `json:"average_los"`
	MedianLos   float64         `json:"median_los"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
}

/* ====== DateRange helper (inklusif kalendar) ====== */

type DateRange struct {
	From time.Time
	To   time.Time // inclusive (hari terakhir)
}

func (dr DateRange) norm() (time.Time, time.Time) {
	loc := dr.From.Location()
	from := time.Date(dr.From.Year(), dr.From.Month(), dr.From.Day(), 0, 0, 0, 0, loc)
	toExclusive := time.Date(dr.To.Year(), dr.To.Month(), dr.To.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
	return from, toExclusive
}

/* ====== Repository ====== */

type PatientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) *PatientRepository {
	return &PatientRepository{db: db}
}

/* ====== SQL Templates (CTE) ====== */

// rawat inap aktif (tanpa epsdisdatetime)
const monitoringInpatientSQL = `
WITH base AS (
  SELECT
    rm.epspatid                    AS patient_id,
    rm.epspatname                  AS patient_name,
    rm.epsepisode					AS episode,
    rm.epsadmsvctypedesc			AS admission_service_type,
    rm.cliadmdiag1					AS diagnosis,
	rm.epsadmreasondesc			   AS admission_reason,
    rm.provname                    AS dpjp,
    p.nama_penjamin,
    MIN(rm.epsadmdatetime)         AS start_date,
    SUM(p.debit)                   AS total_debit,
    SUM(p.credit)                  AS total_credit
  FROM "rm058" rm
  LEFT JOIN pendapatan p ON rm.epspatid = p.patid
  WHERE rm.epsipop = 'I'
    AND rm.epsdisdatetime IS NULL
  GROUP BY
    rm.epspatid, 
	rm.epspatname, 
	rm.epsepisode,
    rm.epsadmsvctypedesc, 
	rm.cliadmdiag1, 
	rm.epsadmreasondesc, 
	rm.provname, 
	p.nama_penjamin
)
SELECT
  patient_id,
  patient_name,
  episode,
  admission_service_type,
  diagnosis,
  admission_reason,
  dpjp,
  nama_penjamin,
  start_date,
  GREATEST(
    FLOOR(EXTRACT(EPOCH FROM ((NOW() AT TIME ZONE 'Asia/Jakarta') - start_date))/86400.0)::int + 1,
    1
  ) AS los_days,
  EXTRACT(EPOCH FROM ((NOW() AT TIME ZONE 'Asia/Jakarta') - start_date))/3600 AS los_hours_exact,
  total_debit,
  total_credit
FROM base
ORDER BY patient_id, episode
LIMIT ? OFFSET ?
`

const monitoringInpatientCountSQL = `
WITH base AS (
  SELECT
    rm.epspatid                    AS patient_id,
    rm.epspatname                  AS patient_name,
    rm.epsepisode					AS episode,
    rm.epsadmsvctypedesc			AS admission_service_type,
    rm.cliadmdiag1					AS diagnosis,
	rm.epsadmreasondesc			   AS admission_reason,
    rm.provname                    AS dpjp,
    p.nama_penjamin,
    MIN(rm.epsadmdatetime)         AS start_date,
    SUM(p.debit)                   AS total_debit,
    SUM(p.credit)                  AS total_credit
  FROM "rm058" rm
  LEFT JOIN pendapatan p ON rm.epspatid = p.patid
  WHERE rm.epsipop = 'I'
    AND rm.epsdisdatetime IS NULL
  GROUP BY
    rm.epspatid, rm.epspatname, rm.epsepisode,
    rm.epsadmsvctypedesc, rm.cliadmdiag1, rm.epsadmreasondesc, rm.provname, p.nama_penjamin
)
SELECT COUNT(*) FROM base
`

// pasien pulang (punya end_date), nanti difilter by end_date range di luar (WHERE pada outer SELECT)
const dischargedBaseSQL = `
WITH base AS (
  SELECT
    rm.epspatid                    AS patient_id,
    rm.epspatname                  AS patient_name,
    rm.epsepisode					AS episode,
    rm.epsadmsvctypedesc			AS admission_service_type,
    rm.cliadmdiag1					AS diagnosis,
	rm.epsadmreasondesc			   AS admission_reason,
    rm.provname                    AS dpjp,
    p.nama_penjamin,
    MIN(rm.epsadmdatetime)         AS start_date,
    MAX(rm.epsdisdatetime)         AS end_date,
    SUM(p.debit)                   AS total_debit,
    SUM(p.credit)                  AS total_credit
  FROM "rm058" rm
  LEFT JOIN pendapatan p ON rm.epspatid = p.patid
  WHERE rm.epsipop = 'I'
    AND rm.epsdisdatetime IS NOT NULL
  GROUP BY
    rm.epspatid, rm.epspatname, rm.epsepisode,
    rm.epsadmsvctypedesc, rm.cliadmdiag1,
	rm.epsadmreasondesc, rm.provname, p.nama_penjamin
)
`

/* ====== Methods ====== */

// Monitoring rawat inap aktif (paginasi)
func (r *PatientRepository) MonitoringInpatient(ctx context.Context, limit, offset int) ([]InpatientMonitoringRow, int64, error) {
	var rows []InpatientMonitoringRow
	if err := r.db.WithContext(ctx).
		Raw(monitoringInpatientSQL, limit, offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Raw(monitoringInpatientCountSQL).
		Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// Pasien pulang dengan filter end_date (range), paginasi
func (r *PatientRepository) DischargedPatients(ctx context.Context, dr DateRange, limit, offset int) ([]DischargedRow, DischargedSummary, error) {
	from, to := dr.norm()

	// data
	var rows []DischargedRow
	dataSQL := dischargedBaseSQL + `
		SELECT
		patient_id,
		patient_name,
		episode,
		admission_service_type,
		diagnosis,
		admission_reason,
		dpjp,
		nama_penjamin,
		start_date,
		end_date,
		GREATEST(
			FLOOR(EXTRACT(EPOCH FROM (end_date - start_date))/86400.0)::int + 1,
			1
		) AS los_days,
		EXTRACT(EPOCH FROM (end_date - start_date))/3600 AS los_hours_exact,
		total_debit,
		total_credit
		FROM base
		WHERE end_date >= ? AND end_date < ?
		ORDER BY patient_id, episode
		LIMIT ? OFFSET ?
		`
	if err := r.db.WithContext(ctx).
		Raw(dataSQL, from, to, limit, offset).
		Scan(&rows).Error; err != nil {
		return nil, DischargedSummary{}, err
	}

	// count
	var summary DischargedSummary
	countSQL := dischargedBaseSQL + `
		SELECT 
			COUNT(*) AS total,
			COALESCE(AVG(los), 0.0) AS average_los,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY los), 0.0) AS median_los,
			SUM(total_debit) AS total_debit,
			SUM(total_credit) AS total_credit
			FROM (
			SELECT 
				GREATEST(
					FLOOR(EXTRACT(EPOCH FROM (end_date - start_date))/86400.0)::int + 1,
					1
				) AS los,
				total_debit,
				total_credit,
				end_date
			FROM base
			) x
		WHERE end_date >= ? AND end_date < ?;

		`
	if err := r.db.WithContext(ctx).
		Raw(countSQL, from, to).
		Scan(&summary).Error; err != nil {
		return nil, DischargedSummary{}, err
	}

	fmt.Println(summary)

	return rows, summary, nil
}
