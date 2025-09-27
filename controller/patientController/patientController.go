package patientController

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/service/patientService"

	"github.com/gin-gonic/gin"
)

type PatientController struct {
	svc *patientService.PatientService
}

func NewPatientController(svc *patientService.PatientService) *PatientController {
	return &PatientController{svc: svc}
}

/* ===== Helpers ===== */

func qInt(c *gin.Context, key string, def int) int {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func parseDateLocal(dateStr string, loc *time.Location) (time.Time, error) {
	// expect YYYY-MM-DD
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

/* ===== Routes =====

GET /patient/inpatient/monitoring?page=1&limit=50
GET /patient/discharged?from=YYYY-MM-DD&to=YYYY-MM-DD&page=1&limit=50
*/

func (h *PatientController) MonitoringInpatient(c *gin.Context) {
	page := qInt(c, "page", 1)
	limit := qInt(c, "limit", 50)

	res, err := h.svc.MonitoringInpatient(c.Request.Context(), page, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "PATIENT_MONITORING_FETCH_FAILED", "fail get inpatient monitoring", err, nil)
		return
	}

	apiresponse.OK(c, res.Items, "ok", apiresponse.PageMeta{
		Page:  page,
		Limit: limit,
		Total: res.Total,
	})
}

func (h *PatientController) DischargedPatient(c *gin.Context) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	fromStr := strings.TrimSpace(c.Query("from"))
	toStr := strings.TrimSpace(c.Query("to"))

	if fromStr == "" || toStr == "" {
		apiresponse.BadRequest(c, "PATIENT_DISCHARGED_BAD_RANGE", "from/to is required", nil, nil)
		return
	}

	from, err := parseDateLocal(fromStr, loc)
	if err != nil {
		apiresponse.BadRequest(c, "PATIENT_DISCHARGED_BAD_RANGE", "invalid from date", err, nil)
		return
	}
	to, err := parseDateLocal(toStr, loc)
	if err != nil {
		apiresponse.BadRequest(c, "PATIENT_DISCHARGED_BAD_RANGE", "invalid to date", err, nil)
		return
	}

	page := qInt(c, "page", 1)
	limit := qInt(c, "limit", 50)

	res, err := h.svc.DischargedPatients(c.Request.Context(), from, to, page, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "PATIENT_DISCHARGED_FETCH_FAILED", "fail get discharged patient", err, map[string]any{
			"from": fromStr, "to": toStr,
		})
		return
	}

	apiresponse.OK(c, res.Items, "ok", map[string]any{
		"page": page, "limit": limit, "total": res.Summary.Total,
		"average_los": res.Summary.AverageLos, "median_los": res.Summary.MedianLos,
		"total_debit": res.Summary.TotalDebit, "total_credit": res.Summary.TotalCredit,
		"from": fromStr, "to": toStr, "timezone": "Asia/Jakarta",
	})
}
